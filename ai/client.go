package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	anthropicOption "github.com/anthropics/anthropic-sdk-go/option"
	openai "github.com/openai/openai-go"
	openaiOption "github.com/openai/openai-go/option"
	"github.com/scottyeager/pal/config"
)

const commandFileName = "expansions.txt"

type Client struct {
	openaiClient    *openai.Client
	anthropicClient *anthropic.Client
	provider        config.Provider
	model           string
	providerName    string
}

func NewClient(cfg *config.Config) (*Client, error) {
	parts := strings.SplitN(cfg.SelectedModel, "/", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("Model name %s isn't valid. Please use /models to select a model again.", cfg.SelectedModel)
	}
	providerName, model := parts[0], parts[1]
	provider := cfg.Providers[providerName]

	var openaiClient *openai.Client
	var anthropicClient *anthropic.Client

	switch providerName {
	case "anthropic":
		anthropicClient = anthropic.NewClient(
			anthropicOption.WithAPIKey(provider.APIKey),
			anthropicOption.WithBaseURL(provider.URL),
		)
	default:
		openaiClient = openai.NewClient(
			openaiOption.WithAPIKey(provider.APIKey),
			openaiOption.WithBaseURL(provider.URL),
		)
	}

	return &Client{
		openaiClient:    openaiClient,
		anthropicClient: anthropicClient,
		provider:        provider,
		model:           model,
		providerName:    providerName,
	}, nil
}

func GetStoredCompletion() (string, error) {
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

func StoreCompletion(completion string) error {
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

func (c *Client) GetCompletion(ctx context.Context, system_prompt string, prompt string, storeCompletion bool, temperature float64, formatMarkdown bool) (string, error) {
	var completion string
	var err error

	if c.providerName == "anthropic" {
		message, err := c.anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.F(c.model),
			MaxTokens: anthropic.F(int64(1024)),
			System: anthropic.F([]anthropic.TextBlockParam{
				anthropic.NewTextBlock(system_prompt),
			}),
			Messages: anthropic.F([]anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			}),
			Temperature: anthropic.F(temperature),
		})
		if err != nil {
			return "", fmt.Errorf("failed to get completion: %w", err)
		}

		// Extract text content from message
		for _, block := range message.Content {
			if textBlock, ok := block.AsUnion().(anthropic.TextBlock); ok {
				completion += textBlock.Text
			}
		}
	} else {
		resp, err := c.openaiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(system_prompt),
				openai.UserMessage(prompt),
			}),
			Model:       openai.F(c.model),
			Temperature: openai.F(temperature),
		})
		if err != nil {
			return "", fmt.Errorf("failed to get completion: %w", err)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("no completion choices returned")
		}

		completion = resp.Choices[0].Message.Content
	}

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

	if formatMarkdown {
		r, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(80),
		)
		formatted, err := r.Render(completion)
		if err != nil {
			return completion, nil
		}

		// Remove the two space margin included in Glamour's default styles
		// Since there's color codes included before the actual spaces, we need
		// to remove what's before the spaces too. Yeah, it would be cleaner to
		// actually ship updated styles, but this is easy and seems to work
		lines := strings.Split(formatted, "\n")
		for i, line := range lines {
			parts := strings.SplitN(line, "  ", 2)
			if len(parts) > 1 {
				lines[i] = parts[1]
			} else {
				lines[i] = parts[0]
			}
		}
		formatted = strings.Join(lines, "\n")

		// Also trim one newline from the end, again to adjust default style
		if strings.HasSuffix(formatted, "\n") {
			formatted = formatted[:len(formatted)-1]
		}
		return formatted, nil
	}
	return completion, nil
}
