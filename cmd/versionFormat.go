package cmd

import (
	"fmt"
	"os"

	"minepack/core/project"
	"minepack/util"

	"github.com/spf13/cobra"
)

var versionFormatCmd = &cobra.Command{
	Use:   "format [semver|breakver|increment|custom]",
	Short: "Set the version format",
	Long:  `Set the version format for the project (semver, breakver, increment, or custom)`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("could not determine project root: %v\n"), err)
			return
		}

		format := args[0]
		var versionFormat project.VersionFormat

		switch format {
		case "semver":
			versionFormat = project.VersionFormatSemVer
		case "breakver":
			versionFormat = project.VersionFormatBreakVer
		case "increment":
			versionFormat = project.VersionFormatIncrement
		case "custom":
			versionFormat = project.VersionFormatCustom
		default:
			fmt.Printf(util.FormatError("invalid format '%s'. Must be one of: semver, breakver, increment, custom\n"), format)
			return
		}

		err = project.SetVersionFormat(projectRoot, versionFormat)
		if err != nil {
			fmt.Printf(util.FormatError("failed to set version format: %v\n"), err)
			return
		}

		fmt.Printf(util.FormatSuccess("version format set to %s\n"), format)
	},
}

func init() {
	versionCmd.AddCommand(versionFormatCmd)
}
