package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scottyeager/pal/abbr"
	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/config"
)

func Configure() {
	cfgPath, err := config.GetConfigPath()
	if err != nil {
		fmt.Printf("Error getting config path: %v\n", err)
		return
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("Error creating config directory: %v\n", err)
		return
	}

	existingCfg, err := config.LoadConfig()
	// Check for configured providers
	providers := make(map[string]config.Provider)
	if existingCfg != nil && existingCfg.Providers != nil {
		providers = existingCfg.Providers
		fmt.Println("Already configured providers:")
		for name := range existingCfg.Providers {
			fmt.Println(name)
		}
		fmt.Println()
	}

	// Display available provider templates
	templates := make([]string, 0, len(config.ProviderTemplates))
	for name := range config.ProviderTemplates {
		templates = append(templates, name)
	}

	for {
		fmt.Println("\nAvailable provider templates:")
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
			if data != "" {
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

	cfg := &config.Config{
		Providers:          providers,
		ZshAbbreviations:   enableZshAbbreviations,
		AbbreviationPrefix: prefix,
	}

	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		return
	}

	fmt.Printf("\nConfig saved successfully at %s\n", cfgPath)
}
