package cmd

import (
	"testing"
)

func TestPreparse(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected int
	}{
		{
			name:     "no flags",
			args:     []string{"pal", "some", "input"},
			expected: 1,
		},
		{
			name:     "short temperature flag with space and /cmd",
			args:     []string{"pal", "-t", "0.5", "/cmd", "input"},
			expected: 4,
		},
		{
			name:     "short temperature flag without space",
			args:     []string{"pal", "-t0.5", "/cmd", "input"},
			expected: 3,
		},
		{
			name:     "long temperature flag with space",
			args:     []string{"pal", "--temperature", "0.5", "/cmd", "input"},
			expected: 4,
		},
		{
			name:     "long temperature flag with equals",
			args:     []string{"pal", "--temperature=0.5", "/cmd", "input"},
			expected: 3,
		},
		{
			name:     "command starting with /",
			args:     []string{"pal", "/command", "arg"},
			expected: 2,
		},
		{
			name:     "only flags. the flag is treated as user message",
			args:     []string{"pal", "-t0.7"},
			expected: 1,
		},
		{
			name:     "only command",
			args:     []string{"pal", "/foo"},
			expected: 2,
		},
		{
			name:     "flags and command",
			args:     []string{"pal", "-t0.7", "/foo"},
			expected: 3,
		},
		{
			name:     "command and flags in input",
			args:     []string{"pal", "-t0.7", "/ask", "what", "is", "-t"},
			expected: 3,
		},
		{
			name:     "/model command",
			args:     []string{"pal", "/model", "sooperAI/pal"},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := preparse(tt.args)
			if actual != tt.expected {
				t.Errorf("preparse(%v) = %d; want %d", tt.args, actual, tt.expected)
			}
		})
	}
}
