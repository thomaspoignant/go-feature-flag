package linter

import (
	"fmt"

	helper "github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
)

type Linter struct {
	InputFile   string
	InputFormat string
}

func (l *Linter) Lint() []error {
	flags, err := helper.LoadConfigFile(
		l.InputFile,
		l.InputFormat,
		helper.ConfigFileDefaultLocations,
	)
	if err != nil {
		return []error{err}
	}
	errs := make([]error, 0)
	for _, flagErr := range helper.ValidateFlags(flags) {
		errs = append(errs, fmt.Errorf("%s: invalid flag %s: %w", l.InputFile, flagErr.Flag, flagErr.Err))
	}
	return errs
}
