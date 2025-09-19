/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"minepack/core/project"
	"minepack/util"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:     "query",
	Short:   "checks if a mod is in your project",
	Long:    `searches your current project to see if a mod is already included`,
	Aliases: []string{"find", "contains", "has"},
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

		// validate arguments
		if len(args) < 1 {
			fmt.Println("please provide a search query.")
			return
		}

		// join all args as query
		query := ""
		for i, arg := range args {
			if i > 0 {
				query += " "
			}
			query += arg
		}

		// search for mod in project
		found := false
		mod := &project.ContentData{}

		// get all content first so we can use it in multiple search attempts
		allContent, err := packData.GetAllContent()
		if err != nil {
			fmt.Printf(util.FormatError("error getting all content: %s"), err)
			return
		}

		// first, try the basic slug/id match that packData.HasContent does
		if packData.HasMod(query) {
			found = true
			mod, _ = packData.GetContent(query)
		}

		// if not found, try searching via the mod's name (exact match)
		if !found {
			for _, c := range allContent {
				if c.Name != "" && c.Name == query {
					found = true
					mod = &c
					break
				}
			}
		}

		// if STILL not found, try fuzzy search and ask the user to pick one from the top 5 results
		if !found {
			var options []string
			contentMap := make(map[string]project.ContentData)
			for _, c := range allContent {
				if c.Name != "" {
					options = append(options, c.Name)
					contentMap[c.Name] = c
				}
			}

			if len(options) == 0 {
				fmt.Print(util.FormatError("no mods found in project to search through\n"))
				return
			}

			matches := fuzzy.Find(query, options)
			if len(matches) > 0 {
				// limit to top 5 matches
				if len(matches) > 5 {
					matches = matches[:5]
				}

				// prompt user to select one of the matches
				var stringMatches []string
				for _, m := range matches {
					stringMatches = append(stringMatches, m.Str)
				}

				var optionsList []huh.Option[string]
				for _, s := range stringMatches {
					optionsList = append(optionsList, huh.NewOption(s, s))
				}
				// Add cancel option
				optionsList = append(optionsList, huh.NewOption("Cancel", "cancel"))

				var selectedQuery string
				prompt := huh.NewForm(
					huh.NewGroup(
						huh.NewSelect[string]().
							Title("found fuzzy results, pick the most relevant one:").
							Options(optionsList...).
							Value(&selectedQuery),
					),
				)

				err = prompt.Run()
				if err != nil {
					fmt.Printf("prompt failed %v\n", err)
					return
				}

				// Check if user cancelled
				if selectedQuery == "cancel" {
					fmt.Println("Query cancelled")
					return
				}

				if selectedMod, exists := contentMap[selectedQuery]; exists {
					found = true
					mod = &selectedMod
				}
			}
		}

		// if we found a mod, print its details
		if found {
			formatted := util.FormatContentData(*mod)
			fmt.Println(formatted)
			return
		}

		// if we didn't find a mod, print a message
		fmt.Printf(util.FormatError("no mod found for query: %s\n"), query)
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// queryCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// queryCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
