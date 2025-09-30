package cmd

import (
	"fmt"
	"os"
	"time"

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
			Padding(1, 2)

		// Build content
		var content string
		content += labelStyle.Render("Current Version: ") + valueStyle.Render(history.Current) + "\n"
		content += labelStyle.Render("Format: ") + valueStyle.Render(string(history.Format)) + "\n"

		if len(history.Entries) > 0 {
			content += "\n" + labelStyle.Render("Version History:") + "\n\n"

			// Show last 10 entries
			start := 0
			if len(history.Entries) > 10 {
				start = len(history.Entries) - 10
			}

			for i := len(history.Entries) - 1; i >= start; i-- {
				entry := history.Entries[i]
				timestamp := entry.Timestamp.Format(time.RFC3339)
				content += fmt.Sprintf("  %s - %s\n", valueStyle.Render(entry.Version), entry.Message)
				content += fmt.Sprintf("    %s (%s)\n", entry.CommitSHA[:8], timestamp)
			}

			if start > 0 {
				content += fmt.Sprintf("\n  ... and %d more entries\n", start)
			}
		} else {
			content += "\n" + labelStyle.Render("No version history yet\n")
		}

		fmt.Println(boxStyle.Render(content))
	},
}

func init() {
	versionCmd.AddCommand(versionShowCmd)
}
