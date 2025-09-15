package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const Version = "v6"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "minepack",
	Version: Version, // Set the version directly
	Short:   "a command line tool for managing minecraft modpacks",
	Long: `minepack is a command line tool that provides various commands for 
managing and processing minecraft modpacks.`,
}

// GetRootCmd returns the root command for use with fang
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.minepack.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
