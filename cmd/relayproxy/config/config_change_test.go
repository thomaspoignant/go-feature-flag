package config_test

import (
	"context"
	"io"
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
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

const (
	timeoutCallbackMsg = "Timeout waiting for callback to be called"
	apiKey1            = "apikey-1"
	apiKey2            = "apikey-2"
	testFlagName       = "test-flag"
	fooFlagName        = "foo-flag"
	localhostURL       = "http://localhost:"
	allFlagsEndpoint   = "/v1/allflags"
	phaseBeforeChange  = "before change"
	phaseAfterChange   = "after change"
)

// httpRequestCheck defines a check to perform on an HTTP request
type httpRequestCheck struct {
	apiKey          string   // API key to use in the request (empty means no API key header)
	expectedStatus  int      // expected HTTP status code
	bodyContains    []string // strings that should be in the response body (optional)
	bodyNotContains []string // strings that should NOT be in the response body (optional)
}

func TestConfigChangeDefaultMode(t *testing.T) {
	tests := []struct {
		name               string
		port               string
		initialConfigFile  string
		modifiedConfigFile string
		checksBeforeChange []httpRequestCheck
		checksAfterChange  []httpRequestCheck
	}{
		{
			name:               "change authorized keys from test to test2",
			port:               "41031",
			initialConfigFile:  "../testdata/config/change-authorized-keys-from-test-to-test2.yaml",
			modifiedConfigFile: "../testdata/config/change-authorized-keys-from-test-to-test2-MODIFIED.yaml",
			checksBeforeChange: []httpRequestCheck{
				{apiKey: "", expectedStatus: http.StatusUnauthorized},
				{apiKey: "test", expectedStatus: http.StatusOK},
			},
			checksAfterChange: []httpRequestCheck{
				{apiKey: "test", expectedStatus: http.StatusUnauthorized},
				{apiKey: "test2", expectedStatus: http.StatusOK},
			},
		},
		{
			name:               "remove authorized keys should allow all requests",
			port:               "41032",
			initialConfigFile:  "../testdata/config/remove-authorized-keys-should-allow-all-requests.yaml",
			modifiedConfigFile: "../testdata/config/remove-authorized-keys-should-allow-all-requests-MODIFIED.yaml",
			checksBeforeChange: []httpRequestCheck{
				{apiKey: "test", expectedStatus: http.StatusOK},
			},
			checksAfterChange: []httpRequestCheck{
				{apiKey: "", expectedStatus: http.StatusOK},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlAPIAllFlags := localhostURL + tt.port + allFlagsEndpoint
			configFile := testutils.CopyFileToNewTempFile(t, tt.initialConfigFile)
			defer func() {
				_ = configFile.Close()
				_ = os.Remove(configFile.Name())
			}()

			callbackCalled := make(chan bool, 1)
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)
			s := newAPIServerWithLogger(t, configFile, logger, func(newConfig *config.Config) {
				callbackCalled <- true
			})
			defer s.Stop(context.Background())
			time.Sleep(100 * time.Millisecond) // wait for the server to start

			body := `{"evaluationContext":{"key":"08b5ffb7-7109-42f4-a6f2-b85560fbd20f"}}`

			// Run checks before config change
			for _, check := range tt.checksBeforeChange {
				doHTTPRequestAndCheck(t, urlAPIAllFlags, body, check, phaseBeforeChange)
			}

			// Modify the config file
			_ = testutils.CopyFileToExistingTempFile(t, tt.modifiedConfigFile, configFile)

			// Wait for the callback to be called
			select {
			case <-callbackCalled:
				time.Sleep(100 * time.Millisecond)
			case <-time.After(500 * time.Millisecond):
				require.Fail(t, timeoutCallbackMsg)
			}

			// Run checks after config change
			for _, check := range tt.checksAfterChange {
				doHTTPRequestAndCheck(t, urlAPIAllFlags, body, check, phaseAfterChange)
			}
		})
	}
}

func TestConfigChangeFlagsetModeAPIKeyChanges(t *testing.T) {
	tests := []struct {
		name               string
		port               string
		initialConfigFile  string
		modifiedConfigFile string
		checksBeforeChange []httpRequestCheck
		checksAfterChange  []httpRequestCheck
	}{
		{
			name:               "change API key from apikey-1 to apikey-2",
			port:               "41033",
			initialConfigFile:  "../testdata/config/flagset-change-api-key-from-key1-to-key2.yaml",
			modifiedConfigFile: "../testdata/config/flagset-change-api-key-from-key1-to-key2-MODIFIED.yaml",
			checksBeforeChange: []httpRequestCheck{
				{apiKey: apiKey1, expectedStatus: http.StatusOK},
				{apiKey: apiKey2, expectedStatus: http.StatusUnauthorized},
			},
			checksAfterChange: []httpRequestCheck{
				{apiKey: apiKey1, expectedStatus: http.StatusUnauthorized},
				{apiKey: apiKey2, expectedStatus: http.StatusOK},
			},
		},
		{
			name:               "add a second API key to flagset",
			port:               "41034",
			initialConfigFile:  "../testdata/config/flagset-add-api-key.yaml",
			modifiedConfigFile: "../testdata/config/flagset-add-api-key-MODIFIED.yaml",
			checksBeforeChange: []httpRequestCheck{
				{apiKey: apiKey1, expectedStatus: http.StatusOK},
				{apiKey: apiKey2, expectedStatus: http.StatusUnauthorized},
			},
			checksAfterChange: []httpRequestCheck{
				{apiKey: apiKey1, expectedStatus: http.StatusOK},
				{apiKey: apiKey2, expectedStatus: http.StatusOK},
			},
		},
		{
			name:               "move API key from one flagset to another",
			port:               "41035",
			initialConfigFile:  "../testdata/config/flagset-move-api-key-between-flagsets.yaml",
			modifiedConfigFile: "../testdata/config/flagset-move-api-key-between-flagsets-MODIFIED.yaml",
			checksBeforeChange: []httpRequestCheck{
				{
					apiKey:          apiKey1,
					expectedStatus:  http.StatusOK,
					bodyContains:    []string{testFlagName},
					bodyNotContains: []string{fooFlagName},
				},
			},
			checksAfterChange: []httpRequestCheck{
				{
					apiKey:          apiKey1,
					expectedStatus:  http.StatusOK,
					bodyContains:    []string{fooFlagName},
					bodyNotContains: []string{testFlagName},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlAPIAllFlags := localhostURL + tt.port + allFlagsEndpoint
			configFile := testutils.CopyFileToNewTempFile(t, tt.initialConfigFile)
			defer func() {
				_ = configFile.Close()
				_ = os.Remove(configFile.Name())
			}()

			callbackCalled := make(chan bool, 1)
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)
			s := newAPIServerWithLogger(t, configFile, logger, func(newConfig *config.Config) {
				callbackCalled <- true
			})
			defer s.Stop(context.Background())
			time.Sleep(100 * time.Millisecond) // wait for the server to start

			body := `{"evaluationContext":{"key":"08b5ffb7-7109-42f4-a6f2-b85560fbd20d"}}`

			// Run checks before config change
			for _, check := range tt.checksBeforeChange {
				doHTTPRequestAndCheck(t, urlAPIAllFlags, body, check, phaseBeforeChange)
			}

			// Modify the config file
			_ = testutils.CopyFileToExistingTempFile(t, tt.modifiedConfigFile, configFile)

			// Wait for the callback to be called
			select {
			case <-callbackCalled:
				time.Sleep(100 * time.Millisecond)
			case <-time.After(500 * time.Millisecond):
				require.Fail(t, timeoutCallbackMsg)
			}

			// Run checks after config change
			for _, check := range tt.checksAfterChange {
				doHTTPRequestAndCheck(t, urlAPIAllFlags, body, check, phaseAfterChange)
			}
		})
	}
}
func TestConfigChangeFlagsetInvalidChanges(t *testing.T) {
	// Tests for invalid config changes that should be rejected
	invalidConfigTests := []struct {
		name               string
		port               string
		initialConfigFile  string
		modifiedConfigFile string
		checksBeforeChange []httpRequestCheck
		checksAfterChange  []httpRequestCheck // should be same as before (config rejected)
	}{
		{
			name:               "duplicate API key in config should be rejected",
			port:               "41036",
			initialConfigFile:  "../testdata/config/flagset-duplicate-api-key.yaml",
			modifiedConfigFile: "../testdata/config/flagset-duplicate-api-key-MODIFIED.yaml",
			checksBeforeChange: []httpRequestCheck{
				{apiKey: apiKey1, expectedStatus: http.StatusOK, bodyContains: []string{testFlagName}},
				{apiKey: apiKey2, expectedStatus: http.StatusOK},
			},
			checksAfterChange: []httpRequestCheck{
				{apiKey: apiKey1, expectedStatus: http.StatusOK, bodyContains: []string{testFlagName}},
				{apiKey: apiKey2, expectedStatus: http.StatusOK},
			},
		},
		{
			name:               "flagset with no API key should be rejected",
			port:               "41037",
			initialConfigFile:  "../testdata/config/flagset-empty-api-key.yaml",
			modifiedConfigFile: "../testdata/config/flagset-empty-api-key-MODIFIED.yaml",
			checksBeforeChange: []httpRequestCheck{
				{apiKey: apiKey1, expectedStatus: http.StatusOK},
				{apiKey: apiKey2, expectedStatus: http.StatusOK},
			},
			checksAfterChange: []httpRequestCheck{
				{apiKey: apiKey1, expectedStatus: http.StatusOK},
				{apiKey: apiKey2, expectedStatus: http.StatusOK}, // should still work, config rejected
			},
		},
	}

	for _, tt := range invalidConfigTests {
		t.Run(tt.name, func(t *testing.T) {
			urlAPIAllFlags := localhostURL + tt.port + allFlagsEndpoint
			configFile := testutils.CopyFileToNewTempFile(t, tt.initialConfigFile)
			defer func() {
				_ = configFile.Close()
				_ = os.Remove(configFile.Name())
			}()

			callbackCalled := make(chan bool, 1)
			// Create observed logger to capture error logs
			core, observedLogs := observer.New(zapcore.ErrorLevel)
			observedLogger := zap.New(core)
			s := newAPIServerWithLogger(t, configFile, observedLogger, func(newConfig *config.Config) {
				callbackCalled <- true
			})
			defer s.Stop(context.Background())
			time.Sleep(100 * time.Millisecond) // wait for the server to start

			body := `{"evaluationContext":{"key":"08b5ffb7-7109-42f4-a6f2-b85560fbd20f"}}`

			// Run checks before config change
			for _, check := range tt.checksBeforeChange {
				doHTTPRequestAndCheck(t, urlAPIAllFlags, body, check, phaseBeforeChange)
			}

			// Modify the config to create an invalid config
			_ = testutils.CopyFileToExistingTempFile(t, tt.modifiedConfigFile, configFile)

			// The callback should be called even for invalid configs (validation happens after)
			select {
			case <-callbackCalled:
				time.Sleep(100 * time.Millisecond)
			case <-time.After(500 * time.Millisecond):
				require.Fail(t, timeoutCallbackMsg)
			}

			// Verify that the error log was displayed for the invalid config
			assert.GreaterOrEqual(t, observedLogs.Len(), 1, "Expected at least one error log message")
			errorLogs := observedLogs.FilterMessage("the new configuration is invalid, it will not be applied")
			assert.GreaterOrEqual(t, errorLogs.Len(), 1, "Expected error log about invalid configuration")

			// After the invalid config is rejected, the original config should still work
			for _, check := range tt.checksAfterChange {
				doHTTPRequestAndCheck(t, urlAPIAllFlags, body, check, phaseAfterChange)
			}
		})
	}
}

func newAPIServerWithLogger(
	t *testing.T,
	configFile *os.File,
	logger *zap.Logger,
	callback func(newConfig *config.Config),
) api.Server {
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	f.String("config", "", "Location of your config file")
	err := f.Parse([]string{"--config=" + configFile.Name()})
	require.NoError(t, err)

	c, err := config.New(f, logger, "toto")
	require.NoError(t, err)

	flagsetManager, err := service.NewFlagsetManager(c, logger, nil)
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  service.NewWebsocketService(),
		FlagsetManager:    flagsetManager,
		Metrics:           metric.Metrics{},
	}

	c.AttachConfigChangeCallback(callback)

	s := api.New(c, services, logger)
	go func() { s.StartWithContext(context.Background()) }()
	return s
}

// doHTTPRequestAndCheck performs an HTTP POST request and verifies the response against expected checks
func doHTTPRequestAndCheck(t *testing.T, url, body string, check httpRequestCheck, phase string) {
	t.Helper()
	request, err := http.NewRequest("POST", url, strings.NewReader(body))
	require.NoError(t, err)
	request.Header.Set(helper.ContentTypeHeader, helper.ContentTypeValueJSON)
	if check.apiKey != "" {
		request.Header.Set(helper.XAPIKeyHeader, check.apiKey)
	}
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()

	assert.Equal(t, check.expectedStatus, response.StatusCode,
		"%s: expected status %d for apiKey=%q, got %d",
		phase, check.expectedStatus, check.apiKey, response.StatusCode)

	// Check body content if needed
	if len(check.bodyContains) > 0 || len(check.bodyNotContains) > 0 {
		responseBody, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		bodyStr := string(responseBody)

		for _, expected := range check.bodyContains {
			assert.Contains(t, bodyStr, expected,
				"%s: expected body to contain %q for apiKey=%q", phase, expected, check.apiKey)
		}
		for _, notExpected := range check.bodyNotContains {
			assert.NotContains(t, bodyStr, notExpected,
				"%s: expected body NOT to contain %q for apiKey=%q", phase, notExpected, check.apiKey)
		}
	}
}
