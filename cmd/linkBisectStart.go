/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"minepack/core/bisect"
	"minepack/core/project"
	"minepack/util"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// linkBisectStartCmd represents the bisect start command
var linkBisectStartCmd = &cobra.Command{
	Use:   "start",
	Short: "start a new bisection session",
	Long:  `begin a new bisection to find problematic mods in your linked instance`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// get current working directory and parse project
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current working directory: %w", err)
		}

		projectRoot := cwd

		// Check if there's already an active bisection
		if _, err := bisect.LoadBisectState(projectRoot); err == nil {
			fmt.Print(util.FormatError("there's already an active bisection. use 'minepack link bisect finish' to end it first\n"))
			return nil
		}

		// Load linked folders
		linked, err := loadLinkedFolders(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to load linked folders: %w", err)
		}

		if len(linked.Links) == 0 {
			fmt.Print(util.FormatError("no linked instances found. use 'minepack link add' to link an instance first\n"))
			return nil
		}

		// Prompt user to select a linked instance
		var selectedInstance string
		var instanceOpts []huh.Option[string]
		for _, link := range linked.Links {
			instanceName := filepath.Base(link)
			instanceOpts = append(instanceOpts, huh.NewOption(fmt.Sprintf("%s (%s)", instanceName, link), link))
		}

		selectForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("select linked instance for bisection").
					Description("choose which linked instance to run the bisection on").
					Options(instanceOpts...).
					Value(&selectedInstance),
			),
		)

		if err := selectForm.Run(); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		// Load project content
		packData, err := project.ParseProject(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to parse project: %w", err)
		}

		// Get all content
		allContent, err := packData.GetAllContent()
		if err != nil {
			return fmt.Errorf("failed to get all content: %w", err)
		}

		if len(allContent) == 0 {
			fmt.Print(util.FormatError("no mods found in modpack\n"))
			return nil
		}

		// Create bisection state
		state, err := bisect.CreateBisectState(projectRoot, selectedInstance, allContent)
		if err != nil {
			return fmt.Errorf("failed to create bisection state: %w", err)
		}

		// Set creation time
		state.Created = time.Now().Format("2006-01-02 15:04:05")

		// Save the initial state
		if err := bisect.SaveBisectState(projectRoot, state); err != nil {
			return fmt.Errorf("failed to save bisection state: %w", err)
		}

		fmt.Printf(util.FormatSuccess("bisection started with %d mods\n"), len(allContent))
		fmt.Printf("target instance: %s\n", filepath.Base(selectedInstance))
		fmt.Println()
		fmt.Println("now:")
		fmt.Println("1. reproduce your issue to confirm it exists with all mods enabled")
		fmt.Println("2. run 'minepack link bisect next' to start the bisection process")
		fmt.Println()
		fmt.Println("tip: each step will disable about half the remaining candidate mods.")
		fmt.Println("     if the issue persists, the problem is in the enabled mods.")
		fmt.Println("     if the issue disappears, the problem is in the disabled mods.")

		return nil
	},
}

func init() {
	linkBisectCmd.AddCommand(linkBisectStartCmd)
}
