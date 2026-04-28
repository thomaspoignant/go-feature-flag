package manifest

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/thomaspoignant/go-feature-flag/cmd/cli/helper"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/manifest"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/manifest/model"
	"github.com/thomaspoignant/go-feature-flag/model/dto"
	dtoCore "github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
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
	var logs []string
	logx := fflog.ConvertToFFLogger(log.New(&sliceWriter{logs: &logs}, "", log.LstdFlags))
	flags := make(map[string]flag.InternalFlag)
	for k, v := range m.dtos {
		flags[k] = dtoCore.ConvertDtoToInternalFlag(v)
	}
	output := helper.Output{}
	definitions, err := manifest.GenerateDefinition(flags, *logx)
	if len(logs) > 0 {
		output.Add(strings.Join(logs, "\n"), helper.WarnLevel)
	}
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

// sliceWriter is an io.Writer that captures each Write call as an entry
// in the underlying string slice. It is used to redirect a *log.Logger's
// output into an in-memory array instead of stdout/stderr.
type sliceWriter struct {
	logs *[]string
}

func (w *sliceWriter) Write(p []byte) (int, error) {
	*w.logs = append(*w.logs, strings.TrimRight(string(p), "\n"))
	return len(p), nil
}
