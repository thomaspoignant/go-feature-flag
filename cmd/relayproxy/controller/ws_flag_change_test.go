package controller_test

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"go.uber.org/zap"
	"net/http/httptest"
	"strings"
	"testing"
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
					"my-flag": notifier.DiffUpdated{
						Before: &flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"A": testconvert.Interface(true),
								"B": testconvert.Interface(false),
							},
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("A"),
							},
						},
						After: &flag.InternalFlag{
							Variations: &map[string]*interface{}{
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
						Variations: &map[string]*interface{}{
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
						Variations: &map[string]*interface{}{
							"A": testconvert.Interface(true),
							"B": testconvert.Interface(false),
						},
						DefaultRule: &flag.Rule{
							VariationResult: testconvert.String("A"),
						},
					},
				},
				Updated: map[string]notifier.DiffUpdated{
					"my-flag": notifier.DiffUpdated{
						Before: &flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"A": testconvert.Interface(true),
								"B": testconvert.Interface(false),
							},
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("A"),
							},
						},
						After: &flag.InternalFlag{
							Variations: &map[string]*interface{}{
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
			websocketService := service.NewWebsocketService()
			defer websocketService.Close()
			log := zap.L()
			ctrl := controller.NewWsFlagChange(websocketService, log)

			e := echo.New()
			e.GET("/ws/v1/flag/change", ctrl.Handler)
			testServer := httptest.NewServer(e)
			defer testServer.Close()
			url := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws/v1/flag/change"
			ws, _, err := websocket.DefaultDialer.Dial(url, nil)
			if err != nil {
				t.Fatalf("Failed to connect to WebSocket: %v", err)
			}
			defer func() { _ = ws.Close() }()
			websocketService.BroadcastFlagChanges(tt.flagChange)
			_, receivedMessage, err := ws.ReadMessage()
			assert.NoError(t, err)

			expectedMessage, err := json.Marshal(tt.flagChange)
			assert.NoError(t, err)
			assert.JSONEq(t, string(expectedMessage), string(receivedMessage))
		})
	}
}
