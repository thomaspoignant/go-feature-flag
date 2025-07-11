package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
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
					Retriever: &config.RetrieverConf{
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
							Retriever: &config.RetrieverConf{
								Kind: "file",
								Path: flagConfig,
							},
						},
						ApiKeys: []string{"test-api-key"},
					},
				},
			},
			logger:    zap.NewNop(),
			notifiers: []notifier.Notifier{},
			wantErr:   false,
		},
		{
			name: "invalid flagsets should fallback to default manager",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{
						Name: "invalid-flagset",
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &config.RetrieverConf{
								Kind: "invalid",
								Path: flagConfig,
							},
						},
						ApiKeys: []string{"test-api-key"},
					},
				},
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &config.RetrieverConf{
						Kind: "file",
						Path: flagConfig,
					},
				},
			},
			logger:    zap.NewNop(),
			notifiers: []notifier.Notifier{},
			wantErr:   false,
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
						Retriever: &config.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					ApiKeys: []string{"test-api-key"},
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
	})

	// Test default mode
	t.Run("default mode", func(t *testing.T) {
		config := &config.Config{
			FlagSets: []config.FlagSet{},
			CommonFlagSet: config.CommonFlagSet{
				Retriever: &config.RetrieverConf{
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
						Retriever: &config.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					ApiKeys: []string{"test-api-key"},
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
				Retriever: &config.RetrieverConf{
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
						Retriever: &config.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					ApiKeys: []string{"api-key-1"},
				},
				{
					Name: "test-flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &config.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					ApiKeys: []string{"api-key-2"},
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
						Retriever: &config.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					ApiKeys: []string{"api-key-1"},
				},
				{
					Name: "test-flagset-2",
					CommonFlagSet: config.CommonFlagSet{
						Retriever: &config.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					ApiKeys: []string{"api-key-2"},
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
				Retriever: &config.RetrieverConf{
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
				Retriever: &config.RetrieverConf{
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
						Retriever: &config.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					ApiKeys: []string{"test-api-key"},
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
				Retriever: &config.RetrieverConf{
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
						Retriever: &config.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					ApiKeys: []string{"test-api-key"},
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
				Retriever: &config.RetrieverConf{
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
						Retriever: &config.RetrieverConf{
							Kind: "file",
							Path: flagConfig,
						},
					},
					ApiKeys: []string{"test-api-key"},
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
