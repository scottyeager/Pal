package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/scottyeager/pal/inout"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cmdCmd)
}

var cmdCmd = &cobra.Command{
	Use:   "/cmd",
	Short: "Get command suggestions (default)",
	Annotations: map[string]string{
		"takes_user_message": "true",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return Commands(*cmd, args)
	},
}

func Commands(cmd cobra.Command, query []string) error {
	// Read stdin
	if err := globalSession.ReadStdin(); err != nil {
		return fmt.Errorf("error reading stdin: %v", err)
	}

	if len(globalSession.UserMessage) == 0 && globalSession.StdinInput == "" {
		return fmt.Errorf("No input detected")
	}

	// Load and check configuration
	if err := globalSession.LoadConfig(); err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	if err := globalSession.CheckConfig(); err != nil {
		return err
	}

	// Get user input
	question := globalSession.GetUserInput()

	// Get model for cmd command
	cmdModel := globalSession.GetSelectedModel("cmd")

	system_prompt := "You are a helpful assistant that suggests shell commands. Each command is a single line that can run in the shell. Respond with three command options, one per line. Don't add anything extra, no context, no explanations, no formatting, no code blocks."

	t := 0.0
	if cmd.Flags().Changed("temperature") {
		t = globalSession.Temperature
	}

	response, err := globalSession.GetCompletion(context.Background(), system_prompt, question, cmdModel)
	if err != nil {
		return fmt.Errorf("error getting completion: %v", err)
	}

	// Remove any blank lines from completion before storing
	lines := strings.Split(response, "\n")
	var nonEmptyLines []string
	for _, line := range lines {
		if len(strings.TrimSpace(line)) > 0 {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}
	cleanedResponse := strings.Join(nonEmptyLines, "\n")

	// Store the completion
	err = inout.StoreCommands(cleanedResponse)
	if err != nil {
		return fmt.Errorf("failed to write to disk: %w", err)
	}

	fmt.Println(cleanedResponse)
	return nil
}
