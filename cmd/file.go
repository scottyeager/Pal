package cmd

import (
	"fmt"
	"strings"

	"github.com/scottyeager/pal/session"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fileCmd)
}

var fileCmd = &cobra.Command{
	Use:   "/file",
	Short: "Generate file contents based on a description",
	Long:  `Generate file contents based on a description. The output is sanitized for direct use.`,
	Annotations: map[string]string{
		"takes_user_message": "true",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		systemPrompt := "You are a helpful assistant that generates file contents. Provide only the raw file content without any additional commentary, explanations, or markdown formatting. Do not wrap the content in code blocks (```)."

		params := session.TurnParams{
			Command:      "file",
			SystemPrompt: systemPrompt,
			Prompt:       "",
			Temperature:  1,
			Markdown:     false,
			FormatFunc:   sanitizeFileContent,
		}

		response, err := sesh.ExecuteTurn(params)
		if err != nil {
			return fmt.Errorf("failed to execute turn: %w", err)
		}

		fmt.Println(response)
		return nil
	},
}

func sanitizeFileContent(input string) string {
	// Strip markdown code block delimiters and any language specifier
	lines := strings.Split(input, "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "```") {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.HasPrefix(lines[len(lines)-1], "```") {
		lines = lines[:len(lines)-1]
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
