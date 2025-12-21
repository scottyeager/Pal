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

// buildMessages builds message arrays for both Anthropic and OpenAI from a thread
func (c *Client) buildMessages(thread *inout.Thread, prompt string) ([]anthropic.MessageParam, []openai.ChatCompletionMessageParamUnion, []anthropic.TextBlockParam) {
	var anthropicMessages []anthropic.MessageParam
	var openaiMessages []openai.ChatCompletionMessageParamUnion
	var systemBlocks []anthropic.TextBlockParam

	// Add system prompt if it exists
	if thread != nil && thread.SystemPrompt != "" {
		systemBlocks = append(systemBlocks, anthropic.NewTextBlock(thread.SystemPrompt))
		openaiMessages = append(openaiMessages, openai.SystemMessage(thread.SystemPrompt))
	}

	// Add history turns if thread exists
	if thread != nil {
		for _, turn := range thread.Turns {
			anthropicMessages = append(anthropicMessages,
				anthropic.NewUserMessage(anthropic.NewTextBlock(turn.Prompt)))
			anthropicMessages = append(anthropicMessages,
				anthropic.NewAssistantMessage(anthropic.NewTextBlock(turn.Response)))

			openaiMessages = append(openaiMessages, openai.UserMessage(turn.Prompt))
			openaiMessages = append(openaiMessages, openai.AssistantMessage(turn.Response))
		}
	}

	// Add current prompt
	anthropicMessages = append(anthropicMessages,
		anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)))
	openaiMessages = append(openaiMessages, openai.UserMessage(prompt))

	return anthropicMessages, openaiMessages, systemBlocks
}

// getCompletionInternal is the core completion logic for both providers
func (c *Client) getCompletionInternal(ctx context.Context, systemPrompt string, prompt string, thread *inout.Thread, temperature float64) (string, error) {
	var completion string

	if c.providerName == "anthropic" {
		anthropicMessages, _, systemBlocks := c.buildMessages(thread, prompt)

		// If no thread provided but systemPrompt exists, use it
		if thread == nil && systemPrompt != "" {
			systemBlocks = []anthropic.TextBlockParam{anthropic.NewTextBlock(systemPrompt)}
		}

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
				completion += textBlock.Text
			}
		}
	} else {
		_, openaiMessages, _ := c.buildMessages(thread, prompt)

		// If no thread provided but systemPrompt exists, add it
		if thread == nil && systemPrompt != "" {
			openaiMessages = append([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemPrompt),
			}, openaiMessages...)
		}

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
			return "", fmt.Errorf("no command choices returned")
		}

		completion = resp.Choices[0].Message.Content
	}

	return completion, nil
}

// formatCompletion handles markdown formatting
func (c *Client) formatCompletion(completion string) string {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	formatted, err := r.Render(completion)
	if err != nil {
		return completion
	}

	// Remove the two space margin included in Glamour's default styles
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
	return formatted
}

// cleanCompletion removes <tool_call> blocks
func (c *Client) cleanCompletion(completion string) string {
	// Remove <tool_call> block if present
	if thinkStart := strings.Index(completion, "<tool_call>"); thinkStart != -1 {
		if thinkEnd := strings.Index(completion[thinkStart:], "</tool_call>"); thinkEnd != -1 {
			completion = completion[:thinkStart] + strings.TrimSpace(completion[thinkStart+thinkEnd+len("</tool_call>"):])
		}
	}
	return completion
}

// GetCompletionWithHistory starts a new conversation thread with history tracking
func (c *Client) GetCompletionWithHistory(ctx context.Context, commandName string, systemPrompt string, prompt string, temperature float64, formatMarkdown bool, model string) (string, *inout.Thread, error) {
	// Create history manager
	history, err := inout.NewHistory()
	if err != nil {
		return "", nil, fmt.Errorf("failed to create history manager: %w", err)
	}

	// Start a new thread with command name
	thread, err := history.StartNewThread(commandName)
	if err != nil {
		return "", nil, fmt.Errorf("failed to start new thread: %w", err)
	}

	// Update the thread with the actual system prompt content
	if systemPrompt != "" {
		if err := history.UpdateSystemPrompt(thread, systemPrompt); err != nil {
			// Log but don't fail
			fmt.Printf("Warning: failed to store system prompt: %v\n", err)
		}
	}

	// Get completion using internal function
	completion, err := c.getCompletionInternal(ctx, systemPrompt, prompt, nil, temperature)
	if err != nil {
		return "", nil, err
	}

	// Clean the completion (remove <tool_call> blocks) for storage
	cleanCompletion := c.cleanCompletion(completion)

	// Format for display
	var formattedCompletion string
	if formatMarkdown {
		formattedCompletion = c.formatCompletion(completion)
	} else {
		formattedCompletion = completion
	}

	// Store the clean (unformatted) completion in history
	if err := history.AddTurn(thread, prompt, cleanCompletion); err != nil {
		// Don't fail the request if history storage fails, just log it
		fmt.Printf("Warning: failed to store history: %v\n", err)
	}

	return formattedCompletion, thread, nil
}

// ContinueCompletion continues an existing conversation thread
func (c *Client) ContinueCompletion(ctx context.Context, thread *inout.Thread, prompt string, temperature float64, formatMarkdown bool, model string) (string, error) {
	completion, err := c.getCompletionInternal(ctx, "", prompt, thread, temperature)
	if err != nil {
		return "", err
	}

	// Clean the completion (remove <tool_call> blocks) for storage
	cleanCompletion := c.cleanCompletion(completion)

	// Format for display
	var formattedCompletion string
	if formatMarkdown {
		formattedCompletion = c.formatCompletion(completion)
	} else {
		formattedCompletion = completion
	}

	// Store the clean (unformatted) completion in history
	history, err := inout.NewHistory()
	if err != nil {
		// Don't fail the request if history manager creation fails
		fmt.Printf("Warning: failed to create history manager: %v\n", err)
		return formattedCompletion, nil
	}

	if err := history.AddTurn(thread, prompt, cleanCompletion); err != nil {
		// Don't fail the request if history storage fails, just log it
		fmt.Printf("Warning: failed to store history: %v\n", err)
	}

	return formattedCompletion, nil
}

// GetCompletion gets a completion from the AI model without history tracking
func (c *Client) GetCompletion(ctx context.Context, systemPrompt string, prompt string, temperature float64, formatMarkdown bool, model string) (string, error) {
	completion, err := c.getCompletionInternal(ctx, systemPrompt, prompt, nil, temperature)
	if err != nil {
		return "", err
	}

	// Clean the completion (remove <tool_call> blocks)
	cleanCompletion := c.cleanCompletion(completion)

	// Format for display
	var formattedCompletion string
	if formatMarkdown {
		formattedCompletion = c.formatCompletion(cleanCompletion)
	} else {
		formattedCompletion = cleanCompletion
	}

	return formattedCompletion, nil
}
