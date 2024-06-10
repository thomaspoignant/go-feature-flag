package controller

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
	http "net/http"
)

type FlagChangeAPICtrl struct {
	goFF    *ffclient.GoFeatureFlag
	metrics metric.Metrics
}

func NewAPIFlagChange(goFF *ffclient.GoFeatureFlag, metrics metric.Metrics) Controller {
	return &FlagChangeAPICtrl{
		goFF:    goFF,
		metrics: metrics,
	}
}

type FlagChangeResponse struct {
	Hash uint32 `json:"hash"`
}

// Handler is the endpoint to poll if you want to know if there is a configuration change in the flags
// @Summary    Endpoint to poll if you want to know if there is a configuration change in the flags
// @Tags GO Feature Flag Evaluation API
// @Description Making a **GET** request to the URL `/v1/flag/change` will give you the hash of the current
// @Description configuration, you can use this hash to know if the configuration has changed.
// @Security    ApiKeyAuth
// @Produce     json
// @Accept      json
// @Param       If-None-Match header string false "The request will be processed only if ETag doesn't match."
// @Success     200  {object} FlagChangeResponse "Success"
// @Success     304 {string} string "Etag: \"117-0193435c612c50d93b798619d9464856263dbf9f\""
// @Failure     500 {object}  modeldocs.HTTPErrorDoc "Internal server error"
// @Router      /v1/flag/change [get]
func (h *FlagChangeAPICtrl) Handler(c echo.Context) error {
	flags, err := h.goFF.GetFlagsFromCache()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	res, err := json.Marshal(flags)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, FlagChangeResponse{Hash: utils.Hash(string(res))})
}
