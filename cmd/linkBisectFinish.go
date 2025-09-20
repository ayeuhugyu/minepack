/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"minepack/core/bisect"
	"minepack/util"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// linkBisectFinishCmd represents the bisect finish command
var linkBisectFinishCmd = &cobra.Command{
	Use:     "finish",
	Aliases: []string{"stop", "end"},
	Short:   "finish the bisection and restore all mods",
	Long:    `complete the bisection process, restore all mods to enabled state, and clean up bisection data`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current working directory: %w", err)
		}

		projectRoot := cwd

		// Load bisection state
		state, err := bisect.LoadBisectState(projectRoot)
		if err != nil {
			fmt.Print(util.FormatError("no active bisection found\n"))
			return nil
		}

		// Show bisection summary
		candidates := state.GetCurrentCandidates()
		fmt.Printf("bisection summary:\n")
		fmt.Printf("- started with %d mods\n", len(state.AllMods))
		fmt.Printf("- completed %d steps\n", len(state.History))

		if len(candidates) == 0 {
			fmt.Printf("- result: no problematic mods found\n")
		} else if len(candidates) == 1 {
			fmt.Printf("- result: problematic mod identified - %s\n", candidates[0])
		} else {
			fmt.Printf("- result: narrowed down to %d candidate mods: %v\n", len(candidates), candidates)
		}
		fmt.Println()

		// Confirm finish
		var confirm bool
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("finish bisection?").
					Description("this will restore all mods and delete bisection data").
					Value(&confirm),
			),
		)

		if err := confirmForm.Run(); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		if !confirm {
			fmt.Println("bisection not finished - use 'minepack link bisect next' to continue")
			return nil
		}

		// Restore all mods
		if err := state.RestoreAllMods(); err != nil {
			return fmt.Errorf("failed to restore mods: %w", err)
		}

		// Delete bisection state
		if err := bisect.DeleteBisectState(projectRoot); err != nil {
			return fmt.Errorf("failed to clean up bisection data: %w", err)
		}

		fmt.Print(util.FormatSuccess("bisection finished and all mods restored\n"))
		fmt.Printf("target instance: %s\n", filepath.Base(state.LinkedInstance))

		if len(candidates) == 1 {
			fmt.Println()
			fmt.Printf("🎯 problematic mod: %s\n", candidates[0])
			fmt.Println("you can now remove this mod or investigate further")
		} else if len(candidates) > 1 {
			fmt.Println()
			fmt.Printf("🤔 multiple candidates remain: %v\n", candidates)
			fmt.Println("consider running another bisection or investigating these mods manually")
		}

		return nil
	},
}

func init() {
	linkBisectCmd.AddCommand(linkBisectFinishCmd)
}
