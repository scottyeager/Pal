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
)

//go:embed zsh-abbr/zsh-abbr.zsh
var zshAbbrEmbed string

//go:embed zsh-abbr/zsh-job-queue/zsh-job-queue.zsh
var zshJobQueueEmbed string

//go:embed abbr.fish
var fishAbbrEmbed string

func readStdin() (string, error) {
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

func checkConfiguration(cfg *config.Config) error {
	if len(cfg.Providers) == 0 {
		return fmt.Errorf("No providers configured. Run 'pal /config' to set up a provider")
	}
	if cfg.SelectedModel == "" {
		return fmt.Errorf("No model selected. Run 'pal /models' to select a model")
	}
	return nil
}

func main() {
	stdinInput, err := readStdin()
	if err != nil {
		fmt.Printf("Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 && stdinInput == "" {
		fmt.Println("No input detected.")
		fmt.Println("Try 'pal /help' for more information")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	var command string
	if len(os.Args) > 1 {
		command = os.Args[1]
	} else {
		command = ""
	}

	if !strings.HasPrefix(command, "/") && !strings.HasPrefix(command, "-") {
		// If no command is specified, treat the entire input as a question
		var question string
		if stdinInput != "" && len(os.Args) > 1 {
			question = stdinInput + "\nThat concludes the stdin contents. Now here's the query from the user:\n" + strings.Join(os.Args[1:], " ")
		} else if stdinInput != "" {
			question = stdinInput
		} else {
			question = strings.Join(os.Args[1:], " ")
		}

		if err := checkConfiguration(cfg); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		aiClient, err := ai.NewClient(cfg)
		if err != nil {
			fmt.Printf("Error creating AI client: %v\n", err)
			os.Exit(1)
		}

		system_prompt := "You are a helpful assistant that suggests shell commands. Each command is a single line that can run in the shell. Respond three command options, one per line. Don't add anything extra, no context, no explanations, no formatting."

		response, err := aiClient.GetCompletion(context.Background(), system_prompt, question, true)
		if err != nil {
			fmt.Printf("Error getting completion: %v\n", err)
			os.Exit(1)
		}

		if cfg.ZshAbbreviations {
			prefix := cfg.AbbreviationPrefix
			if err := abbr.UpdateZshAbbreviations(prefix, prefix, response); err != nil {
				fmt.Printf("%v\n", err)
				return
			}
		}

		fmt.Println(response)
		return
	}

	switch command {
	case "/help":
		showHelp()
	case "/models":
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
		cmd.Models(cfg)
	case "/config":
		cmd.Configure()
	case "/show":
		data, err := ai.GetStoredCompletion()
		if err != nil {
			fmt.Printf("Error reading data from disk: %v\n", err)
			os.Exit(1)
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
				fmt.Printf("%v\n", err)
				return
			}
		}

	case "/ask":
		var question string
		if stdinInput != "" && len(os.Args) > 2 {
			question = stdinInput + "\nThat concludes the stdin contents. Now here's the query from the user:\n" + strings.Join(os.Args[2:], " ")
		} else if stdinInput != "" {
			question = stdinInput
		} else if stdinInput == "" && len(os.Args) < 3 {
			fmt.Println("Usage: pal /ask <question>")
			os.Exit(1)
		} else {
			question = strings.Join(os.Args[2:], " ")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		if err := checkConfiguration(cfg); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		aiClient, err := ai.NewClient(cfg)
		if err != nil {
			fmt.Printf("Error creating AI client: %v\n", err)
			os.Exit(1)
		}

		system_prompt := "You are a helpful assistant that runs in the users shell but can answer on any topic. Keep responses concise and avoid using Markdown formatting that won't render in a shell. Lists and bullets are fine, but avoid headings, bold, and italic text."

		response, err := aiClient.GetCompletion(context.Background(), system_prompt, question, false)
		if err != nil {
			fmt.Printf("Error getting completion: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(response)

	case "/commit":
		if err := checkConfiguration(cfg); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		aiClient, err := ai.NewClient(cfg)
		if err != nil {
			fmt.Printf("Error creating AI client: %v\n", err)
			os.Exit(1)
		}

		stdinInput, err := readStdin()
		if err != nil {
			fmt.Printf("Error reading stdin: %v\n", err)
			os.Exit(1)
		}

		var message string
		if stdinInput != "" {
			// If there's stdin, use it as the commit message
			message = stdinInput
		} else {
			// Otherwise generate a commit message
			message, err = cmd.Commit(cfg, aiClient)
		}
		if err != nil {
			fmt.Printf("Error generating commit: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(message)

	case "--zsh-abbr":
		// Check for existing temp dir
		pattern := os.TempDir() + "/pal-zsh-abbr-*"
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Printf("Error checking for existing temp dir: %v\n", err)
			os.Exit(1)
		}

		var tmpDir string
		if len(matches) > 0 {
			tmpDir = matches[0]
		} else {
			tmpDir, err = os.MkdirTemp("", "pal-zsh-abbr-")
			if err != nil {
				fmt.Printf("Error creating temp dir: %v\n", err)
				os.Exit(1)
			}

			err = os.MkdirAll(tmpDir+"/zsh-job-queue", 0755)
			if err != nil {
				fmt.Printf("Error creating job queue dir: %v\n", err)
				os.Exit(1)
			}

			err = os.WriteFile(tmpDir+"/zsh-job-queue/zsh-job-queue.zsh", []byte(zshJobQueueEmbed), 0755)
			if err != nil {
				fmt.Printf("Error writing job queue file: %v\n", err)
				os.Exit(1)
			}

			err = os.WriteFile(tmpDir+"/zsh-abbr.zsh", []byte(zshAbbrEmbed), 0755)
			if err != nil {
				fmt.Printf("Error writing abbr file: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Println(tmpDir + "/zsh-abbr.zsh")

	case "--fish-abbr":
		fmt.Println(`set -l pal_prefix "` + cfg.AbbreviationPrefix + `"`)
		fmt.Print(fishAbbrEmbed)
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

func showHelp() {
	cmd.ShowHelp()
}
