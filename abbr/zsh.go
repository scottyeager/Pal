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
		isPrefix := false
		for i := 0; i <= 9; i++ {
			if strings.HasPrefix(line, fmt.Sprintf("abbr %s%d=", removePrefix, i)) {
				isPrefix = true
				break
			}
		}
		if !isPrefix && line != "" {
			newLines = append(newLines, line)
		}
	}

	// Add new abbreviations with addPrefix
	commandLines := strings.Split(commands, "\n")
	for i, line := range commandLines {
		if line != "" {
			abbr := fmt.Sprintf(`abbr %s%d="%s"`, addPrefix, i+1, line)
			newLines = append(newLines, abbr)
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
