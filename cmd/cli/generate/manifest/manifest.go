package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/thomaspoignant/go-feature-flag/cmd/cli/helper"
	configHelper "github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
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
	var logs []string
	handler := slogHandlerForLogs(&logs)
	logx := fflog.FFLogger{
		LeveledLogger: slog.New(handler),
	}
	flags := make(map[string]flag.InternalFlag)
	for k, v := range m.dtos {
		flags[k] = dtoCore.ConvertDtoToInternalFlag(v)
	}
	output := helper.Output{}
	definitions, err := manifest.GenerateDefinition(flags, logx)
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

// captureHandler is a slog.Handler that appends each record's message to logs
// (no timestamps/levels in the string; CLI applies WARNING/INFO via helper.Output).
type captureHandler struct {
	logs *[]string
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	*h.logs = append(*h.logs, r.Message)
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *captureHandler) WithGroup(string) slog.Handler { return h }

// slogHandlerForLogs returns a slog.Handler that records log messages into logs.
func slogHandlerForLogs(logs *[]string) slog.Handler {
	return &captureHandler{logs: logs}
}
