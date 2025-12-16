package controller_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
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

			websocketService := service.NewWebsocketService()
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
