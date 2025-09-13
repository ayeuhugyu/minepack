package cmd

import (
	"errors"
	"fmt"
	"strings"

	"minepack/core"
	"minepack/util"

	"github.com/charmbracelet/huh"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	name              string
	description       string
	author            string
	gameVersion       string
	selectedModloader string
)

// initCmd represents the init command

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Minepack project",
	Long:  `Initialize a new Minepack project in the current directory.`,
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
					Title("Name").
					Description("the name that will be used for the project").
					Value(&name).
					Validate(validateMeta),
				huh.NewInput().
					Title("Description").
					Description("a brief description of the project").
					Value(&description).
					Validate(validateMeta),
				huh.NewInput().
					Title("Author").
					Description("the author of the project").
					Value(&author).
					Validate(validateMeta),
			),
		)

		err := metaForm.Run()

		if err != nil {
			fmt.Printf(util.FormatError("Prompt failed: %v\n"), err)
			return
		}

		fmt.Printf(util.FormatSuccess("Project Name: %s\n"), name)
		fmt.Printf(util.FormatSuccess("Description: %s\n"), description)
		fmt.Printf(util.FormatSuccess("Author: %s\n"), author)

		fmt.Printf("\n")
		// fetch minecraft versions
		gameVersionSpinner := progressbar.NewOptions(-1,
			progressbar.OptionSetDescription("Fetching Minecraft versions..."),
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
			fmt.Printf(util.FormatError("Failed to fetch Minecraft versions: %v\n"), err)
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
					Title("Game Version").
					Description("Enter Minecraft version (default: 1.20.1)").
					Placeholder("1.20.1").
					Value(&inputGameVersion).
					Validate(versionValidator),
			),
		)

		err = versionForm.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		if inputGameVersion == "" {
			gameVersion = "1.20.1"
		} else {
			gameVersion = inputGameVersion
		}

		fmt.Printf(util.FormatSuccess("Game Version: %s\n"), gameVersion)

		fmt.Printf("\n")
		// fetch modloader versions

		modloaderVersionSpinner := progressbar.NewOptions(-1,
			progressbar.OptionSetDescription("Fetching modloader versions..."),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionClearOnFinish(),
		)
		modloaderVersionSpinner.Add(1)

		allModloaderVersions := core.GetAllLatestVersions(gameVersion)

		modloaderVersionSpinner.Finish()

		if err != nil {
			fmt.Printf(util.FormatError("Failed to fetch modloader versions: %v\n"), err)
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
					Title("Modloader").
					Description("Choose a modloader (default: Fabric)").
					Options(huh.NewOptions(availableModloaderNames...)...).
					Value(&selectedModloader),
			),
		)

		err = modloaderForm.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		if selectedModloader == "" {
			selectedModloader = "Fabric"
		}

		fmt.Printf(util.FormatSuccess("Modloader: %s\n"), selectedModloader)
		fmt.Printf(util.FormatSuccess("Modloader Version: %s\n"), allModloaderVersions[selectedModloader])

		fmt.Printf("\n")
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
