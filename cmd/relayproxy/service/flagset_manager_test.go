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
			expectedLen: 1, // Only valid flagset should be created
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := service.NewFlagsetManager(tt.config, tt.logger, nil)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, manager)

			// Test interface methods
			flagSets, err := manager.GetFlagSets()
			assert.NoError(t, err)
			assert.NotNil(t, flagSets)

			// Test GetFlagSet with empty API key
			flagSet, err := manager.GetFlagSet("")
			assert.NoError(t, err)
			assert.NotNil(t, flagSet)

			// Test IsDefaultFlagSet
			if tt.config != nil && len(tt.config.FlagSets) == 0 {
				assert.True(t, manager.IsDefaultFlagSet())
			} else {
				assert.False(t, manager.IsDefaultFlagSet())
			}

			// Test GetFlagSets
			flagSetsMap, err := manager.GetFlagSets()
			assert.NoError(t, err)
			if tt.config != nil && len(tt.config.FlagSets) > 0 {
				if !manager.IsDefaultFlagSet() {
					assert.NotEmpty(t, flagSetsMap)
				} else {
					assert.Empty(t, flagSetsMap)
				}
			}
		})
	}
}

func TestFlagsetManager_GetFlagSet(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		logger         *zap.Logger
		apiKey         string
		expectedResult bool // Whether we expect a non-nil result
	}{
		{
			name: "default mode with empty API key should return default flagset",
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
			logger:         zap.NewNop(),
			apiKey:         "",
			expectedResult: true,
		},
		{
			name: "flagsets mode with valid API key should return flagset",
			config: &config.Config{
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
						ApiKeys: []string{"test-api-key"},
					},
				},
			},
			logger:         zap.NewNop(),
			apiKey:         "test-api-key",
			expectedResult: true,
		},
		{
			name: "flagsets mode with invalid API key should return nil",
			config: &config.Config{
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
						ApiKeys: []string{"test-api-key"},
					},
				},
			},
			logger:         zap.NewNop(),
			apiKey:         "invalid-api-key",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := service.NewFlagsetManager(tt.config, tt.logger, nil)
			require.NoError(t, err)
			require.NotNil(t, manager)

			result, err := manager.GetFlagSet(tt.apiKey)
			if tt.expectedResult {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			} else {
				if err == nil {
					assert.Nil(t, result)
				} else {
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestFlagsetManager_GetFlagSets(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		logger         *zap.Logger
		expectedLength int
	}{
		{
			name: "default mode should return single flagset",
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
			logger:         zap.NewNop(),
			expectedLength: 1,
		},
		{
			name: "flagsets mode should return multiple flagsets",
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
			},
			logger:         zap.NewNop(),
			expectedLength: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := service.NewFlagsetManager(tt.config, tt.logger, nil)
			require.NoError(t, err)
			require.NotNil(t, manager)

			flagSets, err := manager.GetFlagSets()
			assert.NoError(t, err)
			assert.Len(t, flagSets, tt.expectedLength)
		})
	}
}

func TestFlagsetManager_IsDefaultFlagSet(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		logger         *zap.Logger
		expectedResult bool
	}{
		{
			name: "default mode should return true for IsDefaultFlagSet",
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
			logger:         zap.NewNop(),
			expectedResult: true,
		},
		{
			name: "flagsets mode should return false for IsDefaultFlagSet",
			config: &config.Config{
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
			},
			logger:         zap.NewNop(),
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := service.NewFlagsetManager(tt.config, tt.logger, nil)
			require.NoError(t, err)
			require.NotNil(t, manager)

			result := manager.IsDefaultFlagSet()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestFlagsetManager_GetDefaultFlagSet(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		logger         *zap.Logger
		expectedResult bool
	}{
		{
			name: "default mode should return default flagset",
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
			logger:         zap.NewNop(),
			expectedResult: true,
		},
		{
			name: "flagsets mode should return nil default flagset",
			config: &config.Config{
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
			},
			logger:         zap.NewNop(),
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := service.NewFlagsetManager(tt.config, tt.logger, nil)
			require.NoError(t, err)
			require.NotNil(t, manager)

			defaultFlagSet := manager.GetDefaultFlagSet()
			if tt.expectedResult {
				assert.NotNil(t, defaultFlagSet)
			} else {
				assert.Nil(t, defaultFlagSet)
			}
		})
	}
}
