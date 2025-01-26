package cmd

import (
	"fmt"
	"os"

	"github.com/scottyeager/pal/config"
	"gopkg.in/yaml.v3"
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

	cfg := config.Config{
		APIKey: apiKey,
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		fmt.Printf("Error marshaling config: %v\n", err)
		return
	}

	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		fmt.Printf("Error writing config file: %v\n", err)
		return
	}

	fmt.Printf("Config saved to %s\n", cfgPath)
}
