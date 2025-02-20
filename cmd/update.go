package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/scottyeager/pal/config"
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

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading config: %v", err)
		}

		fmt.Printf("Current version: %s\n", version)
		fmt.Printf("New version available: %s\n\n", latestVersion)
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("error getting executable path: %w", err)
		}
		var binaryName string
		switch runtime.GOOS {
		case "darwin":
			if runtime.GOARCH == "arm64" {
				binaryName = "pal-darwin-arm64"
			} else {
				binaryName = "pal-darwin-amd64"
			}
		case "linux":
			if runtime.GOARCH == "arm64" {
				binaryName = "pal-linux-arm64"
			} else {
				binaryName = "pal-linux-amd64"
			}
		default:
			return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
		}
		updateCmd := fmt.Sprintf(`wget -q https://github.com/scottyeager/Pal/releases/latest/download/%s -O %s && chmod +x %s`, binaryName, execPath, execPath)
		err = inout.StorePrefix0Command(updateCmd)
		if err != nil {
			return fmt.Errorf("error storing update command: %w", err)
		}
		fmt.Printf("If you have abbreviations enabled, you can expand the following command with %s0 to update pal:\n\n", cfg.AbbreviationPrefix)
		fmt.Println(updateCmd)
		return nil
	},
}
