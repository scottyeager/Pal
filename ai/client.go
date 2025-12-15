package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"

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

func (c *Client) GetCompletion(ctx context.Context, system_prompt string, prompt string, temperature float64, formatMarkdown bool, model string) (string, error) {
	var completion string

	if c.providerName == "anthropic" {
		message, err := c.anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.F(c.model),
			MaxTokens: anthropic.F(int64(4096)),
			System: anthropic.F([]anthropic.TextBlockParam{
				anthropic.NewTextBlock(system_prompt),
			}),
			Messages: anthropic.F([]anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			}),
			Temperature: anthropic.F(temperature),
		})
		if err != nil {
			return "", fmt.Errorf("failed to get completion from anthropic: %w", err)
		}

		// Extract text content from message
		for _, block := range message.Content {
			if textBlock, ok := block.AsUnion().(anthropic.TextBlock); ok {
				completion += textBlock.Text
			}
		}
	} else {
		resp, err := c.openaiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(system_prompt),
				openai.UserMessage(prompt),
			},
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
			return "", fmt.Errorf("no command choices returned")
		}

		completion = resp.Choices[0].Message.Content
	}

	// Remove <tool_call> block if present
	if thinkStart := strings.Index(completion, "<tool_call>"); thinkStart != -1 {
		if thinkEnd := strings.Index(completion[thinkStart:], "</tool_call>"); thinkEnd != -1 {
			completion = completion[:thinkStart] + strings.TrimSpace(completion[thinkStart+thinkEnd+len("</tool_call>"):])
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

// GetCompletionWithHistory starts a new conversation thread with history tracking
func (c *Client) GetCompletionWithHistory(ctx context.Context, system_prompt string, prompt string, temperature float64, formatMarkdown bool, model string) (string, *inout.Thread, error) {
	// Create history manager
	history, err := inout.NewHistory()
	if err != nil {
		return "", nil, fmt.Errorf("failed to create history manager: %w", err)
	}

	// Start a new thread with system prompt
	thread, err := history.StartNewThread(system_prompt)
	if err != nil {
		return "", nil, fmt.Errorf("failed to start new thread: %w", err)
	}

	// Get completion
	completion, err := c.GetCompletion(ctx, system_prompt, prompt, temperature, formatMarkdown, model)
	if err != nil {
		return "", nil, err
	}

	// Store the turn in history
	if err := history.AddTurn(thread, prompt, completion); err != nil {
		// Don't fail the request if history storage fails, just log it
		fmt.Printf("Warning: failed to store history: %v\n", err)
	}

	return completion, thread, nil
}

// ContinueCompletion continues an existing conversation thread
func (c *Client) ContinueCompletion(ctx context.Context, thread *inout.Thread, prompt string, temperature float64, formatMarkdown bool, model string) (string, error) {
	// Create history manager
	history, err := inout.NewHistory()
	if err != nil {
		return "", fmt.Errorf("failed to create history manager: %w", err)
	}

	var completion string

	if c.providerName == "anthropic" {
		// Build messages from history
		var messages []anthropic.MessageParam

		// Add system prompt if it exists
		var systemBlocks []anthropic.TextBlockParam
		if thread.SystemPrompt != "" {
			systemBlocks = append(systemBlocks, anthropic.NewTextBlock(thread.SystemPrompt))
		}

		// Add history turns
		for _, turn := range thread.Turns {
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(turn.Prompt)))
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(turn.Response)))
		}

		// Add current prompt
		messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)))

		message, err := c.anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
			Model:       anthropic.F(c.model),
			MaxTokens:   anthropic.F(int64(4096)),
			System:      anthropic.F(systemBlocks),
			Messages:    anthropic.F(messages),
			Temperature: anthropic.F(temperature),
		})
		if err != nil {
			return "", fmt.Errorf("failed to get completion from anthropic: %w", err)
		}

		// Extract text content from message
		for _, block := range message.Content {
			if textBlock, ok := block.AsUnion().(anthropic.TextBlock); ok {
				completion += textBlock.Text
			}
		}
	} else {
		// Build messages from history
		var messages []openai.ChatCompletionMessageParamUnion

		// Add system prompt if it exists
		if thread.SystemPrompt != "" {
			messages = append(messages, openai.SystemMessage(thread.SystemPrompt))
		}

		// Add history turns
		for _, turn := range thread.Turns {
			messages = append(messages, openai.UserMessage(turn.Prompt))
			messages = append(messages, openai.AssistantMessage(turn.Response))
		}

		// Add current prompt
		messages = append(messages, openai.UserMessage(prompt))

		resp, err := c.openaiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Messages:    messages,
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
			return "", fmt.Errorf("no command choices returned")
		}

		completion = resp.Choices[0].Message.Content
	}

	// Remove <tool_call> block if present
	if thinkStart := strings.Index(completion, "<tool_call>"); thinkStart != -1 {
		if thinkEnd := strings.Index(completion[thinkStart:], "</tool_call>"); thinkEnd != -1 {
			completion = completion[:thinkStart] + strings.TrimSpace(completion[thinkStart+thinkEnd+len("</tool_call>"):])
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
		completion = formatted
	}

	// Store the turn in history
	if err := history.AddTurn(thread, prompt, completion); err != nil {
		// Don't fail the request if history storage fails, just log it
		fmt.Printf("Warning: failed to store history: %v\n", err)
	}

	return completion, nil
}
