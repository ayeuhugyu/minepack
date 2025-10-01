package cmd

import (
	"fmt"
	"os"
	"strconv"

	"minepack/core/project"
	"minepack/util"

	"github.com/spf13/cobra"
)

var versionMajorCmd = &cobra.Command{
	Use:   "major [add|subtract|set] [value]",
	Short: "Update the major version",
	Long:  `Update the major version number for semver and breakver format projects`,
	Args:  cobra.RangeArgs(1, 2),
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

		if history.Format != project.VersionFormatSemVer && history.Format != project.VersionFormatBreakVer {
			fmt.Println(util.FormatError("major version command is only available for semver and breakver formats"))
			return
		}

		operation := args[0]
		// if args[1] is not provided, default to 1
		if len(args) == 1 {
			args = append(args, "1")
		}

		value, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf(util.FormatError("invalid value '%s': must be a number\n"), args[1])
			return
		}

		var newVersion string
		if history.Format == project.VersionFormatSemVer {
			newVersion, err = project.UpdateSemVerMajor(history.Current, operation, value)
		} else {
			newVersion, err = project.UpdateBreakVerMajor(history.Current, operation, value)
		}
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

var versionMinorCmd = &cobra.Command{
	Use:   "minor [add|subtract|set] [value]",
	Short: "Update the minor version",
	Long:  `Update the minor version number for semver and breakver format projects`,
	Args:  cobra.RangeArgs(1, 2),
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

		if history.Format != project.VersionFormatSemVer && history.Format != project.VersionFormatBreakVer {
			fmt.Println(util.FormatError("minor version command is only available for semver and breakver formats"))
			return
		}

		operation := args[0]
		// if args[1] is not provided, default to 1
		if len(args) == 1 {
			args = append(args, "1")
		}
		value, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf(util.FormatError("invalid value '%s': must be a number\n"), args[1])
			return
		}

		var newVersion string
		if history.Format == project.VersionFormatSemVer {
			newVersion, err = project.UpdateSemVerMinor(history.Current, operation, value)
		} else {
			newVersion, err = project.UpdateBreakVerMinor(history.Current, operation, value)
		}
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

var versionPatchCmd = &cobra.Command{
	Use:   "patch [add|subtract|set] [value]",
	Short: "Update the patch version (semver only)",
	Long:  `Update the patch version number for semver format projects`,
	Args:  cobra.RangeArgs(1, 2),
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

		if history.Format == project.VersionFormatBreakVer {
			fmt.Println(util.FormatError("patch version command is not available for breakver format (breakver only uses major.minor)"))
			return
		}

		if history.Format != project.VersionFormatSemVer {
			fmt.Println(util.FormatError("patch version command is only available for semver format"))
			return
		}

		operation := args[0]
		// if args[1] is not provided, default to 1
		if len(args) == 1 {
			args = append(args, "1")
		}
		value, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf(util.FormatError("invalid value '%s': must be a number\n"), args[1])
			return
		}

		newVersion, err := project.UpdateSemVerPatch(history.Current, operation, value)
		if err != nil {
			fmt.Printf(util.FormatError("failed to update patch version: %v\n"), err)
			return
		}

		message, _ := cmd.Flags().GetString("message")
		if message == "" {
			message = fmt.Sprintf("Update patch version: %s -> %s", history.Current, newVersion)
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
	versionMajorCmd.Flags().StringP("message", "m", "", "Commit message for this version")
	versionMinorCmd.Flags().StringP("message", "m", "", "Commit message for this version")
	versionPatchCmd.Flags().StringP("message", "m", "", "Commit message for this version")

	versionCmd.AddCommand(versionMajorCmd)
	versionCmd.AddCommand(versionMinorCmd)
	versionCmd.AddCommand(versionPatchCmd)
}
