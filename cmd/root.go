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
)

var fish bool
var fishAbbr bool
var zshAbbr bool
var fishCompletion bool
var zshCompletion bool
var temperature float64

func init() {
	rootCmd.Flags().BoolVar(&fish, "fish", false, "Print fish abbreviation script and completion script, then exit. Output is meant to be sourced by fish.")
	rootCmd.Flags().BoolVar(&zshAbbr, "fish-abbr", false, "Print fish abbreviation script and exit. Output is meant to be sourced by fish.")
	rootCmd.Flags().BoolVar(&zshAbbr, "zsh-abbr", false, "Print zsh abbreviation script and exit. Output is meant to be sourced by zsh.")
	rootCmd.Flags().BoolVar(&fishCompletion, "fish-completion", false, "Print fish autocompletion script and exit. Output is meant to be sourced by fish.")
	rootCmd.Flags().BoolVar(&zshCompletion, "zsh-completion", false, "Print zsh autocompletion script and exit. Output is meant to be sourced by zsh.")
	rootCmd.PersistentFlags().Float64VarP(&temperature, "temperature", "t", 0, "Set the temperature for the AI model (higher values make output more random)")

	// Disable help command. --help still works
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}

var rootCmd = &cobra.Command{
	Use:   "pal",
	Short: "pal is a command-line tool that suggests shell commands",
	Long: `pal is a command-line tool that suggests shell commands based on your input.
It uses AI to generate commands and can also manage shell abbreviations.`,
	CompletionOptions: cobra.CompletionOptions{
		// DisableDefaultCmd: true,
	},
	// SilenceErrors: true,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		stdinInput, err := io.ReadStdin()
		if err != nil {
			return fmt.Errorf("error reading stdin: %v", err)
		}

		flagsSet := 0
		for _, flag := range []bool{fish, fishAbbr, zshAbbr, fishCompletion, zshCompletion} {
			if flag {
				flagsSet++
			}
		}

		if len(args) < 1 && stdinInput == "" && flagsSet == 0 {
			return fmt.Errorf("No input or commands detected")
		} else if len(args) > 0 && strings.HasPrefix(args[0], "/") {
			return fmt.Errorf("Invalid command.")
		} else if flagsSet > 1 {
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
			return nil
		} else if fishAbbr {
			fmt.Println(abbr.GetFishAbbrScript(cfg.AbbreviationPrefix))
			return nil
		} else if fishCompletion {
			cmd.GenFishCompletion(os.Stdout, true)
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
		if flagsSet > 0 {
			return nil
		}

		var question string
		if stdinInput != "" && len(args) > 0 {
			question = stdinInput + "\nThat concludes the stdin contents. Now here's the query from the user:\n" + strings.Join(args, " ")
		} else if stdinInput != "" {
			question = stdinInput
		} else {
			question = strings.Join(args, " ")
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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// fmt.Println(err)
		os.Exit(1)
	}
}
