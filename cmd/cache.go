/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"minepack/util"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// formatFileSize formats bytes into human readable format
func formatCacheSize(bytes int64) string {
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

// getCacheSize calculates the total size of the cache directory
func getCacheSize(cachePath string) (int64, error) {
	var size int64
	err := filepath.Walk(cachePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// countCacheFiles counts the number of files in the cache directory
func countCacheFiles(cachePath string) (int, error) {
	var count int
	err := filepath.Walk(cachePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}

// cacheCmd represents the cache command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "manage modpack download cache",
	Long:  `view cache information and clear cached mod files`,
}

// cacheClearCmd represents the cache clear command
var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clear the modpack download cache",
	Long:  `removes all cached mod files to free up disk space`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("error getting current working directory: %s"), err)
			return
		}

		cachePath := filepath.Join(cwd, ".mpcache")

		// Check if cache directory exists
		if _, err := os.Stat(cachePath); os.IsNotExist(err) {
			fmt.Println("no cache directory found")
			return
		}

		// Get cache information
		cacheSize, err := getCacheSize(cachePath)
		if err != nil {
			fmt.Printf(util.FormatError("error calculating cache size: %s"), err)
			return
		}

		fileCount, err := countCacheFiles(cachePath)
		if err != nil {
			fmt.Printf(util.FormatError("error counting cache files: %s"), err)
			return
		}

		if fileCount == 0 {
			fmt.Println("cache is already empty")
			return
		}

		// Show cache info and ask for confirmation
		fmt.Printf("cache information:\n")
		fmt.Printf("- location: %s\n", cachePath)
		fmt.Printf("- files: %d\n", fileCount)
		fmt.Printf("- size: %s\n\n", formatCacheSize(cacheSize))

		var proceed bool
		huh.NewConfirm().
			Title("clear cache?").
			Description("this will delete all cached mod files").
			Affirmative("yes, clear it").
			Negative("no, keep it").
			Value(&proceed).
			Run()

		if !proceed {
			fmt.Println(util.FormatError("operation cancelled"))
			return
		}

		// Remove the cache directory
		if err := os.RemoveAll(cachePath); err != nil {
			fmt.Printf(util.FormatError("error clearing cache: %s"), err)
			return
		}

		fmt.Print(util.FormatSuccess("cache cleared successfully!"))
	},
}

// cacheInfoCmd represents the cache info command
var cacheInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "show cache information",
	Long:  `displays cache size, file count, and location`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("error getting current working directory: %s"), err)
			return
		}

		cachePath := filepath.Join(cwd, ".mpcache")

		// Check if cache directory exists
		if _, err := os.Stat(cachePath); os.IsNotExist(err) {
			fmt.Println("no cache directory found")
			return
		}

		// Get cache information
		cacheSize, err := getCacheSize(cachePath)
		if err != nil {
			fmt.Printf(util.FormatError("error calculating cache size: %s"), err)
			return
		}

		fileCount, err := countCacheFiles(cachePath)
		if err != nil {
			fmt.Printf(util.FormatError("error counting cache files: %s"), err)
			return
		}

		fmt.Printf("cache information:\n")
		fmt.Printf("- location: %s\n", cachePath)
		fmt.Printf("- files: %d\n", fileCount)
		fmt.Printf("- size: %s\n", formatCacheSize(cacheSize))

		if fileCount == 0 {
			fmt.Println("\ncache is empty")
		}
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheInfoCmd)
}
