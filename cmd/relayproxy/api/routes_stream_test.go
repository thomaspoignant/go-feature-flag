package api_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

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

func TestDeprecatedAliasHeaders(t *testing.T) {
	z, err := zap.NewProduction()
	require.NoError(t, err)

	c := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{Kind: "file", Path: "../../../testdata/flag-config.yaml"},
			},
		},
		Server: config.Server{
			Mode: config.ServerModeHTTP,
			Port: testutils.GetFreePort(t),
		},
	}

	flagsetManager, err := service.NewFlagsetManager(c, z, []notifier.Notifier{}, nil)
	require.NoError(t, err)

	apiServer := api.New(c, service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  stream.NewWebsocketService(),
		SSEService:        stream.NewSSEService(),
		FlagsetManager:    flagsetManager,
		Metrics:           metric.Metrics{},
	}, z)

	go apiServer.StartWithContext(context.Background())
	defer apiServer.Stop(context.Background())
	time.Sleep(1 * time.Second)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ws/v1/flag/change", c.ServerPort(z)))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "true", resp.Header.Get("Deprecation"))
	assert.Equal(t, `</stream/v1/ws/flag/change>; rel="successor-version"`, resp.Header.Get("Link"))
}
