package inout

import (
	"fmt"
	"io"
	"os"
)

func ReadStdin() (string, error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped to stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("error reading from stdin: %w", err)
		}
		return "Here is some input from stdin. This might be file contents, error messages, or other command output that the user wanted to include with their query:\n" + string(data), nil
	}
	return "", nil
}
