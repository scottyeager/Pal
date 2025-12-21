package cmd

import (
	"fmt"

	"github.com/scottyeager/pal/session"
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
		systemPrompt := "You are a helpful assistant that runs in the users shell but can answer on any topic. Keep responses concise"

		params := session.TurnParams{
			Command:      "ask",
			SystemPrompt: systemPrompt,
			Prompt:       "",
			Temperature:  1,
			Markdown:     true,
			FormatFunc:   nil,
		}

		response, err := sesh.ExecuteTurn(params)
		if err != nil {
			return fmt.Errorf("failed to execute turn: %w", err)
		}

		fmt.Println(response)
		return nil
	},
}
