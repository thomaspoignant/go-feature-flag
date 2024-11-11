package linter

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func RunLint(_ *cobra.Command, args []string, lintFlagFormat string) error {
	if len(args) <= 0 || args[0] == "" {
		return fmt.Errorf("missing configuration file location argument, please provide the location of the configuration file")
	}
	l := Linter{
		InputFile:   args[0],
		InputFormat: lintFlagFormat,
	}
	if errs := l.Lint(); errs != nil && len(errs) > 0 {
		for _, err := range errs {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		os.Exit(len(errs))
	}
	fmt.Println("Configuration is valid")
	return nil
}
