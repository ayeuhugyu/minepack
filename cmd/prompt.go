/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/manifoldco/promptui"
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

		prompt := promptui.Prompt{
			Label:    "Number",
			Validate: validate,
		}

		result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		fmt.Printf("You choose %q\n", result)

		prompt2 := promptui.Select{
			Label: "Select Day",
			Items: []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday",
				"Saturday", "Sunday"},
		}

		_, result2, err2 := prompt2.Run()

		if err2 != nil {
			fmt.Printf("Prompt failed %v\n", err2)
			return
		}

		fmt.Printf("You choose %q\n", result2)
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
