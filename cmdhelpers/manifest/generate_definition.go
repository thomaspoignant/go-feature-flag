package manifest

import (
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
			logger.Error("invalid configuration for flag %s: %s", k, err.Error())
			continue
		}

		metadata := v.Metadata
		if metadata == nil {
			m := make(map[string]any)
			metadata = &m
		}

		defaultValue, ok := (*metadata)["defaultValue"]
		if !ok {
			logger.Error("flag %s ignored: no default value provided", k)
			continue
		}

		description, ok := (*metadata)["description"].(string)
		if !ok {
			description = ""
		}

		defintion := model.FlagDefinition{
			DTO:          dto.ConvertInternalFlagToDto(v),
			FlagType:     model.FlagType(flagType),
			DefaultValue: defaultValue,
			Description:  description,
		}
		definitions[k] = defintion
	}

	return model.FlagManifest{Flags: definitions}, nil
}
