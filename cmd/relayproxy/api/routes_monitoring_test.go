package api_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/notifier"
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
				MonitoringPort: tt.MonitoringPort,
				ListenPort:     1031,
				EnablePprof:    tt.EnablePprof,
			}

			flagsetManager, err := service.NewFlagsetManager(c, z, []notifier.Notifier{})
			require.NoError(t, err)
			apiServer := api.New(c, service.Services{
				MonitoringService: service.NewMonitoring(flagsetManager),
				WebsocketService:  service.NewWebsocketService(),
				FlagsetManager:    flagsetManager,
				Metrics:           metric.Metrics{},
			}, z)

			portToCheck := c.ListenPort
			if tt.MonitoringPort != 0 {
				portToCheck = tt.MonitoringPort
			}

			go apiServer.StartWithContext(context.TODO())
			defer apiServer.Stop(context.Background())
			time.Sleep(1 * time.Second) // waiting for the apiServer to start
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/debug/pprof/heap", portToCheck))
			require.NoError(t, err)
			require.Equal(t, tt.expectedStatusCode, resp.StatusCode)
		})
	}
}
