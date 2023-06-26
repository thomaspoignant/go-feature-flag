package controller

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// NewWsFlagChange is the constructor to create a new controller to handle websocket
// request to be notified about flag changes.
func NewWsFlagChange(websocketService service.WebsocketService, logger *zap.Logger) Controller {
	return &wsFlagChange{
		websocketService: websocketService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		logger: logger,
	}
}

// wsFlagChange is the implementation of the controller
type wsFlagChange struct {
	websocketService service.WebsocketService
	upgrader         websocket.Upgrader
	logger           *zap.Logger
}

// TODO: add doc for swagger
func (f *wsFlagChange) Handler(c echo.Context) error {
	conn, err := f.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	f.websocketService.Register(conn)
	f.logger.Debug("registering new websocket connection", zap.Any("connection", conn))

	// Start the ping pong loop
	go f.pingPongLoop(conn)
	isOpen := true
	for isOpen {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			f.websocketService.Deregister(conn)
			f.logger.Debug("closing websocket connection", zap.Error(err), zap.Any("connection", conn))
			isOpen = false
		}
		// TODO: remove line bellow
		fmt.Println(string(msg))
	}
	return nil
}

// pingPongLoop is a keep-alive call to the client.
// It calls the client to ensure that the connection is still active.
// If the ping is not working we are closing the session.
func (f *wsFlagChange) pingPongLoop(conn *websocket.Conn) {
	// Ping interval duration
	pingInterval := 10 * time.Second
	// Create a ticker to send pings at regular intervals
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	// nolint: gosimple
	for {
		select {
		case <-ticker.C:
			// Send a ping message to the client
			err := conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				f.websocketService.Deregister(conn)
				f.logger.Debug("closing websocket connection", zap.Error(err), zap.Any("connection", conn))
				return
			}
		}
	}
}
