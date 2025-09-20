/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"minepack/core/api/curseforge"
	"minepack/core/api/modrinth"
	"minepack/core/project"
	"minepack/util"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export your modpack to various formats",
	Long:  `export your modpack to different formats like modrinth (.mrpack) or simple zip`,
}

// copyDirExport recursively copies a directory
func copyDirExport(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path from src
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Construct destination path
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

// modSideDataToString converts ModSideData to string for modrinth format
func modSideDataToString(side project.ModSideData) string {
	switch side {
	case project.SideRequired:
		return "required"
	case project.SideOptional:
		return "optional"
	case project.SideUnsupported:
		return "unsupported"
	case project.SideUnknown:
		return "optional" // default to optional for unknown
	default:
		return "optional"
	}
}

// downloadContent downloads a single content item based on its source
func downloadContent(content project.ContentData, destPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	switch content.Source {
	case project.Modrinth:
		return modrinth.DownloadContent(content, destPath)
	case project.Curseforge:
		return curseforge.DownloadContent(content, destPath)
	case project.Custom:
		// For custom content, we need to copy from the overrides folder
		// The content should already be in overrides, so we don't download it
		return fmt.Errorf("custom content should be handled separately")
	default:
		return fmt.Errorf("unsupported content source: %v", content.Source)
	}
}

// getContentPath returns the destination path for content based on its type
func getContentPath(content project.ContentData) string {
	var folder string
	switch content.ContentType {
	case project.Mod:
		folder = "mods"
	case project.Resourcepack:
		folder = "resourcepacks"
	case project.Shaderpack:
		folder = "shaderpacks"
	default:
		folder = "mods" // default to mods folder
	}
	return filepath.Join(folder, content.File.Filename)
}

// exportModrinthCmd represents the export modrinth command
var exportModrinthCmd = &cobra.Command{
	Use:   "modrinth",
	Short: "export as Modrinth .mrpack format",
	Long:  `exports your modpack as a Modrinth .mrpack file (zip containing modrinth.index.json and overrides folder)`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get current working directory and parse project
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

		// Get all content
		allContent, err := packData.GetAllContent()
		if err != nil {
			fmt.Printf(util.FormatError("error getting all content: %s"), err)
			return
		}

		// Export as .mrpack
		outputName := fmt.Sprintf("%s.mrpack", packData.Name)
		if err := exportModrinthPack(packData, allContent, outputName); err != nil {
			fmt.Printf(util.FormatError("failed to export modrinth pack: %s"), err)
			return
		}

		successMsg := fmt.Sprintf("successfully exported to %s", outputName)
		fmt.Println(util.FormatSuccess(successMsg))
	},
}

// exportModrinthPack exports the pack as a .mrpack file
func exportModrinthPack(packData *project.Project, allContent []project.ContentData, outputName string) error {
	// Create temporary directory for building the pack
	tempDir, err := os.MkdirTemp("", "minepack-export-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Copy overrides folder if it exists
	overridesPath := filepath.Join(packData.Root, "overrides")
	if _, err := os.Stat(overridesPath); err == nil {
		tempOverridesPath := filepath.Join(tempDir, "overrides")
		if err := copyDirExport(overridesPath, tempOverridesPath); err != nil {
			return fmt.Errorf("failed to copy overrides: %w", err)
		}
	}

	// Create modrinth.index.json
	modrinthIndex := createModrinthIndex(packData, allContent)
	indexData, err := json.MarshalIndent(modrinthIndex, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal modrinth index: %w", err)
	}

	indexPath := filepath.Join(tempDir, "modrinth.index.json")
	if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
		return fmt.Errorf("failed to write modrinth.index.json: %w", err)
	}

	// Download non-modrinth content to overrides
	for _, content := range allContent {
		if content.Source != project.Modrinth {
			// Download to overrides folder
			destPath := filepath.Join(tempDir, "overrides", getContentPath(content))
			if err := downloadContent(content, destPath); err != nil && content.Source != project.Custom {
				return fmt.Errorf("failed to download %s: %w", content.Name, err)
			}
		}
	}

	// Create the .mrpack zip file
	return createZipFile(tempDir, outputName)
}

// createModrinthIndex creates a modrinth.index.json structure
func createModrinthIndex(packData *project.Project, allContent []project.ContentData) map[string]interface{} {
	// Build dependencies
	dependencies := map[string]string{
		"minecraft": packData.Versions.Game,
	}

	// Add modloader dependency
	switch packData.Versions.Loader.Name {
	case "fabric":
		dependencies["fabric-loader"] = packData.Versions.Loader.Version
	case "quilt":
		dependencies["quilt-loader"] = packData.Versions.Loader.Version
	case "forge":
		dependencies["forge"] = packData.Versions.Loader.Version
	case "neoforge":
		dependencies["neoforge"] = packData.Versions.Loader.Version
	}

	// Build files list (only include Modrinth content)
	var files []map[string]interface{}
	for _, content := range allContent {
		if content.Source == project.Modrinth {
			// Filter out empty hashes
			hashes := make(map[string]string)
			if content.File.Hashes.Sha1 != "" {
				hashes["sha1"] = content.File.Hashes.Sha1
			}
			if content.File.Hashes.Sha256 != "" {
				hashes["sha256"] = content.File.Hashes.Sha256
			}
			if content.File.Hashes.Sha512 != "" {
				hashes["sha512"] = content.File.Hashes.Sha512
			}
			if content.File.Hashes.Md5 != "" {
				hashes["md5"] = content.File.Hashes.Md5
			}

			file := map[string]interface{}{
				"path":      getContentPath(content),
				"hashes":    hashes,
				"downloads": []string{content.DownloadUrl},
				"fileSize":  content.File.Filesize,
				"env": map[string]string{
					"client": modSideDataToString(content.Side.Client),
					"server": modSideDataToString(content.Side.Server),
				},
			}
			files = append(files, file)
		}
	}

	return map[string]interface{}{
		"formatVersion": 1,
		"game":          "minecraft",
		"versionId":     packData.Versions.Minepack,
		"name":          packData.Name,
		"summary":       fmt.Sprintf("%s by %s - %s", packData.Name, packData.Author, packData.Description),
		"files":         files,
		"dependencies":  dependencies,
	}
}

// createZipFile creates a zip file from a directory
func createZipFile(sourceDir, outputName string) error {
	file, err := os.Create(outputName)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == sourceDir {
			return nil
		}

		// Get relative path for zip entry
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Convert Windows paths to forward slashes for zip compatibility
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		if info.IsDir() {
			// Create directory entry
			_, err := zipWriter.Create(relPath + "/")
			return err
		}

		// Create file entry
		zipFile, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		_, err = io.Copy(zipFile, srcFile)
		return err
	})
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.AddCommand(exportModrinthCmd)
}
