/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"minepack/core/api/curseforge"
	"minepack/core/project"

	"github.com/spf13/cobra"
)

// curseforgeCmd represents the curseforge command
var curseforgeCmd = &cobra.Command{
	Use:   "curseforge",
	Short: "test curseforge api functions",
	Long:  `tests curseforge api functions like searching, fetching mod info, and downloading files.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("CurseForge API demo:")

		// search for mods using curseforge API
		var templateProject = project.Project{
			Versions: project.ProjectVersions{
				Game: "1.20.1",
				Loader: project.ModloaderVersion{
					Name: "fabric",
				},
			},
		}
		results, err := curseforge.SearchProjects("journeymap", templateProject, true)
		if err != nil {
			fmt.Printf("search error: %s\n", err)
			return
		} else {
			fmt.Printf("search results for 'journeymap': %d found\n", len(results))
			for _, r := range results {
				fmt.Printf("- %s (ID: %d)\n", r.Name, r.ID)
			}
		}

		if len(results) == 0 {
			fmt.Println("no results found, skipping detailed tests")
			return
		}

		firstId := fmt.Sprintf("%d", results[0].ID)

		// fetch detailed project info
		projectInfo, err := curseforge.GetProject(firstId)
		if err != nil {
			fmt.Println("GetProject error:", err)
		} else {
			fmt.Printf("project info for ID %s: Name=%s, Description=%s\n", firstId, projectInfo.Name, projectInfo.Summary)
		}

		// fetch versions for the first result
		versions, err := curseforge.GetProjectVersions(firstId, templateProject)
		if err != nil {
			fmt.Println("GetProjectVersions error:", err)
		} else {
			fmt.Printf("versions for project id %s:\n", firstId)
			for _, v := range versions {
				fmt.Printf("- %s (ID: %d)\n", v.DisplayName, v.ID)
			}
		}

		if len(versions) == 0 {
			fmt.Println("no compatible versions found, skipping conversion and download tests")
			return
		}

		// convert the first search result to ContentData
		contentData := curseforge.ConvertModToContentData(projectInfo, &versions[0])
		fmt.Printf("converted ContentData: %+v\n", contentData)

		// download the first version's file to testproject/
		err = curseforge.DownloadContent(contentData, "testproject/"+contentData.File.Filename)
		if err != nil {
			fmt.Println("download error:", err)
		} else {
			fmt.Println("download successful!")
		}
	},
}

func init() {
	testCmd.AddCommand(curseforgeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// curseforgeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// curseforgeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
