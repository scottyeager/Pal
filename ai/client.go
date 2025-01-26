package ai

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/scottyeager/pal/config"
)

type Client struct {
	client *openai.Client
	model  string
}

func NewClient(cfg *config.Config) (*Client, error) {
	client := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL("https://api.deepseek.com"),
	)

	return &Client{
		client: client,
		model:  "deepseek-chat",
	}, nil
}

func (c *Client) GetCompletion(ctx context.Context, prompt string) (string, error) {
	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a helpful assistant that suggests shell commands"),
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

	return resp.Choices[0].Message.Content, nil
}
