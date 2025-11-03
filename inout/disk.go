package inout

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scottyeager/pal/config"
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

func StorePrefix0Command(command string) error {
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

	// Read existing content to preserve expansions
	existingContent := ""
	if _, err := os.Stat(storagePath); err == nil {
		content, err := os.ReadFile(storagePath)
		if err != nil {
			return fmt.Errorf("failed to read existing commands: %w", err)
		}
		existingContent = string(content)
	}

	// Split content at the first newline to get expansions
	expansions := ""
	if parts := strings.SplitN(existingContent, "\n", 2); len(parts) > 1 {
		expansions = parts[1]
	}

	// Write new content with prefix0 command and preserved expansions
	newContent := command + "\n" + expansions
	if err := os.WriteFile(storagePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write prefix0 command to disk: %w", err)
	}

	return nil
}

func StoreCommands(completion string) error {
	basePath, err := config.GetBasePath()
	if err != nil {
		return fmt.Errorf("failed to get base path: %w", err)
	}
	storagePath := filepath.Join(basePath, commandFileName)

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

	// Previously we stored data in ~/.local/share/pal_helper. Since we
	// simplified to put everything under .config (or per user's
	// XDG_CONFIG_HOME), we can remove this folder entirely. This should be fast
	// enough to run every time without worry but a reasonable TODO might be to
	// somehow only run this if we detect an update to a new version
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get home dir, skip cleanup
		return nil
	}
	oldDir := filepath.Join(homeDir, ".local", "share", "pal_helper")
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		os.RemoveAll(oldDir)
	}

	return nil
}
