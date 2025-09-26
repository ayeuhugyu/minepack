/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"minepack/core/project"
	"minepack/util"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
)

// finds all mods that depend on a given mod (recursively)
func findDependentMods(targetSlug string, allContent []project.ContentData, visited map[string]bool) []project.ContentData {
	if visited == nil {
		visited = make(map[string]bool)
	}

	if visited[targetSlug] {
		return []project.ContentData{}
	}
	visited[targetSlug] = true

	var dependents []project.ContentData

	for _, content := range allContent {
		// check if this mod has our target mod in its Dependencies list
		// (meaning this mod depends on our target mod)
		for _, dep := range content.Dependencies {
			if dep.Slug == targetSlug {
				dependents = append(dependents, content)
				// recursively find mods that depend on this dependent
				subDependents := findDependentMods(content.Slug, allContent, visited)
				dependents = append(dependents, subDependents...)
				break
			}
		}
	}

	return dependents
}

// finds dependencies that would become orphaned if we remove the given mods
func findOrphanedDependencies(modsToRemove []project.ContentData, allContent []project.ContentData) []project.ContentData {
	var orphaned []project.ContentData
	seen := make(map[string]bool) // to prevent duplicates

	// create a set of mods being removed for quick lookup
	removingSet := make(map[string]bool)
	for _, mod := range modsToRemove {
		removingSet[mod.Slug] = true
	}

	// for each mod being removed
	for _, targetMod := range modsToRemove {
		// for each dependency of the target mod
		for _, dep := range targetMod.Dependencies {
			if dep.DependencyType != project.Required {
				continue // only care about required dependencies
			}

			// find the actual content data for this dependency
			var depContent *project.ContentData
			for _, content := range allContent {
				if content.Slug == dep.Slug || content.Id == dep.Id {
					depContent = &content
					break
				}
			}

			if depContent == nil || seen[depContent.Slug] || removingSet[depContent.Slug] {
				continue // dependency not found, already processed, or being removed anyway
			}

			// check if this dependency was added as a dependency and would become unused
			if depContent.AddedAsDependency {
				// count how many mods in RequiredBy will NOT be removed
				remainingDependents := 0
				for _, reqBy := range depContent.RequiredBy {
					if !removingSet[reqBy.Slug] {
						remainingDependents++
					}
				}

				// if no mods will depend on this after removal, it becomes orphaned
				if remainingDependents == 0 {
					orphaned = append(orphaned, *depContent)
					seen[depContent.Slug] = true
				}
			}
		}
	}

	return orphaned
}

// removes a mod from the RequiredBy lists of its dependencies
func updateDependencyRequiredByOnRemoval(removedMod *project.ContentData, packData *project.Project) error {
	// for each dependency of the removed mod
	for _, dep := range removedMod.Dependencies {
		// get the dependency's content data
		depContent, err := packData.GetContent(dep.Slug)
		if err != nil {
			continue // dependency not found, skip
		}

		// remove the removed mod from the dependency's RequiredBy list
		var updatedRequiredBy []project.RequiredBy
		for _, reqBy := range depContent.RequiredBy {
			if reqBy.Slug != removedMod.Slug {
				updatedRequiredBy = append(updatedRequiredBy, reqBy)
			}
		}

		// update the dependency's RequiredBy field
		depContent.RequiredBy = updatedRequiredBy

		// save the updated dependency
		err = packData.UpdateContent(*depContent)
		if err != nil {
			return fmt.Errorf("failed to update dependency %s: %v", depContent.Name, err)
		}
	}

	return nil
}

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:     "remove",
	Short:   "remove a mod from your modpack",
	Long:    `removes a mod from your modpack, handling dependencies intelligently`,
	Aliases: []string{"rm", "delete", "uninstall"},
	Run: func(cmd *cobra.Command, args []string) {
		// get current working directory and parse project
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("error getting current working directory: %s"), err)
			return
		}

		packData, err := project.ParseProject(cwd)
		if err != nil {
			fmt.Printf(util.FormatError("error parsing project: %s"), err)
			return
		}

		// validate arguments
		if len(args) < 1 {
			fmt.Println("please provide a mod to remove.")
			return
		}

		// join all args as query
		query := ""
		for i, arg := range args {
			if i > 0 {
				query += " "
			}
			query += arg
		}

		// search for mod in project (same logic as query command)
		found := false
		mod := &project.ContentData{}

		// get all content first so we can use it in multiple search attempts
		allContent, err := packData.GetAllContent()
		if err != nil {
			fmt.Printf(util.FormatError("error getting all content: %s"), err)
			return
		}

		// first, try the basic slug/id match
		if packData.HasMod(query) {
			found = true
			mod, _ = packData.GetContent(query)
		}

		// if not found, try searching via the mod's name (exact match)
		if !found {
			for _, c := range allContent {
				if c.Name != "" && c.Name == query {
					found = true
					mod = &c
					break
				}
			}
		}

		// if STILL not found, try fuzzy search
		if !found {
			var options []string
			contentMap := make(map[string]project.ContentData)
			for _, c := range allContent {
				if c.Name != "" {
					options = append(options, c.Name)
					contentMap[c.Name] = c
				}
			}

			if len(options) == 0 {
				fmt.Print(util.FormatError("no mods found in project to search through\n"))
				return
			}

			matches := fuzzy.Find(query, options)
			if len(matches) > 0 {
				// limit to top 5 matches
				if len(matches) > 5 {
					matches = matches[:5]
				}

				// prompt user to select one of the matches
				var stringMatches []string
				for _, m := range matches {
					stringMatches = append(stringMatches, m.Str)
				}

				var optionsList []huh.Option[string]
				for _, s := range stringMatches {
					optionsList = append(optionsList, huh.NewOption(s, s))
				}
				// Add cancel option
				optionsList = append(optionsList, huh.NewOption("Cancel", "cancel"))

				var selectedQuery string
				prompt := huh.NewForm(
					huh.NewGroup(
						huh.NewSelect[string]().
							Title("found fuzzy results, pick the mod to remove:").
							Options(optionsList...).
							Value(&selectedQuery),
					),
				)

				err = prompt.Run()
				if err != nil {
					fmt.Printf("prompt failed %v\n", err)
					return
				}

				// Check if user cancelled
				if selectedQuery == "cancel" {
					fmt.Println("Operation cancelled")
					return
				}

				if selectedMod, exists := contentMap[selectedQuery]; exists {
					found = true
					mod = &selectedMod
				}
			}
		}

		if !found {
			fmt.Printf(util.FormatError("no mod found for query: %s\n"), query)
			return
		}

		// display the mod we're about to remove
		formatted := util.FormatContentData(*mod)
		fmt.Printf("found mod to remove:\n%s\n", formatted)

		// check for dependent mods (mods that require this one)
		dependentMods := findDependentMods(mod.Slug, allContent, nil)
		var modsToRemove []project.ContentData
		modsToRemove = append(modsToRemove, *mod)

		if len(dependentMods) > 0 {
			fmt.Printf(util.FormatWarning("warning: the following mods depend on %s:\n"), mod.Name)
			for _, dep := range dependentMods {
				fmt.Printf("- %s (%s)\n", dep.Name, dep.Slug)
			}

			var userChoice string
			choiceForm := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("what would you like to do?").
						Description("choose how to handle dependent mods").
						Options(
							huh.NewOption("remove all dependent mods as well", "remove_all"),
							huh.NewOption("remove anyway (may break dependencies)", "force_remove"),
							huh.NewOption("cancel the removal", "cancel"),
						).
						Value(&userChoice),
				),
			)

			err = choiceForm.Run()
			if err != nil {
				fmt.Printf(util.FormatError("prompt failed: %v\n"), err)
				return
			}

			switch userChoice {
			case "remove_all":
				modsToRemove = append(modsToRemove, dependentMods...)
				fmt.Printf(util.FormatSuccess("will remove %s and all %d dependent mods\n"), mod.Name, len(dependentMods))
			case "force_remove":
				fmt.Printf(util.FormatWarning("removing %s anyway (this may cause dependency issues)\n"), mod.Name)
			case "cancel":
				fmt.Println("removal cancelled.")
				return
			}
		}

		// check for orphaned dependencies (dependencies that would become unused)
		orphanedDeps := findOrphanedDependencies(modsToRemove, allContent)
		if len(orphanedDeps) > 0 {
			fmt.Print(util.FormatWarning("the following dependencies would become unused:\n"))
			for _, dep := range orphanedDeps {
				fmt.Printf("- %s (%s)\n", dep.Name, dep.Slug)
			}

			var removeOrphans bool
			huh.NewConfirm().
				Title("remove unused dependencies as well?").
				Description("these mods were added as dependencies and are no longer needed").
				Affirmative("yes, remove them").
				Negative("no, keep them").
				Value(&removeOrphans).
				Run()

			if removeOrphans {
				modsToRemove = append(modsToRemove, orphanedDeps...)
				fmt.Printf(util.FormatSuccess("will also remove %d orphaned dependencies\n"), len(orphanedDeps))
			}
		}

		// final confirmation
		var confirm bool
		var modNames []string
		for _, m := range modsToRemove {
			modNames = append(modNames, m.Name)
		}

		fmt.Printf("about to remove: %s\n", strings.Join(modNames, ", "))
		huh.NewConfirm().
			Title("are you sure you want to proceed?").
			Affirmative("yes, remove").
			Negative("cancel").
			Value(&confirm).
			Run()

		if !confirm {
			fmt.Println("removal cancelled.")
			return
		}

		// Load link state to track removed files
		linkState, err := LoadLinkState(cwd)
		if err != nil {
			fmt.Printf(util.FormatWarning("warning: failed to load link state: %s\n"), err)
			linkState = &LinkState{
				RemovedFiles:   []string{},
				OverridesFiles: make(map[string]string),
				Version:        "1.0",
			}
		}

		// perform the removal
		successCount := 0
		for _, modToRemove := range modsToRemove {
			// Track the file path for removal from linked instances
			linkState.AddRemovedFile(modToRemove.File.Filepath)

			// first, update the RequiredBy fields of this mod's dependencies
			err = updateDependencyRequiredByOnRemoval(&modToRemove, packData)
			if err != nil {
				fmt.Printf(util.FormatWarning("warning: failed to update dependencies for %s: %s\n"), modToRemove.Name, err)
			}

			// then remove the mod itself
			err = packData.RemoveContent(modToRemove.Slug)
			if err != nil {
				fmt.Printf(util.FormatError("error removing %s: %s\n"), modToRemove.Name, err)
			} else {
				fmt.Printf(util.FormatSuccess("removed %s\n"), modToRemove.Name)
				successCount++
			}
		}

		// Save link state with tracked removed files
		if err := SaveLinkState(cwd, linkState); err != nil {
			fmt.Printf(util.FormatWarning("warning: failed to save link state: %s\n"), err)
		}

		fmt.Printf(util.FormatSuccess("successfully removed %d of %d mods\n"), successCount, len(modsToRemove))
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
