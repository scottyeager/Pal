package session

import (
	"context"
	"fmt"
	"strings"

	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/config"
	"github.com/scottyeager/pal/inout"
)

// Session holds the values of the CLI args
type Session struct {
	Temperature *float64
	Markdown    bool
	Chat        bool
}

// We duplicate some fields here because we want to centralize the logic for
// determining the final behavior in this module. The config, CLI args, and the
// individual command params (defaults or preferences) will be reconciled here
type TurnParams struct {
	Command     string
	Temperature float64
	Markdown    bool
	Chat        bool
	UserMessage []string
	formatFunc  func(string) string
}

func (s *Session) ExecuteTurn(params TurnParams) (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("error loading config: %w", err)
	}

	if err := config.CheckConfiguration(cfg); err != nil {
		return "", fmt.Errorf("configuration check failed: %w", err)
	}
	selectedModel := config.GetSelectedModel(cfg, params.Command)

	aiClient, err := ai.NewClient(cfg, selectedModel)
	if err != nil {
		return "", fmt.Errorf("error creating AI client: %w", err)
	}

	stdInput, err := inout.ReadStdin()
	if err != nil {
		return "", fmt.Errorf("error reading stdin: %w", err)
	}

	var userInput string
	if stdInput != "" && len(params.UserMessage) > 0 {
		userInput = stdInput + "\nThat concludes the stdin contents. Now here's the query from the user:\n" + strings.Join(params.UserMessage, " ")
	} else if stdInput != "" {
		userInput = stdInput
	} else {
		userInput = strings.Join(params.UserMessage, " ")
	}

	// Temp set on CLI take precedence, otherwise use default temp
	var temperature float64
	if s.Temperature == nil {
		temperature = params.Temperature
	} else {
		temperature = *s.Temperature
	}

	// The markdown param indicates whether a command would ever use Markdown
	// for the output. From there the config is the default and CLI toggles the
	// config value
	var markdown bool
	if params.Markdown {
		if (cfg.FormatMarkdown && !s.Markdown) || (!cfg.FormatMarkdown && s.Markdown) {
			markdown = true
		} else {
			markdown = false
		}
	} else {
		markdown = false
	}

	// Commands don't usually specify whether to use chat mode. The /chat
	// command is the exception
	var chat bool
	if params.Chat || s.Chat {
		chat = true
	} else {
		chat = false
	}

	ctx := context.Background()
	if chat {
		return aiClient.ContinueCompletion(ctx, nil, userInput, selectedModel)
	} else {
		return aiClient.GetCompletion(ctx, "", userInput, temperature, markdown, selectedModel)
	}
}
