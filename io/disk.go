package io

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const commandFileName = "expansions.txt"

func GetStoredCommands() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	storagePath := filepath.Join(homeDir, ".local", "share", "pal_helper", commandFileName)

	// Ensure directory exists
	storageDir := filepath.Dir(storagePath)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Create file if it doesn't exist
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		file, err := os.Create(storagePath)
		if err != nil {
			return "", fmt.Errorf("failed to create storage file: %w", err)
		}
		file.Close()
	}

	// Read file contents
	content, err := os.ReadFile(storagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read completion file: %w", err)
	}

	return string(content), nil
}

func StoreCommands(completion string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	storagePath := filepath.Join(homeDir, ".local", "share", "pal_helper", commandFileName)

	// Ensure directory exists
	storageDir := filepath.Dir(storagePath)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Read existing content to preserve prefix0 if it exists
	existingContent := ""
	if _, err := os.Stat(storagePath); err == nil {
		content, err := os.ReadFile(storagePath)
		if err != nil {
			return fmt.Errorf("failed to read existing commands: %w", err)
		}
		existingContent = string(content)
	}

	// Split content at the first newline to preserve prefix0
	parts := strings.SplitN(existingContent, "\n", 2)
	prefix0 := ""
	if len(parts) > 0 {
		prefix0 = parts[0] + "\n"
	}

	// Write new content with preserved prefix0
	newContent := prefix0 + completion
	if err := os.WriteFile(storagePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write commands to disk: %w", err)
	}

	// Clean up the old file if it exists. This shouldn't slow us down if the
	// file no longer exists, since stat hits cached data
	oldPath := filepath.Join(homeDir, ".local", "share", "pal_helper", "completions.txt")
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		// If both exist, remove the old one
		os.Remove(oldPath)
	}

	return nil
}
