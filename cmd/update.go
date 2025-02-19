package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/scottyeager/pal/inout"
	"github.com/spf13/cobra"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "/update",
	Short: "Update Pal to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get latest release from GitHub
		resp, err := http.Get("https://api.github.com/repos/scottyeager/Pal/releases/latest")
		if err != nil {
			return fmt.Errorf("error checking for updates: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading update information: %w", err)
		}

		var release GitHubRelease
		if err := json.Unmarshal(body, &release); err != nil {
			return fmt.Errorf("error parsing update information: %w", err)
		}

		latestVersion := strings.TrimPrefix(release.TagName, "v")
		currentVersion := strings.TrimPrefix(version, "v")

		if latestVersion == currentVersion {
			fmt.Println("You're already running the latest version:", version)
			return nil
		}

		fmt.Printf("New version available: %s (you have %s)\n", latestVersion, version)
		updateCmd := `wget https://github.com/scottyeager/Pal/releases/latest/download/pal-linux-amd64 -O /usr/local/bin/pal && chmod +x /usr/local/bin/pal`
		err = inout.StorePrefix0Command(updateCmd)
		if err != nil {
			return fmt.Errorf("error storing update command: %w", err)
		}
		fmt.Println("Update command stored as prefix0. Run 'pal /show' to see it.")
		return nil
	},
}
