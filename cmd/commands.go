package cmd

import (
	"fmt"
	"strings"

	"github.com/scottyeager/pal/inout"
	"github.com/scottyeager/pal/session"
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
	systemPrompt := "You are a helpful assistant that suggests shell commands. Each command is a single line that can run in the shell. Respond with three command options, one per line. Don't add anything extra, no context, no explanations, no formatting, no code blocks."

	// Remove any blank lines from completion before storing
	formatFunc := func(response string) string {
		lines := strings.Split(response, "\n")
		var nonEmptyLines []string
		for _, line := range lines {
			if len(strings.TrimSpace(line)) > 0 {
				nonEmptyLines = append(nonEmptyLines, line)
			}
		}
		return strings.Join(nonEmptyLines, "\n")
	}

	params := session.TurnParams{
		Command:      "cmd",
		SystemPrompt: systemPrompt,
		Temperature:  0,
		Markdown:     false,
		FormatFunc:   formatFunc,
	}

	response, err := sesh.ExecuteTurn(params)
	if err != nil {
		return fmt.Errorf("failed to execute turn: %w", err)
	}

	// Store the commands
	err = inout.StoreCommands(response)
	if err != nil {
		return fmt.Errorf("failed to write commands to disk: %w", err)
	}

	fmt.Println(response)
	return nil
}
