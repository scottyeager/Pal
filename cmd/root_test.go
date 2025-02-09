package cmd

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestPreparse(t *testing.T) {
	// Mock the persistent flags for the root command
	rootCmd.PersistentFlags().Float64P("temperature", "t", 0, "Temperature flag")

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
			name:     "short temperature flag with space",
			args:     []string{"pal", "-t", "0.5", "some", "input"},
			expected: 3,
		},
		{
			name:     "short temperature flag without space",
			args:     []string{"pal", "-t0.5", "some", "input"},
			expected: 2,
		},
		{
			name:     "long temperature flag with space",
			args:     []string{"pal", "--temperature", "0.5", "some", "input"},
			expected: 3,
		},
		{
			name:     "long temperature flag with equals",
			args:     []string{"pal", "--temperature=0.5", "some", "input"},
			expected: 2,
		},
		{
			name:     "help flag",
			args:     []string{"pal", "--help", "some", "input"},
			expected: 2,
		},
		{
			name:     "short help flag",
			args:     []string{"pal", "-h", "some", "input"},
			expected: 2,
		},
		{
			name:     "version flag",
			args:     []string{"pal", "--version", "some", "input"},
			expected: 2,
		},
		{
			name:     "short version flag",
			args:     []string{"pal", "-v", "some", "input"},
			expected: 2,
		},
		{
			name:     "command starting with /",
			args:     []string{"pal", "/command", "arg"},
			expected: 2,
		},
		{
			name:     "multiple flags",
			args:     []string{"pal", "-t", "0.7", "--temperature=0.8", "-h", "/foo"},
			expected: 6, // Corrected expected value
		},
		{
			name:     "only flags",
			args:     []string{"pal", "-t0.7"},
			expected: 2,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Need to re-initialize flags for each test since they are global
			rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
				rootCmd.PersistentFlags().Set(f.Name, f.DefValue)
			})

			actual := preparse(tt.args)
			if actual != tt.expected {
				t.Errorf("preparse(%v) = %d; want %d", tt.args, actual, tt.expected)
			}
		})
	}
}
