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
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"go.uber.org/zap"
)

func TestPprofEndpointsStarts(t *testing.T) {
	type test struct {
		name               string
		MonitoringPort     int
		EnablePprof        bool
		expectedStatusCode int
	}
	tests := []test{
		{
			name:               "pprof available in proxy port",
			EnablePprof:        true,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "pprof available in monitoring port",
			EnablePprof:        true,
			MonitoringPort:     1032,
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z, err := zap.NewProduction()
			require.NoError(t, err)
			c := &config.Config{
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
				MonitoringPort: tt.MonitoringPort,
				EnablePprof:    tt.EnablePprof,
			}

			flagsetManager, err := service.NewFlagsetManager(c, z, []notifier.Notifier{}, nil)
			require.NoError(t, err)
			apiServer := api.New(c, service.Services{
				MonitoringService: service.NewMonitoring(flagsetManager),
				WebsocketService:  stream.NewWebsocketService(),
				FlagsetManager:    flagsetManager,
				Metrics:           metric.Metrics{},
			}, z)

			portToCheck := c.ServerPort(z)
			if tt.MonitoringPort != 0 {
				portToCheck = tt.MonitoringPort
			}

			go apiServer.StartWithContext(context.Background())
			defer apiServer.Stop(context.Background())
			waitForServer(t, fmt.Sprintf("http://localhost:%d", portToCheck))
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/debug/pprof/heap", portToCheck))
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()
			require.Equal(t, tt.expectedStatusCode, resp.StatusCode)
		})
	}
}

// Test_VersionHeader_On_MonitoringServer checks that the version header middleware is attached
// to the monitoring server when it runs on a dedicated port, and that it honours
// disableVersionHeader on both servers.
func Test_VersionHeader_On_MonitoringServer(t *testing.T) {
	tests := []struct {
		name                 string
		disableVersionHeader bool
		expectedVersion      string
	}{
		{
			name:                 "version header enabled",
			disableVersionHeader: false,
			expectedVersion:      "1.2.3",
		},
		{
			name:                 "version header disabled",
			disableVersionHeader: true,
			expectedVersion:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := testConfig(t)
			conf.Version = "1.2.3"
			conf.DisableVersionHeader = tt.disableVersionHeader
			conf.Server.MonitoringPort = testutils.GetFreePort(t)
			baseURL := startTestServer(t, conf)
			monitoringURL := fmt.Sprintf("http://localhost:%d", conf.Server.MonitoringPort)

			monitoringResp := doRequest(t, http.MethodGet, monitoringURL+"/health", nil)
			assert.Equal(t, http.StatusOK, monitoringResp.StatusCode)
			assert.Equal(t,
				tt.expectedVersion,
				monitoringResp.Header.Get("X-GOFEATUREFLAG-VERSION"),
				"monitoring server")

			apiResp := doRequest(t, http.MethodGet, baseURL+"/v1/flag/change", nil)
			assert.Equal(t, http.StatusOK, apiResp.StatusCode)
			assert.Equal(t,
				tt.expectedVersion,
				apiResp.Header.Get("X-GOFEATUREFLAG-VERSION"),
				"api server")
		})
	}
}
