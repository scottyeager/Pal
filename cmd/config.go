package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	var apiKey string
	if err == nil && existingCfg.APIKey != "" {
		fmt.Println("Only DeepSeek is supported as an LLM provider for now. More coming soon.")
		fmt.Printf("Found existing API key. Press enter to keep it, or enter a new one: ")
		fmt.Scanln(&apiKey)
		if apiKey == "" {
			apiKey = existingCfg.APIKey
		}
	} else {
		fmt.Println("Only DeepSeek is supported as an LLM provider for now. More coming soon.")
		fmt.Print("Enter your API key: ")
		fmt.Scanln(&apiKey)
	}

	shell := os.Getppid()
	bytes, err := os.ReadFile("/proc/" + fmt.Sprint(shell) + "/comm")
	processName := strings.TrimSpace(string(bytes))

	var enableZshAbbreviations bool
	if filepath.Base(processName) == "zsh" {
		fmt.Print("Do you want to enable zsh abbreviations? This requires the zsh-abbr plugin. (y/N): ")
		var response string
		fmt.Scanln(&response)
		enableZshAbbreviations = response == "y" || response == "Y"
	}
	cfg := &config.Config{
		APIKey:           apiKey,
		ZshAbbreviations: enableZshAbbreviations,
	}

	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		return
	}

	fmt.Printf("\nConfig saved successfully at %s\n", cfgPath)
}
