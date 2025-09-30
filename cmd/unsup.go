package cmd

import (
	"fmt"
	"minepack/util"

	"github.com/spf13/cobra"
)

// unsupCmd represents the unsup command
var unsupCmd = &cobra.Command{
	Use:   "unsup",
	Short: "manage unsup manifest data",
	Long: `allows you to manage and export unsup's manifest format. this command can also be used to create flavor groups n stuff.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(util.FormatSuccess("hi"))
	},
}

func init() {
	rootCmd.AddCommand(unsupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// unsupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// unsupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
