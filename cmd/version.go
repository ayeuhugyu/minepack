package cmd

import (
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage project versions",
	Long:  `Manage project versions including setting format and updating version numbers`,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
