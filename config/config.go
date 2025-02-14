package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ZshAbbreviations   bool                `yaml:"zsh_abbreviations"`
	AbbreviationPrefix string              `yaml:"abbreviation_prefix"`
	Providers          map[string]Provider `yaml:"providers"`
	SelectedModel      string              `yaml:"selected_model"`
}

func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "pal_helper", "config.yaml"), nil
}

func LoadConfig() (*Config, error) {
	cfgPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func LoadConfigOrExit() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Println("error loading config: %w", err)
		os.Exit(1)
	}
	return cfg
}

func SaveConfig(cfg *Config) error {
	cfgPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func CheckConfiguration(cfg *Config) error {
	if len(cfg.Providers) == 0 {
		return fmt.Errorf("No providers configured. Run 'pal /config' to set up a provider")
	}
	if cfg.SelectedModel == "" {
		return fmt.Errorf("No model selected. Run 'pal /models' to select a model")
	}

	modelFound := false
	for provider_name, provider := range cfg.Providers {
		for _, model := range provider.Models {
			if provider_name+"/"+model == cfg.SelectedModel {
				modelFound = true
				break
			}
		}
		if modelFound {
			break
		}
	}

	if !modelFound {
		return fmt.Errorf("Selected model '%s' not found in current configuration. Run 'pal /models' to select a valid model", cfg.SelectedModel)
	}

	return nil
}
