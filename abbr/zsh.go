package abbr

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed zsh-abbr/zsh-abbr.zsh
var ZshAbbrEmbed string

//go:embed zsh-abbr/zsh-job-queue/zsh-job-queue.zsh
var ZshJobQueueEmbed string

func InstallZshAbbr() (string, error) {
	// Check for existing temp dir
	pattern := os.TempDir() + "/pal-zsh-abbr-*"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("error checking for existing temp dir: %w", err)
	}

	var tmpDir string
	if len(matches) > 0 {
		tmpDir = matches[0]
	} else {
		tmpDir, err = os.MkdirTemp("", "pal-zsh-abbr-")
		if err != nil {
			return "", fmt.Errorf("error creating temp dir: %w", err)
		}

		err = os.MkdirAll(tmpDir+"/zsh-job-queue", 0755)
		if err != nil {
			return "", fmt.Errorf("error creating job queue dir: %w", err)
		}

		err = os.WriteFile(tmpDir+"/zsh-job-queue/zsh-job-queue.zsh", []byte(ZshJobQueueEmbed), 0755)
		if err != nil {
			return "", fmt.Errorf("error writing job queue file: %w", err)
		}

		err = os.WriteFile(tmpDir+"/zsh-abbr.zsh", []byte(ZshAbbrEmbed), 0755)
		if err != nil {
			return "", fmt.Errorf("error writing abbr file: %w", err)
		}
	}

	return tmpDir + "/zsh-abbr.zsh", nil
}

func generatePermutations(newLines *[]string, prefix string, commands []string, max int, digits int, current string) {
	if digits == 0 {
		// Build the command from the digits
		var cmdLines []string
		for _, d := range current {
			idx := int(d - '1')
			if idx >= 0 && idx < len(commands) {
				cmdLines = append(cmdLines, commands[idx])
			}
		}
		if len(cmdLines) > 0 {
			// Escape quotes in the command lines
			escapedCmds := make([]string, len(cmdLines))
			for i, cmd := range cmdLines {
				escapedCmds[i] = strings.ReplaceAll(cmd, `"`, `\"`)
			}
			// Join with ; instead of newlines for better shell compatibility
			abbr := fmt.Sprintf(`abbr %s%s="%s"`, prefix, current, strings.Join(escapedCmds, "; "))
			*newLines = append(*newLines, abbr)
		}
		return
	}

	for i := 1; i <= max; i++ {
		generatePermutations(newLines, prefix, commands, max, digits-1, current + fmt.Sprintf("%d", i))
	}
}

func UpdateZshAbbreviations(removePrefix string, addPrefix string, commands string) error {
	abbrFile := os.Getenv("ABBR_USER_ABBREVIATIONS_FILE")
	if abbrFile == "" {
		home, _ := os.UserHomeDir()
		abbrFile = filepath.Join(home, ".config/zsh-abbr/user-abbreviations")
	}

	// Read existing abbreviations
	content, err := os.ReadFile(abbrFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error reading abbreviations file: %v", err)
	}

	// Filter out existing abbreviations with removePrefix
	var newLines []string
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, fmt.Sprintf("abbr %s", removePrefix)) && 
		   strings.Contains(line, "=") {
			// Skip lines with our prefix
			continue
		}
		if line != "" {
			newLines = append(newLines, line)
		}
	}

	// Add new abbreviations with addPrefix
	commandLines := strings.Split(commands, "\n")
	// Generate all permutations of numbers for multi-line commands
	for i := 1; i <= len(commandLines); i++ {
		// Generate all number combinations from 1 to i digits
		for digits := 1; digits <= 3; digits++ {
			generatePermutations(&newLines, addPrefix, commandLines, i, digits, "")
		}
	}

	// Write back to file
	err = os.MkdirAll(filepath.Dir(abbrFile), 0755)
	if err != nil {
		return fmt.Errorf("error creating directories: %v", err)
	}

	err = os.WriteFile(abbrFile, []byte(strings.Join(newLines, "\n")+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("error writing abbreviations file: %v", err)
	}

	// Reload the abbreviations in the current shell
	cmd := exec.Command("zsh", "-c", "source ~/.zshrc && abbr load")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error reloading abbreviations: %v", err)
	}

	return nil
}
