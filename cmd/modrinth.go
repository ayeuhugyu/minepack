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
	Long:  `tests modrinth api functions like searching, fetching mod info, and downloading files.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Modrinth API demo:")

		// search for mods using the package-level client
		var templateProject = project.Project{
			Versions: project.ProjectVersions{
				Game: "1.20.1",
				Loader: project.ModloaderVersion{
					Name: "fabric",
				},
			},
		}
		results, err := modrinth.SearchProjects("journeymap", templateProject, true)
		if err != nil {
			fmt.Println("search error:", err)
		} else {
			fmt.Printf("search results for 'map': %d found\n", len(results))
			for _, r := range results {

				// handle pointer dereferences safely
				title := "Unknown"
				if r.Title != nil {
					title = *r.Title
				}

				id := "Unknown"
				if r.ProjectID != nil {
					id = *r.ProjectID
				}

				fmt.Printf("- %s (%s)\n", title, id)
			}
		}
		firstId := results[0].ProjectID
		// fetch detailed project info
		projectInfo, err := modrinth.GetProject(*firstId)
		if err != nil {
			fmt.Println("GetProject error:", err)
		} else {
			fmt.Printf("project info for ID %s: Name=%s, Description=%s\n", *firstId, *projectInfo.Title, *projectInfo.Description)
		}

		// fetch versions for the first result
		versions, err := modrinth.GetProjectVersions(*firstId, templateProject)
		if err != nil {
			fmt.Println("GetProjectVersions error:", err)
		} else {
			fmt.Printf("versions for project ID %s:\n", *firstId)
			for _, v := range versions {
				fmt.Printf("- %s (ID: %s)\n", *v.Name, *v.ID)
			}
		}

		// convert the first search result to ContentData
		contentData := modrinth.ConvertProjectToContentData(projectInfo, versions[0])
		fmt.Printf("converted ContentData: %+v\n", contentData)

		// download the first version's file to testProject/
		err = modrinth.DownloadContent(contentData, "testProject/"+contentData.File.Filename)
		if err != nil {
			fmt.Println("download error:", err)
		} else {
			fmt.Println("download successful!")
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
