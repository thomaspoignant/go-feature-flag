package controller

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"net/http"

	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
)

type collectEvalData struct {
	goFF    *ffclient.GoFeatureFlag
	metrics metric.Metrics
}

// NewCollectEvalData initialize the controller for the /data/collector endpoint
func NewCollectEvalData(goFF *ffclient.GoFeatureFlag, metrics metric.Metrics) Controller {
	return &collectEvalData{
		goFF:    goFF,
		metrics: metrics,
	}
}

// Handler is the entry point for the data/collector endpoint
// @Summary      Endpoint to send usage of your flags to be collected
// @Tags GO Feature Flag Evaluation API
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

	tracer := otel.GetTracerProvider().Tracer(config.OtelTracerName)
	_, span := tracer.Start(c.Request().Context(), "collectEventData")
	defer span.End()
	span.SetAttributes(attribute.Int("collectEventData.eventCollectionSize", len(reqBody.Events)))
	for _, event := range reqBody.Events {
		if event.Source == "" {
			event.Source = "PROVIDER_CACHE"
		}
		h.goFF.CollectEventData(event)
	}

	h.metrics.IncCollectEvalData(float64(len(reqBody.Events)))

	return c.JSON(http.StatusOK, model.CollectEvalDataResponse{
		IngestedContentCount: len(reqBody.Events),
	})
}
