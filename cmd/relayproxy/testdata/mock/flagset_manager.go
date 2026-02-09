package mock

import (
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

// MockFlagsetManager is a mock implementation for testing error scenarios
type MockFlagsetManager struct {
	FlagSets            map[string]*ffclient.GoFeatureFlag
	DefaultFlagSet      *ffclient.GoFeatureFlag
	IsDefaultFlagSeItem bool
	GetFlagSetsErr      error
}

func (m *MockFlagsetManager) FlagSet(_ string) (*ffclient.GoFeatureFlag, error) {
	if len(m.FlagSets) == 0 {
		return nil, m.GetFlagSetsErr
	}

	if m.GetFlagSetsErr != nil {
		return nil, m.GetFlagSetsErr
	}

	keys := make([]string, 0, len(m.FlagSets))
	for key := range m.FlagSets {
		keys = append(keys, key)
	}

	return m.FlagSets[keys[0]], m.GetFlagSetsErr
}

func (m *MockFlagsetManager) FlagSetName(_ string) (string, error) {
	return "", nil
}

func (m *MockFlagsetManager) AllFlagSets() (map[string]*ffclient.GoFeatureFlag, error) {
	return m.FlagSets, m.GetFlagSetsErr
}

func (m *MockFlagsetManager) Default() *ffclient.GoFeatureFlag {
	return m.DefaultFlagSet
}

func (m *MockFlagsetManager) IsDefaultFlagSet() bool {
	return m.IsDefaultFlagSeItem
}

func (m *MockFlagsetManager) Close() {
	// nothing to do
}

func (m *MockFlagsetManager) OnConfigChange(_ *config.Config) {
	// nothing to do
}
