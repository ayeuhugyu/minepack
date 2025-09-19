/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"minepack/core/api"
	"minepack/core/api/curseforge"
	"minepack/core/api/modrinth"
	"minepack/core/project"
	"minepack/util"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// dependency resolution context to track what we're processing
type depResolutionContext struct {
	packData         *project.Project
	chooseDeps       bool
	processedDeps    map[string]bool // slug/id -> processed to avoid duplicates
	incompatibleDeps []project.Dependency
	requiredBy       map[string]project.RequiredBy // tracks what requires each dependency
}

// creates a new dependency resolution context
func newDepResolutionContext(packData *project.Project, chooseDeps bool) *depResolutionContext {
	return &depResolutionContext{
		packData:         packData,
		chooseDeps:       chooseDeps,
		processedDeps:    make(map[string]bool),
		incompatibleDeps: []project.Dependency{},
		requiredBy:       make(map[string]project.RequiredBy),
	}
}

// fetches a project and its versions based on the source
func fetchProjectData(identifier string, packData *project.Project) (*project.ContentData, error) {
	spinner := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription(fmt.Sprintf("fetching %s...", identifier)),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionClearOnFinish(),
	)
	spinner.Add(1)
	defer spinner.Finish()

	if packData.DefaultSource == "modrinth" {
		projectData, err := modrinth.GetProject(identifier)
		if err != nil || projectData == nil {
			return nil, fmt.Errorf("failed to fetch modrinth project %s: %w", identifier, err)
		}

		versions, err := modrinth.GetProjectVersions(identifier, *packData)
		if err != nil || len(versions) == 0 {
			return nil, fmt.Errorf("no compatible versions found for modrinth project %s", identifier)
		}

		contentData := modrinth.ConvertProjectToContentData(projectData, versions[0])
		return &contentData, nil
	} else {
		projectData, err := curseforge.GetProject(identifier)
		if err != nil || projectData == nil {
			return nil, fmt.Errorf("failed to fetch curseforge project %s: %w", identifier, err)
		}

		versions, err := curseforge.GetProjectVersions(identifier, *packData)
		if err != nil || len(versions) == 0 {
			return nil, fmt.Errorf("no compatible versions found for curseforge project %s", identifier)
		}

		contentData := curseforge.ConvertModToContentData(projectData, &versions[0])
		return &contentData, nil
	}
}

// handles incompatible dependencies found in the project
func handleIncompatibleDependencies(ctx *depResolutionContext) error {
	if len(ctx.incompatibleDeps) == 0 {
		return nil
	}

	// find which incompatible deps are actually in the project
	incompatInProject := []project.Dependency{}
	for _, dep := range ctx.incompatibleDeps {
		if ctx.packData.HasMod(dep.Slug) || ctx.packData.HasMod(dep.Id) {
			incompatInProject = append(incompatInProject, dep)
		}
	}

	if len(incompatInProject) == 0 {
		return nil
	}

	fmt.Println(util.FormatWarning("warning: the following incompatible dependencies were found in your modpack:\n"))
	for _, dep := range incompatInProject {
		fmt.Printf("- %s (%s)\n", dep.Name, dep.Slug)
	}

	var userChoice string
	choiceForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("incompatible dependencies found").
				Description("choose an option to proceed").
				Options(
					huh.NewOption("remove all incompatible dependencies", "remove"),
					huh.NewOption("continue adding the mod (not recommended)", "continue"),
					huh.NewOption("cancel the add operation", "cancel"),
					huh.NewOption("pick which incompatible dependencies to remove", "pick"),
				).
				Value(&userChoice),
		),
	)

	if err := choiceForm.Run(); err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	switch userChoice {
	case "remove":
		for _, dep := range incompatInProject {
			if err := ctx.packData.RemoveContent(dep.Slug); err != nil {
				fmt.Printf(util.FormatError("error removing incompatible dependency %s: %s\n"), dep.Name, err)
			} else {
				fmt.Printf(util.FormatSuccess("removed incompatible dependency %s\n"), dep.Name)
			}
		}
	case "continue":
		fmt.Println(util.FormatWarning("continuing to add the mod with incompatible dependencies present (not recommended)"))
	case "cancel":
		return fmt.Errorf("add operation cancelled")
	case "pick":
		var depsToRemove []string
		var opts []huh.Option[string]
		for _, dep := range incompatInProject {
			opts = append(opts, huh.NewOption(fmt.Sprintf("%s (%s)", dep.Name, dep.Slug), dep.Slug))
		}

		pickForm := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("pick incompatible dependencies to remove").
					Description("select the incompatible dependencies you want to remove").
					Options(opts...).
					Value(&depsToRemove),
			),
		)

		if err := pickForm.Run(); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		for _, slug := range depsToRemove {
			if err := ctx.packData.RemoveContent(slug); err != nil {
				fmt.Printf(util.FormatError("error removing incompatible dependency %s: %s\n"), slug, err)
			} else {
				fmt.Printf(util.FormatSuccess("removed incompatible dependency %s\n"), slug)
			}
		}
	}

	return nil
}

// recursively resolves dependencies, handling deps of deps
func resolveDependenciesRecursively(ctx *depResolutionContext, contentData *project.ContentData, depth int) error {
	if depth > 10 { // prevent infinite recursion
		fmt.Print(util.FormatWarning("maximum dependency depth reached, stopping recursive resolution\n"))
		return nil
	}

	// collect all dependencies from this content
	var allDeps []project.Dependency
	for _, dep := range contentData.Dependencies {
		allDeps = append(allDeps, dep)

		// track incompatible dependencies
		if dep.DependencyType == project.Incompatible {
			ctx.incompatibleDeps = append(ctx.incompatibleDeps, dep)
		}
	}

	// filter dependencies based on user choice or requirement
	depsToProcess := []project.Dependency{}
	if ctx.chooseDeps && len(allDeps) > 0 {
		// let user choose which dependencies to add
		var depsToAddSlugs []string
		var opts []huh.Option[string]
		for _, dep := range allDeps {
			key := dep.Slug
			if key == "" {
				key = dep.Id
			}
			if !ctx.packData.HasMod(dep.Slug) && !ctx.packData.HasMod(dep.Id) && !ctx.processedDeps[key] {
				opts = append(opts, huh.NewOption(fmt.Sprintf("%s (%s) - %s", dep.Name, dep.Slug, project.DependencyTypeToString(dep.DependencyType)), key))
			}
		}

		if len(opts) > 0 {
			pickForm := huh.NewForm(
				huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("pick dependencies to add").
						Description("select the dependencies you want to add").
						Options(opts...).
						Value(&depsToAddSlugs),
				),
			)

			if err := pickForm.Run(); err != nil {
				return fmt.Errorf("dependency selection prompt failed: %w", err)
			}

			for _, key := range depsToAddSlugs {
				for _, dep := range allDeps {
					depKey := dep.Slug
					if depKey == "" {
						depKey = dep.Id
					}
					if depKey == key {
						depsToProcess = append(depsToProcess, dep)
					}
				}
			}
		}
	} else {
		// automatically add required dependencies
		for _, dep := range allDeps {
			if dep.DependencyType == project.Required {
				depsToProcess = append(depsToProcess, dep)
			}
		}
	}

	// process each dependency
	for _, dep := range depsToProcess {
		key := dep.Slug
		if key == "" {
			key = dep.Id
		}

		// skip if already processed or already in modpack
		if ctx.processedDeps[key] || ctx.packData.HasMod(dep.Slug) || ctx.packData.HasMod(dep.Id) {
			// update existing dependency's RequiredBy field if it's in the modpack
			if ctx.packData.HasMod(dep.Slug) || ctx.packData.HasMod(dep.Id) {
				if err := updateDependencyRequiredBy(ctx, dep, contentData); err != nil {
					fmt.Printf(util.FormatError("error updating dependency %s: %s\n"), dep.Name, err)
				}
			}
			continue
		}

		// mark as processed to avoid circular dependencies
		ctx.processedDeps[key] = true

		// fetch the dependency's data
		identifier := dep.Slug
		if identifier == "" {
			identifier = dep.Id
		}

		depContentData, err := fetchProjectData(identifier, ctx.packData)
		if err != nil {
			fmt.Printf(util.FormatError("error fetching dependency %s: %s\n"), dep.Name, err)
			continue
		}

		// set up the RequiredBy relationship
		depContentData.RequiredBy = []project.RequiredBy{
			{
				Name: contentData.Name,
				Slug: contentData.Slug,
				Id:   contentData.Id,
			},
		}

		// display what we're adding
		formatted := util.FormatContentData(*depContentData)
		fmt.Printf("%s\n", formatted)

		// add the dependency to the modpack
		if err := ctx.packData.AddContent(*depContentData); err != nil {
			fmt.Printf(util.FormatError("error adding dependency %s: %s\n"), depContentData.Name, err)
			continue
		}
		fmt.Printf(util.FormatSuccess("successfully added dependency %s\n"), depContentData.Name)

		// recursively resolve this dependency's dependencies
		if err := resolveDependenciesRecursively(ctx, depContentData, depth+1); err != nil {
			fmt.Printf(util.FormatError("error resolving sub-dependencies for %s: %s\n"), depContentData.Name, err)
		}
	}

	return nil
}

// updates an existing dependency's RequiredBy field
func updateDependencyRequiredBy(ctx *depResolutionContext, dep project.Dependency, requiredBy *project.ContentData) error {
	depContentData, err := ctx.packData.GetContent(dep.Slug)
	if err != nil {
		return err
	}

	// check if RequiredBy already contains this mod
	alreadyRequired := false
	for _, reqBy := range depContentData.RequiredBy {
		if reqBy.Id == requiredBy.Id {
			alreadyRequired = true
			break
		}
	}

	if !alreadyRequired {
		depContentData.RequiredBy = append(depContentData.RequiredBy, project.RequiredBy{
			Name: requiredBy.Name,
			Slug: requiredBy.Slug,
			Id:   requiredBy.Id,
		})

		if err := ctx.packData.UpdateContent(*depContentData); err != nil {
			return err
		}
		fmt.Printf(util.FormatSuccess("successfully updated dependency %s in modpack\n"), dep.Name)
	}

	return nil
}

// writes incompatible dependencies to incompat.mp.sum.yaml
func writeIncompatibleSummary(ctx *depResolutionContext) error {
	if len(ctx.incompatibleDeps) == 0 {
		return nil
	}

	var sums []project.SummaryObject
	for _, dep := range ctx.incompatibleDeps {
		sums = append(sums, project.SummaryObject{
			Slug:        dep.Slug,
			Id:          dep.Id,
			ContentType: project.Mod,
		})
	}

	if err := project.WriteIncompatSum(sums, ctx.packData.Root); err != nil {
		return fmt.Errorf("error writing incompat.mp.sum.yaml: %w", err)
	}

	fmt.Println(util.FormatSuccess("successfully updated incompat.mp.sum.yaml"))
	return nil
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:     "add",
	Short:   "add a mod to your modpack",
	Long:    `allows you to add mods to your modpack by providing a link, slug, id, or search term.`,
	Aliases: []string{"install"},
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
			fmt.Println("please provide a mod to add.")
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

		// search for the mod
		searchSpinner := progressbar.NewOptions(-1,
			progressbar.OptionSetDescription("searching for mods..."),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionClearOnFinish(),
		)
		searchSpinner.Add(1)

		result, err := api.SearchAll(query, *packData)
		searchSpinner.Finish()

		if err != nil {
			fmt.Printf(util.FormatError("search failed: %s"), err)
			return
		}
		if result == nil {
			fmt.Println(util.FormatError("no results found"))
			return
		}

		// display the main mod we're adding
		formatted := util.FormatContentData(*result)
		fmt.Printf("%s\n", formatted)

		// set up dependency resolution context
		chooseDeps, _ := cmd.Flags().GetBool("choose-dependencies")
		ctx := newDepResolutionContext(packData, chooseDeps)

		// handle incompatible dependencies first
		for _, dep := range result.Dependencies {
			if dep.DependencyType == project.Incompatible {
				ctx.incompatibleDeps = append(ctx.incompatibleDeps, dep)
			}
		}

		if err := handleIncompatibleDependencies(ctx); err != nil {
			fmt.Printf(util.FormatError("dependency conflict resolution failed: %s"), err)
			return
		}

		// resolve all dependencies recursively
		if err := resolveDependenciesRecursively(ctx, result, 0); err != nil {
			fmt.Printf(util.FormatError("dependency resolution failed: %s"), err)
			return
		}

		// add the main mod
		if err := packData.AddContent(*result); err != nil {
			fmt.Printf(util.FormatError("error adding mod %s: %s\n"), result.Name, err)
			return
		}
		fmt.Printf(util.FormatSuccess("successfully added mod %s\n"), result.Name)

		// write incompatible dependencies summary
		if err := writeIncompatibleSummary(ctx); err != nil {
			fmt.Printf(util.FormatError("failed to write incompatible summary: %s"), err)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	addCmd.Flags().BoolP("choose-dependencies", "d", false, "when enabled, you will manually choose which dependencies to add (if applicable). by default, all required dependencies are added automatically.")
}
