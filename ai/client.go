package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/scottyeager/pal/config"
)

type Client struct {
	client   *openai.Client
	provider config.Provider
	model    string
}

func NewClient(cfg *config.Config) (*Client, error) {
	parts := strings.SplitN(cfg.SelectedModel, "/", 2)
	provider_name, model := parts[0], parts[1]
	provider := cfg.Providers[provider_name]
	client := openai.NewClient(
		option.WithAPIKey(provider.APIKey),
		option.WithBaseURL(provider.URL),
	)

	return &Client{
		client:   client,
		provider: provider,
		model:    model,
	}, nil
}

func GetStoredCompletion() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	storagePath := filepath.Join(homeDir, ".local", "share", "pal_helper", "completions.txt")

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

func StoreCompletion(completion string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	storagePath := filepath.Join(homeDir, ".local", "share", "pal_helper", "completions.txt")

	// Ensure directory exists
	storageDir := filepath.Dir(storagePath)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Write completion to file
	if err := os.WriteFile(storagePath, []byte(completion), 0644); err != nil {
		return fmt.Errorf("failed to write completion: %w", err)
	}

	return nil
}

func (c *Client) GetCompletion(ctx context.Context, system_prompt string, prompt string, storeCompletion bool) (string, error) {
	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system_prompt),
			openai.UserMessage(prompt),
		}),
		Model: openai.F(c.model),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	completion := resp.Choices[0].Message.Content

	// Remove <think> block if present
	if thinkStart := strings.Index(completion, "<think>"); thinkStart != -1 {
		if thinkEnd := strings.Index(completion, "</think>"); thinkEnd != -1 {
			completion = completion[:thinkStart] + strings.TrimSpace(completion[thinkEnd+len("</think>"):])
		}
	}
	if storeCompletion {
		// Store the completion
		err = StoreCompletion(completion)
		if err != nil {
			return completion, fmt.Errorf("failed to write to disk: %w", err)
		}
	}

	return completion, nil
}
