package manifest

import (
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/cmd/cli/helper"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/manifest/model"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

func GenerateDefinition(flags map[string]flag.InternalFlag, logger fflog.FFLogger) (
	model.FlagManifest, error) {
	definitions := make(map[string]model.FlagDefinition)

	for k, v := range flags {
		flagType, err := helper.FlagTypeFromVariations(*v.Variations)
		if err != nil {
			return model.FlagManifest{}, fmt.Errorf("invalid configuration for flag %s: %w", k, err)
		}

		metadata := v.Metadata
		if metadata == nil {
			m := make(map[string]any)
			metadata = &m
		}

		defaultValue, ok := (*metadata)["defaultValue"]
		if !ok {
			logger.Error(fmt.Sprintf("flag %s ignored: no default value provided", k))
			continue
		}

		description, ok := (*metadata)["description"].(string)
		if !ok {
			description = ""
		}

		definition := model.FlagDefinition{
			DTO:          dto.ConvertInternalFlagToDto(v),
			FlagType:     flagType,
			DefaultValue: defaultValue,
			Description:  description,
		}
		definitions[k] = definition
	}

	return model.FlagManifest{Flags: definitions}, nil
}
