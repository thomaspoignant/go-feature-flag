package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
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

func TestFlagsetManager_FlagSet(t *testing.T) {
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
			flagset, err := manager.FlagSet("test-api-key")
			assert.NoError(t, err)
			assert.NotNil(t, flagset)
		})
		t.Run("invalid api key", func(t *testing.T) {
			flagset, err := manager.FlagSet("invalid-key")
			assert.Error(t, err)
			assert.Nil(t, flagset)
		})
		t.Run("empty api key", func(t *testing.T) {
			flagset, err := manager.FlagSet("")
			assert.Error(t, err)
			assert.Nil(t, flagset)
		})

		t.Run("empty api key", func(t *testing.T) {
			flagset, err := manager.FlagSet("")
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
			flagset, err := manager.FlagSet("")
			assert.NoError(t, err)
			assert.NotNil(t, flagset)
		})
	})
}

func TestFlagsetManager_FlagSetName(t *testing.T) {
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
			name, err := manager.FlagSetName("test-api-key")
			assert.NoError(t, err)
			assert.Equal(t, "test-flagset", name)
		})
		t.Run("non-existing api key", func(t *testing.T) {
			name, err := manager.FlagSetName("invalid-key")
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
			name, err := manager.FlagSetName("")
			assert.NoError(t, err)
			assert.Equal(t, "default", name)
		})
	})
}

func TestFlagsetManager_AllFlagSets(t *testing.T) {
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

		flagsets, err := manager.AllFlagSets()
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

		flagsets, err := manager.AllFlagSets()
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

		flagsets, err := manager.AllFlagSets()
		assert.NoError(t, err)
		assert.Len(t, flagsets, 1)
		assert.Contains(t, flagsets, "default")
	})
}

func TestFlagsetManager_Default(t *testing.T) {
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

		defaultFlagset := manager.Default()
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

		defaultFlagset := manager.Default()
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

// setupManager is a helper function to create a manager and capture log messages
func setupManager(t *testing.T, cfg *config.Config) (service.FlagsetManager, *observer.ObservedLogs) {
	t.Helper()
	obs, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(obs)
	manager, err := service.NewFlagsetManager(cfg, logger, nil)
	require.NoError(t, err)
	t.Cleanup(func() { manager.Close() })
	return manager, logs
}
func TestFlagsetManager_OnConfigChange(t *testing.T) {
	flagConfig := "../testdata/controller/configuration_flags.yaml"

	t.Run("should reject switching from default to flagsets mode", func(t *testing.T) {
		// Create manager in default mode
		cfg := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
		}
		manager, logs := setupManager(t, cfg)

		// Try to switch to flagsets mode
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "new-flagset",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"new-key"},
				},
			},
		}

		assert.NotPanics(t, func() {
			manager.OnConfigChange(newConfig)
		})

		// Verify that the error log was displayed
		assert.Equal(t, 1, logs.Len(), "Expected exactly one error log message")
		logEntry := logs.All()[0]
		assert.Equal(t, zap.ErrorLevel, logEntry.Level)
		assert.Contains(t, logEntry.Message, "switching from default to flagsets mode (or the opposite) is not supported during runtime")
	})

	t.Run("should reject switching from flagsets to default mode", func(t *testing.T) {
		// Create manager in flagsets mode
		cfg := &config.Config{
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
		manager, logs := setupManager(t, cfg)

		// Try to switch to default mode
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
		}

		// Should not panic and should reject the change
		assert.NotPanics(t, func() {
			manager.OnConfigChange(newConfig)
		})
		// Verify that the error log was displayed
		assert.Equal(t, 1, logs.Len(), "Expected exactly one error log message")
		logEntry := logs.All()[0]
		assert.Equal(t, zap.ErrorLevel, logEntry.Level)
		assert.Contains(t, logEntry.Message, "switching from default to flagsets mode (or the opposite) is not supported during runtime")
	})

	t.Run("should update AuthorizedKeys in default mode", func(t *testing.T) {
		// Create manager in default mode
		cfg := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
			AuthorizedKeys: config.APIKeys{
				Evaluation: []string{"old-key"},
			},
		}
		manager, _ := setupManager(t, cfg)

		// Update AuthorizedKeys
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
			AuthorizedKeys: config.APIKeys{
				Evaluation: []string{"new-key"},
			},
		}

		manager.OnConfigChange(newConfig)

		// Verify the key was updated - ForceReloadAPIKeys resets and reloads
		assert.False(t, cfg.APIKeyExists("old-key"))
		assert.True(t, cfg.APIKeyExists("new-key"))
	})

	t.Run("should update APIKeys in default mode", func(t *testing.T) {
		// Create manager in default mode
		cfg := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
			APIKeys: []string{"old-api-key"},
		}
		manager, _ := setupManager(t, cfg)

		// Update APIKeys
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
			APIKeys: []string{"new-api-key"},
		}

		manager.OnConfigChange(newConfig)

		// Verify the key was updated - ForceReloadAPIKeys resets and reloads
		assert.False(t, cfg.APIKeyExists("old-api-key"))
		assert.True(t, cfg.APIKeyExists("new-api-key"))
	})

	t.Run("should update both AuthorizedKeys and APIKeys in default mode", func(t *testing.T) {
		// Create manager in default mode
		cfg := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
			AuthorizedKeys: config.APIKeys{
				Evaluation: []string{"old-authorized-key"},
			},
			APIKeys: []string{"old-api-key"},
		}
		manager, _ := setupManager(t, cfg)

		// Update both
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
			AuthorizedKeys: config.APIKeys{
				Evaluation: []string{"new-authorized-key"},
			},
			APIKeys: []string{"new-api-key"},
		}

		manager.OnConfigChange(newConfig)

		// Verify both were updated
		assert.False(t, cfg.APIKeyExists("old-authorized-key"))
		assert.False(t, cfg.APIKeyExists("old-api-key"))
		assert.True(t, cfg.APIKeyExists("new-authorized-key"))
		assert.True(t, cfg.APIKeyExists("new-api-key"))
	})

	t.Run("should reject invalid configuration in default mode - missing retriever", func(t *testing.T) {
		// Create manager in default mode
		cfg := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: flagConfig,
				},
			},
			APIKeys: []string{"old-api-key"},
		}
		// Preload API keys to initialize the internal API key set
		cfg.ForceReloadAPIKeys()
		manager, logs := setupManager(t, cfg)

		// Try to update with invalid config (no retriever)
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: nil,
			},
			APIKeys: []string{"new-api-key"},
		}

		assert.NotPanics(t, func() {
			manager.OnConfigChange(newConfig)
		})

		// Verify that the error log was displayed
		assert.Equal(t, 1, logs.Len(), "Expected exactly one error log message")
		logEntry := logs.All()[0]
		assert.Equal(t, zap.ErrorLevel, logEntry.Level)
		assert.Contains(t, logEntry.Message, "the new configuration is invalid, it will not be applied")

		// Verify the original config was not changed
		assert.True(t, cfg.APIKeyExists("old-api-key"))
		assert.False(t, cfg.APIKeyExists("new-api-key"))
	})

	t.Run("should reject invalid configuration in flagsets mode - flagset with no API keys", func(t *testing.T) {
		// Create manager in flagsets mode
		cfg := &config.Config{
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
		manager, logs := setupManager(t, cfg)

		// Try to update with invalid config (flagset with no API keys)
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{}, // Invalid: no API keys
				},
			},
		}

		assert.NotPanics(t, func() {
			manager.OnConfigChange(newConfig)
		})

		// Verify that the error log was displayed
		assert.Equal(t, 1, logs.Len(), "Expected exactly one error log message")
		logEntry := logs.All()[0]
		assert.Equal(t, zap.ErrorLevel, logEntry.Level)
		assert.Contains(t, logEntry.Message, "the new configuration is invalid, it will not be applied")

		// Verify the original config was not changed
		flagset, err := manager.FlagSet("test-api-key")
		assert.NoError(t, err)
		assert.NotNil(t, flagset)
	})

	t.Run("should reject invalid configuration in flagsets mode - duplicate API keys", func(t *testing.T) {
		// Create manager in flagsets mode
		cfg := &config.Config{
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
		manager, logs := setupManager(t, cfg)

		// Try to update with invalid config (duplicate API keys across flagsets)
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"duplicate-key"},
				},
				{
					Name: "test-flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"duplicate-key"}, // Invalid: duplicate API key
				},
			},
		}

		assert.NotPanics(t, func() {
			manager.OnConfigChange(newConfig)
		})

		// Verify that the error log was displayed
		assert.Equal(t, 1, logs.Len(), "Expected exactly one error log message")
		logEntry := logs.All()[0]
		assert.Equal(t, zap.ErrorLevel, logEntry.Level)
		assert.Contains(t, logEntry.Message, "the new configuration is invalid, it will not be applied")

		// Verify the original config was not changed
		flagset1, err := manager.FlagSet("api-key-1")
		assert.NoError(t, err)
		assert.NotNil(t, flagset1)
		flagset2, err := manager.FlagSet("api-key-2")
		assert.NoError(t, err)
		assert.NotNil(t, flagset2)
	})

	t.Run("should successfully update API keys in flagsets mode", func(t *testing.T) {
		// Create manager in flagsets mode
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"old-api-key"},
				},
			},
		}
		manager, _ := setupManager(t, cfg)

		// Verify old key works before change
		flagset, err := manager.FlagSet("old-api-key")
		assert.NoError(t, err)
		assert.NotNil(t, flagset)

		// Update API key
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"new-api-key"},
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// Verify the key was updated
		_, err = manager.FlagSet("old-api-key")
		assert.Error(t, err, "old API key should no longer work")

		flagset, err = manager.FlagSet("new-api-key")
		assert.NoError(t, err, "new API key should work")
		assert.NotNil(t, flagset)
	})

	t.Run("should add new API keys to existing flagset", func(t *testing.T) {
		// Create manager in flagsets mode with single key
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"api-key-1"},
				},
			},
		}
		manager, _ := setupManager(t, cfg)

		// Verify initial state
		_, err := manager.FlagSet("api-key-1")
		assert.NoError(t, err)
		_, err = manager.FlagSet("api-key-2")
		assert.Error(t, err, "api-key-2 should not work yet")

		// Add second API key
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"api-key-1", "api-key-2"},
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// Verify both keys work now
		flagset1, err := manager.FlagSet("api-key-1")
		assert.NoError(t, err)
		assert.NotNil(t, flagset1)

		flagset2, err := manager.FlagSet("api-key-2")
		assert.NoError(t, err)
		assert.NotNil(t, flagset2)

		// Both should point to the same flagset
		name1, _ := manager.FlagSetName("api-key-1")
		name2, _ := manager.FlagSetName("api-key-2")
		assert.Equal(t, name1, name2)
	})

	t.Run("should remove API keys from flagset", func(t *testing.T) {
		// Create manager with two API keys
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"api-key-1", "api-key-2"},
				},
			},
		}
		manager, _ := setupManager(t, cfg)

		// Verify both keys work initially
		_, err := manager.FlagSet("api-key-1")
		assert.NoError(t, err)
		_, err = manager.FlagSet("api-key-2")
		assert.NoError(t, err)

		// Remove one API key
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "test-flagset",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"api-key-1"},
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// Verify api-key-1 still works
		flagset, err := manager.FlagSet("api-key-1")
		assert.NoError(t, err)
		assert.NotNil(t, flagset)

		// Verify api-key-2 no longer works
		_, err = manager.FlagSet("api-key-2")
		assert.Error(t, err, "api-key-2 should no longer work")
	})

	t.Run("should move API key between flagsets", func(t *testing.T) {
		// Create manager with two flagsets
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"api-key-1", "api-key-move"},
				},
				{
					Name: "flagset-2",
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
		manager, _ := setupManager(t, cfg)

		// Verify initial state
		name, err := manager.FlagSetName("api-key-move")
		assert.NoError(t, err)
		assert.Equal(t, "flagset-1", name)

		// Move api-key-move from flagset-1 to flagset-2
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
					APIKeys: []string{"api-key-1"},
				},
				{
					Name: "flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"api-key-2", "api-key-move"},
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// Verify api-key-move now points to flagset-2
		name, err = manager.FlagSetName("api-key-move")
		assert.NoError(t, err)
		assert.Equal(t, "flagset-2", name)

		// Verify other keys still work correctly
		name1, _ := manager.FlagSetName("api-key-1")
		assert.Equal(t, "flagset-1", name1)
		name2, _ := manager.FlagSetName("api-key-2")
		assert.Equal(t, "flagset-2", name2)
	})

	t.Run("should not update when config is unchanged in flagsets mode", func(t *testing.T) {
		// Create manager in flagsets mode
		cfg := &config.Config{
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
		obs, logs := observer.New(zap.InfoLevel)
		logger := zap.New(obs)
		manager, err := service.NewFlagsetManager(cfg, logger, nil)
		require.NoError(t, err)
		t.Cleanup(func() { manager.Close() })

		// Apply the same config again
		sameConfig := &config.Config{
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

		manager.OnConfigChange(sameConfig)

		// Verify no "Configuration changed" log was produced
		configChangeLogs := logs.FilterMessage("Configuration changed: updating the APIKeys for flagset")
		assert.Equal(t, 0, configChangeLogs.Len(), "No config change log should be produced when config is unchanged")

		// Verify the key still works
		flagset, err := manager.FlagSet("test-api-key")
		assert.NoError(t, err)
		assert.NotNil(t, flagset)
	})

	t.Run("should not update API keys for flagset with empty name", func(t *testing.T) {
		// Create manager with a flagset that has an empty name (gets auto-generated)
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "", // Empty name - will be auto-generated
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"original-key"},
				},
			},
		}
		obs, logs := observer.New(zap.WarnLevel)
		logger := zap.New(obs)
		manager, err := service.NewFlagsetManager(cfg, logger, nil)
		require.NoError(t, err)
		t.Cleanup(func() { manager.Close() })

		// Verify the warning about empty name was logged
		warnLogs := logs.FilterMessageSnippet("no flagset name provided")
		assert.Equal(t, 1, warnLogs.Len(), "Expected warning about empty flagset name")

		// Verify original key works
		flagset, err := manager.FlagSet("original-key")
		assert.NoError(t, err)
		assert.NotNil(t, flagset)

		// Try to update config with empty name flagset - should be ignored
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "", // Still empty
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"new-key"}, // Try to change key
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// Original key should still work because empty name flagsets can't be updated
		flagset, err = manager.FlagSet("original-key")
		assert.NoError(t, err)
		assert.NotNil(t, flagset)
	})

	t.Run("should not update API keys for flagset named default", func(t *testing.T) {
		// Create manager with a flagset named "default" (gets auto-generated)
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "default", // Reserved name - will be auto-generated
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"original-key"},
				},
			},
		}
		obs, logs := observer.New(zap.WarnLevel)
		logger := zap.New(obs)
		manager, err := service.NewFlagsetManager(cfg, logger, nil)
		require.NoError(t, err)
		t.Cleanup(func() { manager.Close() })

		// Verify the warning about using 'default' name was logged
		warnLogs := logs.FilterMessageSnippet("using 'default' as a flagset name")
		assert.Equal(t, 1, warnLogs.Len(), "Expected warning about using 'default' as flagset name")

		// Verify original key works
		flagset, err := manager.FlagSet("original-key")
		assert.NoError(t, err)
		assert.NotNil(t, flagset)

		// Try to update config with "default" name flagset - should be ignored
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "default",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"new-key"},
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// Original key should still work because "default" name flagsets can't be updated
		flagset, err = manager.FlagSet("original-key")
		assert.NoError(t, err)
		assert.NotNil(t, flagset)
	})

	t.Run("should update multiple flagsets API keys simultaneously", func(t *testing.T) {
		// Create manager with multiple flagsets
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"old-key-1"},
				},
				{
					Name: "flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"old-key-2"},
				},
			},
		}
		manager, _ := setupManager(t, cfg)

		// Verify initial state
		name1, _ := manager.FlagSetName("old-key-1")
		assert.Equal(t, "flagset-1", name1)
		name2, _ := manager.FlagSetName("old-key-2")
		assert.Equal(t, "flagset-2", name2)

		// Update both flagsets
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
					APIKeys: []string{"new-key-1"},
				},
				{
					Name: "flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					APIKeys: []string{"new-key-2"},
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// Verify old keys no longer work
		_, err := manager.FlagSet("old-key-1")
		assert.Error(t, err)
		_, err = manager.FlagSet("old-key-2")
		assert.Error(t, err)

		// Verify new keys work and point to correct flagsets
		name1, err = manager.FlagSetName("new-key-1")
		assert.NoError(t, err)
		assert.Equal(t, "flagset-1", name1)

		name2, err = manager.FlagSetName("new-key-2")
		assert.NoError(t, err)
		assert.Equal(t, "flagset-2", name2)
	})
}

func TestFlagsetManager_ConfigurationInheritance(t *testing.T) {
	flagConfig := "../testdata/controller/configuration_flags.yaml"

	// Test that flagsets inherit configuration from top level
	config := &config.Config{
		// Top-level configuration
		CommonFlagSet: config.CommonFlagSet{
			FileFormat:          "yaml",
			PollingInterval:     120,
			Environment:         "production",
			EnablePollingJitter: true,
		},
		FlagSets: []config.FlagSet{
			{
				Name:    "test-flagset",
				APIKeys: []string{"test-api-key"},
				// Only overrides some fields, should inherit others from top level
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "file",
						Path: flagConfig,
					},
					PollingInterval: 60, // Override polling interval
					// Should inherit: FileFormat="json", Environment="production", EnablePollingJitter=true
				},
			},
			{
				Name:    "empty-flagset",
				APIKeys: []string{"empty-api-key"},
				// Should inherit all from top level
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "file",
						Path: flagConfig,
					},
				},
			},
		},
	}

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	notifiers := []notifier.Notifier{}
	manager, err := service.NewFlagsetManager(config, logger, notifiers)
	if err != nil {
		t.Fatalf("failed to create FlagsetManager: %v", err)
	}
	assert.NotNil(t, manager)
	defer manager.Close()

	// Verify flagsets are accessible
	t.Run("flagsets should be accessible", func(t *testing.T) {
		flagset1, err := manager.FlagSet("test-api-key")
		assert.NoError(t, err)
		assert.NotNil(t, flagset1)

		flagset2, err := manager.FlagSet("empty-api-key")
		assert.NoError(t, err)
		assert.NotNil(t, flagset2)
	})

	// Verify flagset names are returned correctly
	t.Run("flagset names should be correct", func(t *testing.T) {
		name1, err := manager.FlagSetName("test-api-key")
		assert.NoError(t, err)
		assert.Equal(t, "test-flagset", name1)

		name2, err := manager.FlagSetName("empty-api-key")
		assert.NoError(t, err)
		assert.Equal(t, "empty-flagset", name2)
	})

	// Verify all flagsets are available
	t.Run("all flagsets should be available", func(t *testing.T) {
		allFlagsets, err := manager.AllFlagSets()
		assert.NoError(t, err)
		assert.Len(t, allFlagsets, 2)

		// Verify flagsets exist by name
		assert.Contains(t, allFlagsets, "test-flagset")
		assert.Contains(t, allFlagsets, "empty-flagset")
	})
}
