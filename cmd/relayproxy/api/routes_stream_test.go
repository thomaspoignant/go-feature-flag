package api_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/websocket"
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

	flagsetManager, err := service.NewFlagsetManager(c, z, []notifier.Notifier{}, nil, nil)
	require.NoError(t, err)

	sseService := stream.NewSSEService()
	defer sseService.Close()
	apiServer := api.New(c, service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  stream.NewWebsocketService(),
		SSEService:        sseService,
		FlagsetManager:    flagsetManager,
		Metrics:           metric.Metrics{},
	}, z)

	go apiServer.StartWithContext(context.Background())
	defer apiServer.Stop(context.Background())
	waitForServer(t, fmt.Sprintf("http://localhost:%d", c.ServerPort(z)))

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ws/v1/flag/change", c.ServerPort(z)))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "true", resp.Header.Get("Deprecation"))
	assert.Equal(t, `</stream/v1/ws/flag/change>; rel="successor-version"`, resp.Header.Get("Link"))
}

// writeFlagFile writes a flag file containing a single flag, value is used to make
// every write a flag change.
func writeFlagFile(t *testing.T, path, flagName, value string) {
	t.Helper()
	content := fmt.Sprintf(`%s:
  variations:
    enabled: %q
    disabled: "disabled"
  defaultRule:
    variation: enabled
`, flagName, value)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}

// readWSFrames reads the text frames of the connection in the background.
// The channel is closed when the connection is closed.
func readWSFrames(ws *websocket.Conn) <-chan string {
	frames := make(chan string, 16)
	go func() {
		defer close(frames)
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				return
			}
			frames <- string(msg)
		}
	}()
	return frames
}

// changeFlagUntilFrame changes the flag file until a frame is received, and returns it.
// Retrying (instead of sleeping one polling interval) covers a change that is picked up
// before the websocket connection is registered.
func changeFlagUntilFrame(t *testing.T, path, flagName string, frames <-chan string) string {
	t.Helper()
	deadline := time.After(20 * time.Second)
	for {
		writeFlagFile(t, path, flagName, fmt.Sprintf("value-%d", time.Now().UnixNano()))
		select {
		case frame, ok := <-frames:
			require.True(t, ok, "websocket connection closed before receiving a frame")
			return frame
		case <-deadline:
			t.Fatalf("no websocket frame received after changing %s", flagName)
		case <-time.After(3 * time.Second):
		}
	}
}

func TestWebsocketFlagChangeWithFlagsets(t *testing.T) {
	dir := t.TempDir()
	tenantAFile := filepath.Join(dir, "tenant-a.yaml")
	tenantBFile := filepath.Join(dir, "tenant-b.yaml")
	writeFlagFile(t, tenantAFile, "flag-tenant-a", "initial")
	writeFlagFile(t, tenantBFile, "flag-tenant-b", "initial")

	flagset := func(name, apiKey, path string) config.FlagSet {
		return config.FlagSet{
			Name:    name,
			APIKeys: []string{apiKey},
			CommonFlagSet: config.CommonFlagSet{
				Retrievers:      &[]retrieverconf.RetrieverConf{{Kind: "file", Path: path}},
				PollingInterval: 200,
			},
		}
	}
	c := &config.Config{
		FlagSets: []config.FlagSet{
			flagset("tenant-a", "tenant-A-api-key", tenantAFile),
			flagset("tenant-b", "tenant-B-api-key", tenantBFile),
		},
		AuthorizedKeys: config.APIKeys{Admin: []string{"admin-api-key"}},
		Server: config.Server{
			Mode: config.ServerModeHTTP,
			Port: testutils.GetFreePort(t),
		},
	}
	c.ForceReloadAPIKeys()
	z := zap.NewNop()

	wsService := stream.NewWebsocketService()
	defer wsService.Close()
	flagsetManager, err := service.NewFlagsetManager(c, z, []notifier.Notifier{}, nil, wsService)
	require.NoError(t, err)
	defer flagsetManager.Close()

	apiServer := api.New(c, service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metric.Metrics{},
	}, z)
	go apiServer.StartWithContext(context.Background())
	defer apiServer.Stop(context.Background())
	waitForServer(t, fmt.Sprintf("http://localhost:%d", c.ServerPort(z)))
	wsURL := fmt.Sprintf("ws://localhost:%d", c.ServerPort(z))

	endpoints := []struct {
		name string
		path string
	}{
		{name: "stream endpoint", path: "/stream/v1/ws/flag/change"},
		{name: "legacy alias", path: "/ws/v1/flag/change"},
	}

	for _, tt := range endpoints {
		t.Run(tt.name+": changes are only sent to the clients of the flagset", func(t *testing.T) {
			dial := func(apiKey string) <-chan string {
				ws, resp, err := websocket.DefaultDialer.Dial(wsURL+tt.path+"?apiKey="+apiKey, nil)
				require.NoError(t, err)
				_ = resp.Body.Close()
				t.Cleanup(func() { _ = ws.Close() })
				return readWSFrames(ws)
			}
			tenantAFrames := dial("tenant-A-api-key")
			tenantBFrames := dial("tenant-B-api-key")

			// The tenant-b client receives the change of its flagset.
			frame := changeFlagUntilFrame(t, tenantBFile, "flag-tenant-b", tenantBFrames)
			assert.Contains(t, frame, "flag-tenant-b")
			assert.NotContains(t, frame, "flag-tenant-a")

			// The tenant-b change has been broadcast: if it had been sent to the tenant-a
			// client, it would be the first frame read here, before the tenant-a change.
			frame = changeFlagUntilFrame(t, tenantAFile, "flag-tenant-a", tenantAFrames)
			assert.Contains(t, frame, "flag-tenant-a")
			assert.NotContains(t, frame, "flag-tenant-b")
		})
	}

	rejected := []struct {
		name   string
		apiKey string
		want   int
	}{
		{name: "key not associated to a flagset", apiKey: "admin-api-key", want: http.StatusBadRequest},
		{name: "unknown key", apiKey: "unknown-api-key", want: http.StatusUnauthorized},
	}
	for _, endpoint := range endpoints {
		for _, tt := range rejected {
			t.Run(endpoint.name+": "+tt.name+" is rejected before the upgrade", func(t *testing.T) {
				ws, resp, err := websocket.DefaultDialer.Dial(wsURL+endpoint.path+"?apiKey="+tt.apiKey, nil)
				if ws != nil {
					_ = ws.Close()
				}
				require.ErrorIs(t, err, websocket.ErrBadHandshake)
				defer func() { _ = resp.Body.Close() }()
				assert.Equal(t, tt.want, resp.StatusCode)
			})
		}
	}
}
