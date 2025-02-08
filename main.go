package main

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/scottyeager/pal/abbr"
	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/cmd"
	"github.com/scottyeager/pal/config"
	"github.com/spf13/cobra"
)

//go:embed zsh-abbr/zsh-abbr.zsh
var zshAbbrEmbed string

//go:embed zsh-abbr/zsh-job-queue/zsh-job-queue.zsh
var zshJobQueueEmbed string

//go:embed abbr.fish
var fishAbbrEmbed string

func init() {
	var rootCmd = &cobra.Command{
		Use:   "pal",
		Short: "pal is a command-line tool that suggests shell commands",
		Long: `pal is a command-line tool that suggests shell commands based on your input.
It uses AI to generate commands and can also manage shell abbreviations.`,
		SilenceUsage: true, // Suppress usage on Run errors
	}

	// Helper functions
	readStdin := func() (string, error) {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Data is being piped to stdin
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", fmt.Errorf("error reading from stdin: %w", err)
			}
			return "Here is some input from stdin. This might be file contents, error messages, or other command output that the user wanted to include with their query:\n" + string(data), nil
		}
		return "", nil
	}

	checkConfiguration := func(cfg *config.Config) error {
		if len(cfg.Providers) == 0 {
			return fmt.Errorf("No providers configured. Run 'pal /config' to set up a provider")
		}
		if cfg.SelectedModel == "" {
			return fmt.Errorf("No model selected. Run 'pal /models' to select a model")
		}
		return nil
	}

	showHelp := func() {
		cmd.ShowHelp()
	}

	// Subcommands
	var helpCmd = &cobra.Command{
		Use:   "help",
		Short: "Display help information",
		Run: func(cmd *cobra.Command, args []string) {
			showHelp()
		},
	}

	var modelsCmd = &cobra.Command{
		Use:   "models",
		Short: "List available AI models",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}
			cmd.Models(cfg)
			return nil
		},
	}

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Configure pal",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Configure()
		},
	}

	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show the last generated commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			data, err := ai.GetStoredCompletion()
			if err != nil {
				return fmt.Errorf("error reading data from disk: %w", err)
			}

			commands := strings.Split(string(data), "\n")
			for i, cmd := range commands {
				if cmd != "" {
					fmt.Printf("%d: %s\n", i+1, cmd)
				}
			}

			if cfg.ZshAbbreviations {
				prefix := cfg.AbbreviationPrefix
				if err := abbr.UpdateZshAbbreviations(prefix, prefix, string(data)); err != nil {
					return fmt.Errorf("error updating zsh abbreviations: %w", err)
				}
			}
			return nil
		},
	}

	var askCmd = &cobra.Command{
		Use:   "ask",
		Short: "Ask a question to the AI",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			if err := checkConfiguration(cfg); err != nil {
				return err
			}

			aiClient, err := ai.NewClient(cfg)
			if err != nil {
				return fmt.Errorf("error creating AI client: %w", err)
			}

			stdinInput, err := readStdin()
			if err != nil {
				return err
			}

			var question string
			if stdinInput != "" {
				question = stdinInput + "\nThat concludes the stdin contents. Now here's the query from the user:\n" + strings.Join(args, " ")
			} else {
				question = strings.Join(args, " ")
			}

			system_prompt := "You are a helpful assistant that runs in the users shell but can answer on any topic. Keep responses concise and avoid using Markdown formatting that won't render in a shell. Lists and bullets are fine, but avoid headings, bold, and italic text."

			response, err := aiClient.GetCompletion(context.Background(), system_prompt, question, false, 1.0)
			if err != nil {
				return fmt.Errorf("error getting completion: %w", err)
			}

			fmt.Println(response)
			return nil
		},
	}

	var commitCmd = &cobra.Command{
		Use:   "commit",
		Short: "Generate a commit message",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			if err := checkConfiguration(cfg); err != nil {
				return err
			}

			aiClient, err := ai.NewClient(cfg)
			if err != nil {
				return fmt.Errorf("error creating AI client: %w", err)
			}

			stdinInput, err := readStdin()
			if err != nil {
				return err
			}

			var message string
			if stdinInput != "" {
				message = stdinInput
			} else {
				message, err = cmd.Commit(cfg, aiClient)
				if err != nil {
					return fmt.Errorf("error generating commit message: %w", err)
				}
			}

			fmt.Println(message)
			return nil
		},
	}

	var zshAbbrCmd = &cobra.Command{
		Use:   "--zsh-abbr",
		Short: "Print the path to the zsh-abbr script",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check for existing temp dir
			pattern := os.TempDir() + "/pal-zsh-abbr-*"
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return fmt.Errorf("error checking for existing temp dir: %w", err)
			}

			var tmpDir string
			if len(matches) > 0 {
				tmpDir = matches[0]
			} else {
				tmpDir, err = os.MkdirTemp("", "pal-zsh-abbr-")
				if err != nil {
					return fmt.Errorf("error creating temp dir: %w", err)
				}

				err = os.MkdirAll(tmpDir+"/zsh-job-queue", 0755)
				if err != nil {
					return fmt.Errorf("error creating job queue dir: %w", err)
				}

				err = os.WriteFile(tmpDir+"/zsh-job-queue/zsh-job-queue.zsh", []byte(zshJobQueueEmbed), 0755)
				if err != nil {
					return fmt.Errorf("error writing job queue file: %w", err)
				}

				err = os.WriteFile(tmpDir+"/zsh-abbr.zsh", []byte(zshAbbrEmbed), 0755)
				if err != nil {
					return fmt.Errorf("error writing abbr file: %w", err)
				}
			}

			fmt.Println(tmpDir + "/zsh-abbr.zsh")
			return nil
		},
	}

	var fishAbbrCmd = &cobra.Command{
		Use:   "--fish-abbr",
		Short: "Print the fish abbreviation script",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}
			fmt.Println(`set -l pal_prefix "` + cfg.AbbreviationPrefix + `"`)
			fmt.Print(fishAbbrEmbed)
			return nil
		},
	}

	// Adding the default command
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		stdinInput, err := readStdin()
		if err != nil {
			return fmt.Errorf("error reading stdin: %v", err)
		}

		if len(args) < 1 && stdinInput == "" {
			return fmt.Errorf("no input detected")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading config: %v", err)
		}

		var question string
		if stdinInput != "" && len(args) > 0 {
			question = stdinInput + "\nThat concludes the stdin contents. Now here's the query from the user:\n" + strings.Join(args, " ")
		} else if stdinInput != "" {
			question = stdinInput
		} else {
			question = strings.Join(args, " ")
		}

		if err := checkConfiguration(cfg); err != nil {
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
		response = strings.Join(strings.Fields(response), "\n")

		if cfg.ZshAbbreviations {
			prefix := cfg.AbbreviationPrefix
			if err := abbr.UpdateZshAbbreviations(prefix, prefix, response); err != nil {
				return fmt.Errorf("error updating zsh abbreviations: %w", err)
			}
		}

		fmt.Println(response)
		return nil
	}

	// Add subcommands to the root command
	rootCmd.AddCommand(helpCmd)
	rootCmd.AddCommand(modelsCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(zshAbbrCmd)
	rootCmd.AddCommand(fishAbbrCmd)
	rootCmd.SetHelpCommand(helpCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {}
