package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

func TestXXX(t *testing.T) {
	type test struct {
		name               string
		MonitoringPort     int
		Debug              bool
		expectedStatusCode int
	}
	tests := []test{
		{
			name:               "pprof available in proxy port",
			Debug:              true,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "pprof available in monitoring port",
			Debug:              true,
			MonitoringPort:     46000,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "pprof not available ii debug not enabled",
			Debug:              false,
			MonitoringPort:     46000,
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z, err := zap.NewProduction()
			require.NoError(t, err)
			c := &config.Config{
				Retriever: &config.RetrieverConf{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
				MonitoringPort: tt.MonitoringPort,
				ListenPort:     46001,
				Debug:          tt.Debug,
			}

			goff, err := service.NewGoFeatureFlagClient(c, z, []notifier.Notifier{})
			require.NoError(t, err)
			apiServer := api.New(c, service.Services{
				MonitoringService:    service.NewMonitoring(goff),
				WebsocketService:     service.NewWebsocketService(),
				GOFeatureFlagService: goff,
				Metrics:              metric.Metrics{},
			}, z)

			portToCheck := c.ListenPort
			if tt.MonitoringPort != 0 {
				portToCheck = tt.MonitoringPort
			}

			go apiServer.Start()
			defer apiServer.Stop()
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/debug/pprof/heap", portToCheck))
			require.NoError(t, err)
			require.Equal(t, tt.expectedStatusCode, resp.StatusCode)
		})
	}
}
