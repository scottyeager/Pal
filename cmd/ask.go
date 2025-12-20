package cmd

import (
	"context"
	"fmt"
	"strings"

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
		// Read stdin
		if err := globalSession.ReadStdin(); err != nil {
			return err
		}

		if len(globalSession.UserMessage) == 0 && len(globalSession.StdinInput) == 0 {
			return fmt.Errorf("No input detected. Please write or pipe in a query")
		}

		// Load and check configuration
		if err := globalSession.LoadConfig(); err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		if err := globalSession.CheckConfig(); err != nil {
			return err
		}

		askModel := globalSession.GetSelectedModel("ask")

		var question string
		if globalSession.StdinInput != "" {
			question = globalSession.StdinInput + "\nThat concludes the stdin contents. Now here's the query from the user:\n" + strings.Join(globalSession.UserMessage, " ")
		} else {
			question = strings.Join(globalSession.UserMessage, " ")
		}

		system_prompt := "You are a helpful assistant that runs in the users shell but can answer on any topic. Keep responses concise"

		if !globalSession.Config.FormatMarkdown {
			system_prompt += " and avoid using Markdown formatting that won't render in a shell. Lists and bullets are fine, but avoid headings, bold, and italic text."
		}

		t := 1.0
		if cmd.Flags().Changed("temperature") {
			t = globalSession.Temperature
		}

		var response string
		var thread *inout.Thread
		if globalSession.Chat {
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
			response, err = globalSession.ContinueCompletion(context.Background(), thread, question, askModel)
			if err != nil {
				return fmt.Errorf("error continuing completion: %w", err)
			}

			// Print thread info for reference
			fmt.Printf("[Chat thread: %s/%d]\n", thread.Date.Format("2006-01-02"), thread.ID)
		} else {
			// Create a new thread each time (no -c flag)
			response, thread, err = globalSession.GetCompletionWithHistory(context.Background(), "ask", system_prompt, question, askModel)
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
