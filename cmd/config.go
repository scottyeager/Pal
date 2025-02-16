package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/scottyeager/pal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "/config",
	Short: "Configure pal",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := config.GetConfigPath()
		if err != nil {
			return fmt.Errorf("error getting config path: %v", err)
		}

		// Create config directory if it doesn't exist
		configDir := filepath.Dir(cfgPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("error creating config directory: %v", err)
		}

		existingCfg, err := config.LoadConfig()
		// Check for configured providers
		providers := make(map[string]config.Provider)

		// Get names of providers with a template
		templates := make([]string, 0, len(config.ProviderTemplates))
		for name := range config.ProviderTemplates {
			templates = append(templates, name)
		}

		if existingCfg != nil && existingCfg.Providers != nil {
			providers = existingCfg.Providers
		}

		for {
			if len(providers) > 0 {
				fmt.Println("\nConfigured providers:")
				for name := range providers {
					fmt.Println(name)
				}
			}

			fmt.Println("\nAvailable providers:")
			for i, t := range templates {
				fmt.Printf("%d. %s\n", i+1, t)
			}
			fmt.Printf("\nPress enter when done, or select provider (1-%d): ", len(templates))

			var input string
			fmt.Scanln(&input)
			if input == "" {
				break
			}

			var choice int
			fmt.Sscanf(input, "%d", &choice)
			if choice < 1 || choice > len(templates) {
				fmt.Println("Invalid choice. Please try again.")
				continue
			}

			selectedProvider := templates[choice-1]

			// Prompt for API key
			var apiKey string
			if err == nil && existingCfg != nil && existingCfg.Providers != nil {
				provider, exists := existingCfg.Providers[selectedProvider]
				if exists && provider.APIKey != "" {
					fmt.Printf("Found existing API key for %s. Press enter to keep it, or enter a new one: ", selectedProvider)
					fmt.Scanln(&apiKey)
					if apiKey == "" {
						apiKey = provider.APIKey
					}
				} else {
					fmt.Printf("Enter your %s API key: ", selectedProvider)
					fmt.Scanln(&apiKey)
				}
			} else {
				fmt.Printf("Enter your %s API key: ", selectedProvider)
				fmt.Scanln(&apiKey)
			}

			providers[selectedProvider] = config.NewProvider(selectedProvider, apiKey)
		}

		var prefix string
		if existingCfg != nil && existingCfg.AbbreviationPrefix != "" {
			fmt.Printf("Current abbreviation prefix is '%s'. Press enter to keep it, or enter a new one: ", existingCfg.AbbreviationPrefix)
			fmt.Scanln(&prefix)
			if prefix == "" {
				prefix = existingCfg.AbbreviationPrefix
			}

		} else {
			fmt.Print("Enter abbreviation prefix (default 'pal'): ")
			fmt.Scanln(&prefix)
			if prefix == "" {
				prefix = "pal"
			}
		}

		// Prompt for markdown formatting
		if existingCfg != nil && existingCfg.FormatMarkdown {
			fmt.Printf("Markdown formatting is currently enabled. Keep it enabled? (Y/n): ")
		} else {
			fmt.Printf("Markdown formatting is currently disabled. Enable it? (y/N): ")
		}
		var markdownResponse string
		fmt.Scanln(&markdownResponse)
		var formatMarkdown bool
		if markdownResponse == "" {
			formatMarkdown = existingCfg != nil && existingCfg.FormatMarkdown
		} else {
			formatMarkdown = markdownResponse == "y" || markdownResponse == "Y"
		}

		cfg := &config.Config{
			Providers:          providers,
			AbbreviationPrefix: prefix,
			FormatMarkdown:     formatMarkdown,
		}

		// If there's no model configured but there's a provider configured now,
		// prompt the user to choose a model
		if len(providers) > 0 {
			if existingCfg != nil && existingCfg.SelectedModel != "" {
				cfg.SelectedModel = existingCfg.SelectedModel
			} else if existingCfg.SelectedModel == "" {
				err = config.Models(cfg)
				if err != nil {
					return err
				}
			}
		}

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("error saving config: %v", err)
		}

		fmt.Printf("\nConfig saved successfully at %s\n", cfgPath)
		return nil
	},
}
