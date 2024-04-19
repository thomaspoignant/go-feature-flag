package ofrep

import (
	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"net/http"
)

// Configuration is the entry point to get the configuration for OFREP.
// @Summary OFREP provider configuration
// @Tags OpenFeature Remote Evaluation Protocol (OFREP)
// @Description OFREP configuration to provide information about the remote flag management system, to configure the
// @Description OpenFeature SDK providers.
// @Description
// @Description This endpoint will be called during the initialization of the provider.
// @Security     ApiKeyAuth
// @Produce      json
// @Accept	 	 json
// @Param        If-None-Match header string false "The request will be processed only if ETag doesn't match."
// @Success      200  {object} model.OFREPConfiguration "Success"
// @Success      304 {string} string "Etag: \"117-0193435c612c50d93b798619d9464856263dbf9f\""
// @Failure      401 {object}  modeldocs.HTTPErrorDoc "Unauthorized"
// @Failure      404 {object}  model.OFREPEvaluateErrorResponse "Flag Not Found"
// @Failure      500 {object}  modeldocs.HTTPErrorDoc "Internal server error"
// @Router       /ofrep/v1/configuration [get]
func (h *EvaluateCtrl) Configuration(c echo.Context) error {
	response := model.OFREPConfiguration{
		Name: "GO Feature Flag",
		Capabilities: model.OFREPConfigCapabilities{
			CacheInvalidation: model.OFREPConfigCapabilitiesCacheInvalidation{
				Polling: model.OFREPConfigCapabilitiesCacheInvalidationPolling{
					Enabled: true,
					// MinPollingInterval will always the same as the polling interval of the GoFeatureFlag
					MinPollingInterval: h.goFF.GetPollingInterval(),
				},
			},
		},
	}
	return c.JSON(http.StatusOK, response)
}
