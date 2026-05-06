package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"go.uber.org/zap"
)

// NewWsFlagChange is the constructor to create a new controller to handle websocket
// request to be notified about flag changes.
func NewWsFlagChange(websocketService service.WebsocketService, logger *zap.Logger) *WSFlagChange {
	return &WSFlagChange{
		websocketService: websocketService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
		logger: logger,
	}
}

// WSFlagChange is the implementation of the controller
type WSFlagChange struct {
	websocketService service.WebsocketService
	upgrader         websocket.Upgrader
	logger           *zap.Logger
}

// Handler is the entry point for the websocket endpoint to get notified when a flag has been edited
// @Summary      Websocket endpoint to be notified about flag changes
// @Tags         GO Feature Flag Evaluation Stream API
// @Deprecated
// @Description  Deprecated: use /stream/v1/ws/flag/change instead. This endpoint
// @Description  is a websocket endpoint to be notified about flag changes; every
// @Description  change pushes a notifier.DiffCache message to the client.
// @Produce      json
// @Accept       json
// @Param        apiKey query string false "apiKey to authorize the connection to the relay proxy"
// @Success      200  {object} notifier.DiffCache "Success"
// @Failure      400  {object} modeldocs.HTTPErrorDoc "Bad Request"
// @Failure      401  {object} modeldocs.HTTPErrorDoc "Unauthorized"
// @Failure      500  {object} modeldocs.HTTPErrorDoc "Internal server error"
// @Router       /ws/v1/flag/change [get]
func (f *WSFlagChange) LegacyHandler(c echo.Context) error {
	// This handler is deprecated and we keep it for the documentation.
	return f.Handler(c)
}

// Handler is the entry point for the websocket endpoint to get notified when a flag has been edited
// @Summary      Websocket endpoint to be notified about flag changes
// @Tags         GO Feature Flag Evaluation Stream API
// @Description  This endpoint is a websocket endpoint to be notified about flag changes;
// @Description  every change pushes a notifier.DiffCache message to the client.
// @Produce      json
// @Accept       json
// @Param        apiKey query string false "apiKey to authorize the connection to the relay proxy"
// @Success      200  {object} notifier.DiffCache "Success"
// @Failure      400  {object} modeldocs.HTTPErrorDoc "Bad Request"
// @Failure      401  {object} modeldocs.HTTPErrorDoc "Unauthorized"
// @Failure      500  {object} modeldocs.HTTPErrorDoc "Internal server error"
// @Router       /stream/v1/ws/flag/change [get]
func (f *WSFlagChange) Handler(c echo.Context) error {
	conn, err := f.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	f.websocketService.Register(conn)
	defer f.websocketService.Deregister(conn)
	f.logger.Debug("registering new websocket connection", zap.Any("connection", conn))

	// Create context for cancellation
	ctx, cancel := context.WithCancel(c.Request().Context())
	defer cancel()

	// Start the ping pong loop
	go f.pingPongLoop(ctx, conn)

	const readDeadline = 60 * time.Second

	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(readDeadline))
		return nil
	})

	_ = conn.SetReadDeadline(time.Now().Add(readDeadline))
	for {
		// ReadMessage is needed to process control messages like pongs and close messages.
		if _, _, err := conn.ReadMessage(); err != nil {
			f.logger.Debug(
				"closing websocket connection",
				zap.Error(err),
				zap.Any("connection", conn),
			)
			return nil
		}
	}
}

// pingPongLoop is a keep-alive call to the client.
// It calls the client to ensure that the connection is still active.
// If the ping is not working we are closing the session.
func (f *WSFlagChange) pingPongLoop(ctx context.Context, conn *websocket.Conn) {
	// Ping interval duration
	pingInterval := 1 * time.Second
	// Create a ticker to send pings at regular intervals
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send a ping message to the client
			err := conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				f.logger.Debug(
					"closing websocket connection",
					zap.Error(err),
					zap.Any("connection", conn),
				)
				return
			}
		case <-ctx.Done():
			f.logger.Debug("stopping ping pong loop")
			return
		}
	}
}
