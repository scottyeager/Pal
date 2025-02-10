package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/scottyeager/pal/abbr"
	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/config"
	"github.com/scottyeager/pal/io"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var version string

var fish bool
var fishAbbr bool
var zshAbbr bool
var fishCompletion bool
var zshCompletion bool
var temperature float64

var userMessage []string

func init() {
	rootCmd.Flags().BoolVarP(&fish, "version", "V", false, "Print version number")
	rootCmd.Flags().BoolVar(&fish, "fish", false, "Print fish abbreviation script and completion script, then exit. Output is meant to be sourced by fish.")
	rootCmd.Flags().BoolVar(&zshAbbr, "fish-abbr", false, "Print fish abbreviation script and exit. Output is meant to be sourced by fish.")
	rootCmd.Flags().BoolVar(&zshAbbr, "zsh-abbr", false, "Print zsh abbreviation script and exit. Output is meant to be sourced by zsh.")
	rootCmd.Flags().BoolVar(&fishCompletion, "fish-completion", false, "Print fish autocompletion script and exit. Output is meant to be sourced by fish.")
	rootCmd.Flags().BoolVar(&zshCompletion, "zsh-completion", false, "Print zsh autocompletion script and exit. Output is meant to be sourced by zsh.")
	rootCmd.PersistentFlags().Float64VarP(&temperature, "temperature", "t", 0, "Set the temperature for the AI model, between 0 and 2 (higher values make output more random)")

	// Disable help command. --help still works
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}

var rootCmd = &cobra.Command{
	// Use:   "pal",
	Use:   os.Args[0],
	Short: "pal is a command-line tool that suggests shell commands",
	Long: `pal is a command-line tool that suggests shell commands based on your input.
It uses AI to generate commands and can also manage shell abbreviations.`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Args: cobra.ArbitraryArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if temperature < 0 || temperature > 2 {
			return fmt.Errorf("Temperature must be between 0 and 2")
		} else {
			return nil
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		stdinInput, err := io.ReadStdin()
		if err != nil {
			return fmt.Errorf("error reading stdin: %v", err)
		}

		// We call these command flags because they are really commands hiding
		// as flags (since we can't use -- as a command prefix)
		commandFlagsSet := 0
		for _, flag := range []bool{fish, fishAbbr, zshAbbr, fishCompletion, zshCompletion} {
			if flag {
				commandFlagsSet++
			}
		}

		if len(userMessage) == 0 && stdinInput == "" && commandFlagsSet == 0 {
			return fmt.Errorf("No input or commands detected")
		} else if len(args) > 0 && strings.HasPrefix(args[0], "/") {
			return fmt.Errorf("Invalid command.")
		} else if commandFlagsSet > 1 {
			return fmt.Errorf("Only one flag to print shell feature scripts can be used at once. To enable both completions and abbreviations in one shot for fish, use --fish.")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading config: %v", err)
		}

		// Since we want to have these "commands" use a double dash, they have
		// to be flags on the root command to work with cobra, I think
		if fish {
			fmt.Println(abbr.GetFishAbbrScript(cfg.AbbreviationPrefix))
			cmd.GenFishCompletion(os.Stdout, true)
			// Disables file name completions. Set command name dynamically in
			// case the user changed it
			fmt.Printf("complete -c %s -f", os.Args[0])
			return nil
		} else if fishAbbr {
			fmt.Println(abbr.GetFishAbbrScript(cfg.AbbreviationPrefix))
			return nil
		} else if fishCompletion {
			cmd.GenFishCompletion(os.Stdout, true)
			// Disables file name completions. Set command name dynamically in
			// case the user changed it
			fmt.Printf("complete -c %s -f", os.Args[0])
			return nil
		} else if zshAbbr {
			path, err := abbr.InstallZshAbbr()
			if err != nil {
				return err
			} else {
				fmt.Println(path)
				return nil
			}
		} else if zshCompletion {
			cmd.GenZshCompletionNoDesc(os.Stdout)
			return nil
		}
		// Just to be sure we don't move on if any of these flags are set.
		// Since users will be sourcing the output of these commands and we
		// don't want AI output getting executed!
		if commandFlagsSet > 0 {
			return nil
		}

		var question string
		if stdinInput != "" && len(userMessage) > 0 {
			question = stdinInput + "\nThat concludes the stdin contents. Now here's the query from the user:\n" + strings.Join(userMessage, " ")
		} else if stdinInput != "" {
			question = stdinInput
		} else {
			question = strings.Join(userMessage, " ")
		}

		if err := config.CheckConfiguration(cfg); err != nil {
			return err
		}

		aiClient, err := ai.NewClient(cfg)
		if err != nil {
			return fmt.Errorf("error creating AI client: %v", err)
		}

		system_prompt := "You are a helpful assistant that suggests shell commands. Each command is a single line that can run in the shell. Respond with three command options, one per line. Don't add anything extra, no context, no explanations, no formatting, no code blocks."

		t := 0.0
		if cmd.Flags().Changed("temperature") {
			t = temperature
		}
		response, err := aiClient.GetCompletion(context.Background(), system_prompt, question, true, t)
		if err != nil {
			return fmt.Errorf("error getting completion: %v", err)
		}

		// Remove any blank lines (weaker models tend to return them)
		lines := strings.Split(response, "\n")
		var nonEmptyLines []string
		for _, line := range lines {
			if len(strings.TrimSpace(line)) > 0 {
				nonEmptyLines = append(nonEmptyLines, line)
			}
		}
		response = strings.Join(nonEmptyLines, "\n")

		if cfg.ZshAbbreviations {
			prefix := cfg.AbbreviationPrefix
			if err := abbr.UpdateZshAbbreviations(prefix, prefix, response); err != nil {
				return fmt.Errorf("error updating zsh abbreviations: %w", err)
			}
		}

		fmt.Println(response)
		return nil
	},
}

func preparse(args []string) int {
	boolLongFlags := []string{"help"}
	boolShortFlags := []string{"h", "V"}
	longFlags := []string{}
	shortFlags := []string{}
	rootCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		boolLongFlags = append(boolLongFlags, flag.Name)
		// boolShortFlags = append(shortFlags, flag.Shorthand)
	})
	rootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		longFlags = append(longFlags, flag.Name)
		shortFlags = append(shortFlags, flag.Shorthand)
	})

	i := 1
	for {
		if strings.HasPrefix(args[i], "--") {
			arg := strings.TrimPrefix(args[i], "--")
			matched := false
			for _, flag := range longFlags {
				if strings.HasPrefix(arg, flag) {
					if len(arg) == len(flag) {
						// Consumes a flag like: --temperature 0
						i += 2
						matched = true
						// Consumes a flag like: --temperature=0
					} else {
						i++
						matched = true
					}

				}
			}
			for _, flag := range boolLongFlags {
				if arg == flag {
					i++
					matched = true
				}
			}

			if !matched {
				break
			}
		} else if strings.HasPrefix(args[i], "-") {
			arg := strings.TrimPrefix(args[i], "-")
			matched := false
			for _, flag := range shortFlags {
				if strings.HasPrefix(arg, flag) {
					if len(arg) == 1 {
						// Consumes a flag like: -t 0
						i += 2
						matched = true
					} else {
						// Consumes a flag like: -t0
						i++
						matched = true
					}
				}
			}
			for _, flag := range boolShortFlags {
				if arg == flag {
					i++
					matched = true
				}
			}
			if !matched {
				break
			}
		} else {
			break
		}
		if i >= len(args)-1 {
			break
		}
	}

	if i <= len(args)-1 && strings.HasPrefix(args[i], "/") {
		return i + 1
	} else {
		return i
	}
}

func Execute() {
	rootCmd.Version = version
	// Skip preparsing for hidden commands used to generate completions
	if len(os.Args) > 1 && os.Args[1] != "__complete" && os.Args[1] != "__completeNoDesc" {
		split := preparse(os.Args)
		userMessage = os.Args[split:]
		os.Args = os.Args[:split]
	}

	if err := rootCmd.Execute(); err != nil {
		// fmt.Println(err)
		os.Exit(1)
	}
}
