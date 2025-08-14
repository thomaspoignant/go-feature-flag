package mock

import ffclient "github.com/thomaspoignant/go-feature-flag"

// MockFlagsetManager is a mock implementation for testing error scenarios
type MockFlagsetManager struct {
	FlagSets            map[string]*ffclient.GoFeatureFlag
	DefaultFlagSet      *ffclient.GoFeatureFlag
	IsDefaultFlagSeItem bool
	GetFlagSetsErr      error
}

func (m *MockFlagsetManager) GetFlagSet(apiKey string) (*ffclient.GoFeatureFlag, error) {
	return nil, nil
}

func (m *MockFlagsetManager) GetFlagSetName(apiKey string) (string, error) {
	return "", nil
}

func (m *MockFlagsetManager) GetFlagSets() (map[string]*ffclient.GoFeatureFlag, error) {
	return m.FlagSets, m.GetFlagSetsErr
}

func (m *MockFlagsetManager) GetDefaultFlagSet() *ffclient.GoFeatureFlag {
	return m.DefaultFlagSet
}

func (m *MockFlagsetManager) IsDefaultFlagSet() bool {
	return m.IsDefaultFlagSeItem
}

func (m *MockFlagsetManager) Close() {
	// nothing to do
}
