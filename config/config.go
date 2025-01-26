package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	APIKey string `yaml:"api_key"`
}

func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "pal_helper", "config.yaml"), nil
}

func LoadConfig() (*Config, error) {
	// TODO: Implement config loading
	return &Config{}, nil
}

func SaveConfig(cfg *Config) error {
	// TODO: Implement config saving
	return nil
}
