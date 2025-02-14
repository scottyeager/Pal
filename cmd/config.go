package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scottyeager/pal/abbr"
	"github.com/scottyeager/pal/ai"
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

		// Try to find the shell name, for zsh specific config
		ppid := os.Getppid()
		bytes, err := os.ReadFile("/proc/" + fmt.Sprint(ppid) + "/comm")
		shell := strings.TrimSpace(string(bytes))

		var prefix string
		if existingCfg != nil && existingCfg.AbbreviationPrefix != "" {
			fmt.Printf("Current abbreviation prefix is '%s'. Press enter to keep it, or enter a new one: ", existingCfg.AbbreviationPrefix)
			fmt.Scanln(&prefix)
			if prefix == "" {
				prefix = existingCfg.AbbreviationPrefix
			} else {
				data, err := ai.GetStoredCompletion()
				if err != nil {
					fmt.Printf("Error reading data from disk: %v\n", err)
				}
				// This is the case where we updated the prefix and zsh abbrs were
				// already enabled, thus we should refresh them
				if data != "" && existingCfg.ZshAbbreviations {
					abbr.UpdateZshAbbreviations(existingCfg.AbbreviationPrefix, prefix, data)
				}
			}

		} else {
			fmt.Print("Enter abbreviation prefix (default 'pal'): ")
			fmt.Scanln(&prefix)
			if prefix == "" {
				prefix = "pal"
			}
		}

		var enableZshAbbreviations bool
		if filepath.Base(shell) == "zsh" {
			var defaultYes bool
			if existingCfg != nil {
				defaultYes = existingCfg.ZshAbbreviations
			}
			if defaultYes {
				fmt.Print(`Do you want to enable zsh abbreviations? This requires the zsh-abbr plugin. Any abbreviations with the form "$prefix$i" will be overwritten. (Y/n): `)
			} else {
				fmt.Print(`Do you want to enable zsh abbreviations? This requires the zsh-abbr plugin. Any abbreviations with the form "$prefix$i" will be overwritten. (y/N): `)
			}
			var response string
			fmt.Scanln(&response)
			if defaultYes {
				enableZshAbbreviations = response != "n" && response != "N"
			} else {
				enableZshAbbreviations = response == "y" || response == "Y"
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
			ZshAbbreviations:   enableZshAbbreviations,
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
