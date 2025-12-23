package config_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestConfigLoaderAddConfigChangeCallback(t *testing.T) {
	tests := []struct {
		name          string
		initialConfig string
		updatedConfig string
		wantCallbacks int
		setupCallback func(*testing.T, *config.ConfigLoader) (*[]*config.Config, *sync.Mutex)
	}{
		{
			name: "Single callback triggered on config change",
			initialConfig: `server:
  port: 1031
pollingInterval: 1000
loglevel: info`,
			updatedConfig: `server:
  port: 2031
pollingInterval: 2000
loglevel: debug`,
			wantCallbacks: 1,
			setupCallback: func(t *testing.T, loader *config.ConfigLoader) (*[]*config.Config, *sync.Mutex) {
				var mu sync.Mutex
				receivedConfigs := make([]*config.Config, 0)

				loader.AddConfigChangeCallback(func(newConfig *config.Config) {
					mu.Lock()
					defer mu.Unlock()
					receivedConfigs = append(receivedConfigs, newConfig)
				})

				return &receivedConfigs, &mu
			},
		},
		{
			name: "Multiple callbacks triggered on config change",
			initialConfig: `server:
  port: 1031
pollingInterval: 1000
loglevel: info`,
			updatedConfig: `server:
  port: 3031
pollingInterval: 3000
loglevel: warn`,
			wantCallbacks: 2,
			setupCallback: func(t *testing.T, loader *config.ConfigLoader) (*[]*config.Config, *sync.Mutex) {
				var mu sync.Mutex
				receivedConfigs := make([]*config.Config, 0)

				loader.AddConfigChangeCallback(func(newConfig *config.Config) {
					mu.Lock()
					defer mu.Unlock()
					receivedConfigs = append(receivedConfigs, newConfig)
				})

				loader.AddConfigChangeCallback(func(newConfig *config.Config) {
					mu.Lock()
					defer mu.Unlock()
					receivedConfigs = append(receivedConfigs, newConfig)
				})

				return &receivedConfigs, &mu
			},
		},
		{
			name: "Callback receives updated config values",
			initialConfig: `server:
  port: 1031
pollingInterval: 1000
loglevel: info`,
			updatedConfig: `server:
  port: 9999
pollingInterval: 5000
loglevel: error`,
			wantCallbacks: 1,
			setupCallback: func(t *testing.T, loader *config.ConfigLoader) (*[]*config.Config, *sync.Mutex) {
				var mu sync.Mutex
				receivedConfigs := make([]*config.Config, 0)

				loader.AddConfigChangeCallback(func(newConfig *config.Config) {
					mu.Lock()
					defer mu.Unlock()
					receivedConfigs = append(receivedConfigs, newConfig)
				})

				return &receivedConfigs, &mu
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory and config file
			tempDir := t.TempDir()
			configFile := filepath.Join(tempDir, "goff-proxy.yaml")

			// Write initial config
			err := os.WriteFile(configFile, []byte(tt.initialConfig), 0644)
			require.NoError(t, err)

			// Create flag set and parse config file path
			f := pflag.NewFlagSet("config", pflag.ContinueOnError)
			f.String("config", "", "Location of your config file")
			err = f.Parse([]string{"--config=" + configFile})
			require.NoError(t, err)

			// Create ConfigLoader with watchChanges enabled
			logger := zap.NewNop()
			loader := config.NewConfigLoader(f, logger, "1.0.0", true)

			// Setup callbacks
			receivedConfigs, mu := tt.setupCallback(t, loader)

			// Get initial config to verify it's loaded correctly
			initialConfig, err := loader.ToConfig()
			require.NoError(t, err)
			assert.Equal(t, 1031, initialConfig.Server.Port)

			// Wait a bit to ensure watcher is set up
			time.Sleep(100 * time.Millisecond)

			// Update config file to trigger callback
			err = os.WriteFile(configFile, []byte(tt.updatedConfig), 0644)
			require.NoError(t, err)

			// Wait for callback to be triggered (file watchers may have some delay)
			timeout := time.After(5 * time.Second)
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-timeout:
					t.Fatal("Timeout waiting for callback to be triggered")
				case <-ticker.C:
					mu.Lock()
					callbackCount := len(*receivedConfigs)
					mu.Unlock()

					if callbackCount >= tt.wantCallbacks {
						// Verify callback was called with correct config
						mu.Lock()
						defer mu.Unlock()

						assert.Len(t, *receivedConfigs, tt.wantCallbacks, "Expected %d callbacks, got %d", tt.wantCallbacks, callbackCount)

						// Verify the last received config has updated values
						lastConfig := (*receivedConfigs)[len(*receivedConfigs)-1]
						// Verify that the config changed from initial values
						assert.NotEqual(t, 1031, lastConfig.Server.Port, "Port should have changed from initial value")
						assert.NotEqual(t, 1000, lastConfig.PollingInterval, "PollingInterval should have changed from initial value")
						assert.NotEqual(t, "info", lastConfig.LogLevel, "LogLevel should have changed from initial value")

						// Clean up
						err = loader.StopWatchChanges()
						assert.NoError(t, err)
						return
					}
				}
			}
		})
	}
}

func TestConfigLoaderAddConfigChangeCallbackNoWatchChanges(t *testing.T) {
	// Create temporary directory and config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "goff-proxy.yaml")

	initialConfig := `server:
  port: 1031
pollingInterval: 1000
loglevel: info`

	// Write initial config
	err := os.WriteFile(configFile, []byte(initialConfig), 0644)
	require.NoError(t, err)

	// Create flag set and parse config file path
	f := pflag.NewFlagSet("config", pflag.ContinueOnError)
	f.String("config", "", "Location of your config file")
	err = f.Parse([]string{"--config=" + configFile})
	require.NoError(t, err)

	// Create ConfigLoader with watchChanges disabled
	logger := zap.NewNop()
	loader := config.NewConfigLoader(f, logger, "1.0.0", false)

	// Setup callback
	var mu sync.Mutex
	var receivedConfigs []*config.Config
	loader.AddConfigChangeCallback(func(newConfig *config.Config) {
		mu.Lock()
		defer mu.Unlock()
		receivedConfigs = append(receivedConfigs, newConfig)
	})

	// Update config file
	updatedConfig := `server:
  port: 9999
pollingInterval: 5000
loglevel: debug`
	err = os.WriteFile(configFile, []byte(updatedConfig), 0644)
	require.NoError(t, err)

	// Wait a bit to ensure no callback is triggered
	time.Sleep(1 * time.Second)

	// Verify callback was NOT called
	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, receivedConfigs, 0, "Callback should not be triggered when watchChanges is disabled")
}

func TestConfigLoaderAddConfigChangeCallbackInvalidConfig(t *testing.T) {
	// Create temporary directory and config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "goff-proxy.yaml")

	initialConfig := `server:
  port: 1031
pollingInterval: 1000
loglevel: info`

	// Write initial config
	err := os.WriteFile(configFile, []byte(initialConfig), 0644)
	require.NoError(t, err)

	// Create flag set and parse config file path
	f := pflag.NewFlagSet("config", pflag.ContinueOnError)
	f.String("config", "", "Location of your config file")
	err = f.Parse([]string{"--config=" + configFile})
	require.NoError(t, err)

	// Create ConfigLoader with watchChanges enabled and observer logger to capture logs
	core, logs := observer.New(zapcore.ErrorLevel)
	logger := zap.New(core)
	loader := config.NewConfigLoader(f, logger, "1.0.0", true)

	// Setup callback
	var mu sync.Mutex
	var callbackCalled bool
	loader.AddConfigChangeCallback(func(newConfig *config.Config) {
		mu.Lock()
		defer mu.Unlock()
		callbackCalled = true
	})

	// Write invalid config file
	invalidConfig := `server:
  port: invalid
pollingInterval: 1000
loglevel: info`
	err = os.WriteFile(configFile, []byte(invalidConfig), 0644)
	require.NoError(t, err)

	// Wait a bit for the file watcher to detect the change and attempt to load the config
	time.Sleep(1 * time.Second)

	// Verify callback was NOT called due to invalid config
	mu.Lock()
	callbackNotCalled := !callbackCalled
	mu.Unlock()
	assert.True(t, callbackNotCalled, "Callback should not be triggered when new config is invalid")

	// Verify that an error log was written
	assert.GreaterOrEqual(t, logs.Len(), 1, "Expected at least one error log when config is invalid")
	errorLogs := logs.FilterMessage("error loading new config")
	assert.GreaterOrEqual(t, errorLogs.Len(), 1, "Expected error log with message 'error loading new config'")

	// Clean up
	err = loader.StopWatchChanges()
	assert.NoError(t, err)
}
