package session

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
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

	// Create history manager
	history, err := inout.NewHistory()
	if err != nil {
		return "", fmt.Errorf("failed to create history manager: %w", err)
	}

	var thread *inout.Thread
	if s.Chat {
		// Get or create thread
		thread, err = history.GetLatestThread()
		if err != nil {
			return "", fmt.Errorf("failed to get latest thread: %w", err)
		}
		// Print thread info for reference. Go to stderr as to not interfere
		// with piping /file output directly to a file
		fmt.Fprintf(os.Stderr, "[Chat thread: %s/%d]\n", thread.Date.Format("2006-01-02"), thread.ID)
	} else {
		thread, err = history.StartNewThread(params.Command)
		if err != nil {
			return "", fmt.Errorf("error creating new thread: %w", err)
		}
		history.UpdateSystemPrompt(thread, systemPrompt)
		// Maybe this could go behind a "debug" or "verbose" flag
		// fmt.Printf("[New chat thread: %s/%d]\n", thread.Date.Format("2006-01-02"), thread.ID)
	}

	ctx := context.Background()
	response, err := s.Client.CallLLMWithThread(ctx, thread, userPrompt, temperature, markdown, model)
	if err != nil {
		return "", fmt.Errorf("error continuing completion: %w", err)
	}

	response = cleanToolCalls(response)

	history.AddTurn(thread, userPrompt, response)

	if markdown {
		response = formatMarkdown(response)
	}

	return response, nil
}

func (s *Session) ConstructPrompt() (string, error) {
	stdInput, err := inout.ReadStdin()
	if err != nil {
		return "", fmt.Errorf("error reading stdin: %w", err)
	}

	userMessage := s.UserMessage
	// If the user's message ends with "..." and nothing is piped on stdin,
	// drop into interactive stdin mode so they can paste/type input that
	// would be awkward to pass as CLI args.
	if stdInput == "" && len(userMessage) > 0 && strings.HasSuffix(userMessage[len(userMessage)-1], "...") {
		last := strings.TrimSuffix(userMessage[len(userMessage)-1], "...")
		trimmed := make([]string, len(userMessage))
		copy(trimmed, userMessage)
		if last == "" {
			trimmed = trimmed[:len(trimmed)-1]
		} else {
			trimmed[len(trimmed)-1] = last
		}
		userMessage = trimmed

		stdInput, err = inout.ReadStdinInteractive()
		if err != nil {
			return "", fmt.Errorf("error reading stdin: %w", err)
		}
	}

	var userPrompt string
	if stdInput != "" && len(userMessage) > 0 {
		userPrompt = stdInput + "\nThat concludes the stdin contents. Now here's the query from the user:\n" + strings.Join(userMessage, " ")
	} else if stdInput != "" {
		userPrompt = stdInput
	} else {
		userPrompt = strings.Join(userMessage, " ")
	}

	return userPrompt, nil
}

func formatMarkdown(text string) string {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	formatted, err := r.Render(text)
	if err != nil {
		return text
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

func cleanToolCalls(text string) string {
	// Remove <tool_call> block if present
	if thinkStart := strings.Index(text, "<tool_call>"); thinkStart != -1 {
		if thinkEnd := strings.Index(text[thinkStart:], "</tool_call>"); thinkEnd != -1 {
			text = text[:thinkStart] + strings.TrimSpace(text[thinkStart+thinkEnd+len("</tool_call>"):])
		}
	}
	return text
}
