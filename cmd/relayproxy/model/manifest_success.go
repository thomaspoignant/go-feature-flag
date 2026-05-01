package model

import "github.com/thomaspoignant/go-feature-flag/cmdhelpers/manifest/model"

// FlagManifest is the machine-readable manifest envelope for CLI and tooling.
type ManifestSuccessResponse struct {
	Flags []ManifestDefinitionSuccessResponse `json:"flags"`
}

type ManifestDefinitionSuccessResponse struct {
	Key          string         `json:"key"`
	FlagType     model.FlagType `json:"type"`
	DefaultValue any            `json:"defaultValue"`
	Description  string         `json:"description"`
}
