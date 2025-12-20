package cmd

import (
	"fmt"
	"github.com/scottyeager/pal/inout"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(chatCmd)
}

var chatCmd = &cobra.Command{
	Use:   "/chat",
	Short: "Continue the last conversation in chat mode",
	Annotations: map[string]string{
		"takes_user_message": "true",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the latest thread to determine the original command
		history, err := inout.NewHistory()
		if err != nil {
			return fmt.Errorf("failed to create history manager: %w", err)
		}

		// Get or create thread
		thread, err := history.GetLatestThread()
		if err != nil {
			return fmt.Errorf("failed to get latest thread: %w", err)
		}

		// Get the original command from the thread
		originalCommand := thread.CommandName
		if originalCommand == "" {
			originalCommand = "ask" // Default fallback
		}

		// Find the original command in the command tree
		originalCmd := rootCmd.Commands()
		var targetCmd *cobra.Command
		for _, c := range originalCmd {
			if c.Name() == "/" + originalCommand {
				targetCmd = c
				break
			}
		}

		if targetCmd == nil {
			return fmt.Errorf("original command %s not found", originalCommand)
		}

		// Set the global chat variable to enable chat mode
		globalSession.Chat = true

		// Call the original command's RunE function
		return targetCmd.RunE(targetCmd, args)
	},
}
