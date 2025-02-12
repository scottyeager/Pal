package cmd

import (
	"fmt"

	"github.com/scottyeager/pal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(modelCmd)
}

var modelCmd = &cobra.Command{
	Use:   "/model [model-name]",
	Short: "Switch to the specified model, or print model name.",
	Args:  cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		cfg, err := config.LoadConfig()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		var models []string
		for providerName, provider := range cfg.Providers {
			for _, model := range provider.Models {
				models = append(models, providerName+"/"+model)
			}
		}
		return models, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		if len(args) == 0 {
			fmt.Printf("Currently selected model: %s\n", cfg.SelectedModel)
			return nil
		}

		// Check if model exists in providers
		modelExists := false
		for providerName, provider := range cfg.Providers {
			for _, model := range provider.Models {
				if providerName+"/"+model == args[0] {
					modelExists = true
					break
				}
			}
			if modelExists {
				break
			}
		}

		if !modelExists {
			return fmt.Errorf("model '%s' not found in any provider", args[0])
		}

		cfg.SelectedModel = args[0]
		err = config.SaveConfig(cfg)
		if err != nil {
			return fmt.Errorf("error saving config: %w", err)
		}

		fmt.Printf("Switched to model: %s\n", args[0])
		return nil
	},
}
