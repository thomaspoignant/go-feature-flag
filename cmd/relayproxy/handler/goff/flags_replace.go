package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"go.opentelemetry.io/otel"
)

type replaceFlag struct {
	flagsetManager service.FlagsetManager
	metrics        metric.Metrics
}

// NewReplaceFlag initializes the controller for PUT /v1/flags/{flag_key}.
func NewReplaceFlag(flagsetManager service.FlagsetManager, metrics metric.Metrics) Controller {
	return &replaceFlag{flagsetManager: flagsetManager, metrics: metrics}
}

// Handler is the entry point for the replaceFlag endpoint.
// @Summary      Replace a feature flag
// @Tags Flag Management API
// @Description  Full replacement of the flag definition. Creates the flag (upsert) if it
// @Description  does not exist yet.
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        flag_key path string true "The flag key"
// @Param        If-Match header string false "Optimistic-concurrency guard"
// @Param        data body dto.DTO true "The full flag definition"
// @Success      200  {object} model.FlagResponse "Replaced"
// @Success      201  {object} model.FlagResponse "Created via upsert"
// @Failure      400 {object} model.ErrorResponse "Bad Request"
// @Failure      401 {object} model.ErrorResponse "Unauthorized"
// @Failure      403 {object} model.ErrorResponse "Forbidden"
// @Failure      412 {object} model.ErrorResponse "Precondition Failed"
// @Router       /v1/flags/{flag_key} [put]
func (h *replaceFlag) Handler(httpContext echo.Context) error {
	tracer := otel.GetTracerProvider().Tracer(configfile.OtelTracerName)
	requestContext, span := tracer.Start(httpContext.Request().Context(), "replaceFlag")
	defer span.End()

	key := httpContext.Param("flag_key")
	if err := validateFlagKey(key); err != nil {
		return writeError(httpContext, http.StatusBadRequest, flag.ErrorFlagConfiguration, err.Error())
	}

	flagset, writable, err := resolveWritableFlagSet(httpContext, h.flagsetManager)
	if err != nil {
		return err
	}

	requestBody, err := readBody(httpContext)
	if err != nil {
		return writeError(httpContext, http.StatusBadRequest, flag.ErrorFlagConfiguration, "impossible to read request body: "+err.Error())
	}
	flagDefinition, created, etag, err := validateAndUpsertFlag(httpContext, requestContext, flagset, writable, key, requestBody)
	if err != nil {
		return err
	}

	if created {
		h.metrics.IncFlagCreated(key)
	} else {
		h.metrics.IncFlagUpdated(key)
	}
	h.metrics.IncFlagChange()

	flagResponse, buildErr := buildFlagResponse(key, flagDefinition, writable.Source(), true, etag)
	if buildErr != nil {
		return writeError(httpContext, http.StatusInternalServerError, flag.ErrorCodeGeneral, buildErr.Error())
	}
	httpContext.Response().Header().Set("ETag", etag)
	status := http.StatusOK
	if created {
		status = http.StatusCreated
		httpContext.Response().Header().Set(echo.HeaderLocation, "/v1/flags/"+key)
	}
	return httpContext.JSON(status, flagResponse)
}
