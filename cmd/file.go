package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/config"
	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{
	Use:   "/file",
	Short: "Generate file contents based on a description",
	Long:  `Generate file contents based on a description. The output is sanitized for direct use.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide a description of the file to generate")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		if err := config.CheckConfiguration(cfg); err != nil {
			return err
		}

		aiClient, err := ai.NewClient(cfg)
		if err != nil {
			return fmt.Errorf("error creating AI client: %w", err)
		}

		description := strings.Join(args, " ")
		system_prompt := "You are a helpful assistant that generates file contents. Provide only the raw file content without any additional commentary, explanations, or markdown formatting. Do not wrap the content in code blocks (```)."

		t := 1.0
		if cmd.Flags().Changed("temperature") {
			t = temperature
		}

		response, err := aiClient.GetCompletion(context.Background(), system_prompt, description, false, t, false)
		if err != nil {
			return fmt.Errorf("error getting completion: %w", err)
		}

		content := sanitizeFileContent(response)
		fmt.Println(content)
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

func init() {
	rootCmd.AddCommand(fileCmd)
}
