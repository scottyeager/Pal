package session

import (
	"context"
	"fmt"
	"strings"

	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/config"
	"github.com/scottyeager/pal/inout"
)

// Session holds the values of the CLI args. We normally don't modify these,
// except in the case of /chat, which sets Chat to true.
type Session struct {
	Config      *config.Config
	Client      *ai.Client
	Temperature float64
	Markdown    bool
	Chat        bool
	UserMessage []string
}

// We duplicate some fields here because we want to centralize the logic for
// determining the final behavior in this module. The config, CLI args, and the
// individual command params (defaults or preferences) will be reconciled here
type TurnParams struct {
	Command      string
	SystemPrompt string
	Prompt       string
	Temperature  float64
	Markdown     bool
	FormatFunc   func(string) string
}

// NewTurnParams creates a new TurnParams with default values.
func NewTurnParams() TurnParams {
	return TurnParams{
		Command:      "",
		SystemPrompt: "",
		Prompt:       "",
		Temperature:  -1,
		Markdown:     false,
		FormatFunc:   nil,
	}
}

func (s *Session) ExecuteTurn(params TurnParams) (string, error) {
	var err error
	s.Config, err = config.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("error loading config: %w", err)
	}

	if err := config.CheckConfiguration(s.Config); err != nil {
		return "", fmt.Errorf("configuration check failed: %w", err)
	}
	model := config.GetSelectedModel(s.Config, params.Command)

	s.Client, err = ai.NewClient(s.Config, model)
	if err != nil {
		return "", fmt.Errorf("error creating AI client: %w", err)
	}

	// Commands can implement custom prompt generation, otherwise we fall back
	// to the default
	var userPrompt string
	if params.Prompt != "" {
		userPrompt = params.Prompt
	} else {
		userPrompt, err = s.ConstructPrompt()
		if err != nil {
			return "", fmt.Errorf("error constructing prompt: %w", err)
		}
	}

	// Temp set on CLI take precedence, otherwise use default temp
	var temperature float64
	if s.Temperature == -1 {
		temperature = params.Temperature
	} else {
		temperature = s.Temperature
	}

	// The markdown param indicates whether a command would ever use Markdown
	// for the output. From there the config is the default and CLI toggles the
	// config value
	var markdown bool
	if params.Markdown {
		if (s.Config.FormatMarkdown && !s.Markdown) || (!s.Config.FormatMarkdown && s.Markdown) {
			markdown = true
		} else {
			markdown = false
		}
	} else {
		markdown = false
	}

	systemPrompt := params.SystemPrompt
	if !markdown {
		systemPrompt += " and avoid using Markdown formatting that won't render in a shell. Lists and bullets are fine, but avoid headings, bold, and italic text."
	}

	ctx := context.Background()
	if s.Chat {
		// Create history manager
		history, err := inout.NewHistory()
		if err != nil {
			return "", fmt.Errorf("failed to create history manager: %w", err)
		}

		// Get or create thread
		thread, err := history.GetLatestThread()
		if err != nil {
			return "", fmt.Errorf("failed to get latest thread: %w", err)
		}

		// Continue the conversation with the existing thread
		response, err := s.Client.ContinueCompletion(ctx, thread, userPrompt, temperature, markdown, model)
		if err != nil {
			return "", fmt.Errorf("error continuing completion: %w", err)
		}

		// Print thread info for reference
		fmt.Printf("[Chat thread: %s/%d]\n", thread.Date.Format("2006-01-02"), thread.ID)
		return response, nil

	} else {
		// Create a new thread each time (no -c flag)
		response, thread, err := s.Client.GetCompletionWithHistory(ctx, params.Command, systemPrompt, userPrompt, temperature, markdown, model)
		if err != nil {
			return "", fmt.Errorf("error getting completion with history: %w", err)
		}
		// Print thread info for reference
		if thread != nil {
			fmt.Printf("[New chat thread: %s/%d]\n", thread.Date.Format("2006-01-02"), thread.ID)
		}
		return response, nil
	}
}

func (s *Session) ConstructPrompt() (string, error) {
	stdInput, err := inout.ReadStdin()
	if err != nil {
		return "", fmt.Errorf("error reading stdin: %w", err)
	}

	var userPrompt string
	if stdInput != "" && len(s.UserMessage) > 0 {
		userPrompt = stdInput + "\nThat concludes the stdin contents. Now here's the query from the user:\n" + strings.Join(s.UserMessage, " ")
	} else if stdInput != "" {
		userPrompt = stdInput
	} else {
		userPrompt = strings.Join(s.UserMessage, " ")
	}

	return userPrompt, nil
}
