/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Test prompt features",
	Long:  `A command to test prompt features.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("prompt called")

		validate := func(input string) error {
			_, err := strconv.ParseFloat(input, 64)
			if err != nil {
				return errors.New("invalid number")
			}
			return nil
		}

		var number string
		numberForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Number").
					Description("Enter a number").
					Value(&number).
					Validate(validate),
			),
		)

		err := numberForm.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		fmt.Printf("You choose %q\n", number)

		var day string
		dayForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select Day").
					Description("Choose a day of the week").
					Options(
						huh.NewOption("Monday", "Monday"),
						huh.NewOption("Tuesday", "Tuesday"),
						huh.NewOption("Wednesday", "Wednesday"),
						huh.NewOption("Thursday", "Thursday"),
						huh.NewOption("Friday", "Friday"),
						huh.NewOption("Saturday", "Saturday"),
						huh.NewOption("Sunday", "Sunday"),
					).
					Value(&day),
			),
		)

		err = dayForm.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		fmt.Printf("You choose %q\n", day)
	},
}

func init() {
	testCmd.AddCommand(promptCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// promptCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// promptCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
