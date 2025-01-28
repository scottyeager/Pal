package abbr

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
			abbr := fmt.Sprintf(`abbr %s%d='%s'`, addPrefix, i+1, strings.ReplaceAll(line, "'", "'\\''"))
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
