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

	"github.com/spf13/cobra"
)

// linkBisectPreviousCmd represents the bisect previous command
var linkBisectPreviousCmd = &cobra.Command{
	Use:   "previous",
	Short: "go back to the previous step of bisection",
	Long:  `return to the previous step in the bisection process and reapply that mod configuration`,
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
			fmt.Print(util.FormatError("no active bisection found. use 'minepack link bisect start' to begin\n"))
			return nil
		}

		// Go to previous step
		if err := state.GoToPreviousStep(); err != nil {
			fmt.Print(util.FormatError(fmt.Sprintf("cannot go to previous step: %s\n", err.Error())))
			return nil
		}

		// Apply the previous step
		if err := state.ApplyCurrentStep(); err != nil {
			return fmt.Errorf("failed to apply previous step: %w", err)
		}

		// Save the updated state
		if err := bisect.SaveBisectState(projectRoot, state); err != nil {
			return fmt.Errorf("failed to save bisection state: %w", err)
		}

		currentStep := state.History[state.CurrentStep]
		fmt.Printf(util.FormatSuccess("returned to step %d\n"), state.CurrentStep+1)
		fmt.Printf("disabled %d mods, enabled %d mods\n", len(currentStep.DisabledMods), len(currentStep.EnabledMods))
		fmt.Printf("target instance: %s\n", filepath.Base(state.LinkedInstance))

		if currentStep.TestResult != "unknown" {
			fmt.Printf("previous result: %s\n", currentStep.TestResult)
		}

		fmt.Println()
		fmt.Println("test your minecraft instance again and use 'minepack link bisect next' to continue")

		return nil
	},
}

func init() {
	linkBisectCmd.AddCommand(linkBisectPreviousCmd)
}
