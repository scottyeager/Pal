package cmd

import (
	"fmt"
	"strings"

	"github.com/scottyeager/pal/inout"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(updateCmd)
}

var showCmd = &cobra.Command{
	Use:   "/show",
	Short: "Show the last generated commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := inout.GetStoredCommands()
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

		return nil
	},
}
