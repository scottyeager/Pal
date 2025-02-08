package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/config"
	"github.com/scottyeager/pal/io"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(askCmd)
}

var askCmd = &cobra.Command{
	Use:   "/ask",
	Short: "Ask a question to the AI",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		stdinInput, err := io.ReadStdin()
		if err != nil {
			return err
		}

		var question string
		if stdinInput != "" {
			question = stdinInput + "\nThat concludes the stdin contents. Now here's the query from the user:\n" + strings.Join(args, " ")
		} else {
			question = strings.Join(args, " ")
		}

		system_prompt := "You are a helpful assistant that runs in the users shell but can answer on any topic. Keep responses concise and avoid using Markdown formatting that won't render in a shell. Lists and bullets are fine, but avoid headings, bold, and italic text."

		t := 1.0
		if cmd.PersistentFlags().Changed("temperature") {
			t = temperature
		}
		response, err := aiClient.GetCompletion(context.Background(), system_prompt, question, false, t)
		if err != nil {
			return fmt.Errorf("error getting completion: %w", err)
		}

		fmt.Println(response)
		return nil
	},
}
