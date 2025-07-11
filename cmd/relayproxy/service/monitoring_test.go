package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"go.uber.org/zap"
)

func TestWithDefaultMode(t *testing.T) {
	t.Run("health", func(t *testing.T) {
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
	})

	t.Run("info", func(t *testing.T) {
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
	})

	t.Run("info with error", func(t *testing.T) {
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
	})

	t.Run("with nil config", func(t *testing.T) {
		manager, err := service.NewFlagsetManager(nil, zap.NewNop(), nil)
		assert.Error(t, err, "Expected error when config is nil")
		assert.Nil(t, manager, "Expected manager to be nil")
	})

	t.Run("is default flagset", func(t *testing.T) {
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
	})
}

func TestWithFlagsetsMode(t *testing.T) {
	t.Run("info", func(t *testing.T) {
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
	})

	t.Run("info multiple flagsets", func(t *testing.T) {
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
	})

	t.Run("is default flagset", func(t *testing.T) {
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
	})

	t.Run("empty flagsets", func(t *testing.T) {
		manager, err := service.NewFlagsetManager(&config.Config{
			FlagSets: []config.FlagSet{}, // Empty flagsets
		}, zap.NewNop(), nil)
		assert.Error(t, err, "Expected error due to no retrievers configured")
		assert.Nil(t, manager)
	})

	t.Run("all flagsets fail to initialize", func(t *testing.T) {
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
	})
}
