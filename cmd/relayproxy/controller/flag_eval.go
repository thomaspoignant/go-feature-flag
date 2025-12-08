package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/helper"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type flagEval struct {
	flagsetMngr service.FlagsetManager
	metrics     metric.Metrics
}

func NewFlagEval(flagsetMngr service.FlagsetManager, metrics metric.Metrics) Controller {
	return &flagEval{
		flagsetMngr: flagsetMngr,
		metrics:     metrics,
	}
}

// Handler is the entry point for the flag eval endpoint
// @Summary     Evaluate a feature flag
// @Tags GO Feature Flag Evaluation API
// @Description Making a **POST** request to the URL `/v1/feature/<your_flag_name>/eval` will give you the value of the
// @Description flag for this user.
// @Description
// @Description To get a variation you should provide information about the user:
// @Description - User information in JSON in the request body.
// @Description - A default value in case there is an error while evaluating the flag.
// @Description
// @Description Note that you will always have a usable value in the response, you can use the field `failed` to know if
// @Description an issue has occurred during the validation of the flag, in that case the value returned will be the
// @Description default value.
// @Security     ApiKeyAuth
// @Security     XApiKeyAuth
// @Produce      json
// @Accept	 	 json
// @Param 		 data body model.EvalFlagRequest true "Payload of the user we want to get all the flags from."
// @Param        flag_key path string true "Name of your feature flag"
// @Success      200  {object} modeldocs.EvalFlagDoc "Success"
// @Failure      400 {object}  modeldocs.HTTPErrorDoc "Bad Request"
// @Failure      500 {object}  modeldocs.HTTPErrorDoc "Internal server error"
// @Router       /v1/feature/{flag_key}/eval [post]
func (h *flagEval) Handler(c echo.Context) error {
	flagKey := c.Param("flagKey")
	if flagKey == "" {
		return fmt.Errorf("impossible to find the flag key in the URL")
	}
	h.metrics.IncFlagEvaluation(flagKey)

	reqBody := new(model.EvalFlagRequest)
	if err := c.Bind(reqBody); err != nil {
		return err
	}

	// validation that we have a reqBody key
	if err := assertRequest(&reqBody.AllFlagRequest); err != nil {
		return err
	}
	evaluationCtx, err := evaluationContextFromRequest(&reqBody.AllFlagRequest)
	if err != nil {
		return err
	}

	tracer := otel.GetTracerProvider().Tracer(config.OtelTracerName)
	_, span := tracer.Start(c.Request().Context(), "flagEvaluation")
	defer span.End()

	flagset, httpErr := helper.FlagSet(h.flagsetMngr, helper.APIKey(c))
	if httpErr != nil {
		return httpErr
	}

	flagValue, _ := flagset.RawVariation(flagKey, evaluationCtx, reqBody.DefaultValue)

	span.SetAttributes(
		attribute.String("flagEvaluation.flagName", flagKey),
		attribute.Bool("flagEvaluation.trackEvents", flagValue.TrackEvents),
		attribute.String("flagEvaluation.variant", flagValue.VariationType),
		attribute.Bool("flagEvaluation.failed", flagValue.Failed),
		attribute.String("flagEvaluation.version", flagValue.Version),
		attribute.String("flagEvaluation.reason", flagValue.Reason),
		attribute.String("flagEvaluation.errorCode", flagValue.ErrorCode),
		attribute.Bool("flagEvaluation.cacheable", flagValue.Cacheable),
		// we convert to string because there is no attribute for interface{}
		attribute.String("flagEvaluation.value", fmt.Sprintf("%v", flagValue.Value)),
	)

	if flagsetName, err := h.flagsetMngr.FlagSetName(helper.APIKey(c)); err == nil {
		span.SetAttributes(attribute.String("flagEvaluation.flagSetName", flagsetName))
	}

	return c.JSON(http.StatusOK, flagValue)
}
