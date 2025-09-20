/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"minepack/core/project"
	"minepack/util"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// LinkedFolders represents the structure of linked.mp.yaml
type LinkedFolders struct {
	Links []string `yaml:"links"`
}

// getLinkedFile returns the path to linked.mp.yaml
func getLinkedFile(projectRoot string) string {
	return filepath.Join(projectRoot, "linked.mp.yaml")
}

// loadLinkedFolders loads the current linked folders from linked.mp.yaml
func loadLinkedFolders(projectRoot string) (*LinkedFolders, error) {
	linkedFile := getLinkedFile(projectRoot)

	// If file doesn't exist, return empty list
	if _, err := os.Stat(linkedFile); os.IsNotExist(err) {
		return &LinkedFolders{Links: []string{}}, nil
	}

	data, err := os.ReadFile(linkedFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read linked.mp.yaml: %w", err)
	}

	var linked LinkedFolders
	if err := yaml.Unmarshal(data, &linked); err != nil {
		return nil, fmt.Errorf("failed to parse linked.mp.yaml: %w", err)
	}

	return &linked, nil
}

// saveLinkedFolders saves the linked folders to linked.mp.yaml
func saveLinkedFolders(projectRoot string, linked *LinkedFolders) error {
	linkedFile := getLinkedFile(projectRoot)

	// If the list is empty, delete the file
	if len(linked.Links) == 0 {
		if _, err := os.Stat(linkedFile); err == nil {
			if err := os.Remove(linkedFile); err != nil {
				return fmt.Errorf("failed to remove linked.mp.yaml: %w", err)
			}
		}
		return nil
	}

	data, err := yaml.Marshal(linked)
	if err != nil {
		return fmt.Errorf("failed to marshal linked folders: %w", err)
	}

	if err := os.WriteFile(linkedFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write linked.mp.yaml: %w", err)
	}

	return nil
}

// linkCmd represents the link command
var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "manage linked minecraft instance folders",
	Long:  `link and unlink your modpack project to minecraft instance folders for easier management`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("use 'minepack link add [folder]' to add a link or 'minepack link remove' to remove links")
		fmt.Println("run 'minepack link --help' for more information")
	},
}

// linkAddCmd represents the link add command
var linkAddCmd = &cobra.Command{
	Use:     "add [folder_path]",
	Short:   "link a Minecraft instance folder to this modpack",
	Long:    `adds a link to a Minecraft instance folder, allowing future operations to sync with that instance`,
	Aliases: []string{"new", "link"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// get current working directory and parse project
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("error getting current working directory: %s"), err)
			return
		}

		_, err = project.ParseProject(cwd)
		if err != nil {
			fmt.Printf(util.FormatError("error parsing project: %s"), err)
			return
		}

		folderPath := args[0]

		// Convert to absolute path
		absPath, err := filepath.Abs(folderPath)
		if err != nil {
			fmt.Printf(util.FormatError("error resolving path: %s"), err)
			return
		}

		// Check if the folder exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			fmt.Printf(util.FormatError("folder does not exist: %s"), absPath)
			return
		}

		// Check if it's actually a directory
		info, err := os.Stat(absPath)
		if err != nil {
			fmt.Printf(util.FormatError("error checking folder: %s"), err)
			return
		}
		if !info.IsDir() {
			fmt.Printf(util.FormatError("path is not a directory: %s"), absPath)
			return
		}

		// Load current linked folders
		linked, err := loadLinkedFolders(cwd)
		if err != nil {
			fmt.Printf(util.FormatError("error loading linked folders: %s"), err)
			return
		}

		// Check if already linked
		for _, existingPath := range linked.Links {
			if existingPath == absPath {
				fmt.Printf(util.FormatWarning("folder is already linked: %s"), absPath)
				return
			}
		}

		// Add the new link
		linked.Links = append(linked.Links, absPath)

		// Save the updated list
		if err := saveLinkedFolders(cwd, linked); err != nil {
			fmt.Printf(util.FormatError("error saving linked folders: %s"), err)
			return
		}

		fmt.Printf(util.FormatSuccess("successfully linked folder: %s"), absPath)
	},
}

// linkRemoveCmd represents the link remove command
var linkRemoveCmd = &cobra.Command{
	Use:     "remove",
	Short:   "remove a linked minecraft instance folder",
	Long:    `removes a link to a minecraft instance folder from this modpack`,
	Aliases: []string{"unlink", "rm"},
	Run: func(cmd *cobra.Command, args []string) {
		// get current working directory and parse project
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("error getting current working directory: %s"), err)
			return
		}

		_, err = project.ParseProject(cwd)
		if err != nil {
			fmt.Printf(util.FormatError("error parsing project: %s"), err)
			return
		}

		// Load current linked folders
		linked, err := loadLinkedFolders(cwd)
		if err != nil {
			fmt.Printf(util.FormatError("error loading linked folders: %s"), err)
			return
		}

		if len(linked.Links) == 0 {
			fmt.Println(util.FormatError("no linked folders found"))
			return
		}

		// Create options for selection
		var options []huh.Option[string]
		for _, link := range linked.Links {
			options = append(options, huh.NewOption(link, link))
		}

		// Add cancel option
		options = append(options, huh.NewOption("cancel", "cancel"))

		var selectedLink string
		prompt := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("select linked folder to remove:").
					Options(options...).
					Value(&selectedLink),
			),
		)

		err = prompt.Run()
		if err != nil {
			fmt.Printf(util.FormatError("prompt failed: %v"), err)
			return
		}

		// Check if user cancelled
		if selectedLink == "cancel" {
			fmt.Println(util.FormatWarning("operation cancelled"))
			return
		}

		// Remove the selected link
		var updatedLinks []string
		for _, link := range linked.Links {
			if link != selectedLink {
				updatedLinks = append(updatedLinks, link)
			}
		}

		linked.Links = updatedLinks

		// Save the updated list
		if err := saveLinkedFolders(cwd, linked); err != nil {
			fmt.Printf(util.FormatError("error saving linked folders: %s"), err)
			return
		}

		fmt.Printf(util.FormatSuccess("successfully removed linked folder: %s"), selectedLink)
	},
}

// linkListCmd represents the link list command
var linkListCmd = &cobra.Command{
	Use:     "list",
	Short:   "list all linked minecraft instance folders",
	Long:    `displays all minecraft instance folders linked to this modpack`,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		// get current working directory and parse project
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("error getting current working directory: %s"), err)
			return
		}

		_, err = project.ParseProject(cwd)
		if err != nil {
			fmt.Printf(util.FormatError("error parsing project: %s"), err)
			return
		}

		// Load current linked folders
		linked, err := loadLinkedFolders(cwd)
		if err != nil {
			fmt.Printf(util.FormatError("error loading linked folders: %s"), err)
			return
		}

		if len(linked.Links) == 0 {
			fmt.Println("no linked folders found")
			fmt.Println("use 'minepack link add [folder]' to add a link")
			return
		}

		fmt.Printf("linked minecraft instances (%d):\n", len(linked.Links))
		for i, link := range linked.Links {
			// Check if folder still exists
			if _, err := os.Stat(link); os.IsNotExist(err) {
				fmt.Printf("  %d. %s %s\n", i+1, link, util.FormatWarning("(missing)"))
			} else {
				fmt.Printf("  %d. %s\n", i+1, link)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(linkCmd)
	linkCmd.AddCommand(linkAddCmd)
	linkCmd.AddCommand(linkRemoveCmd)
	linkCmd.AddCommand(linkListCmd)
}
