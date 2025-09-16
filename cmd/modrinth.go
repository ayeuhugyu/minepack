/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"minepack/core/api/modrinth"
)

// modrinthCmd represents the modrinth command
var modrinthCmd = &cobra.Command{
	Use:   "modrinth",
	Short: "test modrinth api functions",
	Long: `tests modrinth api functions like searching, fetching mod info, and downloading files.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Modrinth API demo:")

		// Use a constant mod project ID (e.g., "journeymap")
		const projectID = "journeymap"

		// Search for mods
		results, err := modrinth.SearchMods("map", []string{"fabric"}, []string{"1.20.1"})
		if err != nil {
			fmt.Println("Search error:", err)
		} else {
			fmt.Printf("Search results for 'map': %d found\n", len(results))
			for i, r := range results {
				if i >= 3 { break }
				fmt.Printf("- %s (%s)\n", r.Title, r.ID)
			}
		}

		// Get mod info
		mod, err := modrinth.GetModInfo(projectID)
		if err != nil {
			fmt.Println("GetModInfo error:", err)
		} else {
			fmt.Printf("Mod info: %s by %s\n", mod.Title, mod.Author)
		}

		// Convert to ContentData
		content := modrinth.ConvertModrinthToContentData(mod)
		fmt.Printf("ContentData: %s (%s)\n", content.Name, content.Id)

		// Download the first file (to ./test_download.jar)
		if content.File.Filename != "" && content.DownloadUrl != "" {
			fmt.Println("Downloading file to ./test_download.jar ...")
			err = modrinth.DownloadModrinthContent(content, "./test_download.jar")
			if err != nil {
				fmt.Println("Download error:", err)
			} else {
				fmt.Println("Download successful!")
			}
		} else {
			fmt.Println("No downloadable file found.")
		}
	},
}

func init() {
	testCmd.AddCommand(modrinthCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// modrinthCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// modrinthCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
