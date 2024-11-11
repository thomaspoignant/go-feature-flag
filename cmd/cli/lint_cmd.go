package main

import (
	"github.com/spf13/cobra"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/linter"
)

var lintFlagFormat string

func init() {
	lintCmd := &cobra.Command{
		Use:   "lint <config_file>",
		Short: "Lint GO Feature Flag configuration file.",
		Long:  `Validate GO Feature Flag configuration file, it will return an error if your file is not valid.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return linter.RunLint(cmd, args, lintFlagFormat)
		},
	}
	lintCmd.Flags().StringVarP(&lintFlagFormat, "format", "f", "yaml", "Format of your input file (YAML, JSON or TOML)")
	rootCmd.AddCommand(lintCmd)
}
