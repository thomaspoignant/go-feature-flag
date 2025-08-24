package main

import (
	"github.com/spf13/cobra"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/evaluate"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/generate"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/helper"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/linter"
)

func main() {
	if err := initRootCmd().Execute(); err != nil {
		helper.PrintFatalAndExit(err)
	}
}

func initRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "go-feature-flag-cli",
		Short: "GO Feature Flag CLI tool",
		Long:  `A command line interface for GO Feature Flag operations.`,
	}
	rootCmd.AddCommand(evaluate.NewEvaluateCmd())
	rootCmd.AddCommand(linter.NewLintCmd())
	rootCmd.AddCommand(generate.NewGenerateCmd())
	return rootCmd
}
