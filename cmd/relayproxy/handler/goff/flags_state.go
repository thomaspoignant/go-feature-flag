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

type setFlagState struct {
	flagsetManager service.FlagsetManager
	metrics        metric.Metrics
}

// NewSetFlagState initializes the controller for PATCH /v1/flags/{flag_key}/state.
func NewSetFlagState(flagsetManager service.FlagsetManager, metrics metric.Metrics) Controller {
	return &setFlagState{flagsetManager: flagsetManager, metrics: metrics}
}

// Handler is the entry point for the setFlagState endpoint.
// @Summary      Enable or disable a flag
// @Tags Flag Management API
// @Description  Convenience endpoint that sets the disable field without touching the rest
// @Description  of the flag definition.
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        flag_key path string true "The flag key"
// @Param        If-Match header string false "Optimistic-concurrency guard"
// @Param        data body model.SetFlagStateRequest true "The new state"
// @Success      200  {object} model.FlagResponse "State changed"
// @Failure      400 {object} model.ErrorResponse "Bad Request"
// @Failure      401 {object} model.ErrorResponse "Unauthorized"
// @Failure      404 {object} model.ErrorResponse "Not Found"
// @Failure      412 {object} model.ErrorResponse "Precondition Failed"
// @Router       /v1/flags/{flag_key}/state [patch]
func (h *setFlagState) Handler(httpContext echo.Context) error {
	tracer := otel.GetTracerProvider().Tracer(configfile.OtelTracerName)
	requestContext, span := tracer.Start(httpContext.Request().Context(), "setFlagState")
	defer span.End()

	flagKey := httpContext.Param("flag_key")
	flagset, writable, err := resolveWritableFlagSet(httpContext, h.flagsetManager)
	if err != nil {
		return err
	}

	requestBody, err := readBody(httpContext)
	if err != nil {
		return writeError(httpContext, http.StatusBadRequest, flag.ErrorFlagConfiguration, "impossible to read request body: "+err.Error())
	}
	var stateRequest model.SetFlagStateCommand
	if err := json.Unmarshal(requestBody, &stateRequest); err != nil {
		return writeError(httpContext, http.StatusBadRequest, flag.ErrorFlagConfiguration, "invalid request body: "+err.Error())
	}

	currentFlagBytes, _, getFlagErr := writable.GetFlag(requestContext, flagKey)
	if getFlagErr != nil {
		status, code := mapWriteErr(getFlagErr)
		return writeError(httpContext, status, code, getFlagErr.Error())
	}

	var flagDefinition model.FlagDefinition
	if err := json.Unmarshal(currentFlagBytes, &flagDefinition); err != nil {
		return writeError(httpContext, http.StatusInternalServerError, flag.ErrorCodeGeneral, err.Error())
	}
	flagDefinition.Disable = &stateRequest.Disable

	updatedFlagBytes, err := json.Marshal(flagDefinition)
	if err != nil {
		return writeError(httpContext, http.StatusInternalServerError, flag.ErrorCodeGeneral, err.Error())
	}
	if _, err := validateDefinition(updatedFlagBytes); err != nil {
		return writeError(httpContext, http.StatusBadRequest, flag.ErrorFlagConfiguration, err.Error())
	}

	_, etag, upsertErr := writable.UpsertFlag(requestContext, flagKey, updatedFlagBytes, ifMatchHeader(httpContext))
	if upsertErr != nil {
		status, code := mapWriteErr(upsertErr)
		return writeError(httpContext, status, code, upsertErr.Error())
	}

	reload(httpContext, flagset)
	return finalizeFlagUpdate(httpContext, h.metrics, flagKey, flagDefinition, writable.Source(), etag)
}
