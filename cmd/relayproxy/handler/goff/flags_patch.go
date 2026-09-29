package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	jsonpatch "gopkg.in/evanphx/json-patch.v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"go.opentelemetry.io/otel"
)

type patchFlag struct {
	flagsetManager service.FlagsetManager
	metrics        metric.Metrics
}

// NewPatchFlag initializes the controller for PATCH /v1/flags/{flag_key}.
func NewPatchFlag(flagsetManager service.FlagsetManager, metrics metric.Metrics) Controller {
	return &patchFlag{flagsetManager: flagsetManager, metrics: metrics}
}

// Handler is the entry point for the patchFlag endpoint.
// @Summary      Update parts of a feature flag
// @Tags Flag Management API
// @Description  RFC 7386 merge-patch of the flag definition. Only the fields present in the
// @Description  body are changed; a JSON null deletes a field.
// @Security     ApiKeyAuth
// @Accept       application/merge-patch+json
// @Produce      json
// @Param        flag_key path string true "The flag key"
// @Param        If-Match header string false "Optimistic-concurrency guard"
// @Param        data body dto.DTO true "Merge-patch document"
// @Success      200  {object} model.FlagResponse "Updated"
// @Failure      400 {object} model.ErrorResponse "Bad Request"
// @Failure      401 {object} model.ErrorResponse "Unauthorized"
// @Failure      403 {object} model.ErrorResponse "Forbidden"
// @Failure      404 {object} model.ErrorResponse "Not Found"
// @Failure      412 {object} model.ErrorResponse "Precondition Failed"
// @Router       /v1/flags/{flag_key} [patch]
func (receiver *patchFlag) Handler(httpContext echo.Context) error {
	otelTracerProvider := otel.GetTracerProvider().Tracer(configfile.OtelTracerName)
	otelContext, span := otelTracerProvider.Start(httpContext.Request().Context(), "patchFlag")
	defer span.End()

	flagKey := httpContext.Param("flag_key")
	if err := validateFlagKey(flagKey); err != nil {
		return writeError(httpContext, http.StatusBadRequest, flag.ErrorFlagConfiguration, err.Error())
	}

	flagSet, writable, err := resolveWritableFlagSet(httpContext, receiver.flagsetManager)
	if err != nil {
		return err
	}

	currentFlag, _, getCurrentFlagError := writable.GetFlag(otelContext, flagKey)
	if getCurrentFlagError != nil {
		status, code := mapWriteErr(getCurrentFlagError)
		return writeError(httpContext, status, code, getCurrentFlagError.Error())
	}

	patchedFlag, err := readBody(httpContext)
	if err != nil {
		return writeError(httpContext, http.StatusBadRequest, flag.ErrorFlagConfiguration, "impossible to read request body: "+err.Error())
	}

	mergedFlag, err := jsonpatch.MergePatch(currentFlag, patchedFlag)
	if err != nil {
		return writeError(httpContext, http.StatusBadRequest, flag.ErrorFlagConfiguration, "invalid merge patch: "+err.Error())
	}

	definitionDto, _, flagEtag, err := validateAndUpsertFlag(httpContext, otelContext, flagSet, writable, flagKey, mergedFlag)
	if err != nil {
		return err
	}

	return finalizeFlagUpdate(httpContext, receiver.metrics, flagKey, definitionDto, writable.Source(), flagEtag)
}
