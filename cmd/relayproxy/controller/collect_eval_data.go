package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-viper/mapstructure/v2"
	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/helper"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type collectEvalData struct {
	flagsetManager service.FlagsetManager
	metrics        metric.Metrics
	logger         *zap.Logger
}

// NewCollectEvalData initialize the controller for the /data/collector endpoint
func NewCollectEvalData(
	flagsetManager service.FlagsetManager,
	metrics metric.Metrics,
	logger *zap.Logger,
) Controller {
	return &collectEvalData{
		flagsetManager: flagsetManager,
		metrics:        metrics,
		logger:         logger,
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
	ctx := c.Request().Context()

	tracer := otel.Tracer(config.OtelTracerName)
	ctx, span := tracer.Start(ctx, "collectEventData")
	defer span.End()

	reqBody := new(model.CollectEvalDataRequest)
	if err := c.Bind(reqBody); err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Sprintf("collectEvalData: invalid input data %v", err))
	}
	if reqBody.Events == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "collectEvalData: invalid input data")
	}

	span.SetAttributes(attribute.Int("collectEventData.eventCollectionSize", len(reqBody.Events)))

	flagset, httpErr := helper.GetFlagSet(h.flagsetManager, helper.GetAPIKey(c))
	if httpErr != nil {
		return httpErr
	}

	counterTracking := 0
	counterEvaluation := 0
	for i, event := range reqBody.Events {
		// Check if context is cancelled before processing each event, to avoid
		// long delays on large payloads.
		select {
		case <-ctx.Done():
			err := fmt.Errorf("context cancelled after processing %d/%d events: %w",
				i, len(reqBody.Events), ctx.Err())
			span.SetAttributes(
				attribute.Int("processed_tracking", counterTracking),
				attribute.Int("processed_evaluation", counterEvaluation),
				attribute.Int("total_events", len(reqBody.Events)),
			)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			return err
		default:
			// all good, keep going
		}

		switch event["kind"] {
		case "tracking":
			e, err := convertTrackingEvent(event, h.logger)
			if err != nil {
				h.logger.Error(
					"impossible to convert the event to a tracking event",
					zap.Error(err),
				)
				continue
			}
			flagset.CollectTrackingEventData(e)
			counterTracking++
		default:
			e, err := convertFeatureEvent(event, reqBody.Meta, h.logger)
			if err != nil {
				h.logger.Error("impossible to convert the event to a feature event", zap.Error(err))
				continue
			}
			flagset.CollectEventData(e)
			counterEvaluation++
		}
	}

	span.SetAttributes(attribute.Int("collectEventData.trackingCollectionSize", counterTracking))
	span.SetAttributes(
		attribute.Int("collectEventData.evaluationCollectionSize", counterEvaluation),
	)
	h.metrics.IncCollectEvalData(float64(len(reqBody.Events)))

	return c.JSON(http.StatusOK, model.CollectEvalDataResponse{
		IngestedContentCount: len(reqBody.Events),
	})
}

func convertTrackingEvent(
	event map[string]any,
	logger *zap.Logger,
) (exporter.TrackingEvent, error) {
	var e exporter.TrackingEvent
	marshalled, err := json.Marshal(event)
	if err != nil {
		return exporter.TrackingEvent{}, err
	}
	err = json.Unmarshal(marshalled, &e)
	if err != nil {
		return exporter.TrackingEvent{}, err
	}
	e.CreationDate = formatCreationDate(e.CreationDate, logger)
	return e, nil
}

func convertFeatureEvent(event map[string]any,
	metadata exporter.FeatureEventMetadata,
	logger *zap.Logger) (exporter.FeatureEvent, error) {
	var e exporter.FeatureEvent
	err := mapstructure.Decode(event, &e)
	if err != nil {
		return exporter.FeatureEvent{}, err
	}
	if e.Source == "" {
		e.Source = "PROVIDER_CACHE"
	}
	if metadata != nil {
		e.Metadata = metadata
	}
	e.CreationDate = formatCreationDate(e.CreationDate, logger)
	return e, nil
}

func formatCreationDate(input int64, logger *zap.Logger) int64 {
	if input > 9999999999 {
		logger.Warn(
			"creationDate received is in milliseconds, we convert it to seconds",
			zap.Int64("creationDate", input))
		// if we receive a timestamp in milliseconds, we convert it to seconds
		// but since it is totally possible to have a timestamp in seconds that is bigger than 9999999999
		// we will accept timestamp up to 9999999999 (2286-11-20 18:46:39 +0100 CET)
		converted, _ := strconv.ParseInt(
			strconv.FormatInt(input, 10)[:10], 10, 64)
		return converted
	}
	return input
}
