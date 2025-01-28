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
	provider := cfg.Providers[cfg.SelectedProvider]
	model := cfg.SelectedModel
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

func GetCompletionStoragePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".local", "share", "pal_helper", "completions.txt"), nil
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
		storagePath, err := GetCompletionStoragePath()
		if err != nil {
			return completion, fmt.Errorf("failed to get storage path: %w", err)
		}

		// Create directory if it doesn't exist
		storageDir := filepath.Dir(storagePath)
		if err := os.MkdirAll(storageDir, 0755); err != nil {
			return completion, fmt.Errorf("failed to create storage directory: %w", err)
		}

		// Write completion to file
		file, err := os.OpenFile(storagePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return completion, fmt.Errorf("failed to open storage file: %w", err)
		}
		defer file.Close()

		if _, err := file.WriteString(completion + "\n"); err != nil {
			return completion, fmt.Errorf("failed to write completion: %w", err)
		}
	}

	return completion, nil
}
