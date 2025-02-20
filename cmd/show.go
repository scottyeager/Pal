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
	showCmd.Flags().BoolP("all", "a", false, "Show all expansions including 0")
}

var showCmd = &cobra.Command{
	Use:   "/show",
	Short: "Show the last generated commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := inout.GetStoredCommands()
		if err != nil {
			return fmt.Errorf("error reading data from disk: %w", err)
		}

		showAll, _ := cmd.Flags().GetBool("all")
		commands := strings.Split(string(data), "\n")
		for i, cmd := range commands[1:] {
			if cmd == "" {
				continue
			}
			fmt.Printf("%d: %s\n", i, cmd)
		}

		// Display first command last if showing all
		if showAll && len(commands) > 0 && commands[0] != "" {
			fmt.Printf("0: %s\n", commands[0])
		}

		return nil
	},
}
