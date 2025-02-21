package linter

import (
	"fmt"

	"github.com/spf13/cobra"
)

var lintFlagFormat string

func NewLintCmd() *cobra.Command {
	lintCmd := &cobra.Command{
		Use:   "lint <config_file>",
		Short: "Lint GO Feature Flag configuration file.",
		Long:  `Validate GO Feature Flag configuration file, it will return an error if your file is not valid.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLint(cmd, args, lintFlagFormat)
		},
	}
	lintCmd.Flags().StringVarP(&lintFlagFormat, "format", "f", "yaml", "Format of your input file (YAML, JSON or TOML)")
	return lintCmd
}

func runLint(cmd *cobra.Command, args []string, lintFlagFormat string) error {
	if len(args) == 0 || args[0] == "" {
		return fmt.Errorf("missing configuration file location argument, " +
			"please provide the location of the configuration file")
	}
	l := Linter{
		InputFile:   args[0],
		InputFormat: lintFlagFormat,
	}
	if errs := l.Lint(); len(errs) > 0 {
		for _, err := range errs {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s\n", err)
		}
		cmd.SilenceUsage = true
		return fmt.Errorf("invalid GO Feature Flag configuration")
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), "Valid GO Feature Flag configuration")
	return err
}
