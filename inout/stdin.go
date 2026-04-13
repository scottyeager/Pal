package inout

import (
	"fmt"
	"io"
	"os"
)

const stdinPreamble = "Here is some input from stdin. This might be file contents, error messages, or other command output that the user wanted to include with their query:\n"

func ReadStdin() (string, error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped to stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("error reading from stdin: %w", err)
		}
		return stdinPreamble + string(data), nil
	}
	return "", nil
}

// ReadStdinInteractive prompts the user to type input on stdin, terminated by
// EOF (ctrl-d). It is intended for use when stdin is a terminal and the user
// has explicitly asked to provide input interactively (e.g. by ending their
// command with "...").
func ReadStdinInteractive() (string, error) {
	fmt.Fprintln(os.Stderr, "Reading input from stdin. Type your input and press ctrl-d when done:\n")
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("error reading from stdin: %w", err)
	}
	if len(data) == 0 {
		return "", nil
	}
	return stdinPreamble + string(data), nil
}
