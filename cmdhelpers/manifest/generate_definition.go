package manifest

import (
	"fmt"
	"log/slog"

	"github.com/thomaspoignant/go-feature-flag/cmd/cli/helper"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/manifest/model"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

func GenerateDefinitionFromFlags(flags map[string]flag.Flag) (
	map[string]model.FlagDefinition, error) {
	definitions := make(map[string]model.FlagDefinition)

	for k, v := range flags {
		v, ok := v.(*flag.InternalFlag)
		if !ok || v == nil {
			slog.Error(fmt.Sprintf("unexpected flag type for key %q, expected InternalFlag", k))
			continue
		}
		if v.Variations == nil {
			slog.Error(fmt.Sprintf("flag %s ignored: no variations provided", k))
			continue
		}
		flagType, err := helper.FlagTypeFromVariations(*v.Variations)
		if err != nil {
			return definitions, fmt.Errorf("invalid configuration for flag %s: %w", k, err)
		}

		metadata := v.Metadata
		if metadata == nil {
			m := make(map[string]any)
			metadata = &m
		}

		defaultValue, ok := (*metadata)["defaultValue"]
		if !ok {
			slog.Warn(fmt.Sprintf("flag %s ignored: no default value provided", k))
			continue
		}

		description, ok := (*metadata)["description"].(string)
		if !ok {
			description = ""
		}

		definition := model.FlagDefinition{
			FlagType:     flagType,
			DefaultValue: defaultValue,
			Description:  description,
		}
		definitions[k] = definition
	}

	return definitions, nil
}

func GenerateDefinitionFromInternalFlags(flags map[string]flag.InternalFlag) (
	map[string]model.FlagDefinition, error) {
	asFlag := make(map[string]flag.Flag, len(flags))
	for name := range flags {
		f := flags[name]
		asFlag[name] = &f
	}
	return GenerateDefinitionFromFlags(asFlag)
}
