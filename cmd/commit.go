package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/config"
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

		preamble := prompt + `\nRecent commit history:\n` + string(logOut)

		maxChunks := config.DefaultCommitMaxChunks
		if cfg, cfgErr := config.LoadConfig(); cfgErr == nil && cfg.CommitMaxChunks != 0 {
			maxChunks = cfg.CommitMaxChunks
		}

		response, err := generateCommitMessage(sesh, systemPrompt, preamble, string(diffOut), maxChunks)
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

const summarizeSystemPrompt = `You are summarizing part of a larger git diff. Produce 1-3 concise bullet points describing the changes in this fragment. Do not write a commit message, do not add a preamble or conclusion. Bullets only.`

// generateCommitMessage produces a commit message from diff, falling back
// to a split-and-synthesize strategy when the provider reports a
// context-length error. On overflow the diff is halved (then halved again,
// up to maxChunks total chunks); each chunk is summarized and a final
// synthesis call produces the commit message from the summaries.
func generateCommitMessage(sesh *session.Session, systemPrompt, preamble, diff string, maxChunks int) (string, error) {
	fullPrompt := preamble + `\n\nDiffs for this commit:\n` + diff
	firstParams := session.TurnParams{
		Command:      "commit",
		SystemPrompt: systemPrompt,
		Prompt:       fullPrompt,
		Temperature:  1,
		Markdown:     false,
		FormatFunc:   strings.TrimSpace,
	}

	response, err := sesh.ExecuteTurn(firstParams)
	if err == nil {
		return response, nil
	}
	if !ai.IsContextLengthError(err) || maxChunks < 2 {
		return "", err
	}

	n := 2
	var summaries []string
	for {
		if n > maxChunks {
			return "", fmt.Errorf("diff too large to summarize within commit_max_chunks=%d", maxChunks)
		}
		fmt.Fprintf(os.Stderr, "[commit: diff too large, summarizing in %d chunks]\n", n)
		chunks := splitDiff(diff, n)
		summaries = summaries[:0]
		retry := false
		for i, chunk := range chunks {
			params := session.TurnParams{
				Command:      "commit",
				SystemPrompt: summarizeSystemPrompt,
				Prompt:       fmt.Sprintf("Diff fragment %d of %d:\n%s", i+1, len(chunks), chunk),
				Temperature:  1,
				Markdown:     false,
				FormatFunc:   strings.TrimSpace,
			}
			out, serr := sesh.ExecuteTurn(params)
			if serr != nil {
				if ai.IsContextLengthError(serr) {
					n *= 2
					retry = true
					break
				}
				return "", fmt.Errorf("failed to summarize chunk %d: %w", i+1, serr)
			}
			summaries = append(summaries, out)
		}
		if !retry {
			break
		}
	}

	var b strings.Builder
	b.WriteString(preamble)
	b.WriteString("\n\nSummary of changes (diff was too large to include directly):\n")
	for i, s := range summaries {
		fmt.Fprintf(&b, "--- Fragment %d ---\n%s\n", i+1, s)
	}
	finalParams := session.TurnParams{
		Command:      "commit",
		SystemPrompt: systemPrompt,
		Prompt:       b.String(),
		Temperature:  1,
		Markdown:     false,
		FormatFunc:   strings.TrimSpace,
	}
	return sesh.ExecuteTurn(finalParams)
}

// splitDiff divides d into n chunks of roughly equal size. Split points
// snap to a nearby "\ndiff --git " header when one is available, falling
// back to the nearest newline, so hunks aren't sliced mid-line.
func splitDiff(d string, n int) []string {
	if n <= 1 || len(d) == 0 {
		return []string{d}
	}
	target := len(d) / n
	chunks := make([]string, 0, n)
	start := 0
	for i := 1; i < n; i++ {
		ideal := start + target
		if ideal >= len(d) {
			break
		}
		cut := findSplitPoint(d, start, ideal, len(d))
		if cut <= start {
			continue
		}
		chunks = append(chunks, d[start:cut])
		start = cut
	}
	chunks = append(chunks, d[start:])
	return chunks
}

func findSplitPoint(d string, lo, ideal, hi int) int {
	// Prefer a "diff --git " file header near the ideal point.
	window := (hi - lo) / 2
	if window < 1 {
		window = 1
	}
	searchLo := ideal - window
	if searchLo < lo {
		searchLo = lo
	}
	searchHi := ideal + window
	if searchHi > hi {
		searchHi = hi
	}
	const marker = "\ndiff --git "
	if idx := strings.LastIndex(d[searchLo:ideal], marker); idx >= 0 {
		return searchLo + idx + 1
	}
	if idx := strings.Index(d[ideal:searchHi], marker); idx >= 0 {
		return ideal + idx + 1
	}
	// Fall back to the nearest newline forward from ideal.
	if nl := strings.IndexByte(d[ideal:hi], '\n'); nl >= 0 {
		return ideal + nl + 1
	}
	return ideal
}
