package cmd

import (
	"fmt"

	"github.com/scottyeager/pal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(modelsCmd)
}

var modelsCmd = &cobra.Command{
	Use:   "/models",
	Short: "View and select models",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}
		err = config.Models(cfg)
		if err != nil {
			return fmt.Errorf("error setting model: %w", err)
		}
		return nil
	},
}
