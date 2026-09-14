package api_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/testutils"
)

// testConfig returns a minimal HTTP relay proxy configuration listening on a free port.
func testConfig(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		Server: config.Server{
			Mode: config.ServerModeHTTP,
			Port: testutils.GetFreePort(t),
		},
	}
}

// startTestServer boots a relay proxy from conf and returns its base URL. When a monitoring
// port is configured it also waits for the monitoring server to accept connections, because
// it is started in its own goroutine and serves the endpoints the API server no longer has.
func startTestServer(t *testing.T, conf *config.Config) string {
	t.Helper()
	log := newTestLogger(t)

	metricsV2, err := metric.NewMetrics()
	require.NoError(t, err)
	wsService := stream.NewWebsocketService()
	t.Cleanup(wsService.Close)
	flagsetManager, err := service.NewFlagsetManager(conf, log.ZapLogger, nil, nil)
	require.NoError(t, err)

	s := api.New(conf, service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}, log.ZapLogger)
	go func() { s.StartWithContext(context.Background()) }()
	t.Cleanup(func() { s.Stop(context.Background()) })

	baseURL := fmt.Sprintf("http://localhost:%d", conf.ServerPort(log.ZapLogger))
	if monitoringPort := conf.EffectiveMonitoringPort(log.ZapLogger); monitoringPort != 0 {
		waitForServer(t, fmt.Sprintf("http://localhost:%d", monitoringPort))
	}
	waitForServer(t, baseURL)
	return baseURL
}

// doRequest performs a request with the given headers and returns the response, the body being
// closed for the caller since every assertion here is on headers and status code.
func doRequest(t *testing.T, method, url string, headers map[string]string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	require.NoError(t, err)
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

// Test_CORS_SimpleRequest pins the behaviour of the default echo CORS middleware on the API
// server: any origin is allowed and no CORS header leaks on a request without an Origin.
func Test_CORS_SimpleRequest(t *testing.T) {
	baseURL := startTestServer(t, testConfig(t))

	tests := []struct {
		name     string
		endpoint string
	}{
		{name: "health endpoint", endpoint: "/health"},
		{name: "evaluation endpoint", endpoint: "/v1/flag/change"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("with Origin header", func(t *testing.T) {
				resp := doRequest(t, http.MethodGet, baseURL+tt.endpoint, map[string]string{
					"Origin": "https://example.com",
				})
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
				assert.Contains(t, resp.Header.Values("Vary"), "Origin")
				assert.Empty(t, resp.Header.Get("Access-Control-Allow-Credentials"))
				assert.Empty(t, resp.Header.Get("Access-Control-Expose-Headers"))
			})

			t.Run("without Origin header", func(t *testing.T) {
				resp := doRequest(t, http.MethodGet, baseURL+tt.endpoint, nil)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
			})
		})
	}
}

// Test_CORS_Preflight checks that a browser preflight is answered by the CORS middleware
// before the authentication middleware runs, otherwise no browser could ever call an
// authenticated endpoint of the relay proxy.
func Test_CORS_Preflight(t *testing.T) {
	conf := testConfig(t)
	conf.AuthorizedKeys = config.APIKeys{Evaluation: []string{"my-key"}}
	conf.ForceReloadAPIKeys()
	baseURL := startTestServer(t, conf)

	endpoints := []string{
		"/health",
		"/v1/allflags",
		"/v1/feature/my-flag/eval",
		"/ofrep/v1/evaluate/flags",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			resp := doRequest(t, http.MethodOptions, baseURL+endpoint, map[string]string{
				"Origin":                         "https://example.com",
				"Access-Control-Request-Method":  http.MethodPost,
				"Access-Control-Request-Headers": "content-type,authorization",
			})
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
			assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
			assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), http.MethodPost)
			assert.Equal(t,
				"content-type,authorization",
				resp.Header.Get("Access-Control-Allow-Headers"))
			// MaxAge is not configured, so browsers are never told to cache the preflight.
			assert.Empty(t, resp.Header.Get("Access-Control-Max-Age"))
		})
	}
}

// Test_CORS_MonitoringPort checks that the CORS middleware is also active on the monitoring
// server when it runs on a dedicated port.
func Test_CORS_MonitoringPort(t *testing.T) {
	conf := testConfig(t)
	conf.Server.MonitoringPort = testutils.GetFreePort(t)
	startTestServer(t, conf)
	monitoringURL := fmt.Sprintf("http://localhost:%d", conf.Server.MonitoringPort)

	resp := doRequest(t, http.MethodGet, monitoringURL+"/health", map[string]string{
		"Origin": "https://example.com",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, resp.Header.Values("Vary"), "Origin")
}
