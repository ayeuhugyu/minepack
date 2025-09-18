/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"minepack/core/api"
	"minepack/core/project"
	"minepack/util"

	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please provide a search query.")
			return
		}
		// query is all args joined with space
		query := ""
		for i, arg := range args {
			if i > 0 {
				query += " "
			}
			query += arg
		}
		fmt.Printf("Searching for: %s\n", query)
		var templateProject = project.Project{
			DefaultSource: "modrinth",
			Versions: project.ProjectVersions{
				Game: "1.20.1",
				Loader: project.ModloaderVersion{
					Name: "fabric",
				},
			},
		}
		result, err := api.SearchAll(query, templateProject)
		if err != nil {
			fmt.Println(err)
			return
		}
		if result == nil {
			fmt.Printf("No results found.")
			return
		}
		formatted := util.FormatContentData(*result)
		fmt.Printf("%s\n", formatted)
	},
}

func init() {
	testCmd.AddCommand(searchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// searchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
