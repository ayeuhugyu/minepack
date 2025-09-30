package cmd

import (
	"fmt"
	"os"
	"strconv"

	"minepack/core/project"
	"minepack/util"

	"github.com/spf13/cobra"
)

var versionBreakverMajorCmd = &cobra.Command{
	Use:   "breakmajor [add|subtract|set] [value]",
	Short: "Update the major version (breakver only)",
	Long:  `Update the major version number for breakver format projects`,
	Args:  cobra.ExactArgs(2),
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

		if history.Format != project.VersionFormatBreakVer {
			fmt.Printf(util.FormatError("breakmajor version command is only available for breakver format\n"))
			return
		}

		operation := args[0]
		value, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf(util.FormatError("invalid value '%s': must be a number\n"), args[1])
			return
		}

		newVersion, err := project.UpdateBreakVerMajor(history.Current, operation, value)
		if err != nil {
			fmt.Printf(util.FormatError("failed to update major version: %v\n"), err)
			return
		}

		message, _ := cmd.Flags().GetString("message")
		if message == "" {
			message = fmt.Sprintf("Update major version: %s -> %s", history.Current, newVersion)
		}

		err = project.SetVersion(projectRoot, newVersion, message)
		if err != nil {
			fmt.Printf(util.FormatError("failed to set version: %v\n"), err)
			return
		}

		fmt.Printf(util.FormatSuccess("version updated: %s -> %s\n"), history.Current, newVersion)
	},
}

var versionBreakverMinorCmd = &cobra.Command{
	Use:   "breakminor [add|subtract|set] [value]",
	Short: "Update the minor version (breakver only)",
	Long:  `Update the minor version number for breakver format projects`,
	Args:  cobra.ExactArgs(2),
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

		if history.Format != project.VersionFormatBreakVer {
			fmt.Printf(util.FormatError("breakminor version command is only available for breakver format\n"))
			return
		}

		operation := args[0]
		value, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf(util.FormatError("invalid value '%s': must be a number\n"), args[1])
			return
		}

		newVersion, err := project.UpdateBreakVerMinor(history.Current, operation, value)
		if err != nil {
			fmt.Printf(util.FormatError("failed to update minor version: %v\n"), err)
			return
		}

		message, _ := cmd.Flags().GetString("message")
		if message == "" {
			message = fmt.Sprintf("Update minor version: %s -> %s", history.Current, newVersion)
		}

		err = project.SetVersion(projectRoot, newVersion, message)
		if err != nil {
			fmt.Printf(util.FormatError("failed to set version: %v\n"), err)
			return
		}

		fmt.Printf(util.FormatSuccess("version updated: %s -> %s\n"), history.Current, newVersion)
	},
}

func init() {
	versionBreakverMajorCmd.Flags().StringP("message", "m", "", "Commit message for this version")
	versionBreakverMinorCmd.Flags().StringP("message", "m", "", "Commit message for this version")

	versionCmd.AddCommand(versionBreakverMajorCmd)
	versionCmd.AddCommand(versionBreakverMinorCmd)
}
