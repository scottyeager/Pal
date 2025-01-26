package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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

	fmt.Print("Enter your API key: ")
	var apiKey string
	fmt.Scanln(&apiKey)

	cfg := &config.Config{
		APIKey: apiKey,
	}

	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		return
	}

	fmt.Printf("Config saved successfully\n")
}
