package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(commitCmd)
}

var commitCmd = &cobra.Command{
	Use:   "/commit",
	Short: "`git add` changed files and generate a commit message",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		if err := config.CheckConfiguration(cfg); err != nil {
			return err
		}

		aiClient, err := ai.NewClient(cfg)
		if err != nil {
			return fmt.Errorf("error creating AI client: %w", err)
		}

		// We might want to use this as context, or as a way to pass a commit message, but for now ignore it
		// stdinInput, err := io.ReadStdin()
		// if err != nil {
		// 	return err
		// }

		statusCmd := exec.Command("git", "status", "--porcelain")
		statusOut, err := statusCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get git status: %w", err)
		}

		// Parse modified and staged files
		lines := strings.Split(string(statusOut), "\n")
		hasChanges := false
		for _, line := range lines {
			if len(line) > 0 && line[:2] != "??" {
				hasChanges = true
			}
		}

		if !hasChanges {
			return fmt.Errorf("no changes to commit")
		}

		var filesToCommit []string
		for _, line := range lines {
			if len(line) > 3 {
				// Check for modified (M), unstaged files (second column)
				if line[1] == 'M' {
					filesToCommit = append(filesToCommit, strings.TrimSpace(line[3:]))
				}
			}
		}

		// Add any unstaged changes
		if len(filesToCommit) > 0 {
			addCmd := exec.Command("git", "add")
			for _, path := range filesToCommit {
				addCmd.Args = append(addCmd.Args, ":/:"+path)
			}
			if err := addCmd.Run(); err != nil {
				return fmt.Errorf("failed to add files: %w. Using git command: %s", err, addCmd)
			}
		}

		// Get diff
		diffCmd := exec.Command("git", "diff", "--cached")
		diffOut, err := diffCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get git diff: %w", err)
		}

		// Get last 10 commits
		logCmd := exec.Command("git", "log", "-n", "10", "--oneline")
		logOut, err := logCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get git log: %w", err)
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

		// Higher temperatures seem to maybe create better commit messages
		// Who knew these were more like poetry than code? :P
		t := 1.5
		if cmd.Flags().Changed("temperature") {
			t = temperature
		}

		message, err := aiClient.GetCompletion(context.Background(), systemPrompt, prompt, false, t, false)
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
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
		commitCmd.Run()

		// If the user aborts the commit, it will cause an error. Maybe we want
		// to detect this and respond differently. Maybe it doesn't matter
		// though since if git errors it will print its own error
		// if err := commitCmd.Run(); err != nil {
		// 	return fmt.Errorf("commit failed: %w", err)
		// }

		return nil
	},
}
