package cmd

import (
	"errors"
	"fmt"
	"strings"

	"minepack/core"

	"github.com/manifoldco/promptui"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Minepack project",
	Long:  `Initialize a new Minepack project in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {

		// name the project

		var bannedCharacters = []rune{'/', '\\', ':', '*', '?', '"', '<', '>', '|'}
		validateName := func(input string) error {
			if input == "" {
				return errors.New("name cannot be empty")
			}
			if len(input) > 50 {
				return errors.New("name is too long (max 50 characters)")
			}
			for _, char := range bannedCharacters {
				if strings.ContainsRune(input, char) {
					return fmt.Errorf("name cannot contain '%c'", char)
				}
			}
			return nil
		}

		namePrompt := promptui.Prompt{
			Label:    "Project Name",
			Validate: validateName,
		}

		name, err := namePrompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		// select minecraft version

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
			fmt.Printf("❌ Error fetching latest Minecraft version: %v\n", err)
			return
		}

		versionValidator := func(input string) error {
			if input == "" {
				return nil // allow empty input to use default version
			}
			// error if input does not match pattern like "1.20.1" or "1.19"
			split := strings.Split(input, ".")
			if len(split) < 2 || len(split) > 3 {
				return errors.New("version must be in format 'Major.Minor' or 'Major.Minor.Patch'")
			}

			// check that the inputted version is a real version
			for _, v := range allGameVersionsFlat {
				if v == input {
					return nil
				}
			}

			return errors.New("version not found")
		}

		versionPrompt := promptui.Prompt{
			Label:    "Game Version (1.20.1)",
			Validate: versionValidator,
		}

		inputGameVersion, err := versionPrompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		var gameVersion string
		if inputGameVersion == "" {
			gameVersion = "1.20.1"
		} else {
			gameVersion = inputGameVersion
		}

		fmt.Printf("Creating project '%s' for Minecraft %s...\n", name, gameVersion)
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
