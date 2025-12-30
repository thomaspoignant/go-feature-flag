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
}
