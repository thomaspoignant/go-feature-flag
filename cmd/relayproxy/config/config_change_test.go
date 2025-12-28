package config_test

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"go.uber.org/zap"
)

func TestConfigChangeDefaultMode(t *testing.T) {
	t.Run("change authorized keys from test to test2", func(t *testing.T) {
		file, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer func() {
			_ = file.Close()
			_ = os.Remove(file.Name())
		}()

		configContent := `server:
  port: 41031
  mode: http
retrievers:
  - kind: file
    path: ../../../testdata/flag-config.yaml
authorizedKeys:
  evaluation:
    - test
`
		err = os.WriteFile(file.Name(), []byte(configContent), 0644)
		require.NoError(t, err)

		f := pflag.NewFlagSet("test", pflag.ContinueOnError)
		f.String("config", file.Name(), "config file")

		c, err := config.New(f, zap.NewNop(), "vTest")
		require.NoError(t, err)

		flagsetManager, err := service.NewFlagsetManager(c, zap.NewNop(), nil)
		require.NoError(t, err)
		defer flagsetManager.Close()

		services := service.Services{
			MonitoringService: service.NewMonitoring(flagsetManager),
			WebsocketService:  service.NewWebsocketService(),
			FlagsetManager:    flagsetManager,
			Metrics:           metric.Metrics{},
		}

		// attach a callback to the config to be called when the configuration changes
		callbackCalled := make(chan bool, 1)
		callback := func(newConfig *config.Config) {
			callbackCalled <- true
		}
		c.AttachConfigChangeCallback(callback)

		s := api.New(c, services, zap.NewNop())
		go func() { s.StartWithContext(context.Background()) }()

		// Should have a 401 response without the correct API Keys
		body := `{"evaluationContext":{"key":"08b5ffb7-7109-42f4-a6f2-b85560fbd20f"}}`
		response, err := http.Post("http://localhost:41031/v1/allflags", "application/json", strings.NewReader(body))
		require.NoError(t, err)
		defer func() { _ = response.Body.Close() }()
		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)

		// Should have a 200 response with the correct API Keys
		request, err := http.NewRequest("POST", "http://localhost:41031/v1/allflags", strings.NewReader(body))
		require.NoError(t, err)
		defer func() { _ = request.Body.Close() }()
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-API-Key", "test")
		response2, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, response2.StatusCode)

		configContent = `server:
  port: 41031
  mode: http
retrievers:
  - kind: file
    path: ../../../testdata/flag-config.yaml
authorizedKeys:
  evaluation:
    - test2
`

		err = os.WriteFile(file.Name(), []byte(configContent), 0644)
		require.NoError(t, err)

		// wait for the callback to be called or error out after 500 milliseconds
		select {
		case <-callbackCalled:
			// Callback was called, continue with test
		case <-time.After(500 * time.Millisecond):
			assert.Fail(t, "Timeout waiting for callback to be called")
		}

		response3, err := http.Post("http://localhost:41031/v1/allflags", "application/json", strings.NewReader(body))
		require.NoError(t, err)
		defer func() { _ = response3.Body.Close() }()
		assert.Equal(t, http.StatusUnauthorized, response3.StatusCode)
	})

	t.Run("remove authorized keys should allow all requests", func(t *testing.T) {
		file, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer func() {
			_ = file.Close()
			_ = os.Remove(file.Name())
		}()

		configContent := `server:
  port: 41032
  mode: http
retrievers:
  - kind: file
    path: ../../../testdata/flag-config.yaml
authorizedKeys:
  evaluation:
    - test
`
		err = os.WriteFile(file.Name(), []byte(configContent), 0644)
		require.NoError(t, err)

		f := pflag.NewFlagSet("test", pflag.ContinueOnError)
		f.String("config", file.Name(), "config file")

		c, err := config.New(f, zap.NewNop(), "vTest")
		require.NoError(t, err)

		callbackCalled := make(chan bool, 1)
		callback := func(newConfig *config.Config) {
			callbackCalled <- true
		}
		c.AttachConfigChangeCallback(callback)

		flagsetManager, err := service.NewFlagsetManager(c, zap.NewNop(), nil)
		require.NoError(t, err)
		defer flagsetManager.Close()

		services := service.Services{
			MonitoringService: service.NewMonitoring(flagsetManager),
			WebsocketService:  service.NewWebsocketService(),
			FlagsetManager:    flagsetManager,
			Metrics:           metric.Metrics{},
		}

		s := api.New(c, services, zap.NewNop())
		go func() { s.StartWithContext(context.Background()) }()
		time.Sleep(100 * time.Millisecond)
		defer s.Stop(context.Background())

		// Should have a 200 response with the correct API Keys
		body := `{"evaluationContext":{"key":"08b5ffb7-7109-42f4-a6f2-b85560fbd20f"}}`
		request, err := http.NewRequest("POST", "http://localhost:41032/v1/allflags", strings.NewReader(body))
		require.NoError(t, err)
		defer func() { _ = request.Body.Close() }()
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-API-Key", "test")
		response2, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, response2.StatusCode)

		configContent = `server:
  port: 41032
  mode: http
retrievers:
  - kind: file
    path: ../../../testdata/flag-config.yaml
`
		err = os.WriteFile(file.Name(), []byte(configContent), 0644)
		require.NoError(t, err)

		// wait for the callback to be called or error out after 500 milliseconds
		select {
		case <-callbackCalled:
			// Callback was called, continue with test
		case <-time.After(500 * time.Millisecond):
			assert.Fail(t, "Timeout waiting for callback to be called")
		}

		response3, err := http.Post("http://localhost:41032/v1/allflags", "application/json", strings.NewReader(body))
		require.NoError(t, err)
		defer func() { _ = response3.Body.Close() }()
		assert.Equal(t, http.StatusOK, response3.StatusCode)
	})
}
