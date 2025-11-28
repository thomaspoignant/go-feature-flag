package controller

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/helper"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

type FlagChangeAPICtrl struct {
	flagsetManager service.FlagsetManager
	metrics        metric.Metrics
}

func NewAPIFlagChange(flagsetManager service.FlagsetManager, metrics metric.Metrics) Controller {
	return &FlagChangeAPICtrl{
		flagsetManager: flagsetManager,
		metrics:        metrics,
	}
}

type FlagChangeResponse struct {
	Hash  uint32            `json:"hash"`
	Flags map[string]uint32 `json:"flags"`
}

// Handler is the endpoint to poll if you want to know if there is a configuration change in the flags
// @Summary    Endpoint to poll if you want to know if there is a configuration change in the flags
// @Tags GO Feature Flag Evaluation API
// @Description Making a **GET** request to the URL `/v1/flag/change` will give you the hash of the current
// @Description configuration, you can use this hash to know if the configuration has changed.
// @Security    ApiKeyAuth
// @Security    XApiKeyAuth
// @Produce     json
// @Accept      json
// @Param       If-None-Match header string false "The request will be processed only if ETag doesn't match."
// @Success     200  {object} FlagChangeResponse "Success"
// @Success     304 {string} string "Etag: \"117-0193435c612c50d93b798619d9464856263dbf9f\""
// @Failure     500 {object}  modeldocs.HTTPErrorDoc "Internal server error"
// @Router      /v1/flag/change [get]
func (h *FlagChangeAPICtrl) Handler(c echo.Context) error {
	flagset, httpErr := helper.GetFlagSet(h.flagsetManager, helper.GetAPIKey(c))
	if httpErr != nil {
		return httpErr
	}

	flags, err := flagset.GetFlagsFromCache()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	res, err := json.Marshal(flags)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	flagHashes := map[string]uint32{}
	for key, flag := range flags {
		jsonFlag, _ := json.Marshal(flag)
		flagHashes[key] = utils.Hash(string(jsonFlag))
	}

	return c.JSON(
		http.StatusOK,
		FlagChangeResponse{Hash: utils.Hash(string(res)), Flags: flagHashes},
	)
}
