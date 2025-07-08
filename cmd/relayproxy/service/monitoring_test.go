package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"go.uber.org/zap"
)

// Test the NewMonitoring function
func TestNewMonitoring(t *testing.T) {
	manager, err := service.NewFlagsetManager(&config.Config{
		CommonFlagSet: config.CommonFlagSet{
			PollingInterval: 60000, // 1 minute in milliseconds
			FileFormat:      "yaml",
			Retrievers: &[]config.RetrieverConf{
				{Kind: "file", Path: "../../../testdata/flag-config.yaml"},
			},
		},
	}, zap.NewNop(), nil)
	require.NoError(t, err)
	require.NotNil(t, manager)

	monitoring := service.NewMonitoring(manager)
	assert.NotNil(t, monitoring, "Expected monitoring to not be nil")
}

// Test the Health function of monitoringImpl
func TestMonitoringImpl_Health(t *testing.T) {
	manager, err := service.NewFlagsetManager(&config.Config{
		CommonFlagSet: config.CommonFlagSet{
			PollingInterval: 60000, // 1 minute in milliseconds
			FileFormat:      "yaml",
			Retrievers: &[]config.RetrieverConf{
				{Kind: "file", Path: "../../../testdata/flag-config.yaml"},
			},
		},
	}, zap.NewNop(), nil)
	require.NoError(t, err)
	require.NotNil(t, manager)

	monitoring := service.NewMonitoring(manager)
	health := monitoring.Health()
	assert.True(t, health.Initialized, "Expected initialized to be true, got false")
}

// Test the Info function of monitoringImpl with default flagset
func TestMonitoringImpl_Info_WithDefaultFlagset(t *testing.T) {
	manager, err := service.NewFlagsetManager(&config.Config{
		CommonFlagSet: config.CommonFlagSet{
			PollingInterval: 60000, // 1 minute in milliseconds
			FileFormat:      "yaml",
			Retrievers: &[]config.RetrieverConf{
				{Kind: "file", Path: "../../../testdata/flag-config.yaml"},
			},
		},
	}, zap.NewNop(), nil)
	require.NoError(t, err)
	require.NotNil(t, manager)

	monitoring := service.NewMonitoring(manager)
	info, err := monitoring.Info()
	assert.NoError(t, err, "Expected no error")
	assert.NotNil(t, info.LatestCacheRefresh, "Expected LatestCacheRefresh to not be nil")
	assert.False(
		t,
		info.LatestCacheRefresh.IsZero(),
		"Expected LatestCacheRefresh to not be zero, got zero",
	)
}

// Test the Info function with flagsets mode
func TestMonitoringImpl_Info_WithFlagsets(t *testing.T) {
	manager, err := service.NewFlagsetManager(&config.Config{
		FlagSets: []config.FlagSet{
			{
				Name: "test-flagset",
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000, // 1 minute in milliseconds
					FileFormat:      "yaml",
					Retrievers: &[]config.RetrieverConf{
						{Kind: "file", Path: "../../../testdata/flag-config.yaml"},
					},
				},
				ApiKeys: []string{"test-api-key"},
			},
		},
	}, zap.NewNop(), nil)
	require.NoError(t, err)
	require.NotNil(t, manager)

	monitoring := service.NewMonitoring(manager)
	info, err := monitoring.Info()
	assert.NoError(t, err, "Expected no error")
	assert.NotEmpty(t, info.Flagsets, "Expected flagsets to not be empty")
	assert.Contains(t, info.Flagsets, "test-flagset", "Expected test-flagset to be in flagsets")
}

// Test the Info function with multiple flagsets
func TestMonitoringImpl_Info_WithMultipleFlagsets(t *testing.T) {
	manager, err := service.NewFlagsetManager(&config.Config{
		FlagSets: []config.FlagSet{
			{
				Name: "flagset1",
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000, // 1 minute in milliseconds
					FileFormat:      "yaml",
					Retrievers: &[]config.RetrieverConf{
						{Kind: "file", Path: "../../../testdata/flag-config.yaml"},
					},
				},
				ApiKeys: []string{"api-key-1"},
			},
			{
				Name: "flagset2",
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000, // 1 minute in milliseconds
					FileFormat:      "yaml",
					Retrievers: &[]config.RetrieverConf{
						{Kind: "file", Path: "../../../testdata/flag-config-2nd-file.yaml"},
					},
				},
				ApiKeys: []string{"api-key-2"},
			},
		},
	}, zap.NewNop(), nil)
	require.NoError(t, err)
	require.NotNil(t, manager)

	monitoring := service.NewMonitoring(manager)
	info, err := monitoring.Info()
	assert.NoError(t, err, "Expected no error")
	assert.Len(t, info.Flagsets, 2, "Expected 2 flagsets")
	assert.Contains(t, info.Flagsets, "flagset1", "Expected flagset1 to be in flagsets")
	assert.Contains(t, info.Flagsets, "flagset2", "Expected flagset2 to be in flagsets")
}

// Test the Info function with GetFlagSets error (invalid config)
func TestMonitoringImpl_Info_WithGetFlagSetsError(t *testing.T) {
	manager, err := service.NewFlagsetManager(&config.Config{
		CommonFlagSet: config.CommonFlagSet{
			PollingInterval: 60000, // 1 minute in milliseconds
			FileFormat:      "yaml",
			Retrievers: &[]config.RetrieverConf{
				{Kind: "file", Path: "../../../testdata/non-existent-file.yaml"},
			},
		},
	}, zap.NewNop(), nil)
	assert.Error(t, err, "Expected error due to invalid file path")
	assert.Nil(t, manager)
}

// Test the Info function with default flagset mode but nil default flagset
func TestMonitoringImpl_Info_WithDefaultFlagsetModeButNilDefault(t *testing.T) {
	// Create a manager with nil config to trigger error
	manager, err := service.NewFlagsetManager(nil, zap.NewNop(), nil)
	assert.Error(t, err, "Expected error when config is nil")
	assert.Nil(t, manager, "Expected manager to be nil")
}

// Test the IsDefaultFlagSet function
func TestMonitoringImpl_IsDefaultFlagSet(t *testing.T) {
	// Test default mode
	managerDefault, err := service.NewFlagsetManager(&config.Config{
		CommonFlagSet: config.CommonFlagSet{
			PollingInterval: 60000, // 1 minute in milliseconds
			FileFormat:      "yaml",
			Retrievers: &[]config.RetrieverConf{
				{Kind: "file", Path: "../../../testdata/flag-config.yaml"},
			},
		},
	}, zap.NewNop(), nil)
	require.NoError(t, err)
	assert.True(t, managerDefault.IsDefaultFlagSet())

	// Test flagsets mode
	managerFlagsets, err := service.NewFlagsetManager(&config.Config{
		FlagSets: []config.FlagSet{
			{
				Name: "test-flagset",
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000, // 1 minute in milliseconds
					FileFormat:      "yaml",
					Retrievers: &[]config.RetrieverConf{
						{Kind: "file", Path: "../../../testdata/flag-config.yaml"},
					},
				},
				ApiKeys: []string{"test-api-key"},
			},
		},
	}, zap.NewNop(), nil)
	require.NoError(t, err)
	assert.False(t, managerFlagsets.IsDefaultFlagSet())
}

// Test error case when no flagsets are configured
func TestMonitoringImpl_Info_NoFlagsetsConfigured(t *testing.T) {
	manager, err := service.NewFlagsetManager(&config.Config{
		FlagSets: []config.FlagSet{}, // Empty flagsets
	}, zap.NewNop(), nil)
	assert.Error(t, err, "Expected error due to no retrievers configured")
	assert.Nil(t, manager)
}

// Test error case when all flagsets fail to initialize
func TestMonitoringImpl_Info_AllFlagsetsFail(t *testing.T) {
	manager, err := service.NewFlagsetManager(&config.Config{
		FlagSets: []config.FlagSet{
			{
				Name: "invalid-flagset",
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000, // 1 minute in milliseconds
					FileFormat:      "yaml",
					Retrievers: &[]config.RetrieverConf{
						{Kind: "file", Path: "../../../testdata/non-existent-file.yaml"},
					},
				},
				ApiKeys: []string{"api-key-1"},
			},
		},
	}, zap.NewNop(), nil)
	assert.Error(t, err, "Expected error due to invalid file path")
	assert.Nil(t, manager)
}
