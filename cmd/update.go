package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var updateManagerCmd = &cobra.Command{
	Use:   "update",
	Short: "Update nas-manager to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		if err := updateBinary(); err != nil {
			fmt.Printf("Update failed: %v\n", err)
			os.Exit(1)
		}
	},
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func updateBinary() error {
	fmt.Println("Checking for updates...")
	
	// Get latest release info
	resp, err := http.Get("https://api.github.com/repos/SlashGordon/scripts/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to check for updates: %v", err)
	}
	defer resp.Body.Close()
	
	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release info: %v", err)
	}
	
	// Check if update is needed
	if release.TagName == Version {
		fmt.Printf("Already running the latest version: %s\n", Version)
		return nil
	}
	
	fmt.Printf("New version available: %s (current: %s)\n", release.TagName, Version)
	
	// Find appropriate binary for current platform
	binaryName := fmt.Sprintf("nas-manager-%s-%s", runtime.GOOS, runtime.GOARCH)
	var downloadURL string
	
	for _, asset := range release.Assets {
		if asset.Name == binaryName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	
	if downloadURL == "" {
		return fmt.Errorf("no binary found for platform %s-%s", runtime.GOOS, runtime.GOARCH)
	}
	
	fmt.Printf("Downloading %s...\n", binaryName)
	
	// Download new binary
	resp, err = http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %v", err)
	}
	defer resp.Body.Close()
	
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}
	
	// Create temporary file
	tmpFile := execPath + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer out.Close()
	
	// Copy downloaded content
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to write update: %v", err)
	}
	out.Close()
	
	// Make executable
	if err := os.Chmod(tmpFile, 0755); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to make binary executable: %v", err)
	}
	
	// Replace current binary
	if runtime.GOOS == "windows" {
		// On Windows, rename current binary and replace
		oldFile := execPath + ".old"
		if err := os.Rename(execPath, oldFile); err != nil {
			os.Remove(tmpFile)
			return fmt.Errorf("failed to backup current binary: %v", err)
		}
		if err := os.Rename(tmpFile, execPath); err != nil {
			os.Rename(oldFile, execPath) // Restore backup
			return fmt.Errorf("failed to replace binary: %v", err)
		}
		os.Remove(oldFile)
	} else {
		// On Unix-like systems, use mv command
		if err := exec.Command("mv", tmpFile, execPath).Run(); err != nil {
			os.Remove(tmpFile)
			return fmt.Errorf("failed to replace binary: %v", err)
		}
	}
	
	fmt.Printf("Successfully updated to version %s\n", release.TagName)
	fmt.Println("Please restart the application to use the new version.")
	
	return nil
}

func init() {
	// Only add update command if not running in development
	if !strings.Contains(Version, "dev") {
		rootCmd.AddCommand(updateManagerCmd)
	}
}