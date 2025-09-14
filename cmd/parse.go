/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	proj "minepack/core/project"
	"minepack/util"
	"os"

	"github.com/spf13/cobra"
)

// parseCmd represents the parse command
var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("parse called")
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("failed to get current working directory: %v\n"), err)
			return
		}

		proj, err := proj.ParseProject(currentDir)
		if err != nil {
			fmt.Printf(util.FormatError("failed to parse project: %v\n"), err)
			return
		}

		fmt.Printf("Project Name: %s\n", proj.Name)
		fmt.Printf("Description: %s\n", proj.Description)
		fmt.Printf("Author: %s\n", proj.Author)
		fmt.Printf("Game Version: %s\n", proj.Versions.Game)
		fmt.Printf("Modloader: %s\n", proj.Versions.Loader.Name)
		fmt.Printf("Modloader Version: %v\n", proj.Versions.Loader.Version)
		fmt.Printf("Default Source: %v\n", proj.DefaultSource)
	},
}

func init() {
	testCmd.AddCommand(parseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// parseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// parseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
