package cmd

import (
	"fmt"
	"os"
	"minepack/core/bisect"
	"github.com/spf13/cobra"
)

var linkBisectEnableCmd = &cobra.Command{
	Use:   "enable [slug]",
	Short: "manually enable a mod during bisection",
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
		err = state.ManualEnableMod(slug)
		if err != nil {
			fmt.Printf("failed to enable mod: %v\n", err)
			return
		}
		err = bisect.SaveBisectState(projectRoot, state)
		if err != nil {
			fmt.Printf("failed to save bisection state: %v\n", err)
			return
		}
		fmt.Printf("mod '%s' enabled\n", slug)
	},
}

func init() {
	linkBisectCmd.AddCommand(linkBisectEnableCmd)
}
