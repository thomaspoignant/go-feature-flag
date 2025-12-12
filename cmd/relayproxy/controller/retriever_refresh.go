package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/helper"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type forceFlagsRefresh struct {
	flagsetManager service.FlagsetManager
	metrics        metric.Metrics
}

type retrieverRefreshResponse struct {
	Refreshed bool `json:"refreshed"`
}

// NewForceFlagsRefresh initialize the controller for the /data/collector endpoint
func NewForceFlagsRefresh(flagsetManager service.FlagsetManager, metrics metric.Metrics) Controller {
	return &forceFlagsRefresh{
		flagsetManager: flagsetManager,
		metrics:        metrics,
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

	flagset, httpErr := helper.FlagSet(h.flagsetManager, helper.APIKey(c))
	if httpErr != nil {
		return httpErr
	}

	tracer := otel.GetTracerProvider().Tracer(config.OtelTracerName)
	_, span := tracer.Start(c.Request().Context(), "retrieverRefresh")
	defer span.End()
	forceRefreshStatus := flagset.ForceRefresh()
	resp := retrieverRefreshResponse{
		Refreshed: forceRefreshStatus,
	}
	span.SetAttributes(attribute.Bool("retrieverRefresh.refreshed", forceRefreshStatus))
	return c.JSON(http.StatusOK, resp)
}
