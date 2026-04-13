package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	anthropicOption "github.com/anthropics/anthropic-sdk-go/option"
	openai "github.com/openai/openai-go/v3"
	openaiOption "github.com/openai/openai-go/v3/option"
	"github.com/scottyeager/pal/config"
	"github.com/scottyeager/pal/inout"
)

type Client struct {
	openaiClient    openai.Client
	anthropicClient *anthropic.Client
	provider        config.Provider
	model           string
	providerName    string
}

func NewClient(cfg *config.Config, modelName string) (*Client, error) {
	parts := strings.SplitN(modelName, "/", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("Model name %s isn't valid. Please use /models or /model to select a valid model.", modelName)
	}
	providerName, model := parts[0], parts[1]
	provider := cfg.Providers[providerName]

	var openaiClient openai.Client
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

func (c *Client) callLLM(ctx context.Context, anthropicMessages []anthropic.MessageParam, openaiMessages []openai.ChatCompletionMessageParamUnion, systemBlocks []anthropic.TextBlockParam, temperature float64) (string, error) {
	var response string

	if c.providerName == "anthropic" {
		message, err := c.anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
			Model:       anthropic.F(c.model),
			MaxTokens:   anthropic.F(int64(4096)),
			System:      anthropic.F(systemBlocks),
			Messages:    anthropic.F(anthropicMessages),
			Temperature: anthropic.F(temperature),
		})
		if err != nil {
			return "", fmt.Errorf("failed to get completion from anthropic: %w", err)
		}

		// Extract text content from message
		for _, block := range message.Content {
			if textBlock, ok := block.AsUnion().(anthropic.TextBlock); ok {
				response += textBlock.Text
			}
		}
	} else {
		resp, err := c.openaiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Messages:    openaiMessages,
			Model:       c.model,
			Temperature: openai.Float(temperature),
			MaxTokens:   openai.Int(4096),
		})

		if err != nil {
			var apierr *openai.Error
			if errors.As(err, &apierr) {
				return "", fmt.Errorf("failed to get completion from %s: %w\n%s\n", c.providerName, err, apierr.Response.Body)
			}
			return "", fmt.Errorf("failed to get completion from %s: %w", c.providerName, err)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("no response choices returned")
		}

		response = resp.Choices[0].Message.Content
	}

	return response, nil
}

// IsContextLengthError reports whether err appears to be a provider
// context-length / prompt-too-long error. Detection is by string match
// across Anthropic and OpenAI-style messages; if a provider changes its
// wording we fall through and the caller surfaces the raw error.
func IsContextLengthError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	patterns := []string{
		"prompt is too long",
		"context length",
		"maximum context",
		"context_length_exceeded",
		"string_above_max_length",
		"too many tokens",
		"input is too long",
	}
	for _, p := range patterns {
		if strings.Contains(msg, p) {
			return true
		}
	}
	return false
}

// Start or continue a conversation thread
func (c *Client) CallLLMWithThread(ctx context.Context, thread *inout.Thread, prompt string, temperature float64, formatMarkdown bool, model string) (string, error) {
	var anthropicMessages []anthropic.MessageParam
	var openaiMessages []openai.ChatCompletionMessageParamUnion
	var systemBlocks []anthropic.TextBlockParam

	// Add system prompt from thread
	systemBlocks = append(systemBlocks, anthropic.NewTextBlock(thread.SystemPrompt))
	openaiMessages = append(openaiMessages, openai.SystemMessage(thread.SystemPrompt))

	// Add history turns
	for _, turn := range thread.Turns {
		anthropicMessages = append(anthropicMessages,
			anthropic.NewUserMessage(anthropic.NewTextBlock(turn.Prompt)))
		anthropicMessages = append(anthropicMessages,
			anthropic.NewAssistantMessage(anthropic.NewTextBlock(turn.Response)))

		openaiMessages = append(openaiMessages, openai.UserMessage(turn.Prompt))
		openaiMessages = append(openaiMessages, openai.AssistantMessage(turn.Response))
	}

	// Add current prompt
	anthropicMessages = append(anthropicMessages,
		anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)))
	openaiMessages = append(openaiMessages, openai.UserMessage(prompt))

	response, err := c.callLLM(ctx, anthropicMessages, openaiMessages, systemBlocks, temperature)

	if err != nil {
		return "", err
	}

	return response, nil
}

// This is a one off LLM call with no thread support
func (c *Client) CallLLMWithPrompts(ctx context.Context, systemPrompt string, prompt string, temperature float64, formatMarkdown bool, model string) (string, error) {
	var anthropicMessages []anthropic.MessageParam
	var openaiMessages []openai.ChatCompletionMessageParamUnion
	var systemBlocks []anthropic.TextBlockParam

	// Add system prompt
	systemBlocks = append(systemBlocks, anthropic.NewTextBlock(systemPrompt))
	openaiMessages = append(openaiMessages, openai.SystemMessage(systemPrompt))

	// Add user prompt
	anthropicMessages = append(anthropicMessages,
		anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)))
	openaiMessages = append(openaiMessages, openai.UserMessage(prompt))

	response, err := c.callLLM(ctx, anthropicMessages, openaiMessages, systemBlocks, temperature)
	if err != nil {
		return "", err
	}

	return response, nil
}
