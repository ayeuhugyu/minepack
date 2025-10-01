package cmd

import (
	"fmt"
	"os"
	"strconv"

	"minepack/core/project"
	"minepack/util"

	"github.com/spf13/cobra"
)

var versionAddCmd = &cobra.Command{
	Use:   "add [value]",
	Short: "Add to the version number (increment format only)",
	Long:  `Add to the version number for increment format projects`,
	Args:  cobra.MaximumNArgs(1),
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

		if history.Format != project.VersionFormatIncrement {
			fmt.Println(util.FormatError("add command is only available for increment format"))
			return
		}

		// if args[0] is not provided, default to 1
		if len(args) == 0 {
			args = append(args, "1")
		}

		value, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf(util.FormatError("invalid value '%s': must be a number\n"), args[0])
			return
		}

		newVersion, err := project.UpdateIncrementVersion(history.Current, "add", value)
		if err != nil {
			fmt.Printf(util.FormatError("failed to update version: %v\n"), err)
			return
		}

		message, _ := cmd.Flags().GetString("message")
		if message == "" {
			message = fmt.Sprintf("Increment version: %s -> %s", history.Current, newVersion)
		}

		err = project.SetVersion(projectRoot, newVersion, message)
		if err != nil {
			fmt.Printf(util.FormatError("failed to set version: %v\n"), err)
			return
		}

		fmt.Printf(util.FormatSuccess("version updated: %s -> %s\n"), history.Current, newVersion)
	},
}

var versionSubtractCmd = &cobra.Command{
	Use:   "subtract [value]",
	Short: "Subtract from the version number (increment format only)",
	Long:  `Subtract from the version number for increment format projects`,
	Args:  cobra.MaximumNArgs(1),
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

		if history.Format != project.VersionFormatIncrement {
			fmt.Println(util.FormatError("subtract command is only available for increment format"))
			return
		}

		// if args[0] is not provided, default to 1
		if len(args) == 0 {
			args = append(args, "1")
		}

		value, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf(util.FormatError("invalid value '%s': must be a number\n"), args[0])
			return
		}

		newVersion, err := project.UpdateIncrementVersion(history.Current, "subtract", value)
		if err != nil {
			fmt.Printf(util.FormatError("failed to update version: %v\n"), err)
			return
		}

		message, _ := cmd.Flags().GetString("message")
		if message == "" {
			message = fmt.Sprintf("Decrement version: %s -> %s", history.Current, newVersion)
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
	versionAddCmd.Flags().StringP("message", "m", "", "Commit message for this version")
	versionSubtractCmd.Flags().StringP("message", "m", "", "Commit message for this version")

	versionCmd.AddCommand(versionAddCmd)
	versionCmd.AddCommand(versionSubtractCmd)
}
