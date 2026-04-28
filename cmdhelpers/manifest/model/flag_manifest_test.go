package model_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/manifest/model"
)

func TestFlagManifest_MarshalJSON_includesSchema(t *testing.T) {
	m := model.FlagManifest{Flags: map[string]model.FlagDefinition{}}
	raw, err := json.MarshalIndent(m, "", "  ")
	require.NoError(t, err)
	require.Contains(t, string(raw), `"$schema": "`+model.DefaultFlagManifestSchemaURL+`"`)
}
