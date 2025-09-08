package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	myFlag           bool
	force            bool
	myFlagWithAValue string
	mode             string
)

// commandCmd represents the command command
var commandCmd = &cobra.Command{
	Use:   "command",
	Short: "Example command with various flags",
	Long: `This is an example command that demonstrates how to use different types of flags:
- Boolean flags (--myFlag, -f)
- String flags with values (--myFlagWithAValue, -m)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running 'command' subcommand!")

		if myFlag {
			fmt.Println("✓ myFlag is set")
		}

		if force {
			fmt.Println("✓ force flag (-f) is set")
		}

		if myFlagWithAValue != "" {
			fmt.Printf("✓ myFlagWithAValue: %s\n", myFlagWithAValue)
		}

		if len(mode) > 0 {
			fmt.Printf("✓ mode (-m): %s\n", mode)
		}

		if len(args) > 0 {
			fmt.Printf("Additional arguments: %v\n", args)
		}
	},
}

func init() {
	rootCmd.AddCommand(commandCmd)

	// Boolean flags
	commandCmd.Flags().BoolVar(&myFlag, "myFlag", false, "Example boolean flag")
	commandCmd.Flags().BoolVarP(&force, "force", "f", false, "Force operation")

	// String flags with values
	commandCmd.Flags().StringVar(&myFlagWithAValue, "myFlagWithAValue", "", "Example flag that takes a value")
	commandCmd.Flags().StringVarP(&mode, "mode", "m", "", "Set the mode (takes a value)")
}
