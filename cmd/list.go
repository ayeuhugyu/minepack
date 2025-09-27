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
	"golang.org/x/term"
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

// truncateString truncates a string to a maximum length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return strings.Repeat(".", maxLen)
	}
	return s[:maxLen-3] + "..."
}

// getTerminalWidth returns the terminal width or a default value
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// Default to 80 columns if we can't detect terminal size
		return 80
	}
	return width
}

// formatListItem formats a mod for list display with proper alignment and terminal width awareness
func formatListItem(data project.ContentData, sourceWidth, nameWidth, slugWidth, urlWidth int) string {
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

	// Truncate and format name with styling
	truncatedName := truncateString(data.Name, nameWidth-2) // -2 for padding
	styledName := boldStyle.Render(truncatedName)
	nameStylePadding := nameWidth - len(truncatedName)
	styledNamePadded := styledName + strings.Repeat(" ", nameStylePadding)

	// Truncate and format slug with styling
	maxSlugLen := slugWidth - 4 // -2 for parentheses, -2 for padding
	truncatedSlug := truncateString(data.Slug, maxSlugLen)
	styledSlug := grayStyle.Render("(" + truncatedSlug + ")")
	slugStylePadding := slugWidth - len(truncatedSlug) - 2 // -2 for parentheses
	styledSlugPadded := styledSlug + strings.Repeat(" ", slugStylePadding)

	// Format URL with styling and truncation
	var urlToShow string
	switch data.Source {
	case project.Custom:
		urlToShow = "(no page)"
	default:
		urlToShow = truncateString(data.PageUrl, urlWidth)
	}

	var styledUrl string
	switch data.Source {
	case project.Modrinth:
		styledUrl = modrinthStyle.Render(urlToShow)
	case project.Curseforge:
		styledUrl = curseforgeStyle.Render(urlToShow)
	case project.Custom:
		styledUrl = customStyle.Render(urlToShow)
	default:
		styledUrl = customStyle.Render(urlToShow)
	}

	return fmt.Sprintf("%s %s %s %s", styledSourcePadded, styledNamePadded, styledSlugPadded, styledUrl)
}

// calculateColumnWidths determines the optimal column widths for alignment with terminal width constraints
func calculateColumnWidths(mods []project.ContentData) (sourceWidth, nameWidth, slugWidth, urlWidth int) {
	terminalWidth := getTerminalWidth()
	// Reserve some space for borders and padding from lipgloss styling
	availableWidth := terminalWidth - 8 // -8 for border, padding, and spacing

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

	// Calculate remaining space for URL after fixed columns and spaces between them
	var fixedWidth int
	// fixedWidth := sourceWidth + nameWidth + slugWidth + 3 // +3 for spaces between columns
	// urlWidth = availableWidth - fixedWidth

	// Set reasonable limits to prevent extremely wide columns
	maxNameWidth := availableWidth / 3
	maxSlugWidth := availableWidth / 4
	minUrlWidth := 20

	if nameWidth > maxNameWidth {
		nameWidth = maxNameWidth
	}
	if slugWidth > maxSlugWidth {
		slugWidth = maxSlugWidth
	}

	// Recalculate URL width after applying limits
	fixedWidth = sourceWidth + nameWidth + slugWidth + 3
	urlWidth = availableWidth - fixedWidth
	if urlWidth < minUrlWidth {
		// If URL width is too small, reduce other columns further
		nameWidth = maxNameWidth / 2
		slugWidth = maxSlugWidth / 2
		fixedWidth = sourceWidth + nameWidth + slugWidth + 3
		urlWidth = availableWidth - fixedWidth
		if urlWidth < minUrlWidth {
			urlWidth = minUrlWidth
		}
	}

	return sourceWidth, nameWidth, slugWidth, urlWidth
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
		sourceWidth, nameWidth, slugWidth, urlWidth := calculateColumnWidths(allContent)

		// add each line
		var fullString string
		for i, mod := range allContent {
			formatted := formatListItem(mod, sourceWidth, nameWidth, slugWidth, urlWidth)
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
