package cmd

import (
	"fmt"
	"minepack/core/bisect"
	"os"

	"github.com/spf13/cobra"
)

var linkBisectDisableCmd = &cobra.Command{
	Use:   "disable [slug]",
	Short: "manually disable a mod during bisection",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Println("could not determine project root")
			return
		}
		state, err := bisect.LoadBisectState(projectRoot)
		if err != nil {
			fmt.Println("no active bisection found")
			return
		}
		slug := args[0]
		err = state.ManualDisableMod(slug)
		if err != nil {
			fmt.Printf("failed to disable mod: %v\n", err)
			return
		}
		err = bisect.SaveBisectState(projectRoot, state)
		if err != nil {
			fmt.Printf("failed to save bisection state: %v\n", err)
			return
		}
		fmt.Printf("mod '%s' disabled\n", slug)
	},
}

func init() {
	linkBisectCmd.AddCommand(linkBisectDisableCmd)
}
