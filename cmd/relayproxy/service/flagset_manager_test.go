package service_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewFlagsetManager(t *testing.T) {
	flagConfig := "../testdata/goff/configuration_flags.yaml"
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
			manager, err := service.NewFlagsetManager(tt.config, tt.logger, tt.notifiers, nil)

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

func TestNewFlagsetManager_WithSSEService(t *testing.T) {
	flagConfig := "../testdata/goff/configuration_flags.yaml"

	tests := []struct {
		name   string
		config *config.Config
	}{
		{
			name: "default mode with SSE service attaches SSE notifier",
			config: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "file",
						Path: flagConfig,
					},
				},
			},
		},
		{
			name: "flagsets mode with SSE service attaches SSE notifier per flagset",
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sseService := stream.NewSSEService()
			defer sseService.Close()

			manager, err := service.NewFlagsetManager(tt.config, zap.NewNop(), []notifier.Notifier{}, sseService)
			require.NoError(t, err)
			require.NotNil(t, manager)
			defer manager.Close()
		})
	}
}

func TestFlagsetManager_FlagSet(t *testing.T) {
	flagConfig := "../testdata/goff/configuration_flags.yaml"

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
		manager, err := service.NewFlagsetManager(config, logger, notifiers, nil)
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
		manager, err := service.NewFlagsetManager(config, logger, notifiers, nil)
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
	flagConfig := "../testdata/goff/configuration_flags.yaml"

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
		manager, err := service.NewFlagsetManager(config, logger, notifiers, nil)
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
		manager, err := service.NewFlagsetManager(config, logger, notifiers, nil)
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
	flagConfig := "../testdata/goff/configuration_flags.yaml"

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
		manager, err := service.NewFlagsetManager(config, logger, notifiers, nil)
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
		manager, err := service.NewFlagsetManager(config, logger, notifiers, nil)
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
		manager, err := service.NewFlagsetManager(config, logger, notifiers, nil)
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
	flagConfig := "../testdata/goff/configuration_flags.yaml"

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
		manager, err := service.NewFlagsetManager(config, logger, notifiers, nil)
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
		manager, err := service.NewFlagsetManager(config, logger, notifiers, nil)
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
	flagConfig := "../testdata/goff/configuration_flags.yaml"

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
		manager, err := service.NewFlagsetManager(config, logger, notifiers, nil)
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
		manager, err := service.NewFlagsetManager(config, logger, notifiers, nil)
		if err != nil {
			t.Fatalf("failed to create FlagsetManager: %v", err)
		}
		assert.NotNil(t, manager)
		defer manager.Close()

		assert.False(t, manager.IsDefaultFlagSet())
	})
}

func TestFlagsetManager_Close(t *testing.T) {
	flagConfig := "../testdata/goff/configuration_flags.yaml"

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
		manager, err := service.NewFlagsetManager(config, logger, notifiers, nil)
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
		manager, err := service.NewFlagsetManager(config, logger, notifiers, nil)
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
	manager, err := service.NewFlagsetManager(cfg, logger, nil, nil)
	require.NoError(t, err)
	t.Cleanup(func() { manager.Close() })
	return manager, logs
}
func TestFlagsetManager_OnConfigChange(t *testing.T) {
	flagConfig := "../testdata/goff/configuration_flags.yaml"

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

	t.Run("should rotate API keys across flagsets regardless of processing order", func(t *testing.T) {
		// Cyclic rotation in a single reload: ka: A->B, kb: B->C, kc: C->A.
		// This is order-sensitive with incremental per-flagset updates, so it guards against a
		// regression of the routing rebuild.
		newFlagset := func(name, key string) config.FlagSet {
			return config.FlagSet{
				Name: name,
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
				},
				APIKeys: []string{key},
			}
		}
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				newFlagset("flagset-a", "ka"),
				newFlagset("flagset-b", "kb"),
				newFlagset("flagset-c", "kc"),
			},
		}
		manager, _ := setupManager(t, cfg)

		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				newFlagset("flagset-a", "kc"),
				newFlagset("flagset-b", "ka"),
				newFlagset("flagset-c", "kb"),
			},
		}
		manager.OnConfigChange(newConfig)

		name, err := manager.FlagSetName("ka")
		assert.NoError(t, err)
		assert.Equal(t, "flagset-b", name)
		name, err = manager.FlagSetName("kb")
		assert.NoError(t, err)
		assert.Equal(t, "flagset-c", name)
		name, err = manager.FlagSetName("kc")
		assert.NoError(t, err)
		assert.Equal(t, "flagset-a", name)
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
		manager, err := service.NewFlagsetManager(cfg, logger, nil, nil)
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
		manager, err := service.NewFlagsetManager(cfg, logger, nil, nil)
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
		manager, err := service.NewFlagsetManager(cfg, logger, nil, nil)
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

	t.Run("should add a new flagset at runtime", func(t *testing.T) {
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"key-1"},
				},
			},
		}
		manager, logs := setupManager(t, cfg)

		// The second flagset does not exist yet.
		_, err := manager.FlagSet("key-2")
		assert.Error(t, err)

		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"key-1"},
				},
				{
					Name: "flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"key-2"},
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// The new flagset is now available and the existing one still works.
		flagset2, err := manager.FlagSet("key-2")
		assert.NoError(t, err, "the flagset added at runtime should be available")
		assert.NotNil(t, flagset2)
		name2, err := manager.FlagSetName("key-2")
		assert.NoError(t, err)
		assert.Equal(t, "flagset-2", name2)

		flagset1, err := manager.FlagSet("key-1")
		assert.NoError(t, err)
		assert.NotNil(t, flagset1)

		all, err := manager.AllFlagSets()
		assert.NoError(t, err)
		assert.Len(t, all, 2)

		// Adding a flagset must not produce any error log.
		assert.Equal(t, 0, logs.Len(), "adding a flagset should not log any error")
	})

	t.Run("should remove a flagset at runtime", func(t *testing.T) {
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"key-1"},
				},
				{
					Name: "flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"key-2"},
				},
			},
		}
		manager, logs := setupManager(t, cfg)

		// Both flagsets exist initially.
		_, err := manager.FlagSet("key-1")
		assert.NoError(t, err)
		_, err = manager.FlagSet("key-2")
		assert.NoError(t, err)

		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"key-1"},
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// flagset-2 is gone, flagset-1 still works.
		flagset1, err := manager.FlagSet("key-1")
		assert.NoError(t, err)
		assert.NotNil(t, flagset1)

		_, err = manager.FlagSet("key-2")
		assert.Error(t, err, "the removed flagset should no longer be available")

		all, err := manager.AllFlagSets()
		assert.NoError(t, err)
		assert.Len(t, all, 1)

		assert.Equal(t, 0, logs.Len(), "removing a flagset should not log any error")
	})

	t.Run("should reject modifying an existing flagset configuration", func(t *testing.T) {
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever:       &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
						PollingInterval: 0,
					},
					APIKeys: []string{"key-1"},
				},
			},
		}
		manager, logs := setupManager(t, cfg)

		// Change the polling interval (a non-API-key field) -> forbidden.
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever:       &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
						PollingInterval: 5000,
					},
					APIKeys: []string{"key-1"},
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// The flagset is still served (unchanged) and an error is logged.
		flagset, err := manager.FlagSet("key-1")
		assert.NoError(t, err)
		assert.NotNil(t, flagset)

		rejectedLogs := logs.FilterMessageSnippet("modifying a flagset is not supported at runtime")
		assert.Equal(t, 1, rejectedLogs.Len(), "expected an error log about the rejected modification")
		assert.Contains(t, rejectedLogs.All()[0].Message, "next time the relay proxy is restarted")
		assert.Equal(t, "flagset-1", rejectedLogs.All()[0].ContextMap()["flagset"])
		// No bundled API-key error since the API keys did not change.
		assert.Equal(t, 0, logs.FilterMessageSnippet("API keys change for this flagset is also ignored").Len())
	})

	t.Run("should reject a bundled API key and configuration change", func(t *testing.T) {
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever:       &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
						PollingInterval: 0,
					},
					APIKeys: []string{"old-key"},
				},
			},
		}
		manager, logs := setupManager(t, cfg)

		// Change both the polling interval and the API key -> the whole change is rejected.
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-1",
					CommonFlagSet: config.CommonFlagSet{
						Retriever:       &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
						PollingInterval: 5000,
					},
					APIKeys: []string{"new-key"},
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// The old key still works and the new key is NOT applied.
		flagset, err := manager.FlagSet("old-key")
		assert.NoError(t, err, "the old API key must still work because the change is rejected")
		assert.NotNil(t, flagset)
		_, err = manager.FlagSet("new-key")
		assert.Error(t, err, "the bundled API key change must not be applied")

		// Two error logs: the modification rejection and the explicit bundled-key rejection.
		assert.Equal(t, 1, logs.FilterMessageSnippet("modifying a flagset is not supported at runtime").Len())
		assert.Equal(t, 1, logs.FilterMessageSnippet("API keys change for this flagset is also ignored").Len())
	})

	t.Run("should apply add, remove and key change while rejecting a modification in one reload", func(t *testing.T) {
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-a",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"key-a"},
				},
				{
					Name: "flagset-b",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"key-b"},
				},
				{
					Name: "flagset-c",
					CommonFlagSet: config.CommonFlagSet{
						Retriever:       &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
						PollingInterval: 0,
					},
					APIKeys: []string{"key-c"},
				},
			},
		}
		manager, logs := setupManager(t, cfg)

		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				// flagset-a: unchanged
				{
					Name: "flagset-a",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"key-a"},
				},
				// flagset-b: removed (absent)
				// flagset-c: forbidden modification (polling interval changed)
				{
					Name: "flagset-c",
					CommonFlagSet: config.CommonFlagSet{
						Retriever:       &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
						PollingInterval: 5000,
					},
					APIKeys: []string{"key-c"},
				},
				// flagset-d: added
				{
					Name: "flagset-d",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"key-d"},
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// a unchanged, b removed, c kept (modification rejected), d added.
		_, err := manager.FlagSet("key-a")
		assert.NoError(t, err, "flagset-a should be unchanged")
		_, err = manager.FlagSet("key-b")
		assert.Error(t, err, "flagset-b should have been removed")
		_, err = manager.FlagSet("key-c")
		assert.NoError(t, err, "flagset-c should be kept because its modification was rejected")
		_, err = manager.FlagSet("key-d")
		assert.NoError(t, err, "flagset-d should have been added")

		all, err := manager.AllFlagSets()
		assert.NoError(t, err)
		assert.Len(t, all, 3, "expected flagset-a, flagset-c and flagset-d")

		// Only flagset-c's modification is rejected.
		assert.Equal(t, 1, logs.FilterMessageSnippet("modifying a flagset is not supported at runtime").Len())
	})

	t.Run("should leave unnamed flagsets untouched and warn when their count changes", func(t *testing.T) {
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-named",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"named-key"},
				},
				{
					Name: "", // unnamed -> gets a generated name at startup
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"unnamed-key"},
				},
			},
		}
		obs, logs := observer.New(zap.WarnLevel)
		logger := zap.New(obs)
		manager, err := service.NewFlagsetManager(cfg, logger, nil, nil)
		require.NoError(t, err)
		t.Cleanup(func() { manager.Close() })

		// New config drops the unnamed flagset (which cannot be live-removed).
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "flagset-named",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"named-key"},
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// The unnamed flagset is left running (not removed) and a warning is logged.
		flagset, err := manager.FlagSet("unnamed-key")
		assert.NoError(t, err, "the unnamed flagset must keep working until a restart")
		assert.NotNil(t, flagset)

		warnLogs := logs.FilterMessageSnippet("without an explicit name is not supported at runtime")
		assert.Equal(t, 1, warnLogs.Len(), "expected a warning about unnamed flagsets")
	})

	t.Run("should fail closed when an accepted key move collides with a rejected modification", func(t *testing.T) {
		// fs-a gains "k-move" (accepted) while fs-b — which still holds "k-move" — also gets a
		// forbidden retriever change (rejected, so it keeps its old keys). The reconciled config
		// then has "k-move" on BOTH flagsets even though the incoming config was valid. The key
		// must not route non-deterministically: it is left unrouted (fail closed).
		otherFlagConfig := "../../../testdata/flag-config.yaml"
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "fs-a",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"k-a"},
				},
				{
					Name: "fs-b",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"k-b", "k-move"},
				},
			},
		}
		manager, logs := setupManager(t, cfg)

		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "fs-a",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"k-a", "k-move"}, // gains k-move (accepted)
				},
				{
					Name: "fs-b",
					CommonFlagSet: config.CommonFlagSet{
						// Forbidden retriever change -> the whole fs-b change is rejected, so it
						// keeps its old keys [k-b, k-move].
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: otherFlagConfig},
					},
					APIKeys: []string{"k-b"},
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// Non-colliding keys still route deterministically.
		_, err := manager.FlagSet("k-a")
		assert.NoError(t, err)
		_, err = manager.FlagSet("k-b")
		assert.NoError(t, err)
		// The colliding key is failed closed: it routes to no flagset.
		_, err = manager.FlagSet("k-move")
		assert.Error(t, err, "a key present on two flagsets after reconciliation must not route")

		assert.GreaterOrEqual(t,
			logs.FilterMessageSnippet("configured on multiple flagsets after reconciliation").Len(), 1,
			"expected an error log about the colliding API key")
		assert.GreaterOrEqual(t,
			logs.FilterMessageSnippet("modifying a flagset is not supported at runtime").Len(), 1,
			"expected the forbidden modification of fs-b to be rejected")
	})

	t.Run("should skip a flagset added at runtime whose client cannot be created", func(t *testing.T) {
		cfg := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "fs-keep",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"keep-key"},
				},
			},
		}
		manager, logs := setupManager(t, cfg)

		// The added flagset references a file that does not exist: the config is valid (the kind
		// is "file") but NewGoFeatureFlagClient fails when it tries to load the flags.
		newConfig := &config.Config{
			FlagSets: []config.FlagSet{
				{
					Name: "fs-keep",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					},
					APIKeys: []string{"keep-key"},
				},
				{
					Name: "fs-broken",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &retrieverconf.RetrieverConf{
							Kind: "file",
							Path: "/does/not/exist/flags.yaml",
						},
					},
					APIKeys: []string{"broken-key"},
				},
			},
		}

		manager.OnConfigChange(newConfig)

		// The broken flagset is not registered; the existing flagset keeps working.
		_, err := manager.FlagSet("broken-key")
		assert.Error(t, err, "a flagset whose client failed to build must not be reachable")
		_, err = manager.FlagSet("keep-key")
		assert.NoError(t, err, "the pre-existing flagset must keep working")

		all, err := manager.AllFlagSets()
		assert.NoError(t, err)
		assert.Len(t, all, 1)

		assert.GreaterOrEqual(t,
			logs.FilterMessageSnippet("failed to create the flagset added at runtime").Len(), 1,
			"expected an error log about the failed flagset creation")
	})
}

// TestFlagsetManager_ConcurrentReadDuringReconcile exercises the flagsetsMutex: readers hit
// FlagSet/FlagSetName/AllFlagSets while a writer repeatedly adds and removes a flagset via
// OnConfigChange. It is meant to run under -race; the stable flagset must always resolve and
// nothing may panic.
func TestFlagsetManager_ConcurrentReadDuringReconcile(t *testing.T) {
	flagConfig := "../../../testdata/flag-config.yaml"
	stable := config.FlagSet{
		Name: "fs-stable",
		CommonFlagSet: config.CommonFlagSet{
			Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
		},
		APIKeys: []string{"stable-key"},
	}
	churn := config.FlagSet{
		Name: "fs-churn",
		CommonFlagSet: config.CommonFlagSet{
			Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
		},
		APIKeys: []string{"churn-key"},
	}
	manager, _ := setupManager(t, &config.Config{FlagSets: []config.FlagSet{stable, churn}})

	withChurn := &config.Config{FlagSets: []config.FlagSet{stable, churn}}
	withoutChurn := &config.Config{FlagSets: []config.FlagSet{stable}}

	done := make(chan struct{})
	var stableErrors atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					if _, err := manager.FlagSet("stable-key"); err != nil {
						stableErrors.Add(1)
					}
					_, _ = manager.AllFlagSets()
					_, _ = manager.FlagSetName("churn-key")
				}
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			if i%2 == 0 {
				manager.OnConfigChange(withoutChurn)
			} else {
				manager.OnConfigChange(withChurn)
			}
		}
		close(done)
	}()

	wg.Wait()

	// fs-stable is present in every configuration, so its key must never fail to resolve.
	assert.Equal(t, int64(0), stableErrors.Load(),
		"the stable flagset must resolve throughout concurrent add/remove of another flagset")
}

// TestFlagsetManager_RemoveFlagsetGracefullyCloses verifies that removing a flagset at runtime
// gracefully closes its GoFeatureFlag client: the data exporter is stopped and its buffered
// events are flushed (no data loss) instead of being dropped.
func TestFlagsetManager_RemoveFlagsetGracefullyCloses(t *testing.T) {
	flagConfig := "../../../testdata/flag-config.yaml"

	// Webhook server that records how many times the exporter flushes to it.
	var exportCalls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		exportCalls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		FlagSets: []config.FlagSet{
			{
				Name: "flagset-to-remove",
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
					// Bulk webhook exporter with a huge flush interval/buffer so it only flushes
					// when the flagset is closed (and not earlier during the test).
					Exporter: &config.ExporterConf{
						Kind:             config.WebhookExporter,
						EndpointURL:      server.URL,
						FlushInterval:    600000,
						MaxEventInMemory: 1000000,
					},
				},
				APIKeys: []string{"key-remove"},
			},
			{
				// Keeper flagset so that removing the other one is not a mode switch.
				Name: "flagset-keep",
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
				},
				APIKeys: []string{"key-keep"},
			},
		},
	}
	manager, _ := setupManager(t, cfg)

	// Capture the client of the flagset we are going to remove.
	all, err := manager.AllFlagSets()
	require.NoError(t, err)
	client := all["flagset-to-remove"]
	require.NotNil(t, client)

	// Generate one evaluation event; it is buffered (bulk exporter, not flushed yet).
	_, err = client.RawVariation("test-flag", ffcontext.NewEvaluationContext("user-key"), false)
	require.NoError(t, err)
	require.Equal(t, int64(0), exportCalls.Load(),
		"the event must still be buffered before the flagset is removed")

	// Remove the flagset at runtime (keeping flagset-keep so the mode does not change).
	newConfig := &config.Config{
		FlagSets: []config.FlagSet{
			{
				Name: "flagset-keep",
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{Kind: "file", Path: flagConfig},
				},
				APIKeys: []string{"key-keep"},
			},
		},
	}
	manager.OnConfigChange(newConfig)

	// The removed flagset is no longer reachable...
	_, err = manager.FlagSet("key-remove")
	assert.Error(t, err)

	// ...and its exporter was flushed during the graceful close, so the buffered event was
	// exported (no data loss). Close() flushes synchronously, so the call has completed here.
	assert.GreaterOrEqual(t, exportCalls.Load(), int64(1),
		"removing a flagset must gracefully close it and flush its buffered events")
}
