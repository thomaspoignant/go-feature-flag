package controller_test

import (
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_XXX(t *testing.T) {
	wsS := service.NewWebsocketService()
	log := zap.L()
	fr := controller.NewFlagReload(wsS, log)

	// Create a new Echo instance
	e := echo.New()

	// Define the WebSocket route
	e.GET("/websocket", fr.Handler)

	// Create a test server with the Echo instance
	server := httptest.NewServer(e)
	defer server.Close()

	// Convert the server URL to WebSocket URL
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/websocket"

	// Create a WebSocket connection
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer ws.Close()

	// Write a message to the WebSocket connection
	wsS.BroadcastFlagChanges(notifier.DiffCache{
		Deleted: nil,
		Added:   nil,
		Updated: nil,
	})
	message := "Hello, WebSocket!"
	err = ws.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		t.Fatalf("Failed to write message to WebSocket: %v", err)
	}

	// Read the response from the WebSocket connection
	_, receivedMessage, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message from WebSocket: %v", err)
	}

	// Check if the received message matches the expected message
	expectedMessage := "Hello, Client!"
	if string(receivedMessage) != expectedMessage {
		t.Errorf("Received message %q, expected %q", receivedMessage, expectedMessage)
	}
}
