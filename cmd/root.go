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

var fishAbbr bool
var zshAbbr bool

func init() {
	rootCmd.Flags().BoolVar(&zshAbbr, "fish-abbr", false, "Print fish abbreviation script and exit. Output is meant to be sourced by fish.")
	rootCmd.Flags().BoolVar(&zshAbbr, "zsh-abbr", false, "Print zsh abbreviation script and exit. Output is meant to be sourced by zsh.")

	// Disable help command. --help still works
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}

var rootCmd = &cobra.Command{
	Use:   "pal",
	Short: "pal is a command-line tool that suggests shell commands",
	Long: `pal is a command-line tool that suggests shell commands based on your input.
It uses AI to generate commands and can also manage shell abbreviations.`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	// SilenceErrors: true,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		stdinInput, err := io.ReadStdin()
		if err != nil {
			return fmt.Errorf("error reading stdin: %v", err)
		}

		if len(args) < 1 && stdinInput == "" && !fishAbbr && !zshAbbr {
			return fmt.Errorf("No input or commands detected")
		} else if len(args) > 0 && strings.HasPrefix(args[0], "/") {
			return fmt.Errorf("Invalid command.")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading config: %v", err)
		}

		// Since we want to have these "commands" use a double dash, they have
		// to be flags on the root command to work with cobra, I think
		if fishAbbr && zshAbbr {
			return fmt.Errorf("Only one of --fish-abbr and --zsh-abbr can be used at once")
		} else if fishAbbr {
			fmt.Println(`set -l pal_prefix "` + cfg.AbbreviationPrefix + `"`)
			fmt.Print(abbr.FishAbbrEmbed)
			return nil
		} else if zshAbbr {
			path, err := abbr.InstallZshAbbr()
			if err != nil {
				return err
			} else {
				fmt.Println(path)
				return nil
			}
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

		response, err := aiClient.GetCompletion(context.Background(), system_prompt, question, true, 0)
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
