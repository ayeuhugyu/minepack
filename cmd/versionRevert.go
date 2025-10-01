package cmd

import (
	"fmt"
	"os"

	"minepack/core/project"
	"minepack/util"

	"github.com/spf13/cobra"
)

var versionRevertCmd = &cobra.Command{
	Use:     "revert [version]",
	Aliases: []string{"goto", "jump"},
	Short:   "Revert to a specific version",
	Long:    `Revert the project to a specific version from the version history`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("could not determine project root: %v\n"), err)
			return
		}

		targetVersion := args[0]

		err = project.RevertToVersion(projectRoot, targetVersion)
		if err != nil {
			fmt.Printf(util.FormatError("failed to jump to version %s: %v\n"), targetVersion, err)
			return
		}

		fmt.Printf(util.FormatSuccess("jumped to version %s\n"), targetVersion)
	},
}

func init() {
	versionCmd.AddCommand(versionRevertCmd)
}
