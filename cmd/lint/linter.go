package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/thomaspoignant/go-feature-flag/internal/dto"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type Linter struct {
	InputFile   string
	InputFormat string
}

func (l *Linter) Lint() []error {
	dat, err := os.ReadFile(l.InputFile)
	if err != nil {
		return []error{err}
	}

	var flags map[string]dto.DTO
	switch strings.ToLower(l.InputFormat) {
	case "toml":
		err = toml.Unmarshal(dat, &flags)
	case "json":
		err = json.Unmarshal(dat, &flags)
	case "yaml":
		err = yaml.Unmarshal(dat, &flags)
	default:
		return []error{fmt.Errorf("%s: invalid input format: %s", l.InputFile, l.InputFormat)}
	}
	if err != nil {
		return []error{fmt.Errorf("%s: could not parse file: %w", l.InputFile, err)}
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
