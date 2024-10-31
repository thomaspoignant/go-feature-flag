package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "go-feature-flag-cli",
	Short: "GO Feature Flag CLI tool",
	Long:  `A command line interface for GO Feature Flag operations including flag evaluation and linting.`,
}





