package cmd

import (
	"fmt"
	"os"
	"strings"

	"minepack/core/project"
	"minepack/util"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/spf13/cobra"
)

var versionShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current version and history",
	Long:  `Display the current version, format, and version history`,
	Run: func(cmd *cobra.Command, args []string) {
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("could not determine project root: %v\n"), err)
			return
		}

		history, err := project.ParseVersionHistory(projectRoot)
		if err != nil {
			fmt.Printf(util.FormatError("failed to read version history: %v\n"), err)
			return
		}

		// Styling
		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7FF")).
			Bold(true)

		valueStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 1)

		// Build content
		var content string
		content += labelStyle.Render("Current Version: ") + valueStyle.Render(history.Current) + "\n"
		content += labelStyle.Render("Format: ") + valueStyle.Render(string(history.Format)) + "\n"
		// if there is a message associated with this version, show it
		if len(history.Entries) > 0 && history.Entries[len(history.Entries)-1].Version == history.Current {
			content += labelStyle.Render("Message: ") + valueStyle.Render(history.Entries[len(history.Entries)-1].Message) + "\n"
		}
		// if the --history flag is set, show the version history
		showHistory, _ := cmd.Flags().GetBool("history")
		if showHistory && len(history.Entries) > 0 {
			content += "\n" + labelStyle.Render("Version History:") + "\n\n"

			// Show last 10 entries
			start := 0
			if len(history.Entries) > 10 {
				start = len(history.Entries) - 10
			}

			for i := len(history.Entries) - 1; i >= start; i-- {
				entry := history.Entries[i]
				content += fmt.Sprintf("  %s - %s", valueStyle.Render(entry.Version), entry.Message) + "\n"
			}

			if start > 0 {
				content += fmt.Sprintf("\n  ... and %d more entries", start) + "\n"
			}
		} else if showHistory {
			content += "\n" + labelStyle.Render("No version history yet") + "\n"
		}
		// if the content ends in 1 or more newlines, remove all of them
		content = strings.TrimRight(content, "\n")
		fmt.Println(boxStyle.Render(content))
	},
}

func init() {
	versionCmd.AddCommand(versionShowCmd)
	// -h is reserved for help, so we use -i
	versionShowCmd.Flags().BoolP("history", "i", false, "show version history")
}
