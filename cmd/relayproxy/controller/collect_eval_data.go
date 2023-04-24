package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
)

type collectEvalData struct {
	goFF *ffclient.GoFeatureFlag
}

// NewCollectEvalData initialize the controller for the /data/collector endpoint
func NewCollectEvalData(goFF *ffclient.GoFeatureFlag) Controller {
	return &collectEvalData{
		goFF: goFF,
	}
}

// Handler is the entry point for the data/collector endpoint
// @Summary      Endpoint to send usage of your flags to be collected
// @Description  This endpoint is receiving the events of your flags usage to send them in the data collector.
// @Description
// @Description  It is used by the different Open Feature providers to send in bulk all the cached events to avoid
// @Description  to lose track of what happen when a cached flag is used.
// @Security     ApiKeyAuth
// @Produce      json
// @Accept		 json
// @Param 		 data body model.CollectEvalDataRequest true "List of flag evaluation that be passed to the data exporter"
// @Success      200  {object} model.CollectEvalDataResponse "Success"
// @Failure 	 400 {object} modeldocs.HTTPErrorDoc "Bad Request"
// @Failure      500 {object} modeldocs.HTTPErrorDoc "Internal server error"
// @Router       /v1/data/collector [post]
func (h *collectEvalData) Handler(c echo.Context) error {
	reqBody := new(model.CollectEvalDataRequest)
	if err := c.Bind(reqBody); err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Sprintf("collectEvalData: invalid input data %v", err))
	}
	if reqBody == nil || reqBody.Events == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "collectEvalData: invalid input data")
	}

	for _, event := range reqBody.Events {
		h.goFF.CollectEventData(event)
	}

	// send metric
	metrics := c.Get(metric.CustomMetrics).(*metric.Metrics)
	metrics.IncCollectEvalData(float64(len(reqBody.Events)))

	return c.JSON(http.StatusOK, model.CollectEvalDataResponse{
		IngestedContentCount: len(reqBody.Events),
	})
}
