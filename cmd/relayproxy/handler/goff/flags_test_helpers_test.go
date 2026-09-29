package controller_test

import (
	"encoding/json"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

// decodeJSON is a small helper to unmarshal a recorded response body in tests.
func decodeJSON(body []byte, v any) error {
	return json.Unmarshal(body, v)
}

// fakeFlagsetManager is a minimal service.FlagsetManager test double that always resolves to
// a single, preconfigured flagset regardless of the API key.
type fakeFlagsetManager struct {
	flagset *ffclient.GoFeatureFlag
	err     error
}

func (f *fakeFlagsetManager) FlagSet(_ string) (*ffclient.GoFeatureFlag, error) {
	return f.flagset, f.err
}

func (f *fakeFlagsetManager) FlagSetName(_ string) (string, error) {
	return "default", nil
}

func (f *fakeFlagsetManager) AllFlagSets() (map[string]*ffclient.GoFeatureFlag, error) {
	return map[string]*ffclient.GoFeatureFlag{"default": f.flagset}, nil
}

func (f *fakeFlagsetManager) Default() *ffclient.GoFeatureFlag {
	return f.flagset
}

func (f *fakeFlagsetManager) IsDefaultFlagSet() bool {
	return true
}

func (f *fakeFlagsetManager) Close() {}

func (f *fakeFlagsetManager) OnConfigChange(_ *config.Config) {}
