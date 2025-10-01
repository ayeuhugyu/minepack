package cmd

import (
	"fmt"
	"os"

	"minepack/core/project"
	"minepack/util"

	"github.com/spf13/cobra"
)

var versionSetCmd = &cobra.Command{
	Use:   "set [version]",
	Short: "Set the project version",
	Long:  `Set the project version to a specific value`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("could not determine project root: %v\n"), err)
			return
		}

		newVersion := args[0]
		message, _ := cmd.Flags().GetString("message")

		if message == "" {
			message = fmt.Sprintf("Set version to %s", newVersion)
		}

		err = project.SetVersion(projectRoot, newVersion, message)
		if err != nil {
			fmt.Printf(util.FormatError("failed to set version: %v\n"), err)
			return
		}

		fmt.Printf(util.FormatSuccess("version set to %s\n"), newVersion)
	},
}

func init() {
	versionSetCmd.Flags().StringP("message", "m", "", "Commit message for this version")
	versionCmd.AddCommand(versionSetCmd)
}
