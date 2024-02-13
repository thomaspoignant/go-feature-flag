package ofrep

import (
	"fmt"
	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/ofrep/customerr"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	http "net/http"
)

type ofrepEvaluateCtrl struct {
	goFF    *ffclient.GoFeatureFlag
	metrics metric.Metrics
}

func NewOFREPEvaluate(goFF *ffclient.GoFeatureFlag, metrics metric.Metrics) Controller {
	return &ofrepEvaluateCtrl{
		goFF:    goFF,
		metrics: metrics,
	}
}

// OFREPHandler is the entry point to evaluate a flag using the OpenFeature Remote Evaluation Protocol
// @Summary     Evaluate a feature flag using the OpenFeature Remote Evaluation Protocol
// @Description Making a **POST** request to the URL `/ofrep/v1/evaluate/<your_flag_name>` will give you the value of
// @Description the flag for this evaluation context
// @Description
// @Security     ApiKeyAuth
// @Produce      json
// @Accept	 	 json
// @Param 		 data body model.OFREPEvalFlagRequest true "Evaluation Context for this API call"
// @Param        flag_key path string true "Name of your feature flag"
// @Success      200  {object} model.OFREPEvaluateSuccessResponse "Success"
// @Failure      400 {object}  model.OFREPEvaluateErrorResponse "Bad Request"
// @Failure      401 {object}  modeldocs.HTTPErrorDoc "Unauthorized"
// @Failure      404 {object}  model.OFREPEvaluateErrorResponse "Flag Not Found"
// @Failure      500 {object}  modeldocs.HTTPErrorDoc "Internal server error"
// @Router       /ofrep/v1/evaluate/{flag_key} [post]
func (h *ofrepEvaluateCtrl) OFREPHandler(c echo.Context) error {
	flagKey := c.Param("flagKey")
	if flagKey == "" {
		return c.JSON(
			http.StatusBadRequest,
			customerr.NewOFREPEvaluateError(flagKey, flag.ErrorCodeGeneral,
				"No key provided in the URL").ToOFRErrorResponse())
	}
	h.metrics.IncFlagEvaluation(flagKey)

	reqBody := new(model.OFREPEvalFlagRequest)
	if err := c.Bind(reqBody); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			customerr.NewOFREPEvaluateError(flagKey, flag.ErrorCodeInvalidContext, err.Error()).ToOFRErrorResponse())
	}
	if err := assertOFREPEvaluateRequest(flagKey, reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, err.ToOFRErrorResponse())
	}
	evalCtx, err := evaluationContextFromOFREPRequest(reqBody.Context)
	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			err)
	}

	tracer := otel.GetTracerProvider().Tracer(config.OtelTracerName)
	_, span := tracer.Start(c.Request().Context(), "flagEvaluation")
	defer span.End()
	defaultValue := "thisisadefaultvaluethatItest1233%%"
	flagValue, _ := h.goFF.RawVariation(flagKey, evalCtx, defaultValue)

	if flagValue.Reason == flag.ReasonError {
		httpStatus := http.StatusBadRequest
		if flagValue.ErrorCode == flag.ErrorCodeFlagNotFound {
			httpStatus = http.StatusNotFound
		}
		return c.JSON(
			httpStatus,
			customerr.NewOFREPEvaluateError(flagKey, flagValue.ErrorCode,
				fmt.Sprintf("Error while evaluating the flag: %s", flagValue.ErrorCode)).ToOFRErrorResponse())
	}

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

	return c.JSON(http.StatusOK, model.OFREPEvaluateSuccessResponse{
		Key:      flagKey,
		Value:    flagValue.Value,
		Reason:   flagValue.Reason,
		Variant:  flagValue.VariationType,
		Metadata: flagValue.Metadata,
	})
}

func assertOFREPEvaluateRequest(key string, ofrepEvalReq *model.OFREPEvalFlagRequest) *customerr.OfrepEvaluateError {
	if ofrepEvalReq.Context == nil || ofrepEvalReq.Context["targetingKey"] == "" {
		return customerr.NewOFREPEvaluateError(key,
			"TARGETING_KEY_MISSING", "GO Feature Flag MUST have a targeting key in the request.")
	}

	return nil
}

func evaluationContextFromOFREPRequest(ctx map[string]any) (ffcontext.Context, error) {
	if targetingKey, ok := ctx["targetingKey"].(string); ok {
		delete(ctx, "targetingKey")
		evalCtx := utils.ConvertEvaluationCtxFromRequest(targetingKey, ctx)
		return evalCtx, nil
	}
	return ffcontext.EvaluationContext{}, customerr.NewOFREPEvaluateError("",
		"TARGETING_KEY_MISSING", "GO Feature Flag has received a targetingKey that is not a string.")
}
