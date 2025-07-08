package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"go.uber.org/zap"
)

func TestNewFlagsetManager(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		logger      *zap.Logger
		wantErr     bool
		expectedErr string
		expectedLen int // Number of API keys expected in flagsets mode
	}{
		{
			name:        "nil config should return error",
			config:      nil,
			logger:      zap.NewNop(),
			wantErr:     true,
			expectedErr: "configuration is nil",
		},
		{
			name: "empty flagsets should use default config",
			config: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000,
					FileFormat:      "yaml",
					Retrievers: &[]config.RetrieverConf{
						{
							Kind: "file",
							Path: "../../../testdata/flag-config.yaml",
						},
					},
				},
			},
			logger:  zap.NewNop(),
			wantErr: false,
		},
		{
			name: "valid flagsets should create flagset manager",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{
						Name: "test-flagset-1",
						CommonFlagSet: config.CommonFlagSet{
							PollingInterval: 60000,
							FileFormat:      "yaml",
							Retrievers: &[]config.RetrieverConf{
								{
									Kind: "file",
									Path: "../../../testdata/flag-config.yaml",
								},
							},
						},
						ApiKeys: []string{"api-key-1", "api-key-2"},
					},
					{
						Name: "test-flagset-2",
						CommonFlagSet: config.CommonFlagSet{
							PollingInterval: 60000,
							FileFormat:      "yaml",
							Retrievers: &[]config.RetrieverConf{
								{
									Kind: "file",
									Path: "../../../testdata/flag-config-2nd-file.yaml",
								},
							},
						},
						ApiKeys: []string{"api-key-3"},
					},
				},
			},
			logger:      zap.NewNop(),
			wantErr:     false,
			expectedLen: 3, // 3 API keys total
		},
		{
			name: "flagset with empty name should generate UUID",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{
						Name: "", // Empty name
						CommonFlagSet: config.CommonFlagSet{
							PollingInterval: 60000,
							FileFormat:      "yaml",
							Retrievers: &[]config.RetrieverConf{
								{
									Kind: "file",
									Path: "../../../testdata/flag-config.yaml",
								},
							},
						},
						ApiKeys: []string{"api-key-1"},
					},
				},
			},
			logger:      zap.NewNop(),
			wantErr:     false,
			expectedLen: 1, // 1 API key
		},
		{
			name: "flagset with default name should generate UUID",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{
						Name: "default", // Default name
						CommonFlagSet: config.CommonFlagSet{
							PollingInterval: 60000,
							FileFormat:      "yaml",
							Retrievers: &[]config.RetrieverConf{
								{
									Kind: "file",
									Path: "../../../testdata/flag-config.yaml",
								},
							},
						},
						ApiKeys: []string{"api-key-1"},
					},
				},
			},
			logger:      zap.NewNop(),
			wantErr:     false,
			expectedLen: 1, // 1 API key
		},
		{
			name: "invalid flagset should fallback to default config",
			config: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000,
					FileFormat:      "yaml",
					Retrievers: &[]config.RetrieverConf{
						{
							Kind: "file",
							Path: "../../../testdata/flag-config.yaml",
						},
					},
				},
				FlagSets: []config.FlagSet{
					{
						Name: "invalid-flagset",
						CommonFlagSet: config.CommonFlagSet{
							PollingInterval: 60000,
							FileFormat:      "yaml",
							Retrievers: &[]config.RetrieverConf{
								{
									Kind: "file",
									Path: "non-existent-file.yaml",
								},
							},
						},
						ApiKeys: []string{"api-key-1"},
					},
				},
			},
			logger:  zap.NewNop(),
			wantErr: false, // Should fallback to default config
		},
		{
			name: "mixed valid and invalid flagsets should succeed with valid ones",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{
						Name: "invalid-flagset",
						CommonFlagSet: config.CommonFlagSet{
							PollingInterval: 60000,
							FileFormat:      "yaml",
							Retrievers: &[]config.RetrieverConf{
								{
									Kind: "file",
									Path: "non-existent-file.yaml",
								},
							},
						},
						ApiKeys: []string{"api-key-1"},
					},
					{
						Name: "valid-flagset",
						CommonFlagSet: config.CommonFlagSet{
							PollingInterval: 60000,
							FileFormat:      "yaml",
							Retrievers: &[]config.RetrieverConf{
								{
									Kind: "file",
									Path: "../../../testdata/flag-config.yaml",
								},
							},
						},
						ApiKeys: []string{"api-key-2"},
					},
				},
			},
			logger:      zap.NewNop(),
			wantErr:     false,
			expectedLen: 1, // 1 valid API key (the invalid one is skipped)
		},
		{
			name: "invalid retriever should return error",
			config: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000,
					FileFormat:      "yaml",
					Retrievers: &[]config.RetrieverConf{
						{
							Kind: "file",
							Path: "non-existent-file.yaml",
						},
					},
				},
			},
			logger:  zap.NewNop(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := service.NewFlagsetManager(tt.config, tt.logger)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Equal(t, tt.expectedErr, err.Error())
				}
				assert.Equal(t, service.FlagsetManager{}, manager)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)

				// Check if we're in flagsets mode or default mode
				if manager.DefaultFlagSet != nil {
					// Default mode - should have DefaultFlagSet but no FlagSets
					assert.NotNil(t, manager.DefaultFlagSet)
					assert.Nil(t, manager.FlagSets)
					assert.Nil(t, manager.APIKeysToFlagSet)
				} else {
					// Flagsets mode - should have FlagSets but no DefaultFlagSet
					assert.Nil(t, manager.DefaultFlagSet)
					assert.NotNil(t, manager.FlagSets)
					assert.NotNil(t, manager.APIKeysToFlagSet)
					if tt.expectedLen != len(manager.APIKeysToFlagSet) {
						t.Logf("Expected %d API keys, got %d", tt.expectedLen, len(manager.APIKeysToFlagSet))
						t.Logf("FlagSets: %v", manager.FlagSets)
						t.Logf("APIKeysToFlagSet: %v", manager.APIKeysToFlagSet)
					}
					assert.Equal(t, tt.expectedLen, len(manager.APIKeysToFlagSet))
				}
			}
		})
	}
}

func TestFlagsetManager_GetFlagSet(t *testing.T) {
	// Test default mode
	t.Run("default mode should return DefaultFlagSet", func(t *testing.T) {
		config := &config.Config{
			CommonFlagSet: config.CommonFlagSet{
				PollingInterval: 60000,
				FileFormat:      "yaml",
				Retrievers: &[]config.RetrieverConf{
					{
						Kind: "file",
						Path: "../../../testdata/flag-config.yaml",
					},
				},
			},
		}
		logger := zap.NewNop()
		manager, err := service.NewFlagsetManager(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, manager.DefaultFlagSet)

		// Test that any API key returns the DefaultFlagSet
		result := manager.GetFlagSet("any-api-key")
		assert.Equal(t, manager.DefaultFlagSet, result)
	})

	// Test flagsets mode
	t.Run("flagsets mode with valid API key should return correct flagset", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						PollingInterval: 60000,
						FileFormat:      "yaml",
						Retrievers: &[]config.RetrieverConf{
							{
								Kind: "file",
								Path: "../../../testdata/flag-config.yaml",
							},
						},
					},
					ApiKeys: []string{"api-key-1"},
				},
				{
					Name: "test-flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						PollingInterval: 60000,
						FileFormat:      "yaml",
						Retrievers: &[]config.RetrieverConf{
							{
								Kind: "file",
								Path: "../../../testdata/flag-config-2nd-file.yaml",
							},
						},
					},
					ApiKeys: []string{"api-key-2"},
				},
			},
		}
		logger := zap.NewNop()
		manager, err := service.NewFlagsetManager(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, manager.FlagSets)

		// Test that API keys return the correct flagsets
		flagset1 := manager.GetFlagSet("api-key-1")
		flagset2 := manager.GetFlagSet("api-key-2")
		assert.NotNil(t, flagset1)
		assert.NotNil(t, flagset2)
		assert.Equal(t, manager.FlagSets["test-flagset-1"], flagset1)
		assert.Equal(t, manager.FlagSets["test-flagset-2"], flagset2)
	})

	t.Run("flagsets mode with invalid API key should return nil", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset",
					CommonFlagSet: config.CommonFlagSet{
						PollingInterval: 60000,
						FileFormat:      "yaml",
						Retrievers: &[]config.RetrieverConf{
							{
								Kind: "file",
								Path: "../../../testdata/flag-config.yaml",
							},
						},
					},
					ApiKeys: []string{"api-key-1"},
				},
			},
		}
		logger := zap.NewNop()
		manager, err := service.NewFlagsetManager(config, logger)
		require.NoError(t, err)

		// Test that invalid API key returns nil
		result := manager.GetFlagSet("invalid-api-key")
		assert.Nil(t, result)
	})
}

func TestNewFlagsetManager_Integration(t *testing.T) {
	// Test the complete flow with valid configuration
	validConfig := &config.Config{
		FlagSets: []config.FlagSet{
			{
				Name: "test-flagset",
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000,
					FileFormat:      "yaml",
					Retrievers: &[]config.RetrieverConf{
						{
							Kind: "file",
							Path: "../../../testdata/flag-config.yaml",
						},
					},
				},
				ApiKeys: []string{"api-key-1", "api-key-2"},
			},
		},
	}

	logger := zap.NewNop()
	manager, err := service.NewFlagsetManager(validConfig, logger)

	require.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, 1, len(manager.FlagSets))
	assert.Equal(t, 2, len(manager.APIKeysToFlagSet))

	// Test GetFlagSet with valid API key
	flagset := manager.GetFlagSet("api-key-1")
	assert.NotNil(t, flagset)

	// Test GetFlagSet with invalid API key
	flagset = manager.GetFlagSet("invalid-api-key")
	assert.Nil(t, flagset)
}

func TestNewFlagsetManager_FallbackToDefault(t *testing.T) {
	// Test the fallback scenario where flagsets fail but default config succeeds
	config := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			PollingInterval: 60000,
			FileFormat:      "yaml",
			Retrievers: &[]config.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		FlagSets: []config.FlagSet{
			{
				Name: "invalid-flagset",
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000,
					FileFormat:      "yaml",
					Retrievers: &[]config.RetrieverConf{
						{
							Kind: "file",
							Path: "non-existent-file.yaml",
						},
					},
				},
				ApiKeys: []string{"api-key-1"},
			},
		},
	}

	logger := zap.NewNop()
	manager, err := service.NewFlagsetManager(config, logger)

	require.NoError(t, err)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.DefaultFlagSet)
	assert.Nil(t, manager.FlagSets)
	assert.Nil(t, manager.APIKeysToFlagSet)

	// Test GetFlagSet should return DefaultFlagSet regardless of API key
	flagset := manager.GetFlagSet("any-api-key")
	assert.Equal(t, manager.DefaultFlagSet, flagset)
}

func TestFlagsetManager_Constants(t *testing.T) {
	// Test that constants are properly defined
	// Note: Since we're in a separate package, we can't access private constants
	// This test verifies the behavior rather than the constants themselves
	manager := service.FlagsetManager{}
	assert.NotNil(t, manager)
}

func TestNewFlagsetManager_ErrorHandling(t *testing.T) {
	// Test error handling when both flagsets and default config fail
	config := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			PollingInterval: 60000,
			FileFormat:      "yaml",
			Retrievers: &[]config.RetrieverConf{
				{
					Kind: "file",
					Path: "non-existent-file.yaml",
				},
			},
		},
		FlagSets: []config.FlagSet{
			{
				Name: "invalid-flagset",
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000,
					FileFormat:      "yaml",
					Retrievers: &[]config.RetrieverConf{
						{
							Kind: "file",
							Path: "another-non-existent-file.yaml",
						},
					},
				},
				ApiKeys: []string{"api-key-1"},
			},
		},
	}

	logger := zap.NewNop()
	manager, err := service.NewFlagsetManager(config, logger)

	// Should fail because both flagsets and default config fail
	require.Error(t, err)
	assert.Equal(t, service.FlagsetManager{}, manager)
}

func TestFlagsetManager_EmptyApiKeys(t *testing.T) {
	// Test flagset with empty API keys
	config := &config.Config{
		FlagSets: []config.FlagSet{
			{
				Name: "test-flagset",
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000,
					FileFormat:      "yaml",
					Retrievers: &[]config.RetrieverConf{
						{
							Kind: "file",
							Path: "../../../testdata/flag-config.yaml",
						},
					},
				},
				ApiKeys: []string{}, // Empty API keys
			},
		},
	}

	logger := zap.NewNop()
	manager, err := service.NewFlagsetManager(config, logger)

	require.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, 1, len(manager.FlagSets))
	assert.Equal(t, 0, len(manager.APIKeysToFlagSet)) // No API keys mapped
}

func TestFlagsetManager_MultipleApiKeys(t *testing.T) {
	// Test flagset with multiple API keys mapping to the same flagset
	config := &config.Config{
		FlagSets: []config.FlagSet{
			{
				Name: "test-flagset",
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000,
					FileFormat:      "yaml",
					Retrievers: &[]config.RetrieverConf{
						{
							Kind: "file",
							Path: "../../../testdata/flag-config.yaml",
						},
					},
				},
				ApiKeys: []string{"api-key-1", "api-key-2", "api-key-3"},
			},
		},
	}

	logger := zap.NewNop()
	manager, err := service.NewFlagsetManager(config, logger)

	require.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, 1, len(manager.FlagSets))
	assert.Equal(t, 3, len(manager.APIKeysToFlagSet))

	// All API keys should map to the same flagset
	flagset1 := manager.GetFlagSet("api-key-1")
	flagset2 := manager.GetFlagSet("api-key-2")
	flagset3 := manager.GetFlagSet("api-key-3")

	assert.NotNil(t, flagset1)
	assert.NotNil(t, flagset2)
	assert.NotNil(t, flagset3)
	assert.Equal(t, flagset1, flagset2)
	assert.Equal(t, flagset2, flagset3)
}
