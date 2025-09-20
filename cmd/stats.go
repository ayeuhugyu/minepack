/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"minepack/core/project"
	"minepack/util"
	"os"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/spf13/cobra"
)

var statsBoxStyle = lipgloss.NewStyle().
	Padding(0, 1).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Margin(0, 1)

var statsTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#874BFD")).
	Padding(0, 0, 1, 0)

var statsLabelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#888888")).
	Bold(true)

var statsValueStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FFFFFF"))

var statsNameStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#48cf7aff"))

var statsDescStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#CCCCCC")).
	Italic(true)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:     "stats",
	Short:   "display basic information about the pack",
	Long:    `shows pack statistics including name, author, description, mod count, and version information`,
	Aliases: []string{"info", "status"},
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

		// get all content to count mods
		allContent, err := packData.GetAllContent()
		if err != nil {
			fmt.Printf(util.FormatError("error getting all content: %s"), err)
			return
		}

		// count mods by source
		modrinthCount := 0
		curseforgeCount := 0
		customCount := 0

		for _, content := range allContent {
			switch content.Source {
			case project.Modrinth:
				modrinthCount++
			case project.Curseforge:
				curseforgeCount++
			case project.Custom:
				customCount++
			}
		}

		// build the stats content
		var statsContent string

		// pack title
		statsContent += statsTitleStyle.Render("PACK STATISTICS") + "\n"

		// pack name (highlighted)
		statsContent += statsLabelStyle.Render("Name: ") + statsNameStyle.Render(packData.Name) + "\n"

		// author
		statsContent += statsLabelStyle.Render("Author: ") + statsValueStyle.Render(packData.Author) + "\n"

		// description (with italics)
		if packData.Description != "" {
			statsContent += statsLabelStyle.Render("Description: ") + statsDescStyle.Render(packData.Description) + "\n"
		}

		statsContent += "\n" // spacing

		// version information
		statsContent += statsLabelStyle.Render("Game Version: ") + statsValueStyle.Render(packData.Versions.Game) + "\n"
		statsContent += statsLabelStyle.Render("Modloader: ") + statsValueStyle.Render(packData.Versions.Loader.Name+" "+packData.Versions.Loader.Version) + "\n"
		statsContent += statsLabelStyle.Render("Minepack Version: ") + statsValueStyle.Render(packData.Versions.Minepack) + "\n"

		statsContent += "\n" // spacing

		// mod counts
		totalMods := len(allContent)
		statsContent += statsLabelStyle.Render("Total Mods: ") + statsValueStyle.Render(fmt.Sprintf("%d", totalMods)) + "\n"

		if totalMods > 0 {
			// breakdown by source
			if modrinthCount > 0 {
				statsContent += statsLabelStyle.Render("  • Modrinth: ") + modrinthStyle.Render(fmt.Sprintf("%d", modrinthCount)) + "\n"
			}
			if curseforgeCount > 0 {
				statsContent += statsLabelStyle.Render("  • CurseForge: ") + curseforgeStyle.Render(fmt.Sprintf("%d", curseforgeCount)) + "\n"
			}
			if customCount > 0 {
				statsContent += statsLabelStyle.Render("  • Custom: ") + customStyle.Render(fmt.Sprintf("%d", customCount)) + "\n"
			}
		}

		// default source
		statsContent += "\n" // spacing
		statsContent += statsLabelStyle.Render("Default Source: ") + statsValueStyle.Render(packData.DefaultSource)

		// render the content in a box
		fmt.Println(statsBoxStyle.Render(statsContent))
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
