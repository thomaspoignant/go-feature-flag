package controller

import (
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
			CheckOrigin: func(_ *http.Request) bool {
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

// Handler is the entry point for the websocket endpoint to get notified when a flag has been edited
// @Summary      Websocket endpoint to be notified about flag changes
// @Description  This endpoint is a websocket endpoint to be notified about flag changes, every change
// @Description  will send a request to the client with a model.DiffCache format.
// @Description
// @Produce      json
// @Accept		 json
// @Param 	     apiKey query string false "apiKey use authorize the connection to the relay proxy"
// @Success      200  {object} notifier.DiffCache "Success"
// @Failure      400 {object} modeldocs.HTTPErrorDoc "Bad Request"
// @Failure      500 {object} modeldocs.HTTPErrorDoc "Internal server error"
// @Router       /ws/v1/flag/change [post]
func (f *wsFlagChange) Handler(c echo.Context) error {
	conn, err := f.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	f.websocketService.Register(conn)
	defer f.websocketService.Deregister(conn)
	f.logger.Debug("registering new websocket connection", zap.Any("connection", conn))

	// Start the ping pong loop
	stopper := make(chan struct{}, 1)
	go f.pingPongLoop(stopper, conn)
	isOpen := true
	for isOpen {
		_, _, err := conn.ReadMessage()
		if err != nil {
			close(stopper)
			f.websocketService.Deregister(conn)
			f.logger.Debug("closing websocket connection", zap.Error(err), zap.Any("connection", conn))
			isOpen = false
		}
	}
	return nil
}

// pingPongLoop is a keep-alive call to the client.
// It calls the client to ensure that the connection is still active.
// If the ping is not working we are closing the session.
func (f *wsFlagChange) pingPongLoop(stopper chan struct{}, conn *websocket.Conn) {
	// Ping interval duration
	pingInterval := 1 * time.Second
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
				f.logger.Debug("closing websocket connection", zap.Error(err), zap.Any("connection", conn))
				f.websocketService.Deregister(conn)
				return
			}
		case <-stopper:
			f.logger.Debug("stopping ping pong loop")
			return
		}
	}
}
