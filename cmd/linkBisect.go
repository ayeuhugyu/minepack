/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// linkBisectCmd represents the bisect command
var linkBisectCmd = &cobra.Command{
	Use:   "bisect",
	Short: "binary search to find problematic mods",
	Long: `use binary search (bisection) to systematically find which mod is causing issues in your modpack.

this is incredibly useful when you have crashes, lag, or other problems but don't know which mod is responsible.
the bisect process will methodically disable half your mods at a time, respecting dependencies, until the
problematic mod is isolated.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("use 'minepack link bisect start' to begin a bisection session")
		fmt.Println("or 'minepack link bisect help' for detailed instructions")
	},
}

func init() {
	linkCmd.AddCommand(linkBisectCmd)
}
