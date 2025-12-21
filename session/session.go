package session

import (
	"context"
	"fmt"
	"strings"

	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/config"
	"github.com/scottyeager/pal/inout"
)

type Session struct {
	Config       *config.Config
	Temperature float64
	Markdown    bool
	Chat        bool
	UserMessage []string
	StdinInput  string
}

func NewSession() *Session {
	return &Session{
		Temperature: 0,
		Markdown:    false,
		Chat:        false,
		UserMessage: []string{},
		StdinInput:  "",
	}
}

func (s *Session) LoadConfig() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}
	s.Config = cfg
	return nil
}

func (s *Session) CheckConfig() error {
	return config.CheckConfiguration(s.Config)
}

func (s *Session) GetSelectedModel(key string) string {
	return config.GetSelectedModel(s.Config, key)
}

func (s *Session) NewAIClient(modelName string) (*ai.Client, error) {
	return ai.NewClient(s.Config, modelName)
}

func (s *Session) ReadStdin() error {
	var err error
	s.StdinInput, err = inout.ReadStdin()
	return err
}

func (s *Session) GetUserInput() string {
	if s.StdinInput != "" && len(s.UserMessage) > 0 {
		return s.StdinInput + "\nThat concludes the stdin contents. Now here's the query from the user:\n" + strings.Join(s.UserMessage, " ")
	} else if s.StdinInput != "" {
		return s.StdinInput
	} else {
		return strings.Join(s.UserMessage, " ")
	}
}

func (s *Session) GetCompletion(ctx context.Context, systemPrompt, prompt string, modelName string) (string, error) {
	client, err := s.NewAIClient(modelName)
	if err != nil {
		return "", fmt.Errorf("error creating AI client: %w", err)
	}

	t := s.Temperature
	if t == 0 {
		t = 1.0
	}

	formatMarkdown := s.Markdown
	if !s.Config.FormatMarkdown {
		formatMarkdown = !formatMarkdown
	}

	return client.GetCompletion(ctx, systemPrompt, prompt, t, formatMarkdown, modelName)
}

func (s *Session) GetCompletionWithHistory(ctx context.Context, commandName, systemPrompt, prompt string, modelName string) (string, *inout.Thread, error) {
	client, err := s.NewAIClient(modelName)
	if err != nil {
		return "", nil, fmt.Errorf("error creating AI client: %w", err)
	}

	t := s.Temperature
	if t == 0 {
		t = 1.0
	}

	formatMarkdown := s.Markdown
	if !s.Config.FormatMarkdown {
		formatMarkdown = !formatMarkdown
	}

	return client.GetCompletionWithHistory(ctx, commandName, systemPrompt, prompt, t, formatMarkdown, modelName)
}

func (s *Session) ContinueCompletion(ctx context.Context, thread *inout.Thread, prompt string, modelName string) (string, error) {
	client, err := s.NewAIClient(modelName)
	if err != nil {
		return "", fmt.Errorf("error creating AI client: %w", err)
	}

	t := s.Temperature
	if t == 0 {
		t = 1.0
	}

	formatMarkdown := s.Markdown
	if !s.Config.FormatMarkdown {
		formatMarkdown = !formatMarkdown
	}

	return client.ContinueCompletion(ctx, thread, prompt, t, formatMarkdown, modelName)
}
