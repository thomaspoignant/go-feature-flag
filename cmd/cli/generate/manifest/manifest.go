package manifest

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/thomaspoignant/go-feature-flag/cmd/cli/generate/manifest/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/helper"
	"github.com/thomaspoignant/go-feature-flag/model/dto"
	dtoCore "github.com/thomaspoignant/go-feature-flag/modules/core/dto"
)

func NewManifest(configFile, configFormat, flagManifestDestination string) (Manifest, error) {
	if flagManifestDestination == "" {
		return Manifest{}, fmt.Errorf("--flag_manifest_destination is mandatory")
	}
	flagDTOs, err := helper.LoadConfigFile(
		configFile,
		configFormat,
		helper.ConfigFileDefaultLocations,
	)
	if err != nil {
		return Manifest{}, err
	}
	return Manifest{
		dtos:                    flagDTOs,
		flagManifestDestination: flagManifestDestination,
	}, nil
}

type Manifest struct {
	dtos                    map[string]dto.DTO
	flagManifestDestination string
}

// Generate the manifest file
func (m *Manifest) Generate() (helper.Output, error) {
	definitions, output, err := m.generateDefinitions(m.dtos)
	if err != nil {
		return output, err
	}
	definitionsJSON, err := m.toJSON(definitions)
	if err != nil {
		return output, err
	}
	// if we have a destination we write the JSON in the file
	err = m.storeManifest(definitionsJSON)
	if err != nil {
		return output, err
	}
	return output.Add("🎉 Manifest has been created", helper.InfoLevel), nil
}

// generateDefinitions will generate the definitions from the flagDTOs
func (m *Manifest) generateDefinitions(flagDTOs map[string]dto.DTO) (
	model.FlagManifest, helper.Output, error) {
	definitions := make(map[string]model.FlagDefinition)
	output := helper.Output{}
	for flagKey, flagDTO := range flagDTOs {
		flag := dtoCore.ConvertDtoToInternalFlag(flagDTO)
		flagType, err := helper.FlagTypeFromVariations(flag.GetVariations())
		if err != nil {
			return model.FlagManifest{}, output,
				fmt.Errorf("invalid configuration for flag %s: %s", flagKey, err.Error())
		}

		metadata := flag.GetMetadata()
		description, ok := metadata["description"].(string)
		if !ok {
			description = ""
		}

		defaultValue, ok := metadata["defaultValue"]
		if !ok {
			output.Add(
				fmt.Sprintf("flag %s ignored: no default value provided", flagKey),
				helper.WarnLevel,
			)
			continue
		}
		definitions[flagKey] = model.FlagDefinition{
			FlagType:     flagType,
			DefaultValue: defaultValue,
			Description:  description,
		}
	}
	return model.FlagManifest{Flags: definitions}, output, nil
}

// toJSON will convert the definitions to a JSON string
func (m *Manifest) toJSON(manifest model.FlagManifest) (string, error) {
	definitionsJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", err
	}
	return string(definitionsJSON), nil
}

func (m *Manifest) storeManifest(definitionsJSON string) error {
	return os.WriteFile(m.flagManifestDestination, []byte(definitionsJSON), 0600)
}
