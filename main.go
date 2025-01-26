package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/cmd"
	"github.com/scottyeager/pal/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pal <command>")
		fmt.Println("Try 'pal /help' for more information")
		os.Exit(1)
	}

	command := os.Args[1]

	if !strings.HasPrefix(command, "/") {
		// If no command is specified, treat the entire input as a question
		question := strings.Join(os.Args[1:], " ")
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		aiClient, err := ai.NewClient(cfg)
		if err != nil {
			fmt.Printf("Error creating AI client: %v\n", err)
			os.Exit(1)
		}

		system_prompt := "You are a helpful assistant that suggests shell commands. Respond only with a one line shell command string. Don't add anything extra, no context and no explanations"
		response, err := aiClient.GetCompletion(context.Background(), system_prompt, question)
		if err != nil {
			fmt.Printf("Error getting completion: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(response)
		return
	}

	switch command {
	case "/help":
		showHelp()
	case "/config":
		cmd.Configure()
	case "/make-completion":
		fmt.Println("bye")
	case "/ask":
		if len(os.Args) < 3 {
			fmt.Println("Usage: pal /ask <question>")
			os.Exit(1)
		}
		question := strings.Join(os.Args[2:], " ")

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		aiClient, err := ai.NewClient(cfg)
		if err != nil {
			fmt.Printf("Error creating AI client: %v\n", err)
			os.Exit(1)
		}

		response, err := aiClient.GetCompletion(context.Background(), "", question)
		if err != nil {
			fmt.Printf("Error getting completion: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(response)
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

func showHelp() {
	cmd.ShowHelp()
}
