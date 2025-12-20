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

		var response string
		var thread *inout.Thread
		if chat {
			// Create history manager
			history, err := inout.NewHistory()
			if err != nil {
				return fmt.Errorf("failed to create history manager: %w", err)
			}
			
			// Get or create thread
			thread, err = history.GetLatestThread()
			if err != nil {
				return fmt.Errorf("failed to get latest thread: %w", err)
			}
			
			// Continue the conversation with the existing thread
			response, err = aiClient.ContinueCompletion(context.Background(), thread, question, t, formatMarkdown, askModel)
			if err != nil {
				return fmt.Errorf("error continuing completion: %w", err)
			}
			
			// Print thread info for reference
			fmt.Printf("[Chat thread: %s/%d]\n", thread.Date.Format("2006-01-02"), thread.ID)
		} else {
			// Create a new thread each time (no -c flag)
			response, thread, err = aiClient.GetCompletionWithHistory(context.Background(), "ask", system_prompt, question, t, formatMarkdown, askModel)
			if err != nil {
				return fmt.Errorf("error getting completion with history: %w", err)
			}
			// Print thread info for reference
			if thread != nil {
				fmt.Printf("[New chat thread: %s/%d]\n", thread.Date.Format("2006-01-02"), thread.ID)
			}
		}

		fmt.Println(response)
		return nil
	},
}
