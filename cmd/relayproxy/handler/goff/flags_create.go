package controller

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"go.opentelemetry.io/otel"
)

type createFlag struct {
	flagsetManager service.FlagsetManager
	metrics        metric.Metrics
}

// NewCreateFlag initializes the controller for POST /v1/flags.
func NewCreateFlag(flagsetManager service.FlagsetManager, metrics metric.Metrics) Controller {
	return &createFlag{flagsetManager: flagsetManager, metrics: metrics}
}

// Handler is the entry point for the createFlag endpoint.
// @Summary      Create a feature flag
// @Tags Flag Management API
// @Description  Only available when the caller's flagset is backed by exactly one writable
// @Description  (PostgreSQL) retriever with an explicit flagset configured.
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        data body model.CreateFlagRequest true "The flag to create"
// @Success      201  {object} model.FlagResponse "Created"
// @Failure      400 {object} model.ErrorResponse "Bad Request"
// @Failure      401 {object} model.ErrorResponse "Unauthorized"
// @Failure      403 {object} model.ErrorResponse "Forbidden"
// @Failure      409 {object} model.ErrorResponse "Conflict"
// @Router       /v1/flags [post]
func (h *createFlag) Handler(httpContext echo.Context) error {
	tracer := otel.GetTracerProvider().Tracer(configfile.OtelTracerName)
	requestContext, span := tracer.Start(httpContext.Request().Context(), "createFlag")
	defer span.End()

	flagset, writable, err := resolveWritableFlagSet(httpContext, h.flagsetManager)
	if err != nil {
		return err
	}

	requestBody, err := readBody(httpContext)
	if err != nil {
		return writeError(httpContext, http.StatusBadRequest, flag.ErrorFlagConfiguration, "impossible to read request body: "+err.Error())
	}

	var createFlagCommand model.CreateFlagCommand
	if err := json.Unmarshal(requestBody, &createFlagCommand); err != nil {
		return writeError(httpContext, http.StatusBadRequest, flag.ErrorFlagConfiguration, "invalid request body: "+err.Error())
	}
	if err := validateFlagKey(createFlagCommand.Key); err != nil {
		return writeError(httpContext, http.StatusBadRequest, flag.ErrorFlagConfiguration, err.Error())
	}

	flagCommandDefinition, err := json.Marshal(createFlagCommand.Definition)
	if err != nil {
		return writeError(httpContext, http.StatusBadRequest, flag.ErrorFlagConfiguration, err.Error())
	}
	d, err := validateDefinition(flagCommandDefinition)
	if err != nil {
		return writeError(httpContext, http.StatusBadRequest, flag.ErrorFlagConfiguration, err.Error())
	}

	etag, createErr := writable.CreateFlag(requestContext, createFlagCommand.Key, flagCommandDefinition)
	if createErr != nil {
		status, code := mapWriteErr(createErr)
		return writeError(httpContext, status, code, createErr.Error())
	}

	reload(httpContext, flagset)
	h.metrics.IncFlagCreated(createFlagCommand.Key)
	h.metrics.IncFlagChange()

	flagResponse, buildErr := buildFlagResponse(createFlagCommand.Key, d, writable.Source(), true, etag)
	if buildErr != nil {
		return writeError(httpContext, http.StatusInternalServerError, flag.ErrorCodeGeneral, buildErr.Error())
	}
	httpContext.Response().Header().Set(echo.HeaderLocation, "/v1/flags/"+createFlagCommand.Key)
	httpContext.Response().Header().Set("ETag", etag)
	return httpContext.JSON(http.StatusCreated, flagResponse)
}
