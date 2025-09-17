/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"minepack/core/api/modrinth"
	"minepack/core/project"

	"github.com/spf13/cobra"
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
		results, err := modrinth.SearchProjects("map", project.Project{
			Versions: project.ProjectVersions{
				Game:  "1.20.1",
				Loader: project.ModloaderVersion{
					Name: "fabric",
				},
			},
		}, true)
		if err != nil {
			fmt.Println("Search error:", err)
		} else {
			fmt.Printf("Search results for 'map': %d found\n", len(results))
			for i, r := range results {
				if i >= 3 { break }
				fmt.Printf("- %s (%s)\n", r.Title, r.ID)
			}
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
