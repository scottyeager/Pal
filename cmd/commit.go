package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/config"
)

func Commit(cfg *config.Config, aiClient *ai.Client) (string, error) {
	// Get git status to find modified files
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOut, err := statusCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git status: %w", err)
	}

	// Parse modified and staged files
	var filesToCommit []string
	lines := strings.Split(string(statusOut), "\n")
	for _, line := range lines {
		if len(line) > 3 {
			// Check for modified (M), added (A), or renamed (R) files
			// Either staged (first column) or unstaged (second column)
			if line[0] == 'M' || line[0] == 'A' || line[0] == 'R' ||
				line[1] == 'M' {
				filesToCommit = append(filesToCommit, strings.TrimSpace(line[3:]))
			}
		}
	}

	if len(filesToCommit) == 0 {
		return "", fmt.Errorf("no changes to commit")
	}

	// Add any unstaged changes
	addCmd := exec.Command("git", "add")
	addCmd.Args = append(addCmd.Args, filesToCommit...)
	if err := addCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to add files: %w", err)
	}

	// Get diff
	diffCmd := exec.Command("git", "diff", "--cached")
	diffOut, err := diffCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git diff: %w", err)
	}

	// Get last 10 commits
	logCmd := exec.Command("git", "log", "-n", "10", "--oneline")
	logOut, err := logCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git log: %w", err)
	}

	// Generate commit message
	systemPrompt := `You are a helpful assistant who generates concise and complete git commit messages based on code changes in diff format. Use the Conventional Commit style.

	Follow these guidelines:
	- Write a single line under 72 characters
	- Use imperative mood (e.g. "Fix bug" not "Fixed bug")
	- Review the diffs carfully and summarize them at a high level
	- Check for context in the previous commit messages

	Choose one of the following types to begin the message:

	feat: New feature
	fix: Bug fix
	docs: Documentation
	style: Formatting
	refactor: Code restructuring
	ci: Continuous integration
	test: Testing-related
	chore: Build/config/tooling
	perf: Performance improvements

	Respond only with a single line containing the commit message. No explanations, additional formatting, or line breaks, please.`

	prompt := `Recent commit history: ` + string(logOut) + `Diffs for this commit: ` + string(diffOut)

	// Might be interesting to see if lower temp gives better commit msgs
	temperature := 0.5
	message, err := aiClient.GetCompletion(context.Background(), systemPrompt, prompt, false, temperature)
	if err != nil {
		return "", fmt.Errorf("failed to generate commit message: %w", err)
	}

	// Clean up message
	message = strings.TrimSpace(message)
	if len(message) > 72 {
		message = message[:72]
	}

	// Add comment explaining how to abort
	message = message + "\n\n# To abort this commit, delete the commit message and save the file\n" +
		"# If you exit without saving, the pregenerated message will be used"

	// Start interactive commit with prefilled message
	commitCmd := exec.Command("git", "commit", "-m", message, "--edit")
	commitCmd.Stdin = os.Stdin
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr
	if err := commitCmd.Run(); err != nil {
		return "", fmt.Errorf("commit failed: %w", err)
	}

	return message, nil
}
