package main

import (
	"context"
	"os"

	"minepack/cmd"

	"github.com/charmbracelet/fang"
)

func main() {
	// Use fang.Execute with the root command from cmd package
	if err := fang.Execute(context.Background(), cmd.GetRootCmd()); err != nil {
		os.Exit(1)
	}
}
