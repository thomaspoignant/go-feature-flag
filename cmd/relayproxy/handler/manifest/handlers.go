package manifest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/helper"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
	manifestHelper "github.com/thomaspoignant/go-feature-flag/cmdhelpers/manifest"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

const (
	manifestCapabilitiesHeader = "X-Manifest-Capabilities"
	manifestCapabilitiesValue  = "read"
)

func NewManifest(
	flagsetManager service.FlagsetManager, metrics metric.Metrics, logger *zap.Logger) ManifestCtrl {
	return ManifestCtrl{
		flagsetManager: flagsetManager,
		metrics:        metrics,
		logger:         logger,
	}
}

type ManifestCtrl struct {
	flagsetManager service.FlagsetManager
	metrics        metric.Metrics
	logger         *zap.Logger
}

// GetManifest is a GET endpoint to return the flags configured in GO Feature Flag as a OpenFeature flag manifest.
// @Summary   GetManifest is a GET endpoint to return the flags configured in GO Feature Flag.
// @Tags OpenFeature Manisfest API
// @Description **GET** request to the URL `/openfeature/v0/manifest` returns the project manifest
// @Description containing active flags. Archived flags are excluded.
// @Security     ApiKeyAuth
// @Produce      json
// @Accept	 	 json
// @Success      200  {object} model.ManifestSuccessResponse "Success"
// @Failure      401 {object}  model.ManifestError "Unauthorized"
// @Failure      403 {object}  model.ManifestError "Forbidden"
// @Failure      500 {object}  model.ManifestError "Internal server error"
// @Router       /openfeature/v0/manifest [GET]
func (m *ManifestCtrl) GetManifest(c echo.Context) error {
	tracer := otel.GetTracerProvider().Tracer(configfile.OtelTracerName)
	_, span := tracer.Start(c.Request().Context(), "getManifest")
	defer span.End()

	flagset, httpErr := helper.FlagSet(m.flagsetManager, helper.APIKey(c))
	if httpErr != nil {
		return httpErr
	}

	flags, err := flagset.GetFlagsFromCacheWithContext(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error while getting flags from cache: "+err.Error())
	}

	manifest, err := manifestHelper.GenerateDefinitionFromFlags(flags)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error while generating manifest: "+err.Error())
	}
	manifestSuccessResponse := model.ManifestSuccessResponse{
		Flags: func() []model.ManifestDefinitionSuccessResponse {
			responses := make([]model.ManifestDefinitionSuccessResponse, 0, len(manifest))
			for k, v := range manifest {
				responses = append(responses, model.ManifestDefinitionSuccessResponse{
					Key:          k,
					FlagType:     v.FlagType,
					DefaultValue: v.DefaultValue,
					Description:  v.Description,
				})
			}
			return responses
		}(),
	}
	c.Response().Header().Set(manifestCapabilitiesHeader, manifestCapabilitiesValue)
	m.metrics.IncGetManifestCall()
	return c.JSON(http.StatusOK, manifestSuccessResponse)
}
