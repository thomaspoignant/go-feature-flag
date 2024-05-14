package controller

import (
	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"net/http"
)

type forceFlagsRefresh struct {
	goFF    *ffclient.GoFeatureFlag
	metrics metric.Metrics
}

type retrieverRefreshResponse struct {
	Refreshed bool `json:"refreshed"`
}

// NewForceFlagsRefresh initialize the controller for the /data/collector endpoint
func NewForceFlagsRefresh(goFF *ffclient.GoFeatureFlag, metrics metric.Metrics) Controller {
	return &forceFlagsRefresh{
		goFF:    goFF,
		metrics: metrics,
	}
}

// Handler is used to force the refresh of the flags in the cache.
// It will trigger the retrievers to get the latest version of the available flags.
// This endpoint is used when you know explicitly that a flag has changed, and you want to trigger the collection
// of the new version.
// @Summary      This endpoint is used to force the refresh of the flags in the cache.
// @Tags Admin API to manage GO Feature Flag
// @Description  This endpoint is used to force the refresh of the flags in the cache.
// @Description  This endpoint is used when you know explicitly that a flag has changed, and you want to trigger
// @Description the collection of the new versions.
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {object} retrieverRefreshResponse "Success"
// @Failure 	 400 {object} modeldocs.HTTPErrorDoc "Bad Request"
// @Failure      500 {object} modeldocs.HTTPErrorDoc "Internal server error"
// @Router       /admin/v1/retriever/refresh [post]
func (h *forceFlagsRefresh) Handler(c echo.Context) error {
	h.metrics.IncForceRefresh()
	if h.goFF == nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"forceFlagsRefresh: goFF is not initialized")
	}

	tracer := otel.GetTracerProvider().Tracer(config.OtelTracerName)
	_, span := tracer.Start(c.Request().Context(), "retrieverRefresh")
	defer span.End()
	forceRefreshStatus := h.goFF.ForceRefresh()
	resp := retrieverRefreshResponse{
		Refreshed: forceRefreshStatus,
	}
	span.SetAttributes(attribute.Bool("retrieverRefresh.refreshed", forceRefreshStatus))
	return c.JSON(http.StatusOK, resp)
}
