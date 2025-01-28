package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/scottyeager/pal/config"
)

func Models(cfg *config.Config) {
	models := []string{}
	for provider, providerConfig := range cfg.Providers {
		for _, model := range providerConfig.Models {
			models = append(models, fmt.Sprintf("%s/%s", provider, model))
		}
	}
	fmt.Println("Available models:")
	for i, model := range models {
		fmt.Printf("%d. %s\n", i+1, model)
	}
	fmt.Printf("\nCurrently selected: %s/%s\n", cfg.SelectedProvider, cfg.SelectedModel)
	fmt.Print("\nEnter model number or press Enter to keep current: ")
	var selectedNumber string
	fmt.Scanln(&selectedNumber)

	if selectedNumber == "" {
		return
	}

	var modelIndex int
	_, err := fmt.Sscanf(selectedNumber, "%d", &modelIndex)
	if err != nil || modelIndex < 1 || modelIndex > len(models) {
		fmt.Println("Invalid model number")
		os.Exit(1)
	}

	selectedModel := models[modelIndex-1]
	parts := strings.SplitN(selectedModel, "/", 2)
	provider := parts[0]
	model := parts[1]

	providerConfig, exists := cfg.Providers[provider]
	if !exists {
		fmt.Printf("Provider '%s' not found in config\n", provider)
		os.Exit(1)
	}

	modelValid := false
	for _, configModel := range providerConfig.Models {
		if configModel == model {
			modelValid = true
			break
		}
	}
	if !modelValid {
		fmt.Printf("Model '%s' not found for provider '%s'\n", model, provider)
		os.Exit(1)
	}

	cfg.SelectedProvider = provider
	cfg.SelectedModel = model
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Model set to: %s\n", selectedModel)
}
