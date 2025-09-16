package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"minepack/core"
	"minepack/core/project"
	"minepack/util"

	"github.com/charmbracelet/huh"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	name                  string
	description           string
	author                string
	gameVersion           string
	selectedModloader     string
	selectedDefaultSource string
)

// initCmd represents the init command

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize a new minepack project",
	Long:  `initialize a new minepack project`,
	Run: func(cmd *cobra.Command, args []string) {
		var bannedCharacters = []rune{'/', '\\', ':', '*', '?', '"', '<', '>', '|'}

		validateMeta := func(input string) error {
			for _, char := range bannedCharacters {
				if strings.ContainsRune(input, char) {
					return fmt.Errorf("input cannot contain '%c'", char)
				}
			}
			return nil
		}

		metaForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("name").
					Description("the name that will be used for the project").
					Value(&name).
					Validate(validateMeta),
				huh.NewText().
					Title("description").
					Description("a brief description of the project").
					Lines(3).
					Value(&description).
					Validate(validateMeta),
				huh.NewInput().
					Title("author").
					Description("the author of the project").
					Value(&author).
					Validate(validateMeta),
			),
		)

		err := metaForm.Run()

		if err != nil {
			fmt.Printf(util.FormatError("prompt failed: %v\n"), err)
			return
		}

		fmt.Printf(util.FormatSuccess("project name: %s\n"), name)
		fmt.Printf(util.FormatSuccess("description: %s\n"), description)
		fmt.Printf(util.FormatSuccess("author: %s\n"), author)

		fmt.Printf("\n")
		// fetch minecraft versions
		gameVersionSpinner := progressbar.NewOptions(-1,
			progressbar.OptionSetDescription("fetching minecraft versions..."),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionClearOnFinish(),
		)
		gameVersionSpinner.Add(1)

		allGameVersions, err := core.FetchMinecraftVersions()
		var allGameVersionsFlat []string
		for _, v := range allGameVersions.Versions {
			allGameVersionsFlat = append(allGameVersionsFlat, v.ID)
		}

		gameVersionSpinner.Finish()

		if err != nil {
			fmt.Printf(util.FormatError("failed to fetch minecraft versions: %v\n"), err)
			return
		}

		// select game version
		versionValidator := func(input string) error {
			if input == "" {
				return nil // allow empty input to use default version
			}
			split := strings.Split(input, ".")
			if len(split) < 2 || len(split) > 3 {
				return errors.New("version must be in format 'Major.Minor' or 'Major.Minor.Patch'")
			}
			for _, v := range allGameVersionsFlat {
				if v == input {
					return nil
				}
			}
			return errors.New("version not found")
		}

		var inputGameVersion string
		versionForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("game version").
					Description("enter minecraft version (default: 1.20.1)").
					Placeholder("1.20.1").
					Suggestions(allGameVersionsFlat).
					Value(&inputGameVersion).
					Validate(versionValidator),
			),
		)

		err = versionForm.Run()
		if err != nil {
			fmt.Printf(util.FormatError("prompt failed: %v\n"), err)
			return
		}

		if inputGameVersion == "" {
			gameVersion = "1.20.1"
		} else {
			gameVersion = inputGameVersion
		}

		fmt.Printf(util.FormatSuccess("game version: %s\n"), gameVersion)

		fmt.Printf("\n")
		// fetch modloader versions

		modloaderVersionSpinner := progressbar.NewOptions(-1,
			progressbar.OptionSetDescription("fetching modloader versions..."),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionClearOnFinish(),
		)
		modloaderVersionSpinner.Add(1)

		allModloaderVersions := core.GetAllLatestVersions(gameVersion)

		modloaderVersionSpinner.Finish()

		if err != nil {
			fmt.Printf(util.FormatError("failed to fetch modloader versions: %v\n"), err)
			return
		}
		// select modloader

		// use modloaderVersions to determine the items in the select
		var availableModloaderNames []string
		for name := range allModloaderVersions {
			if name == "minecraft" {
				continue
			}
			if strings.HasPrefix(allModloaderVersions[name], "error:") {
				continue
			}
			availableModloaderNames = append(availableModloaderNames, name)
		}

		modloaderForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("modloader").
					Description("choose a modloader (default: fabric)").
					Options(huh.NewOptions(availableModloaderNames...)...).
					Value(&selectedModloader),
			),
		)

		err = modloaderForm.Run()
		if err != nil {
			fmt.Printf(util.FormatError("prompt failed: %v\n"), err)
			return
		}

		if selectedModloader == "" {
			selectedModloader = "fabric"
		}

		fmt.Printf(util.FormatSuccess("modloader: %s\n"), selectedModloader)
		fmt.Printf(util.FormatSuccess("modloader version: %s\n"), allModloaderVersions[selectedModloader])

		// select default source
		var availableDefaultSources = []string{"modrinth", "curseforge"}

		defaultSourceForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("default source").
					Description("choose a default source for mods (default: modrinth)").
					Options(huh.NewOptions(availableDefaultSources...)...).
					Value(&selectedDefaultSource),
			),
		)
		err = defaultSourceForm.Run()
		if err != nil {
			fmt.Printf(util.FormatError("prompt failed: %v\n"), err)
			return
		}
		if selectedDefaultSource == "" {
			selectedDefaultSource = "modrinth"
		}
		// if the selected source is curseforge, verify that the user actually wants to use this
		if selectedDefaultSource == "curseforge" {
			var confirm bool
			fmt.Println("curseforge has a lot of limitations compared to modrinth, such as no support for datapacks or shaderpacks.\nsome features of minepack will not work when curseforge is the default source.")
			huh.NewConfirm().
				Title("are you sure you want to use curseforge as the default source?").
				Affirmative("yup").
				Negative("nah").
				Value(&confirm).
				Run()

			if !confirm {
				selectedDefaultSource = "modrinth"
				fmt.Printf("default source set to modrinth instead\n")
			}
		}

		fmt.Printf(util.FormatSuccess("default source: %s\n"), selectedDefaultSource)

		// root will be the current working directory\normalized version of project name
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("failed to get current working directory: %v\n"), err)
			return
		}
		root := cwd + "/" + strings.ReplaceAll(name, " ", "-")

		fmt.Printf("project root: %s\n", root)

		fmt.Printf("\n")

		projectData := project.Project{
			Name:          name,
			Description:   description,
			Author:        author,
			DefaultSource: selectedDefaultSource,
			Root:          root,
			Versions: project.ProjectVersions{
				Game: gameVersion,
				Loader: project.ModloaderVersion{
					Name:    selectedModloader,
					Version: allModloaderVersions[selectedModloader],
				},
			},
		}

		var confirm bool
		fmt.Print("this will write to " + root + string(filepath.Separator) + "project.mp.yaml. ")
		huh.NewConfirm().
			Title("do you want to continue?").
			Affirmative("yup").
			Negative("nah").
			Value(&confirm).
			Run()

		if !confirm {
			fmt.Printf("aborting project initialization\n")
			return
		}

		err = project.WriteProject(&projectData)
		if err != nil {
			fmt.Printf(util.FormatError("failed to write project files: %v\n"), err)
			return
		}

		fmt.Print(util.FormatSuccess("project initialized!\n"))
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
