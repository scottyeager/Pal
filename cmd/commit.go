package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/scottyeager/pal/session"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(commitCmd)
	commitCmd.Flags().BoolP("yolo", "y", false, "Automatically commit without editing the message")
}

var commitCmd = &cobra.Command{
	Use:   "/commit",
	Short: "`git add` changed files and generate a commit message",
	Annotations: map[string]string{
		"takes_user_message": "true",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		systemPrompt := `You are a helpful assistant who generates concise and complete git commit messages based on code changes in diff format. Use the Conventional Commit style.

		Follow these guidelines:
		- Write a single line of 72 characters or less
		- Use imperative mood (e.g. "Fix bug" not "Fixed bug")
		- Review the diffs carefully and summarize them at a high level
		- Check for context in the previous commit messages

		Choose one of the following types to begin the message, only add it on the first line:

		feat: New feature (or general code changes that don't fit under another category)
		fix: Bug fix
		docs: Documentation (README files, doc strings, comments, not just any string edits in code)
		style: Formatting
		refactor: Code restructuring
		ci: Continuous integration
		test: Testing-related
		chore: Build/config/tooling
		perf: Performance improvements

		Most commits will be a "feat" or "fix". Use the others only when you are sure it's a good fit.

		Respond only with the commit message. No explanations, additional formatting, or line breaks, please.`

		// Add all modified and deleted files (but not untracked files) with verbose output
		addCmd := exec.Command("git", "add", "--update")
		if err := addCmd.Run(); err != nil {
			return fmt.Errorf("failed to add files: %w", err)
		}

		// Get diff
		diffCmd := exec.Command("git", "diff", "--cached")
		diffOut, err := diffCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get git diff: %w", err)
		}

		// Check if diffs are empty
		if len(strings.TrimSpace(string(diffOut))) == 0 {
			fmt.Println("No changes to commit. Aborting.")
			return nil
		}

		// Get last 10 commits
		logCmd := exec.Command("git", "log", "-n", "10", "--oneline")
		logOut, err := logCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get git log: %w", err)
		}

		// We don't really expect the user to give any input, but we include it
		// if they do
		prompt, err := sesh.ConstructPrompt()
		if err != nil {
			return fmt.Errorf("error constructing prompt: %w", err)
		}

		prompt = prompt + `\nRecent commit history:\n` + string(logOut) + `\n\nDiffs for this commit:\n` + string(diffOut)

		// Higher temperatures seem to maybe create better commit messages
		// Who knew these were more like poetry than code? :P
		params := session.TurnParams{
			Command:      "commit",
			SystemPrompt: systemPrompt,
			Prompt:       prompt,
			Temperature:  1,
			Markdown:     false,
			FormatFunc:   strings.TrimSpace,
		}

		// Generate commit message
		response, err := sesh.ExecuteTurn(params)
		if err != nil {
			return fmt.Errorf("failed to execute turn: %w", err)
		}

		// Check if yolo mode is enabled
		yoloMode, _ := cmd.Flags().GetBool("yolo")
		if yoloMode {
			// Auto-commit without editor
			commitCmd := exec.Command("git", "commit", "-m", response)
			commitCmd.Stdout = os.Stdout
			commitCmd.Stderr = os.Stderr
			if err := commitCmd.Run(); err != nil {
				return fmt.Errorf("commit failed: %w", err)
			}
			return nil
		}

		// Add comment explaining how to abort
		message := response + "\n\n# To abort this commit, delete the commit message and save the file\n" +
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
