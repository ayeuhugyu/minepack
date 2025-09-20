/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"minepack/util"
	"minepack/util/version"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// Progress messages for download
type progressMsg float64
type progressErrMsg struct{ err error }
type progressDoneMsg struct{ filename string }

// Download progress writer for selfupdate
type downloadProgressWriter struct {
	total      int64
	downloaded int64
	file       *os.File
	reader     io.Reader
	onProgress func(float64)
}

func (pw *downloadProgressWriter) Start() {
	// TeeReader calls pw.Write() each time a new response is received
	_, err := io.Copy(pw.file, io.TeeReader(pw.reader, pw))
	if err != nil {
		program.Send(progressErrMsg{err})
		return
	}
	program.Send(progressDoneMsg{pw.file.Name()})
}

func (pw *downloadProgressWriter) Write(p []byte) (int, error) {
	pw.downloaded += int64(len(p))
	if pw.total > 0 && pw.onProgress != nil {
		ratio := float64(pw.downloaded) / float64(pw.total)
		pw.onProgress(ratio)
	}
	return len(p), nil
}

// Model for progress display
type downloadModel struct {
	pw       *downloadProgressWriter
	progress progress.Model
	err      error
	done     bool
	filename string
}

func (m downloadModel) Init() tea.Cmd {
	return nil
}

func (m downloadModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case progressMsg:
		// Update the progress percentage and force a re-render
		cmd := m.progress.SetPercent(float64(msg))
		return m, tea.Batch(cmd, func() tea.Msg { return nil })
	case progressErrMsg:
		m.err = msg.err
		return m, tea.Quit
	case progressDoneMsg:
		m.done = true
		m.filename = msg.filename
		return m, tea.Quit
	case tea.KeyMsg:
		// Allow quitting with q or ctrl+c
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// Handle progress model updates
	var cmd tea.Cmd
	progressModel, progressCmd := m.progress.Update(msg)
	m.progress = progressModel.(progress.Model)
	cmd = progressCmd
	return m, cmd
}

func (m downloadModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error downloading: %v\n", m.err)
	}
	if m.done {
		return fmt.Sprintf("Download complete: %s\n", filepath.Base(m.filename))
	}

	// Show current progress percentage
	percent := m.progress.Percent()
	return fmt.Sprintf("\nDownloading... %.1f%%\n%s\n", percent*100, m.progress.View())
}

var program *tea.Program

// GitHub API response structure for releases
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// selfupdateCmd represents the selfupdate command
var selfupdateCmd = &cobra.Command{
	Use:   "selfupdate",
	Short: "update minepack to the latest version",
	Long:  `downloads and installs the latest version of minepack from GitHub releases`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := performSelfUpdate(); err != nil {
			fmt.Printf(util.FormatError("failed to update: %s"), err)
			return
		}

		successMsg := "successfully updated minepack! Please restart to use the new version."
		fmt.Println(util.FormatSuccess(successMsg))
	},
}

// performSelfUpdate handles the self-update process
func performSelfUpdate() error {
	// Get the current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	currentExe, err = filepath.Abs(currentExe)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	fmt.Printf("Current version: %s\n", version.Version)
	fmt.Printf("Checking for updates...\n")

	// Get the latest release info from GitHub
	release, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	fmt.Printf("Latest version: %s\n", release.TagName)

	// Check if we're already on the latest version
	if release.TagName == version.Version {
		fmt.Println("Already on the latest version!")
		return nil
	}

	// Determine the correct asset for this platform
	// Try new Go naming convention first, then legacy TypeScript names
	var downloadURL string
	var assetName string

	// Try new Go naming convention
	newAssetName := getAssetNameForPlatform()
	for _, asset := range release.Assets {
		if asset.Name == newAssetName {
			downloadURL = asset.BrowserDownloadURL
			assetName = asset.Name
			break
		}
	}

	// If not found, try legacy TypeScript naming conventions
	if downloadURL == "" {
		legacyNames := getAssetNameForPlatformLegacy()
		for _, legacyName := range legacyNames {
			for _, asset := range release.Assets {
				if asset.Name == legacyName {
					downloadURL = asset.BrowserDownloadURL
					assetName = asset.Name
					break
				}
			}
			if downloadURL != "" {
				break
			}
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary found for platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("Downloading %s...\n", assetName)
	fmt.Printf("Download URL: %s\n", downloadURL)

	// Download the new binary
	tempFile, err := downloadBinary(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer os.Remove(tempFile)

	// Create backup of current binary
	backupPath := strings.TrimSuffix(currentExe, filepath.Ext(currentExe)) + "-old.exe.bak"
	if err := copyFileUpdate(currentExe, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	fmt.Printf("Created backup at %s\n", backupPath)

	// Replace current binary with new one
	if err := replaceExecutable(tempFile, currentExe); err != nil {
		// Try to restore backup on failure
		copyFileUpdate(backupPath, currentExe)
		return fmt.Errorf("failed to replace executable: %w", err)
	}

	fmt.Printf("Updated to version %s\n", release.TagName)
	return nil
}

// getLatestRelease fetches the latest release info from GitHub API
func getLatestRelease() (*GitHubRelease, error) {
	url := "https://api.github.com/repos/ayeuhugyu/minepack/releases/latest"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// getAssetNameForPlatform returns the expected asset name for the current platform
func getAssetNameForPlatform() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Try new Go naming convention first
	var newName string
	if goos == "windows" {
		newName = fmt.Sprintf("minepack-%s-%s.exe", goos, goarch)
	} else {
		newName = fmt.Sprintf("minepack-%s-%s", goos, goarch)
	}

	// Return both old and new naming conventions for checking
	// The caller will try both
	return newName
}

// getAssetNameForPlatformLegacy returns the legacy TypeScript asset naming
func getAssetNameForPlatformLegacy() []string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var names []string

	switch goos {
	case "windows":
		if goarch == "amd64" {
			names = append(names, "minepack-win-x64.exe")
		}
	case "linux":
		if goarch == "amd64" {
			names = append(names, "minepack-linux-x64", "minepack-linux-x64-musl")
		} else if goarch == "arm64" {
			names = append(names, "minepack-linux-arm64", "minepack-linux-arm64-musl")
		}
	case "darwin":
		if goarch == "amd64" {
			names = append(names, "minepack-mac-x64")
		} else if goarch == "arm64" {
			names = append(names, "minepack-mac-arm64")
		}
	}

	return names
}

// downloadBinary downloads a binary from URL to a temporary file with progress bar
func downloadBinary(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	fmt.Printf("Content-Length: %d bytes\n", resp.ContentLength)

	// Don't show progress bar if we can't determine content length
	if resp.ContentLength <= 0 {
		fmt.Println("No content length available, downloading without progress bar...")
		return downloadBinarySimple(url)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "minepack-update-*")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Set up progress writer
	pw := &downloadProgressWriter{
		total:  resp.ContentLength,
		file:   tempFile,
		reader: resp.Body,
		onProgress: func(ratio float64) {
			program.Send(progressMsg(ratio))
		},
	}

	// Set up progress model
	m := downloadModel{
		pw:       pw,
		progress: progress.New(progress.WithDefaultGradient()),
	}

	// Start Bubble Tea program
	program = tea.NewProgram(m)

	// Start the download in a goroutine
	go pw.Start()

	// Run the progress display
	_, err = program.Run()
	if err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("progress display error: %w", err)
	}

	// Make it executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tempFile.Name(), 0755); err != nil {
			os.Remove(tempFile.Name())
			return "", err
		}
	}

	return tempFile.Name(), nil
}

// downloadBinarySimple downloads without progress bar (fallback)
func downloadBinarySimple(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "minepack-update-*")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Copy downloaded content to temp file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	// Make it executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tempFile.Name(), 0755); err != nil {
			os.Remove(tempFile.Name())
			return "", err
		}
	}

	return tempFile.Name(), nil
}

// copyFileUpdate copies a file from src to dst
func copyFileUpdate(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// replaceExecutable replaces the current executable with a new one
func replaceExecutable(newPath, currentPath string) error {
	// On Windows, we can't replace a running executable directly
	// So we need to move the current one and then move the new one in place
	if runtime.GOOS == "windows" {
		tempOld := currentPath + ".old.tmp"

		// Move current exe to temp location
		if err := os.Rename(currentPath, tempOld); err != nil {
			return fmt.Errorf("failed to move current executable: %w", err)
		}

		// Move new exe to current location
		if err := os.Rename(newPath, currentPath); err != nil {
			// Try to restore original
			os.Rename(tempOld, currentPath)
			return fmt.Errorf("failed to move new executable: %w", err)
		}

		// Clean up temp file - this is the fix for the .tmp file not being deleted
		if err := os.Remove(tempOld); err != nil {
			// Don't fail the operation if we can't clean up, just log it
			fmt.Printf("Warning: failed to clean up temporary file %s: %v\n", tempOld, err)
		}
		return nil
	}

	// On Unix systems, we can replace directly
	return os.Rename(newPath, currentPath)
}

func init() {
	rootCmd.AddCommand(selfupdateCmd)
}
