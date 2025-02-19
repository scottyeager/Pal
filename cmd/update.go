package cmd

import (
	"fmt"

	"github.com/scottyeager/pal/io"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "/update",
	Short: "Update Pal to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		updateCmd := `wget https://github.com/scottyeager/Pal/releases/latest/download/pal-linux-amd64 -O /usr/local/bin/pal && chmod +x /usr/local/bin/pal`
		err := io.StorePrefix0Command(updateCmd)
		if err != nil {
			return fmt.Errorf("error storing update command: %w", err)
		}
		fmt.Println("Update command stored as prefix0. Run 'pal /show' to see it.")
		return nil
	},
}
