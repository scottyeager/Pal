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

var fileCmd = &cobra.Command{
	Use:   "/file",
	Short: "Generate file contents based on a description",
	Long:  `Generate file contents based on a description. The output is sanitized for direct use.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stdinInput, err := inout.ReadStdin()
		if err != nil {
			return err
		}

		if len(args) == 0 && len(stdinInput) == 0 {
			return fmt.Errorf("please provide a description of the file to generate or pipe in content")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		if err := config.CheckConfiguration(cfg); err != nil {
			return err
		}

		fileModel := config.GetSelectedModel(cfg, "file")
		aiClient, err := ai.NewClient(cfg, fileModel)

		if err != nil {
			return fmt.Errorf("error creating AI client: %w", err)
		}

		var description string
		if stdinInput != "" {
			if len(args) > 0 {
				description = stdinInput + "\nThat concludes the stdin contents. Now here's the description from the user:\n" + strings.Join(args, " ")
			} else {
				description = stdinInput
			}
		} else {
			description = strings.Join(args, " ")
		}

		system_prompt := "You are a helpful assistant that generates file contents. Provide only the raw file content without any additional commentary, explanations, or markdown formatting. Do not wrap the content in code blocks (```)."

		t := 1.0
		if cmd.Flags().Changed("temperature") {
			t = temperature
		}

		response, err := aiClient.GetCompletion(context.Background(), system_prompt, description, false, t, false, fileModel)
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
