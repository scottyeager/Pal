package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/config"
	"github.com/scottyeager/pal/inout"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(askCmd)
}

var askCmd = &cobra.Command{
	Use:   "/ask",
	Short: "Ask a question to the AI",
	Annotations: map[string]string{
		"takes_user_message": "true",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		stdinInput, err := inout.ReadStdin()
		if err != nil {
			return err
		}

		if len(userMessage) == 0 && len(stdinInput) == 0 {
			return fmt.Errorf("No input detected. Please write or pipe in a query")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		if err := config.CheckConfiguration(cfg); err != nil {
			return err
		}

		askModel := config.GetSelectedModel(cfg, "ask")

		aiClient, err := ai.NewClient(cfg, askModel)
		if err != nil {
			return fmt.Errorf("error creating AI client: %w", err)
		}

		var question string
		if stdinInput != "" {
			question = stdinInput + "\nThat concludes the stdin contents. Now here's the query from the user:\n" + strings.Join(userMessage, " ")
		} else {
			question = strings.Join(userMessage, " ")
		}

		var formatMarkdown bool
		if markdown {
			formatMarkdown = !cfg.FormatMarkdown
		} else {
			formatMarkdown = cfg.FormatMarkdown
		}

		system_prompt := "You are a helpful assistant that runs in the users shell but can answer on any topic. Keep responses concise"

		if !cfg.FormatMarkdown {
			system_prompt += " and avoid using Markdown formatting that won't render in a shell. Lists and bullets are fine, but avoid headings, bold, and italic text."
		}

		t := 1.0
		if cmd.Flags().Changed("temperature") {
			t = temperature
		}

		response, err := aiClient.GetCompletion(context.Background(), system_prompt, question, false, t, formatMarkdown, askModel)
		if err != nil {
			return fmt.Errorf("error getting completion: %w", err)
		}

		fmt.Println(response)
		return nil
	},
}
