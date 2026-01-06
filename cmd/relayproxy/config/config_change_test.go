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
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/helper"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"go.uber.org/zap"
)

func TestConfigChangeDefaultMode(t *testing.T) {
	t.Run("change authorized keys from test to test2", func(t *testing.T) {
		const urlAPIAllFlags = "http://localhost:41031/v1/allflags"
		configFile := testutils.CopyFileToNewTempFile(t, "../testdata/config/change-authorized-keys-from-test-to-test2.yaml")
		defer func() {
			_ = os.Remove(configFile.Name())
		}()
		callbackCalled := make(chan bool, 1)
		s := newAPIServer(t, configFile, func(newConfig *config.Config) {
			callbackCalled <- true
		})
		defer s.Stop(context.Background())
		time.Sleep(100 * time.Millisecond) // wait for the server to start

		// Should have a 401 response without the correct API Keys
		body := `{"evaluationContext":{"key":"08b5ffb7-7109-42f4-a6f2-b85560fbd20f"}}`
		response, err := http.Post(urlAPIAllFlags, helper.ContentTypeValueJSON, strings.NewReader(body))
		require.NoError(t, err)
		defer func() { _ = response.Body.Close() }()
		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)

		// Should have a 200 response with the correct API Keys
		request, err := http.NewRequest("POST", urlAPIAllFlags, strings.NewReader(body))
		require.NoError(t, err)
		defer func() { _ = request.Body.Close() }()
		request.Header.Set(helper.ContentTypeHeader, helper.ContentTypeValueJSON)
		request.Header.Set(helper.XAPIKeyHeader, "test")
		response2, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, response2.StatusCode)

		_ = testutils.CopyFileToExistingTempFile(t, "../testdata/config/change-authorized-keys-from-test-to-test2-MODIFIED.yaml", configFile)

		// wait for the callback to be called or error out after 500 milliseconds
		select {
		case <-callbackCalled:
			// Callback was called, continue with test
			time.Sleep(100 * time.Millisecond)
		case <-time.After(500 * time.Millisecond):
			require.Fail(t, "Timeout waiting for callback to be called")
		}
		// After reload, the old API key "test" should be invalid.
		requestOld, err := http.NewRequest("POST", urlAPIAllFlags, strings.NewReader(body))
		require.NoError(t, err)
		requestOld.Header.Set(helper.ContentTypeHeader, helper.ContentTypeValueJSON)
		requestOld.Header.Set(helper.XAPIKeyHeader, "test")
		responseOld, err := http.DefaultClient.Do(requestOld)
		require.NoError(t, err)
		defer func() { _ = responseOld.Body.Close() }()
		assert.Equal(t, http.StatusUnauthorized, responseOld.StatusCode)

		// The new API key "test2" should now be valid.
		requestNew, err := http.NewRequest("POST", urlAPIAllFlags, strings.NewReader(body))
		require.NoError(t, err)
		requestNew.Header.Set(helper.ContentTypeHeader, helper.ContentTypeValueJSON)
		requestNew.Header.Set(helper.XAPIKeyHeader, "test2")
		responseNew, err := http.DefaultClient.Do(requestNew)
		require.NoError(t, err)
		defer func() { _ = responseNew.Body.Close() }()
		assert.Equal(t, http.StatusOK, responseNew.StatusCode)
	})

	t.Run("remove authorized keys should allow all requests", func(t *testing.T) {
		const urlAPIAllFlags = "http://localhost:41032/v1/allflags"
		configFile := testutils.CopyFileToNewTempFile(t, "../testdata/config/remove-authorized-keys-should-allow-all-requests.yaml")
		defer func() {
			_ = os.Remove(configFile.Name())
		}()

		callbackCalled := make(chan bool, 1)
		s := newAPIServer(t, configFile, func(newConfig *config.Config) {
			callbackCalled <- true
		})
		defer s.Stop(context.Background())
		time.Sleep(100 * time.Millisecond) // wait for the server to start

		// Should have a 200 response with the correct API Keys
		body := `{"evaluationContext":{"key":"08b5ffb7-7109-42f4-a6f2-b85560fbd20f"}}`
		request, err := http.NewRequest("POST", urlAPIAllFlags, strings.NewReader(body))
		require.NoError(t, err)
		defer func() { _ = request.Body.Close() }()
		request.Header.Set(helper.ContentTypeHeader, helper.ContentTypeValueJSON)
		request.Header.Set("X-API-Key", "test")
		response2, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, response2.StatusCode)
		_ = testutils.CopyFileToExistingTempFile(t, "../testdata/config/remove-authorized-keys-should-allow-all-requests-MODIFIED.yaml", configFile)
		// wait for the callback to be called or error out after 500 milliseconds
		select {
		case <-callbackCalled:
			time.Sleep(100 * time.Millisecond)
			// Callback was called, continue with test
		case <-time.After(500 * time.Millisecond):
			assert.Fail(t, "Timeout waiting for callback to be called")
		}

		response3, err := http.Post(urlAPIAllFlags, helper.ContentTypeValueJSON, strings.NewReader(body))
		require.NoError(t, err)
		defer func() { _ = response3.Body.Close() }()
		assert.Equal(t, http.StatusOK, response3.StatusCode)
	})
}

func newAPIServer(t *testing.T, configFile *os.File, callback func(newConfig *config.Config)) api.Server {
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	f.String("config", "", "Location of your config file")
	err := f.Parse([]string{"--config=" + configFile.Name()})
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	c, err := config.New(f, logger, "toto")
	require.NoError(t, err)

	flagsetManager, err := service.NewFlagsetManager(c, zap.NewNop(), nil)
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  service.NewWebsocketService(),
		FlagsetManager:    flagsetManager,
		Metrics:           metric.Metrics{},
	}

	c.AttachConfigChangeCallback(callback)

	s := api.New(c, services, zap.NewNop())
	go func() { s.StartWithContext(context.Background()) }()
	return s
}
