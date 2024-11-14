package main

import (
	"log"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("error executing command: %v\n", err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "go-feature-flag-cli",
	Short: "GO Feature Flag CLI tool",
	Long:  `A command line interface for GO Feature Flag operations.`,
}
