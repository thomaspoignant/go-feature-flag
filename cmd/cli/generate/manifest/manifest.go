package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/thomaspoignant/go-feature-flag/cmd/cli/generate/manifest/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/helper"
	configHelper "github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/manifest"
	"github.com/thomaspoignant/go-feature-flag/model/dto"
	dtoCore "github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

func NewManifest(configFile, configFormat, flagManifestDestination string) (Manifest, error) {
	if flagManifestDestination == "" {
		return Manifest{}, fmt.Errorf("--flag_manifest_destination is mandatory")
	}
	flagDTOs, err := configHelper.LoadConfigFile(
		configFile,
		configFormat,
		configHelper.ConfigFileDefaultLocations,
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
	var captured []capturedRecord
	flags := make(map[string]flag.InternalFlag)
	for k, v := range m.dtos {
		flags[k] = dtoCore.ConvertDtoToInternalFlag(v)
	}
	output := helper.Output{}
	prev := slog.Default()
	slog.SetDefault(slog.New(slogHandlerForRecords(&captured)))
	defer slog.SetDefault(prev)

	definitions, err := manifest.GenerateDefinitionFromInternalFlags(flags)
	for _, rec := range captured {
		output.Add(rec.message, slogLevelToOutputLevel(rec.level))
	}
	if err != nil {
		return output, err
	}

	manifest := model.FlagManifest{
		Flags: definitions,
	}
	definitionsJSON, err := m.toJSON(manifest)
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

type capturedRecord struct {
	level   slog.Level
	message string
}

// captureHandler records each slog record for replay into helper.Output.
type captureHandler struct {
	records *[]capturedRecord
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	*h.records = append(*h.records, capturedRecord{level: r.Level, message: r.Message})
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *captureHandler) WithGroup(string) slog.Handler { return h }

func slogHandlerForRecords(records *[]capturedRecord) slog.Handler {
	return &captureHandler{records: records}
}

func slogLevelToOutputLevel(l slog.Level) helper.Level {
	switch {
	case l >= slog.LevelError:
		return helper.ErrorLevel
	case l >= slog.LevelWarn:
		return helper.WarnLevel
	case l >= slog.LevelInfo:
		return helper.InfoLevel
	default:
		return helper.DefaultLevel
	}
}
