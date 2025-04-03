package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type collectEvalData struct {
	goFF    *ffclient.GoFeatureFlag
	metrics metric.Metrics
	logger  *zap.Logger
}

// NewCollectEvalData initialize the controller for the /data/collector endpoint
func NewCollectEvalData(
	goFF *ffclient.GoFeatureFlag,
	metrics metric.Metrics,
	logger *zap.Logger,
) Controller {
	return &collectEvalData{
		goFF:    goFF,
		metrics: metrics,
		logger:  logger,
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
		// force the creation date to be a unix timestamp
		if event.CreationDate > 9999999999 {
			h.logger.Warn(
				"creationDate received is in milliseconds, we convert it to seconds",
				zap.Int64("creationDate", event.CreationDate))
			// if we receive a timestamp in milliseconds, we convert it to seconds
			// but since it is totally possible to have a timestamp in seconds that is bigger than 9999999999
			// we will accept timestamp up to 9999999999 (2286-11-20 18:46:39 +0100 CET)
			event.CreationDate, _ = strconv.ParseInt(
				strconv.FormatInt(event.CreationDate, 10)[:10], 10, 64)
		}
		if reqBody.Meta != nil {
			event.Metadata = reqBody.Meta
		}
		h.goFF.CollectEventData(event)
	}
	h.metrics.IncCollectEvalData(float64(len(reqBody.Events)))
	return c.JSON(http.StatusOK, model.CollectEvalDataResponse{
		IngestedContentCount: len(reqBody.Events),
	})
}
