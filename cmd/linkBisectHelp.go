/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// linkBisectHelpCmd represents the bisect help command
var linkBisectHelpCmd = &cobra.Command{
	Use:   "help",
	Short: "detailed guide on using bisection to find problematic mods",
	Long:  `comprehensive help for using the bisection feature to systematically find problematic mods`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(`
🔍 minepack bisection guide

what is bisection?
bisection (also called binary search) is a systematic way to find which mod in your modpack 
is causing problems. instead of manually testing each mod one by one, bisection cuts the 
search space in half at each step, making it much faster.

when should you use bisection?
- your minecraft crashes on startup or during gameplay
- your game is laggy or has performance issues  
- you're experiencing weird bugs or glitches
- something broke after adding new mods
- any problem where you suspect a specific mod but don't know which one

how does it work?
1. start with all your mods enabled and confirm the problem exists
2. disable half of your mods and test again
3. if the problem is gone: the issue is in the disabled mods (search those)
4. if the problem persists: the issue is in the enabled mods (search those)
5. repeat until you find the problematic mod

step-by-step usage:

1. start a bisection:
   minepack link bisect start
   
   this will:
   - let you choose which linked instance to use
   - create a bisect.mp.yaml file to track progress
   - give you initial instructions

2. test your current setup:
   go to minecraft and reproduce your issue to confirm it exists

3. continue the bisection:
   minepack link bisect next
   
   this will:
   - disable about half your mods (considering dependencies)
   - ask you to test again
   - record whether the issue still occurs

4. keep testing:
   repeat step 3, each time telling minepack whether the issue occurred or not.
   with each step, the number of candidate mods gets smaller.

5. finish when done:
   minepack link bisect finish
   
   this will:
   - show you which mod(s) are problematic
   - restore all your mods
   - clean up the bisection data

other useful commands:
- minepack link bisect previous  -- go back one step if you made a mistake
- minepack link bisect finish    -- stop early and restore everything

tips for success:
- be consistent in your testing (same world, same actions)
- if unsure about a test result, err on the side of "bad" (issue occurred)
- don't worry about dependencies - minepack handles them automatically

example session:
$ minepack link bisect start
  # test with all mods, confirm crash exists
$ minepack link bisect next
  # disabled 15/30 mods, crash gone - problem is in disabled mods
$ minepack link bisect next  
  # disabled 7/15 suspicious mods, crash still gone - problem is in other 8
$ minepack link bisect next
  # disabled 4/8 remaining mods, crash is back - problem is in these 4
$ minepack link bisect next
  # disabled 2/4 mods, crash gone - problem is in the other 2
$ minepack link bisect next
  # disabled 1/2 mods, crash persists - found the problematic mod!
$ minepack link bisect finish
  # shows: "problematic mod: badmod" and restores everything

troubleshooting:
- if minecraft won't start: some dependencies might be missing, this is normal
- if you get confused: use 'minepack link bisect previous' to go back
- if you want to stop: use 'minepack link bisect finish' to restore everything

happy debugging!
`)
	},
}

func init() {
	linkBisectCmd.AddCommand(linkBisectHelpCmd)
}
