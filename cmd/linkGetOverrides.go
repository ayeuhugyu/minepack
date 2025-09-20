/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"minepack/core/project"
	"minepack/util"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// linkGetOverridesCmd represents the get-overrides command
var linkGetOverridesCmd = &cobra.Command{
	Use:   "get-overrides",
	Short: "copy files from a linked instance to overrides",
	Long:  `select a linked instance and copy files from it to your project's overrides folder`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// get current working directory and parse project
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current working directory: %w", err)
		}

		projectRoot := cwd

		// Load linked folders
		linked, err := loadLinkedFolders(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to load linked folders: %w", err)
		}

		if len(linked.Links) == 0 {
			fmt.Print(util.FormatError("no linked instances found. use 'minepack link add' to link an instance first\n"))
			return nil
		}

		// Prompt user to select a linked instance
		var selectedInstance string
		var instanceOpts []huh.Option[string]
		for _, link := range linked.Links {
			instanceName := filepath.Base(link)
			instanceOpts = append(instanceOpts, huh.NewOption(fmt.Sprintf("%s (%s)", instanceName, link), link))
		}

		selectForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("select linked instance").
					Description("choose which linked instance to copy files from").
					Options(instanceOpts...).
					Value(&selectedInstance),
			),
		)

		if err := selectForm.Run(); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		// Read all files and folders in the root of the instance
		entries, err := os.ReadDir(selectedInstance)
		if err != nil {
			return fmt.Errorf("failed to read instance directory: %w", err)
		}

		// Build options for multi-select (pre-select all)
		var fileOpts []huh.Option[string]
		var defaultSelected []string
		for _, entry := range entries {
			entryPath := filepath.Join(selectedInstance, entry.Name())
			var displayName string
			if entry.IsDir() {
				displayName = fmt.Sprintf("📁 %s/", entry.Name())
			} else {
				displayName = fmt.Sprintf("📄 %s", entry.Name())
			}
			fileOpts = append(fileOpts, huh.NewOption(displayName, entryPath))
			defaultSelected = append(defaultSelected, entryPath)
		}

		if len(fileOpts) == 0 {
			fmt.Print(util.FormatError("no files or folders found in the selected instance\n"))
			return nil
		}

		// Prompt user to select which files/folders to copy
		var selectedFiles []string
		multiSelectForm := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("select files and folders to copy").
					Description("choose which files and folders to copy to overrides").
					Options(fileOpts...).
					Value(&selectedFiles).
					Height(15),
			),
		)

		// Pre-select all items
		selectedFiles = defaultSelected

		if err := multiSelectForm.Run(); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		if len(selectedFiles) == 0 {
			fmt.Print(util.FormatError("no files selected\n"))
			return nil
		}

		// Load project content to check for conflicts
		packData, err := project.ParseProject(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to parse project: %w", err)
		}

		// Get all content
		allContent, err := packData.GetAllContent()
		if err != nil {
			return fmt.Errorf("failed to get all content: %w", err)
		}

		// Build a list of all file paths that will be downloaded via content data
		contentFilePaths := make(map[string]bool)
		for _, content := range allContent {
			// Normalize path separators to match the OS
			normalizedPath := filepath.FromSlash(content.File.Filepath)
			contentFilePaths[normalizedPath] = true
		}

		// Create overrides directory if it doesn't exist
		overridesDir := filepath.Join(projectRoot, "overrides")
		if err := os.MkdirAll(overridesDir, 0755); err != nil {
			return fmt.Errorf("failed to create overrides directory: %w", err)
		}

		// Check which files would conflict with content data and copy the rest
		for _, selectedFile := range selectedFiles {
			fileName := filepath.Base(selectedFile)
			relPath, err := filepath.Rel(selectedInstance, selectedFile)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}

			// Check if this file path conflicts with any content data file paths
			conflictsWithContent := false
			normalizedRelPath := filepath.FromSlash(relPath)
			if contentFilePaths[normalizedRelPath] {
				// Find the content name for better error message
				var contentName string
				for _, content := range allContent {
					if filepath.FromSlash(content.File.Filepath) == normalizedRelPath {
						contentName = content.Name
						break
					}
				}
				fmt.Printf(util.FormatWarning("skipping %s - conflicts with content file %s\n"), relPath, contentName)
				conflictsWithContent = true
			}

			if conflictsWithContent {
				continue
			}

			// Copy the file/folder to overrides
			destPath := filepath.Join(overridesDir, relPath)

			// Create destination directory if needed
			destDir := filepath.Dir(destPath)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return fmt.Errorf("failed to create destination directory: %w", err)
			}

			// Check if source is a file or directory
			srcInfo, err := os.Stat(selectedFile)
			if err != nil {
				return fmt.Errorf("failed to stat source: %w", err)
			}

			if srcInfo.IsDir() {
				// For directories, we need to recursively check each file inside
				err := filepath.Walk(selectedFile, func(walkPath string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					// Skip directories in the walk
					if info.IsDir() {
						return nil
					}

					// Get relative path from the instance root
					walkRelPath, err := filepath.Rel(selectedInstance, walkPath)
					if err != nil {
						return err
					}

					// Check if this specific file conflicts with content data
					normalizedWalkRelPath := filepath.FromSlash(walkRelPath)
					if contentFilePaths[normalizedWalkRelPath] {
						// Find the content name for better error message
						var contentName string
						for _, content := range allContent {
							if filepath.FromSlash(content.File.Filepath) == normalizedWalkRelPath {
								contentName = content.Name
								break
							}
						}
						fmt.Printf(util.FormatWarning("skipping %s - conflicts with content file %s\n"), walkRelPath, contentName)
						return nil
					}

					// Copy this individual file
					walkDestPath := filepath.Join(overridesDir, walkRelPath)
					walkDestDir := filepath.Dir(walkDestPath)
					if err := os.MkdirAll(walkDestDir, 0755); err != nil {
						return fmt.Errorf("failed to create destination directory: %w", err)
					}

					if err := copyFile(walkPath, walkDestPath); err != nil {
						fmt.Printf(util.FormatError("failed to copy file %s: %v\n"), walkRelPath, err)
						return nil
					}

					fmt.Printf(util.FormatSuccess("copied file %s to overrides\n"), walkRelPath)
					return nil
				})

				if err != nil {
					fmt.Printf(util.FormatError("failed to copy directory %s: %v\n"), fileName, err)
					continue
				}
			} else {
				// Copy file
				if err := copyFile(selectedFile, destPath); err != nil {
					fmt.Printf(util.FormatError("failed to copy file %s: %v\n"), fileName, err)
					continue
				}
				fmt.Printf(util.FormatSuccess("copied file %s to overrides\n"), fileName)
			}
		}

		return nil
	},
}

func init() {
	linkCmd.AddCommand(linkGetOverridesCmd)
}
