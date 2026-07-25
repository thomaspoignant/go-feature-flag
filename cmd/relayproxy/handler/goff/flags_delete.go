package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
	"go.opentelemetry.io/otel"
)

type deleteFlag struct {
	flagsetManager service.FlagsetManager
	metrics        metric.Metrics
}

// NewDeleteFlag initializes the controller for DELETE /v1/flags/{flag_key}.
func NewDeleteFlag(flagsetManager service.FlagsetManager, metrics metric.Metrics) Controller {
	return &deleteFlag{flagsetManager: flagsetManager, metrics: metrics}
}

// Handler is the entry point for the deleteFlag endpoint.
// @Summary      Delete a feature flag
// @Tags Flag Management API
// @Security     ApiKeyAuth
// @Param        flag_key path string true "The flag key"
// @Param        If-Match header string false "Optimistic-concurrency guard"
// @Success      204  "Deleted"
// @Failure      401 {object} model.ErrorResponse "Unauthorized"
// @Failure      403 {object} model.ErrorResponse "Forbidden"
// @Failure      404 {object} model.ErrorResponse "Not Found"
// @Failure      412 {object} model.ErrorResponse "Precondition Failed"
// @Router       /v1/flags/{flag_key} [delete]
func (h *deleteFlag) Handler(httpContext echo.Context) error {
	tracer := otel.GetTracerProvider().Tracer(configfile.OtelTracerName)
	requestContext, span := tracer.Start(httpContext.Request().Context(), "deleteFlag")
	defer span.End()

	flagKey := httpContext.Param("flag_key")
	flagset, writable, err := resolveWritableFlagSet(httpContext, h.flagsetManager)
	if err != nil {
		return err
	}

	if delErr := writable.DeleteFlag(requestContext, flagKey, ifMatchHeader(httpContext)); delErr != nil {
		status, code := mapWriteErr(delErr)
		return writeError(httpContext, status, code, delErr.Error())
	}

	reload(httpContext, flagset)
	h.metrics.IncFlagDeleted(flagKey)
	h.metrics.IncFlagChange()

	return httpContext.NoContent(http.StatusNoContent)
}
