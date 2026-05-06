package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/utils"
	"go.uber.org/zap"
)

// NewSSEFlagChange is the constructor to create a new controller to handle SSE
// requests to be notified about flag changes.
func NewSSEFlagChange(
	sseService service.SSEService,
	flagsetManager service.FlagsetManager,
	logger *zap.Logger,
) Controller {
	return &sseFlagChange{
		sseService:     sseService,
		flagsetManager: flagsetManager,
		logger:         logger,
	}
}

type sseFlagChange struct {
	sseService     service.SSEService
	flagsetManager service.FlagsetManager
	logger         *zap.Logger
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
// @Failure      400  {object} modeldocs.HTTPErrorDoc "Bad Request"
// @Failure      401  {object} modeldocs.HTTPErrorDoc "Unauthorized"
// @Failure      500  {object} modeldocs.HTTPErrorDoc "Internal server error"
// @Router       /stream/v1/sse/flag/change [get]
func (h *sseFlagChange) Handler(c echo.Context) error {
	apiKey := c.QueryParam("apiKey")
	flagsetName, err := h.resolveFlagsetName(apiKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	h.logger.Debug("SSE client connecting", zap.String("flagset", flagsetName))

	// r3labs/sse routes by the "stream" query parameter.
	// We need to set the "stream" query parameter to the flagset name. We also need to delete the "apiKey" query parameter.
	// This is because the r3labs/sse server will use the "stream" query parameter to route the request to the correct flagset.
	// After that point the apiKey is not needed anymore, so we delete it.
	q := c.Request().URL.Query()
	q.Set("stream", flagsetName)
	q.Del("apiKey")
	c.Request().URL.RawQuery = q.Encode()

	h.sseService.ServeHTTP(c.Response(), c.Request())
	return nil
}

// resolveFlagsetName resolves the flagset name from the API key
// If the flagset manager is in default mode, it returns the default flagset name
func (h *sseFlagChange) resolveFlagsetName(apiKey string) (string, error) {
	if h.flagsetManager.IsDefaultFlagSet() {
		return utils.DefaultFlagSetName, nil
	}
	if apiKey == "" {
		return "", fmt.Errorf("apiKey is required when using flagsets")
	}
	return h.flagsetManager.FlagSetName(apiKey)
}
