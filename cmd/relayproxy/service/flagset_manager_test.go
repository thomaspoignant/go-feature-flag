package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

func TestNewFlagsetManager(t *testing.T) {
	flagConfig := "../testdata/controller/configuration_flags.yaml"
	tests := []struct {
		name        string
		config      *config.Config
		logger      *zap.Logger
		notifiers   []notifier.Notifier
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "nil config should return error",
			config:      nil,
			logger:      zap.NewNop(),
			notifiers:   []notifier.Notifier{},
			wantErr:     true,
			expectedErr: "configuration is nil",
		},
		{
			name: "empty flagsets should create default manager",
			config: &config.Config{
				FlagSets: []config.FlagSet{},
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "file",
						Path: flagConfig,
					},
				},
			},
			logger:    zap.NewNop(),
			notifiers: []notifier.Notifier{},
			wantErr:   false,
		},
		{
			name: "valid flagsets should create flagsets manager",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{
						Name: "test-flagset",
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: "file",
								Path: flagConfig,
							},
						},
						APIKeys: []string{"test-api-key"},
					},
				},
			},
			logger:    zap.NewNop(),
			notifiers: []notifier.Notifier{},
			wantErr:   false,
		},
		{
			name: "invalid flagsets should error even if default flagset is valid",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{
						Name: "invalid-flagset",
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: "invalid",
								Path: flagConfig,
							},
						},
						APIKeys: []string{"test-api-key"},
					},
				},
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "file",
						Path: flagConfig,
					},
				},
			},
			logger:    zap.NewNop(),
			notifiers: []notifier.Notifier{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := service.NewFlagsetManager(tt.config, tt.logger, tt.notifiers)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, manager)
				if tt.expectedErr != "" {
					assert.Equal(t, tt.expectedErr, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
				defer manager.Close()
			}
		})
	}
}

func TestFlagsetManager_GetFlagSet(t *testing.T) {
	flagConfig := "../testdata/controller/configuration_flags.yaml"

	// Test flagset mode
	t.Run("flagset mode", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"test-api-key"},
				},
			},
		}
		logger := zap.NewNop()
		notifiers := []notifier.Notifier{}
		manager, err := service.NewFlagsetManager(config, logger, notifiers)
		if err != nil {
			t.Fatalf("failed to create FlagsetManager: %v", err)
		}
		assert.NotNil(t, manager)
		defer manager.Close()

		t.Run("valid api key", func(t *testing.T) {
			flagset, err := manager.GetFlagSet("test-api-key")
			assert.NoError(t, err)
			assert.NotNil(t, flagset)
		})
		t.Run("invalid api key", func(t *testing.T) {
			flagset, err := manager.GetFlagSet("invalid-key")
			assert.Error(t, err)
			assert.Nil(t, flagset)
		})
		t.Run("empty api key", func(t *testing.T) {
			flagset, err := manager.GetFlagSet("")
			assert.Error(t, err)
			assert.Nil(t, flagset)
		})

		t.Run("empty api key", func(t *testing.T) {
			flagset, err := manager.GetFlagSet("")
			assert.Error(t, err)
			assert.Nil(t, flagset)
		})
	})

	// Test default mode
	t.Run("default mode", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
		}
		logger := zap.NewNop()
		notifiers := []notifier.Notifier{}
		manager, err := service.NewFlagsetManager(config, logger, notifiers)
		if err != nil {
			t.Fatalf("failed to create FlagsetManager: %v", err)
		}
		assert.NotNil(t, manager)
		defer manager.Close()

		t.Run("empty api key should work", func(t *testing.T) {
			flagset, err := manager.GetFlagSet("")
			assert.NoError(t, err)
			assert.NotNil(t, flagset)
		})
	})
}

func TestFlagsetManager_GetFlagSetName(t *testing.T) {
	flagConfig := "../testdata/controller/configuration_flags.yaml"

	// Test flagset mode
	t.Run("flagset mode", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"test-api-key"},
				},
			},
		}
		logger := zap.NewNop()
		notifiers := []notifier.Notifier{}
		manager, err := service.NewFlagsetManager(config, logger, notifiers)
		if err != nil {
			t.Fatalf("failed to create FlagsetManager: %v", err)
		}
		assert.NotNil(t, manager)
		defer manager.Close()

		t.Run("existing api key", func(t *testing.T) {
			name, err := manager.GetFlagSetName("test-api-key")
			assert.NoError(t, err)
			assert.Equal(t, "test-flagset", name)
		})
		t.Run("non-existing api key", func(t *testing.T) {
			name, err := manager.GetFlagSetName("invalid-key")
			assert.Error(t, err)
			assert.Equal(t, "", name)
			assert.Equal(t, "no flag set associated to the API key", err.Error())
		})
	})

	// Test default mode
	t.Run("default mode", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
		}
		logger := zap.NewNop()
		notifiers := []notifier.Notifier{}
		manager, err := service.NewFlagsetManager(config, logger, notifiers)
		if err != nil {
			t.Fatalf("failed to create FlagsetManager: %v", err)
		}
		assert.NotNil(t, manager)
		defer manager.Close()

		t.Run("empty api key should return default", func(t *testing.T) {
			name, err := manager.GetFlagSetName("")
			assert.NoError(t, err)
			assert.Equal(t, "default", name)
		})
	})
}

func TestFlagsetManager_GetFlagSets(t *testing.T) {
	flagConfig := "../testdata/controller/configuration_flags.yaml"

	// Test flagset mode
	t.Run("flagset mode", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"api-key-1"},
				},
				{
					Name: "test-flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"api-key-2"},
				},
			},
		}
		logger := zap.NewNop()
		notifiers := []notifier.Notifier{}
		manager, err := service.NewFlagsetManager(config, logger, notifiers)
		if err != nil {
			t.Fatalf("failed to create FlagsetManager: %v", err)
		}
		assert.NotNil(t, manager)
		defer manager.Close()

		flagsets, err := manager.GetFlagSets()
		assert.NoError(t, err)
		assert.Len(t, flagsets, 2)
		assert.Contains(t, flagsets, "test-flagset-1")
		assert.Contains(t, flagsets, "test-flagset-2")
	})

	t.Run("flagset mode using default flagset name", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "default",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"api-key-1"},
				},
				{
					Name: "test-flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"api-key-2"},
				},
			},
		}
		logger := zap.NewNop()
		notifiers := []notifier.Notifier{}
		manager, err := service.NewFlagsetManager(config, logger, notifiers)
		if err != nil {
			t.Fatalf("failed to create FlagsetManager: %v", err)
		}
		assert.NotNil(t, manager)
		defer manager.Close()

		flagsets, err := manager.GetFlagSets()
		assert.NoError(t, err)
		assert.Len(t, flagsets, 2)
		assert.NotContains(t, flagsets, "default")
		assert.Contains(t, flagsets, "test-flagset-2")
	})

	// Test default mode
	t.Run("default mode", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
		}
		logger := zap.NewNop()
		notifiers := []notifier.Notifier{}
		manager, err := service.NewFlagsetManager(config, logger, notifiers)
		if err != nil {
			t.Fatalf("failed to create FlagsetManager: %v", err)
		}
		assert.NotNil(t, manager)
		defer manager.Close()

		flagsets, err := manager.GetFlagSets()
		assert.NoError(t, err)
		assert.Len(t, flagsets, 1)
		assert.Contains(t, flagsets, "default")
	})
}

func TestFlagsetManager_GetDefaultFlagSet(t *testing.T) {
	flagConfig := "../testdata/controller/configuration_flags.yaml"

	// Test default mode
	t.Run("default mode", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
		}
		logger := zap.NewNop()
		notifiers := []notifier.Notifier{}
		manager, err := service.NewFlagsetManager(config, logger, notifiers)
		if err != nil {
			t.Fatalf("failed to create FlagsetManager: %v", err)
		}
		assert.NotNil(t, manager)
		defer manager.Close()

		defaultFlagset := manager.GetDefaultFlagSet()
		assert.NotNil(t, defaultFlagset)
	})

	// Test flagset mode
	t.Run("flagset mode", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"test-api-key"},
				},
			},
		}
		logger := zap.NewNop()
		notifiers := []notifier.Notifier{}
		manager, err := service.NewFlagsetManager(config, logger, notifiers)
		if err != nil {
			t.Fatalf("failed to create FlagsetManager: %v", err)
		}
		assert.NotNil(t, manager)
		defer manager.Close()

		defaultFlagset := manager.GetDefaultFlagSet()
		assert.Nil(t, defaultFlagset)
	})
}

func TestFlagsetManager_IsDefaultFlagSet(t *testing.T) {
	flagConfig := "../testdata/controller/configuration_flags.yaml"

	// Test default mode
	t.Run("default mode", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
		}
		logger := zap.NewNop()
		notifiers := []notifier.Notifier{}
		manager, err := service.NewFlagsetManager(config, logger, notifiers)
		if err != nil {
			t.Fatalf("failed to create FlagsetManager: %v", err)
		}
		assert.NotNil(t, manager)
		defer manager.Close()

		assert.True(t, manager.IsDefaultFlagSet())
	})

	// Test flagset mode
	t.Run("flagset mode", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"test-api-key"},
				},
			},
		}
		logger := zap.NewNop()
		notifiers := []notifier.Notifier{}
		manager, err := service.NewFlagsetManager(config, logger, notifiers)
		if err != nil {
			t.Fatalf("failed to create FlagsetManager: %v", err)
		}
		assert.NotNil(t, manager)
		defer manager.Close()

		assert.False(t, manager.IsDefaultFlagSet())
	})
}

func TestFlagsetManager_Close(t *testing.T) {
	flagConfig := "../testdata/controller/configuration_flags.yaml"

	// Test default mode
	t.Run("default mode", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
		}
		logger := zap.NewNop()
		notifiers := []notifier.Notifier{}
		manager, err := service.NewFlagsetManager(config, logger, notifiers)
		if err != nil {
			t.Fatalf("failed to create FlagsetManager: %v", err)
		}
		assert.NotNil(t, manager)

		assert.NotPanics(t, func() {
			manager.Close()
		})
	})

	// Test flagset mode
	t.Run("flagset mode", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"test-api-key"},
				},
			},
		}
		logger := zap.NewNop()
		notifiers := []notifier.Notifier{}
		manager, err := service.NewFlagsetManager(config, logger, notifiers)
		if err != nil {
			t.Fatalf("failed to create FlagsetManager: %v", err)
		}
		assert.NotNil(t, manager)

		assert.NotPanics(t, func() {
			manager.Close()
		})
	})
}

func TestFlagsetManager_ReloadFlagsets(t *testing.T) {
	flagConfig := "../testdata/controller/configuration_flags.yaml"
	logger := zap.NewNop()
	notifiers := []notifier.Notifier{}

	t.Run("successfully add new flagset", func(t *testing.T) {
		initialConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-1"},
				},
			},
		}

		manager, err := service.NewFlagsetManager(initialConfig, logger, notifiers)
		assert.NoError(t, err)
		defer manager.Close()

		// Verify initial state
		flagsets, err := manager.GetFlagSets()
		assert.NoError(t, err)
		assert.Len(t, flagsets, 1)

		// Reload with new flagset added
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-1"},
				},
				{
					Name: "flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-2"},
				},
			},
		}

		err = manager.ReloadFlagsets(newConfig, logger, notifiers)
		assert.NoError(t, err)

		// Verify new flagset was added
		flagsets, err = manager.GetFlagSets()
		assert.NoError(t, err)
		assert.Len(t, flagsets, 2)
		assert.Contains(t, flagsets, "flagset-1")
		assert.Contains(t, flagsets, "flagset-2")

		// Verify both flagsets are accessible
		flagset1, err := manager.GetFlagSet("key-1")
		assert.NoError(t, err)
		assert.NotNil(t, flagset1)

		flagset2, err := manager.GetFlagSet("key-2")
		assert.NoError(t, err)
		assert.NotNil(t, flagset2)
	})

	t.Run("successfully remove flagset", func(t *testing.T) {
		initialConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-1"},
				},
				{
					Name: "flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-2"},
				},
			},
		}

		manager, err := service.NewFlagsetManager(initialConfig, logger, notifiers)
		assert.NoError(t, err)
		defer manager.Close()

		// Verify initial state
		flagsets, err := manager.GetFlagSets()
		assert.NoError(t, err)
		assert.Len(t, flagsets, 2)

		// Reload with one flagset removed
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-1"},
				},
			},
		}

		err = manager.ReloadFlagsets(newConfig, logger, notifiers)
		assert.NoError(t, err)

		// Verify flagset was removed
		flagsets, err = manager.GetFlagSets()
		assert.NoError(t, err)
		assert.Len(t, flagsets, 1)
		assert.Contains(t, flagsets, "flagset-1")

		// Verify removed flagset is no longer accessible
		_, err = manager.GetFlagSet("key-2")
		assert.Error(t, err)
	})

	t.Run("reject when existing flagset is modified", func(t *testing.T) {
		initialConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
						PollingInterval: 60000,
					},
					APIKeys: []string{"key-1"},
				},
			},
		}

		manager, err := service.NewFlagsetManager(initialConfig, logger, notifiers)
		assert.NoError(t, err)
		defer manager.Close()

		// Try to reload with modified polling interval
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
						PollingInterval: 30000, // Changed
					},
					APIKeys: []string{"key-1"},
				},
			},
		}

		err = manager.ReloadFlagsets(newConfig, logger, notifiers)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "has been modified, reload rejected")
	})

	t.Run("reject when API key moved to different flagset", func(t *testing.T) {
		initialConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-1"},
				},
				{
					Name: "flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-2"},
				},
			},
		}

		manager, err := service.NewFlagsetManager(initialConfig, logger, notifiers)
		assert.NoError(t, err)
		defer manager.Close()

		// Try to reload with API key moved to different flagset
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-1", "key-2"}, // key-2 moved here
				},
				{
					Name: "flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{}, // key-2 removed
				},
			},
		}

		err = manager.ReloadFlagsets(newConfig, logger, notifiers)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API key moved from flagset")
	})

	t.Run("reject when in default mode", func(t *testing.T) {
		initialConfig := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
		}

		manager, err := service.NewFlagsetManager(initialConfig, logger, notifiers)
		assert.NoError(t, err)
		defer manager.Close()

		assert.True(t, manager.IsDefaultFlagSet())

		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-1"},
				},
			},
		}

		err = manager.ReloadFlagsets(newConfig, logger, notifiers)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot reload flagsets in default mode")
	})

	t.Run("reject when new config has no flagsets", func(t *testing.T) {
		initialConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-1"},
				},
			},
		}

		manager, err := service.NewFlagsetManager(initialConfig, logger, notifiers)
		assert.NoError(t, err)
		defer manager.Close()

		newConfig := &config.Config{
			FlagSets: []config.FlagSet{},
		}

		err = manager.ReloadFlagsets(newConfig, logger, notifiers)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot reload: new configuration has no flagsets")
	})

	t.Run("allow API keys to be added to existing flagset", func(t *testing.T) {
		initialConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-1"},
				},
			},
		}

		manager, err := service.NewFlagsetManager(initialConfig, logger, notifiers)
		assert.NoError(t, err)
		defer manager.Close()

		// Reload with additional API key
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-1", "key-2"}, // Added key-2
				},
			},
		}

		err = manager.ReloadFlagsets(newConfig, logger, notifiers)
		assert.NoError(t, err)

		// Verify both API keys work
		flagset1, err := manager.GetFlagSet("key-1")
		assert.NoError(t, err)
		assert.NotNil(t, flagset1)

		flagset2, err := manager.GetFlagSet("key-2")
		assert.NoError(t, err)
		assert.NotNil(t, flagset2)

		// Both should return the same flagset
		assert.Equal(t, flagset1, flagset2)
	})

	t.Run("allow API keys to be removed from existing flagset", func(t *testing.T) {
		initialConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-1", "key-2"},
				},
			},
		}

		manager, err := service.NewFlagsetManager(initialConfig, logger, notifiers)
		assert.NoError(t, err)
		defer manager.Close()

		// Reload with one API key removed
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-1"}, // Removed key-2
				},
			},
		}

		err = manager.ReloadFlagsets(newConfig, logger, notifiers)
		assert.NoError(t, err)

		// Verify key-1 still works
		_, err = manager.GetFlagSet("key-1")
		assert.NoError(t, err)

		// Verify key-2 no longer works
		_, err = manager.GetFlagSet("key-2")
		assert.Error(t, err)
	})

	t.Run("handle flagsets without names", func(t *testing.T) {
		initialConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "", // No name
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-1"},
				},
			},
		}

		manager, err := service.NewFlagsetManager(initialConfig, logger, notifiers)
		assert.NoError(t, err)
		defer manager.Close()

		// Reload with same configuration (no name still)
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "", // Still no name
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"key-1"},
				},
			},
		}

		err = manager.ReloadFlagsets(newConfig, logger, notifiers)
		assert.NoError(t, err)

		// Verify it still works
		_, err = manager.GetFlagSet("key-1")
		assert.NoError(t, err)
	})
}
