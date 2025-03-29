package main

import (
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/cmd/cli/helper"
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
	for key, flagDto := range flags {
		flag := flagDto.Convert()
		if err := flag.IsValid(); err != nil {
			errs = append(errs, fmt.Errorf("%s: invalid flag %s: %w", l.InputFile, key, err))
		}
	}

	return errs
}
