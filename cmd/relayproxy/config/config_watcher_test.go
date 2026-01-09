package config_test

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"go.uber.org/zap"
)

const (
	configFlagName   = "config"
	configFlagUsage  = "Location of your config file"
	configFlagPrefix = "--config="
)

func createTestConfig(t *testing.T, configContent string) (*config.Config, *os.File) {
	configFile := testutils.CopyContentToNewTempFile(t, configContent)
	f := pflag.NewFlagSet("config", pflag.ContinueOnError)
	f.String(configFlagName, "", configFlagUsage)
	err := f.Parse([]string{configFlagPrefix + configFile.Name()})
	require.NoError(t, err)

	cfg, err := config.New(f, zap.NewNop(), "1.0.0")
	require.NoError(t, err)
	return cfg, configFile
}

func TestConfigAttachConfigChangeCallback(t *testing.T) {
	initialConfig := `server:
  port: 1031
pollingInterval: 1000
loglevel: info`

	updatedConfig := `server:
  port: 2031
pollingInterval: 2000
loglevel: debug`

	cfg, configFile := createTestConfig(t, initialConfig)

	var mu sync.Mutex
	var callbackCalled bool
	var receivedConfig *config.Config

	callback := func(newConfig *config.Config) {
		mu.Lock()
		defer mu.Unlock()
		callbackCalled = true
		receivedConfig = newConfig
	}

	cfg.AttachConfigChangeCallback(callback)

	// Update config file
	_ = testutils.CopyContentToExistingTempFile(t, updatedConfig, configFile)

	// Wait for callback to be triggered
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for callback to be triggered")
		case <-ticker.C:
			mu.Lock()
			if callbackCalled {
				mu.Unlock()
				assert.NotNil(t, receivedConfig, "Received config should not be nil")
				assert.Equal(t, 2031, receivedConfig.Server.Port, "Port should be updated")
				assert.Equal(t, 2000, receivedConfig.PollingInterval, "PollingInterval should be updated")
				assert.Equal(t, "debug", receivedConfig.LogLevel, "LogLevel should be updated")

				err := cfg.StopConfigChangeWatcher()
				assert.NoError(t, err)
				return
			}
			mu.Unlock()
		}
	}
}

func TestConfigStopConfigChangeWatcher(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func(*testing.T) *config.Config
		wantErr     bool
	}{
		{
			name: "Successfully stop watcher when configLoader is initialized",
			setupConfig: func(t *testing.T) *config.Config {
				cfg, _ := createTestConfig(t, `server:
  port: 1031
pollingInterval: 1000
loglevel: info`)
				return cfg
			},
			wantErr: false,
		},
		{
			name: "Stop watcher when configLoader is nil - should return nil",
			setupConfig: func(t *testing.T) *config.Config {
				return &config.Config{}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig(t)
			err := cfg.StopConfigChangeWatcher()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigAttachConfigChangeCallbackMultipleCallbacks(t *testing.T) {
	initialConfig := `server:
  port: 1031
pollingInterval: 1000
loglevel: info`

	updatedConfig := `server:
  port: 9999
pollingInterval: 5000
loglevel: error`

	cfg, configFile := createTestConfig(t, initialConfig)

	var mu sync.Mutex
	var callback1Called, callback2Called bool

	callback1 := func(newConfig *config.Config) {
		mu.Lock()
		defer mu.Unlock()
		callback1Called = true
	}

	callback2 := func(newConfig *config.Config) {
		mu.Lock()
		defer mu.Unlock()
		callback2Called = true
	}

	cfg.AttachConfigChangeCallback(callback1)
	cfg.AttachConfigChangeCallback(callback2)

	_ = testutils.CopyContentToExistingTempFile(t, updatedConfig, configFile)

	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for callbacks to be triggered")
		case <-ticker.C:
			mu.Lock()
			if callback1Called && callback2Called {
				mu.Unlock()
				assert.True(t, callback1Called, "First callback should be called")
				assert.True(t, callback2Called, "Second callback should be called")

				err := cfg.StopConfigChangeWatcher()
				assert.NoError(t, err)
				return
			}
			mu.Unlock()
		}
	}
}
