package controller_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	controller "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/goff"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

func Test_websocket_flag_change(t *testing.T) {
	tests := []struct {
		name       string
		flagChange notifier.DiffCache
	}{
		{
			name: "Update 1 flag",
			flagChange: notifier.DiffCache{
				Deleted: nil,
				Added:   nil,
				Updated: map[string]notifier.DiffUpdated{
					"my-flag": {
						Before: &flag.InternalFlag{
							Variations: &map[string]*any{
								"A": testconvert.Interface(true),
								"B": testconvert.Interface(false),
							},
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("A"),
							},
						},
						After: &flag.InternalFlag{
							Variations: &map[string]*any{
								"A": testconvert.Interface(true),
								"B": testconvert.Interface(false),
							},
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("B"),
							},
						},
					},
				},
			},
		},
		{
			name: "Update remove and add flag at the same time",
			flagChange: notifier.DiffCache{
				Deleted: map[string]flag.Flag{
					"flag-1": &flag.InternalFlag{
						Variations: &map[string]*any{
							"A": testconvert.Interface(true),
							"B": testconvert.Interface(false),
						},
						DefaultRule: &flag.Rule{
							VariationResult: testconvert.String("A"),
						},
					},
				},
				Added: map[string]flag.Flag{
					"flag-2": &flag.InternalFlag{
						Variations: &map[string]*any{
							"A": testconvert.Interface(true),
							"B": testconvert.Interface(false),
						},
						DefaultRule: &flag.Rule{
							VariationResult: testconvert.String("A"),
						},
					},
				},
				Updated: map[string]notifier.DiffUpdated{
					"my-flag": {
						Before: &flag.InternalFlag{
							Variations: &map[string]*any{
								"A": testconvert.Interface(true),
								"B": testconvert.Interface(false),
							},
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("A"),
							},
						},
						After: &flag.InternalFlag{
							Variations: &map[string]*any{
								"A": testconvert.Interface(true),
								"B": testconvert.Interface(false),
							},
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("B"),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context with timeout for the entire test
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			websocketService := stream.NewWebsocketService()
			defer func() {
				websocketService.Close()
				// Wait for cleanup to complete to avoid leaking goroutines in tests.
				if err := websocketService.WaitForCleanup(5 * time.Second); err != nil {
					t.Errorf("websocket service cleanup failed: %v", err)
				}
			}()

			log := zap.L()
			ctrl := controller.NewWsFlagChange(websocketService, log)

			e := echo.New()
			e.GET("/ws/v1/flag/change", ctrl.Handler)
			testServer := httptest.NewServer(e)
			defer testServer.Close()

			url := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws/v1/flag/change"

			// Create websocket connection with timeout
			dialer := &websocket.Dialer{
				HandshakeTimeout: 10 * time.Second,
			}
			ws, _, err := dialer.DialContext(ctx, url, nil)
			if err != nil {
				t.Fatalf("Failed to connect to WebSocket: %v", err)
			}
			defer func() {
				ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
				_ = ws.Close()
			}()

			// Set read deadline to prevent hanging
			err = ws.SetReadDeadline(time.Now().Add(10 * time.Second))
			assert.NoError(t, err)

			// Wait a short time to ensure WebSocket registration is complete
			// This prevents the race condition where BroadcastFlagChanges is called
			// before the WebSocket connection is fully registered in the service
			time.Sleep(100 * time.Millisecond)

			// Broadcast the flag change
			websocketService.BroadcastFlagChanges(tt.flagChange)

			// Read message with timeout
			_, receivedMessage, err := ws.ReadMessage()
			assert.NoError(t, err)

			expectedMessage, err := json.Marshal(tt.flagChange)
			assert.NoError(t, err)
			assert.JSONEq(t, string(expectedMessage), string(receivedMessage))
		})
	}
}

// Test_websocket_concurrent_writes is a regression test for issue #5463.
// The ping loop and the flag-change broadcast both write to the same websocket
// connection. Without synchronization gorilla/websocket panics with
// "concurrent write to websocket connection" and crashes the relay proxy.
// Run with -race to surface the data race if the writes are not serialized.
func Test_websocket_concurrent_writes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	websocketService := stream.NewWebsocketService()
	defer func() {
		websocketService.Close()
		if err := websocketService.WaitForCleanup(5 * time.Second); err != nil {
			t.Errorf("websocket service cleanup failed: %v", err)
		}
	}()

	ctrl := controller.NewWsFlagChange(websocketService, zap.L())

	e := echo.New()
	e.GET("/ws/v1/flag/change", ctrl.Handler)
	testServer := httptest.NewServer(e)
	defer testServer.Close()

	url := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws/v1/flag/change"

	dialer := &websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	ws, _, err := dialer.DialContext(ctx, url, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer func() {
		_ = ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_ = ws.Close()
	}()

	// Drain incoming messages so the server-side writes can make progress.
	clientDone := make(chan struct{})
	go func() {
		defer close(clientDone)
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// Wait for the registration to complete to avoid broadcasting before the
	// connection is registered in the service.
	time.Sleep(100 * time.Millisecond)

	diff := notifier.DiffCache{
		Updated: map[string]notifier.DiffUpdated{
			"my-flag": {
				Before: &flag.InternalFlag{
					DefaultRule: &flag.Rule{VariationResult: testconvert.String("A")},
				},
				After: &flag.InternalFlag{
					DefaultRule: &flag.Rule{VariationResult: testconvert.String("B")},
				},
			},
		},
	}

	// Fire many broadcasts concurrently while the 1s ping loop is running.
	// This reproduces both concurrency vectors: broadcast vs ping and
	// broadcast vs broadcast (BroadcastFlagChanges only holds an RLock).
	const broadcasters = 8
	const broadcastsEach = 50
	var wg sync.WaitGroup
	for range broadcasters {
		wg.Go(func() {
			for range broadcastsEach {
				websocketService.BroadcastFlagChanges(diff)
			}
		})
	}
	wg.Wait()

	// Close the connection and make sure the client goroutine stops.
	_ = ws.Close()
	select {
	case <-clientDone:
	case <-time.After(5 * time.Second):
		t.Fatal("client read loop did not stop")
	}
}
