package cmd

import (
	"fmt"
	"strings"

	"github.com/scottyeager/pal/abbr"
	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "/show",
	Short: "Show the last generated commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		data, err := ai.GetStoredCompletion()
		if err != nil {
			return fmt.Errorf("error reading data from disk: %w", err)
		}

		commands := strings.Split(string(data), "\n")
		var firstCommand string
		for i, cmd := range commands {
			if cmd == "" {
				continue
			}
			if i == 0 {
				firstCommand = cmd
			} else {
				fmt.Printf("%d: %s\n", i, cmd)
			}
		}
		if firstCommand != "" {
			fmt.Printf("0: %s\n", firstCommand)
		}

		if cfg.ZshAbbreviations {
			prefix := cfg.AbbreviationPrefix
			if err := abbr.UpdateZshAbbreviations(prefix, prefix, string(data)); err != nil {
				return fmt.Errorf("error updating zsh abbreviations: %w", err)
			}
		}
		return nil
	},
}
