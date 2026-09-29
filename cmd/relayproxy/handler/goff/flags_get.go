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

type getFlagCtrl struct {
	flagsetManager service.FlagsetManager
	metrics        metric.Metrics
}

// NewGetFlag initializes the controller for GET /v1/flags/{flag_key}.
func NewGetFlag(flagsetManager service.FlagsetManager, metrics metric.Metrics) Controller {
	return &getFlagCtrl{flagsetManager: flagsetManager, metrics: metrics}
}

// Handler is the entry point for the getFlag endpoint.
// @Summary      Get a feature flag
// @Tags Flag Management API
// @Security     ApiKeyAuth
// @Produce      json
// @Param        flag_key path string true "The flag key"
// @Success      200  {object} model.FlagResponse "Success"
// @Failure      401 {object} model.ErrorResponse "Unauthorized"
// @Failure      404 {object} model.ErrorResponse "Not Found"
// @Router       /v1/flags/{flag_key} [get]
func (h *getFlagCtrl) Handler(c echo.Context) error {
	tracer := otel.GetTracerProvider().Tracer(configfile.OtelTracerName)
	ctx, span := tracer.Start(c.Request().Context(), "getFlag")
	defer span.End()

	flagset, err := getFlagSet(c, h.flagsetManager)
	if err != nil {
		return err
	}

	key := c.Param("flag_key")
	flags, ffErr := flagset.GetFlagsFromCacheWithContext(ctx)
	if ffErr != nil {
		return writeError(c, http.StatusInternalServerError, flag.ErrorCodeGeneral, ffErr.Error())
	}

	f, found := flags[key]
	if !found {
		return writeError(c, http.StatusNotFound, flag.ErrorCodeFlagNotFound, "flag '"+key+"' not found")
	}
	d, ok := getDtoFromCachedFlag(f)
	if !ok {
		return writeError(c, http.StatusInternalServerError, flag.ErrorCodeGeneral, "impossible to read flag definition")
	}

	writable, editable := flagset.GetWritableRetriever()
	source := ""
	if editable {
		source = writable.Source()
	}

	resp, buildErr := buildFlagResponse(key, d, source, editable, "")
	if buildErr != nil {
		return writeError(c, http.StatusInternalServerError, flag.ErrorCodeGeneral, buildErr.Error())
	}
	c.Response().Header().Set("ETag", resp.ETag)
	return c.JSON(http.StatusOK, resp)
}
