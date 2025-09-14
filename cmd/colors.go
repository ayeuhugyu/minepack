/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"minepack/util"

	"github.com/spf13/cobra"
)

// colorsCmd represents the colors command
var colorsCmd = &cobra.Command{
	Use:   "colors",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("colors called")

		fmt.Println(util.FormatError("This is an error message"))
		fmt.Println(util.FormatSuccess("This is a success message"))
	},
}

func init() {
	testCmd.AddCommand(colorsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// colorsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// colorsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
