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
	stdinInput, err := inout.ReadStdin()
	if err != nil {
		return fmt.Errorf("error reading stdin: %v", err)
	}

	if len(userMessage) == 0 && stdinInput == "" {
		return fmt.Errorf("No input detected")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	var question string
	if stdinInput != "" && len(userMessage) > 0 {
		question = stdinInput + "\nThat concludes the stdin contents. Now here's the query from the user:\n" + strings.Join(userMessage, " ")
	} else if stdinInput != "" {
		question = stdinInput
	} else {
		question = strings.Join(userMessage, " ")
	}

	if err := config.CheckConfiguration(cfg); err != nil {
		return err
	}

	aiClient, err := ai.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("error creating AI client: %v", err)
	}

	system_prompt := "You are a helpful assistant that suggests shell commands. Each command is a single line that can run in the shell. Respond with three command options, one per line. Don't add anything extra, no context, no explanations, no formatting, no code blocks."

	t := 0.0
	if cmd.Flags().Changed("temperature") {
		t = temperature
	}
	response, err := aiClient.GetCompletion(context.Background(), system_prompt, question, true, t, false)
	if err != nil {
		return fmt.Errorf("error getting completion: %v", err)
	}

	// Remove any blank lines (weaker models tend to return them)
	lines := strings.Split(response, "\n")
	var nonEmptyLines []string
	for _, line := range lines {
		if len(strings.TrimSpace(line)) > 0 {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}
	response = strings.Join(nonEmptyLines, "\n")

	fmt.Println(response)
	return nil
}
