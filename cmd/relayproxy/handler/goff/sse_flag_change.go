package controller

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// NewSSEFlagChange is the constructor to create a new controller to handle SSE
// requests to be notified about flag changes.
func NewSSEFlagChange(logger *zap.Logger) Controller {
	return &sseFlagChange{logger: logger}
}

type sseFlagChange struct {
	logger *zap.Logger
}

// Handler is the entry point for the SSE endpoint to be notified of flag changes.
// @Summary      SSE endpoint to be notified about flag changes
// @Tags         GO Feature Flag Evaluation Stream API
// @Description  Server-Sent Events endpoint pushing flag change notifications.
// @Description  Each event payload is a `notifier.DiffCache` JSON document.
// @Description  The full URL (including query string) is sensitive and must not be logged
// @Description  or persisted by intermediaries.
// @Produce      text/event-stream
// @Param        apiKey query string false "apiKey to authorize the connection to the relay proxy"
// @Success      200  {object} notifier.DiffCache "SSE stream of flag change events"
// @Failure      401  {object} modeldocs.HTTPErrorDoc "Unauthorized"
// @Failure      500  {object} modeldocs.HTTPErrorDoc "Internal server error"
// @Router       /stream/v1/sse/flag/change [get]
func (h *sseFlagChange) Handler(c echo.Context) error {
	// TODO: implement SSE streaming.
	return nil
}
