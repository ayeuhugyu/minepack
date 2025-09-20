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

// linkBisectNextCmd represents the bisect next command
var linkBisectNextCmd = &cobra.Command{
	Use:     "next",
	Aliases: []string{"continue"},
	Short:   "continue to the next step of bisection",
	Long:    `move to the next step in the bisection process, disabling half of the remaining candidate mods`,
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

		// If we have a current step, ask for the test result first
		if state.CurrentStep >= 0 && state.CurrentStep < len(state.History) {
			currentStep := state.History[state.CurrentStep]
			if currentStep.TestResult == "unknown" {
				var result string
				resultForm := huh.NewForm(
					huh.NewGroup(
						huh.NewSelect[string]().
							Title("what was the result of your test?").
							Description("did the issue occur with the current mod configuration?").
							Options(
								huh.NewOption("issue still occurs (bad)", "bad"),
								huh.NewOption("issue is fixed (good)", "good"),
								huh.NewOption("cancel", "cancel"),
							).
							Value(&result),
					),
				)

				if err := resultForm.Run(); err != nil {
					return fmt.Errorf("prompt failed: %w", err)
				}

				if result == "cancel" {
					fmt.Println("bisection step cancelled")
					return nil
				}

				// Save the result
				if err := state.AddStepResult(result); err != nil {
					return fmt.Errorf("failed to add step result: %w", err)
				}

				if err := bisect.SaveBisectState(projectRoot, state); err != nil {
					return fmt.Errorf("failed to save bisection state: %w", err)
				}

				fmt.Printf(util.FormatSuccess("recorded result: %s\n"), result)
			}
		}

		// Calculate next step
		disabled, enabled, err := state.GetNextBisectStep()
		if err != nil {
			if err.Error() == "bisection complete" {
				candidates := state.GetCurrentCandidates()
				if len(candidates) == 0 {
					fmt.Print(util.FormatSuccess("bisection complete! no problematic mods found.\n"))
				} else if len(candidates) == 1 {
					fmt.Printf(util.FormatSuccess("bisection complete! problematic mod found: %s\n"), candidates[0])
				} else {
					fmt.Printf(util.FormatWarning("bisection narrowed down to %d mods: %v\n"), len(candidates), candidates)
				}
				fmt.Println("use 'minepack link bisect finish' to clean up and restore all mods")
				return nil
			}
			return fmt.Errorf("failed to calculate next step: %w", err)
		}

		// Add the new step
		state.AddStep(disabled, enabled)

		// Apply the step (rename files)
		if err := state.ApplyCurrentStep(); err != nil {
			return fmt.Errorf("failed to apply bisection step: %w", err)
		}

		// Save the updated state
		if err := bisect.SaveBisectState(projectRoot, state); err != nil {
			return fmt.Errorf("failed to save bisection state: %w", err)
		}

		fmt.Printf(util.FormatSuccess("step %d applied\n"), state.CurrentStep+1)
		fmt.Printf("disabled %d mods, enabled %d mods\n", len(disabled), len(enabled))
		fmt.Printf("target instance: %s\n", filepath.Base(state.LinkedInstance))
		fmt.Println()
		fmt.Println("now test your minecraft instance:")
		fmt.Println("- if the issue still occurs: the problem is in the ENABLED mods")
		fmt.Println("- if the issue is fixed: the problem is in the DISABLED mods")
		fmt.Println()
		fmt.Println("after testing, run 'minepack link bisect next' again to continue")

		return nil
	},
}

func init() {
	linkBisectCmd.AddCommand(linkBisectNextCmd)
}
