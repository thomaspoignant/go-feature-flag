package model

import (
	"encoding/json"

	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/manifest/model"
)

// DefaultFlagManifestSchemaURL is the JSON Schema URL emitted for every serialized
// flag manifest ($schema), per Open Feature CLI tooling.
const DefaultFlagManifestSchemaURL = "https://raw.githubusercontent.com/open-feature/cli/refs/heads/main/schema/v0/flag-manifest.json" //nolint: lll

// FlagManifest is the machine-readable manifest envelope for CLI and tooling.
type FlagManifest struct {
	Flags map[string]model.FlagDefinition `json:"flags"`
}

// MarshalJSON always includes "$schema" so manifest files validate against the
// published Open Feature CLI schema.
func (m FlagManifest) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Schema string                          `json:"$schema"`
		Flags  map[string]model.FlagDefinition `json:"flags"`
	}{
		Schema: DefaultFlagManifestSchemaURL,
		Flags:  m.Flags,
	})
}
