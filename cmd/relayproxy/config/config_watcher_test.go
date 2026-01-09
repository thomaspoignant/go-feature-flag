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
)

const (
	testConfigFileName = "goff-proxy.yaml"
	configFlagName     = "config"
	configFlagUsage    = "Location of your config file"
	configFlagPrefix   = "--config="
)

// syncFile ensures the file is written to disk before returning.
// This is important for file watchers that might detect changes before the file is fully written.
// Without syncing, the file watcher might detect the change before the OS has flushed the write,
// causing the reload to read stale or empty file content.
func syncFile(filePath string) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err == nil {
		file.Sync()
		file.Close()
	}
}

func createTestConfig(t *testing.T, configContent string) (*config.Config, string) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, testConfigFileName)

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	f := pflag.NewFlagSet("config", pflag.ContinueOnError)
	f.String(configFlagName, "", configFlagUsage)
	err = f.Parse([]string{configFlagPrefix + configFile})
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
	err := os.WriteFile(configFile, []byte(updatedConfig), 0644)
	require.NoError(t, err)
	// Sync the file to ensure it's written to disk before the watcher reads it
	syncFile(configFile)

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

				err = cfg.StopConfigChangeWatcher()
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

	err := os.WriteFile(configFile, []byte(updatedConfig), 0644)
	require.NoError(t, err)
	// Sync the file to ensure it's written to disk before the watcher reads it
	syncFile(configFile)

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

				err = cfg.StopConfigChangeWatcher()
				assert.NoError(t, err)
				return
			}
			mu.Unlock()
		}
	}
}
