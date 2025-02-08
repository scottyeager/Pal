package config

import (
	"fmt"
	"sort"
	"strings"
)

func Models(cfg *Config) error {
	models := []string{}
	providers := make([]string, 0, len(cfg.Providers))
	for provider := range cfg.Providers {
		providers = append(providers, provider)
	}
	// For now we sort alphabetically. It would probably be best to preserve
	// the ordering from the config file, but that requires a change to the
	// config structure or some additional parsing
	sort.Strings(providers)

	for _, provider := range providers {
		providerConfig := cfg.Providers[provider]
		for _, model := range providerConfig.Models {
			models = append(models, fmt.Sprintf("%s/%s", provider, model))
		}
	}
	fmt.Println("\nAvailable models:")
	for i, model := range models {
		fmt.Printf("%d. %s\n", i+1, model)
	}

	selectedNumber := ""
	if cfg.SelectedModel != "" {
		fmt.Printf("\nCurrently selected: %s\n", cfg.SelectedModel)
		fmt.Print("\nEnter model number or press Enter to keep current: ")
		fmt.Scanln(&selectedNumber)
		if selectedNumber == "" {
			fmt.Printf("Model set to: %s\n", cfg.SelectedModel)
			return nil
		}
	} else {
		fmt.Print("\nEnter model number or press Enter for default (1): ")
		fmt.Scanln(&selectedNumber)
		if selectedNumber == "" {
			selectedNumber = "1"
		}
	}

	var modelIndex int
	_, err := fmt.Sscanf(selectedNumber, "%d", &modelIndex)
	if err != nil || modelIndex < 1 || modelIndex > len(models) {
		return fmt.Errorf("Invalid model number")
	}

	selectedModel := models[modelIndex-1]
	parts := strings.SplitN(selectedModel, "/", 2)
	provider := parts[0]
	model := parts[1]

	providerConfig, exists := cfg.Providers[provider]
	if !exists {
		return fmt.Errorf("Provider '%s' not found in config\n", provider)
	}

	modelValid := false
	for _, configModel := range providerConfig.Models {
		if configModel == model {
			modelValid = true
			break
		}
	}
	if !modelValid {
		return fmt.Errorf("Model '%s' not found for provider '%s'\n", model, provider)
	}

	cfg.SelectedModel = provider + "/" + model
	if err := SaveConfig(cfg); err != nil {
		return fmt.Errorf("Error saving config: %v\n", err)
	}
	fmt.Printf("Model set to: %s\n", selectedModel)
	return nil
}
