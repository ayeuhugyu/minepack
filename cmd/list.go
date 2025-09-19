/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"minepack/core/project"
	"minepack/util"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/spf13/cobra"
)

var modrinthStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#48cf7aff"))
var curseforgeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ef7d4fff"))
var customStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
var boldStyle = lipgloss.NewStyle().Bold(true)
var grayStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

var listStyle = lipgloss.NewStyle().
	Padding(0, 1).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Margin(0, 1)

// formatListItem formats a mod for list display with proper alignment
func formatListItem(data project.ContentData, sourceWidth, nameWidth, slugWidth int) string {
	// Get source string with appropriate styling
	sourceStr := project.SourceToString(data.Source)
	var styledSource string
	switch data.Source {
	case project.Modrinth:
		styledSource = modrinthStyle.Render(sourceStr)
	case project.Curseforge:
		styledSource = curseforgeStyle.Render(sourceStr)
	case project.Custom:
		styledSource = customStyle.Render(sourceStr)
	default:
		styledSource = customStyle.Render(sourceStr)
	}

	// Pad source to align columns
	styledSourcePadded := "[" + styledSource + "]" + strings.Repeat(" ", sourceWidth-len(sourceStr)-2)

	// Format name and slug with styling
	styledName := boldStyle.Render(data.Name)
	styledSlug := grayStyle.Render("(" + data.Slug + ")")

	// Calculate padding for styled versions
	nameStylePadding := nameWidth - len(data.Name)
	slugStylePadding := slugWidth - len(data.Slug) - 2 // -2 for parentheses

	styledNamePadded := styledName + strings.Repeat(" ", nameStylePadding)
	styledSlugPadded := styledSlug + strings.Repeat(" ", slugStylePadding)

	// Format URL with styling
	var styledUrl string
	switch data.Source {
	case project.Modrinth:
		styledUrl = modrinthStyle.Render(data.PageUrl)
	case project.Curseforge:
		styledUrl = curseforgeStyle.Render(data.PageUrl)
	case project.Custom:
		styledUrl = customStyle.Render("(no page)")
	default:
		styledUrl = customStyle.Render(data.PageUrl)
	}

	return fmt.Sprintf("%s %s %s %s", styledSourcePadded, styledNamePadded, styledSlugPadded, styledUrl)
}

// calculateColumnWidths determines the optimal column widths for alignment
func calculateColumnWidths(mods []project.ContentData) (sourceWidth, nameWidth, slugWidth int) {
	sourceWidth = 0
	nameWidth = 0
	slugWidth = 0

	for _, mod := range mods {
		sourceStr := project.SourceToString(mod.Source)
		sourceLen := len(sourceStr) + 2 // +2 for brackets
		nameLen := len(mod.Name)
		slugLen := len(mod.Slug) + 2 // +2 for parentheses

		if sourceLen > sourceWidth {
			sourceWidth = sourceLen
		}
		if nameLen > nameWidth {
			nameWidth = nameLen
		}
		if slugLen > slugWidth {
			slugWidth = slugLen
		}
	}

	// Add some padding
	sourceWidth += 2
	nameWidth += 2
	slugWidth += 2

	return sourceWidth, nameWidth, slugWidth
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "list all mods in your modpack",
	Long:    `displays a list of all mods currently installed in your modpack`,
	Aliases: []string{"ls", "show"},
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

		// get all content
		allContent, err := packData.GetAllContent()
		if err != nil {
			fmt.Printf(util.FormatError("error getting all content: %s"), err)
			return
		}

		if len(allContent) == 0 {
			fmt.Println(util.FormatError("no mods found in modpack."))
			return
		}

		// calculate column widths for alignment
		sourceWidth, nameWidth, slugWidth := calculateColumnWidths(allContent)

		// add each line
		var fullString string
		for i, mod := range allContent {
			formatted := formatListItem(mod, sourceWidth, nameWidth, slugWidth)
			fullString += formatted
			if i != len(allContent)-1 {
				fullString += "\n"
			}
		}
		fmt.Println(listStyle.Render(fullString))
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
