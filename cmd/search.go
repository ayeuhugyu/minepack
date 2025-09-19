/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"minepack/core/api"
	"minepack/core/project"
	"minepack/util"
	"os"

	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

var validLoaders = []string{"fabric", "forge", "quilt", "neoforge", "liteloader"}

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "search for a mod",
	Long:  `search for a minecraft mod. by default, will use project data from your current directory, otherwise will use default values.`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("error getting current working directory: %s"), err)
			return
		}
		packData, err := project.ParseProject(cwd)
		if err != nil {
			packData = &project.Project{
				DefaultSource: "modrinth",
			}
		}

		// Get flags
		modloader, _ := cmd.Flags().GetString("modloader")
		version, _ := cmd.Flags().GetString("version")
		useModrinth, _ := cmd.Flags().GetBool("modrinth")
		useCurseforge, _ := cmd.Flags().GetBool("curseforge")
		
		// Validate source flags
		if useModrinth && useCurseforge {
			fmt.Printf(util.FormatError("cannot specify both --modrinth and --curseforge flags\n"))
			return
		}
		
		// Override packData default source if flags are provided
		if useModrinth {
			packData.DefaultSource = "modrinth"
		} else if useCurseforge {
			packData.DefaultSource = "curseforge"
		}

		// override packData if flags are provided
		if modloader != "" {
			// check if modloader is valid
			isValid := false
			for _, loader := range validLoaders {
				if modloader == loader {
					isValid = true
					break
				}
			}
			if !isValid {
				fmt.Printf(util.FormatError("invalid modloader: %s. valid options are: %v\n"), modloader, validLoaders)
				return
			}
			packData.Versions.Loader.Name = modloader
		}
		if version != "" {
			packData.Versions.Game = version
		}

		query := ""
		if len(args) < 1 {
			fmt.Println("please provide a search query.")
			return
		}
		// query is all args joined with space
		for i, arg := range args {
			if i > 0 {
				query += " "
			}
			query += arg
		}

		// search for the mod
		var result *project.ContentData
		var searchErr error

		err = spinner.New().
			Title("searching for mods...").
			Type(spinner.Dots).
			Action(func() {
				result, searchErr = api.SearchAll(query, *packData)
			}).
			Run()

		if err != nil {
			fmt.Printf(util.FormatError("spinner error: %s"), err)
			return
		}
		if searchErr != nil {
			fmt.Printf(util.FormatError("search failed: %s"), searchErr)
			return
		}
		if result == nil {
			fmt.Println(util.FormatError("no results found."))
			return
		}
		formatted := util.FormatContentData(*result)

		fmt.Printf("%s\n", formatted)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Add flags for modloader and version
	searchCmd.Flags().StringP("modloader", "m", "", "specify the modloader (e.g. forge, fabric)")
	searchCmd.Flags().StringP("version", "v", "", "specify the Minecraft version (e.g. 1.20.1)")
	searchCmd.Flags().Bool("modrinth", false, "search mods from Modrinth only")
	searchCmd.Flags().Bool("curseforge", false, "search mods from CurseForge only")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// searchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
