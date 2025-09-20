/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"minepack/core/project"
	"minepack/util"
	"net/http"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/spf13/cobra"
)

// formatFileSize formats bytes into human readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// getMissingFiles returns a list of files that are missing from the given link directory
func getMissingFiles(linkPath string, allContent []project.ContentData) ([]project.ContentData, error) {
	var missingFiles []project.ContentData

	for _, content := range allContent {
		// Check if the file exists in the link's mods directory
		modPath := filepath.Join(linkPath, content.File.Filepath)
		if _, err := os.Stat(modPath); os.IsNotExist(err) {
			missingFiles = append(missingFiles, content)
		}
	}

	return missingFiles, nil
}

// downloadFile downloads a file from URL to the specified path
func downloadFile(url, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// copyDir recursively copies a directory from src to dst
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			// Copy file
			return copyFile(path, dstPath)
		}
	})
}

var updateSummaryStyle = lipgloss.NewStyle().
	Padding(0, 1).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Margin(0, 1)

// linkUpdateCmd represents the link update command
var linkUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update all linked Minecraft instances with current modpack content",
	Long:  `downloads missing mods and syncs overrides to all linked Minecraft instance folders`,
	Run: func(cmd *cobra.Command, args []string) {
		// get current working directory and parse project
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("error getting current working directory: %s"), err)
			return
		}

		packData, err := project.ParseProject(cwd)
		if err != nil {
			fmt.Printf(util.FormatError("error parsing project: %s"), err)
			return
		}

		// Load linked folders
		linked, err := loadLinkedFolders(cwd)
		if err != nil {
			fmt.Printf(util.FormatError("error loading linked folders: %s"), err)
			return
		}

		if len(linked.Links) == 0 {
			fmt.Println("No linked folders found. Use 'minepack link add [folder]' to add links first.")
			return
		}

		// Get all content
		allContent, err := packData.GetAllContent()
		if err != nil {
			fmt.Printf(util.FormatError("error getting all content: %s"), err)
			return
		}

		if len(allContent) == 0 {
			fmt.Println("No mods found in modpack.")
			return
		}

		// Calculate what files are missing from each link
		var allMissingFiles []project.ContentData
		missingFilesByLink := make(map[string][]project.ContentData)

		for _, linkPath := range linked.Links {
			missingFiles, err := getMissingFiles(linkPath, allContent)
			if err != nil {
				fmt.Printf(util.FormatWarning("error checking files in %s: %s\n"), linkPath, err)
				continue
			}
			missingFilesByLink[linkPath] = missingFiles

			// Add to total missing files (avoiding duplicates)
			for _, missing := range missingFiles {
				found := false
				for _, existing := range allMissingFiles {
					if existing.Slug == missing.Slug {
						found = true
						break
					}
				}
				if !found {
					allMissingFiles = append(allMissingFiles, missing)
				}
			}
		}

		// Calculate download size and total disk usage
		var downloadSize int64
		for _, content := range allMissingFiles {
			downloadSize += content.File.Filesize
		}

		totalDiskUsage := downloadSize * int64(len(linked.Links))

		// Show summary and ask for confirmation
		var finalString string

		finalString += "update summary:\n"
		finalString += fmt.Sprintf("- linked instances: %d\n", len(linked.Links))
		finalString += fmt.Sprintf("- total mods in modpack: %d\n", len(allContent))
		finalString += fmt.Sprintf("- unique missing files: %d\n", len(allMissingFiles))
		finalString += fmt.Sprintf("- download size: %s\n", formatFileSize(downloadSize))
		finalString += fmt.Sprintf("- total disk usage after sync: %s", formatFileSize(totalDiskUsage))
		fmt.Print(updateSummaryStyle.Render(finalString))
		fmt.Println()

		if len(allMissingFiles) == 0 {
			fmt.Println("all linked instances are up to date!")
		} else {
			// Ask for confirmation
			var proceed bool
			huh.NewConfirm().
				Title("proceed with download and sync?").
				Description("this will download missing files and sync overrides to all linked instances").
				Affirmative("yup!").
				Negative("hell nah").
				Value(&proceed).
				Run()

			if !proceed {
				fmt.Println(util.FormatError("operation cancelled"))
				return
			}

			// Create cache directory
			cacheDir := filepath.Join(cwd, ".mpcache")
			if err := os.MkdirAll(cacheDir, 0755); err != nil {
				fmt.Printf(util.FormatError("error creating cache directory: %s"), err)
				return
			}

			// Download missing files to cache
			if len(allMissingFiles) > 0 {
				fmt.Printf("downloading %d missing file(s)...\n", len(allMissingFiles))

				for i, content := range allMissingFiles {
					var downloadErr error
					err := spinner.New().
						Title(fmt.Sprintf("downloading %s (%d/%d)", content.Name, i+1, len(allMissingFiles))).
						Type(spinner.Dots).
						Action(func() {
							cachePath := filepath.Join(cacheDir, content.File.Filepath)

							// Create directory if needed
							if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
								downloadErr = fmt.Errorf("failed to create cache directory: %w", err)
								return
							}

							// Skip if already exists in cache
							if _, err := os.Stat(cachePath); err == nil {
								return
							}

							downloadErr = downloadFile(content.DownloadUrl, cachePath)
						}).
						Run()

					if err != nil {
						fmt.Printf(util.FormatError("spinner error for %s: %s\n"), content.Name, err)
						continue
					}
					if downloadErr != nil {
						fmt.Printf(util.FormatError("failed to download %s: %s\n"), content.Name, downloadErr)
						continue
					}

					fmt.Printf(util.FormatSuccess("downloaded %s\n"), content.Name)
				}
			}

			// Copy files from cache to each linked instance
			fmt.Println("\nsyncing files to linked instances...")

			for _, linkPath := range linked.Links {
				missingFiles := missingFilesByLink[linkPath]
				if len(missingFiles) == 0 {
					fmt.Printf(util.FormatSuccess("%s (already up to date)\n"), linkPath)
					continue
				}

				// Copy missing files
				successCount := 0
				for _, content := range missingFiles {
					cachePath := filepath.Join(cacheDir, content.File.Filepath)
					destPath := filepath.Join(linkPath, content.File.Filepath)

					// Ensure subdirectory exists
					splitPath := filepath.Dir(destPath)
					if err := os.MkdirAll(splitPath, 0755); err != nil {
						fmt.Printf(util.FormatError("failed to create directory %s: %s\n"), splitPath, err)
						continue
					}

					if err := copyFile(cachePath, destPath); err != nil {
						fmt.Printf(util.FormatError("failed to copy %s to %s: %s\n"), content.Name, linkPath, err)
					} else {
						successCount++
					}
				}

				fmt.Printf(util.FormatSuccess("%s (%d/%d files synced)\n"), linkPath, successCount, len(missingFiles))
			}
		}

		// Sync overrides folder
		overridesPath := filepath.Join(cwd, "overrides")
		if _, err := os.Stat(overridesPath); err == nil {
			fmt.Println("\nsyncing overrides...")

			for _, linkPath := range linked.Links {
				if err := copyDir(overridesPath, linkPath); err != nil {
					fmt.Printf(util.FormatError("failed to sync overrides to %s: %s\n"), linkPath, err)
				} else {
					fmt.Printf(util.FormatSuccess("%s (overrides synced)\n"), linkPath)
				}
			}
		} else {
			fmt.Println("\nno overrides folder found, skipping override sync")
		}

		fmt.Println()
		fmt.Print(util.FormatSuccess("sync completed successfully!"))
	},
}

func init() {
	linkCmd.AddCommand(linkUpdateCmd)
}
