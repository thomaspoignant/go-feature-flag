package ofrep

import (
	"fmt"
	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	http "net/http"
	"sort"
)

type EvaluateCtrl struct {
	goFF    *ffclient.GoFeatureFlag
	metrics metric.Metrics
}

func NewOFREPEvaluate(goFF *ffclient.GoFeatureFlag, metrics metric.Metrics) EvaluateCtrl {
	return EvaluateCtrl{
		goFF:    goFF,
		metrics: metrics,
	}
}

// Evaluate is the entry point to evaluate a flag using the OpenFeature Remote Evaluation Protocol
// @Summary     Evaluate a feature flag using the OpenFeature Remote Evaluation Protocol
// @Tags OpenFeature Remote Evaluation Protocol (OFREP)
// @Description Making a **POST** request to the URL `/ofrep/v1/evaluate/flags/{your_flag_name}` will give you the
// @Description value of the flag for this evaluation context
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
// @Router       /ofrep/v1/evaluate/flags/{flag_key} [post]
func (h *EvaluateCtrl) Evaluate(c echo.Context) error {
	flagKey := c.Param("flagKey")
	if flagKey == "" {
		return c.JSON(
			http.StatusBadRequest,
			NewEvaluateError(flagKey, flag.ErrorCodeGeneral,
				"No key provided in the URL"))
	}
	h.metrics.IncFlagEvaluation(flagKey)

	reqBody := new(model.OFREPEvalFlagRequest)
	if err := c.Bind(reqBody); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			NewEvaluateError(flagKey, flag.ErrorCodeInvalidContext, err.Error()))
	}
	if err := assertOFREPEvaluateRequest(reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, model.OFREPEvaluateErrorResponse{
			OFREPCommonErrorResponse: *err,
			Key:                      flagKey,
		})
	}
	evalCtx, err := evaluationContextFromOFREPRequest(reqBody.Context)
	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			model.OFREPEvaluateErrorResponse{
				OFREPCommonErrorResponse: model.OFREPCommonErrorResponse{
					ErrorCode:    flag.ErrorCodeInvalidContext,
					ErrorDetails: err.Error(),
				},
				Key: flagKey,
			})
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
			NewEvaluateError(flagKey, flagValue.ErrorCode,
				fmt.Sprintf("Error while evaluating the flag: %s", flagKey)))
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

// BulkEvaluate is the entry point to evaluate in bulk flags using the OpenFeature Remote Evaluation Protocol
// @Summary     Open-Feature Remote Evaluation Protocol bulk evaluation API.
// @Tags OpenFeature Remote Evaluation Protocol (OFREP)
// @Description Making a **POST** request to the URL `/ofrep/v1/evaluate/flags` will give you the value of the list
// @Description of feature flags for this evaluation context.
// @Description
// @Description If no flags are provided, the API will evaluate all available flags in the configuration.
// @Security    ApiKeyAuth
// @Produce     json
// @Accept	 	json
// @Param       If-None-Match header string false "The request will be processed only if ETag doesn't match."
// @Param 		data body model.OFREPEvalFlagRequest true "Evaluation Context and list of flag for this API call"
// @Success     200  {object} model.OFREPBulkEvaluateSuccessResponse "OFREP successful evaluation response"
// @Failure     400 {object}  model.OFREPCommonErrorResponse "Bad evaluation request"
// @Failure     401 {object}  modeldocs.HTTPErrorDoc "Unauthorized - You need credentials to access the API"
// @Failure     403 {object}  modeldocs.HTTPErrorDoc "Forbidden - You are not authorized to access the API"
// @Failure     500 {object}  modeldocs.HTTPErrorDoc "Internal server error"
// @Router      /ofrep/v1/evaluate/flags [post]
func (h *EvaluateCtrl) BulkEvaluate(c echo.Context) error {
	h.metrics.IncAllFlag()

	request := new(model.OFREPEvalFlagRequest)
	if err := c.Bind(request); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			NewOFREPCommonError("INVALID_CONTEXT", err.Error()))
	}
	if err := assertOFREPEvaluateRequest(request); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	evalCtx, err := evaluationContextFromOFREPRequest(request.Context)
	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			err)
	}

	// if no flag is not provided, we evaluate all available flags
	response := model.OFREPBulkEvaluateSuccessResponse{
		Flags: make([]model.OFREPFlagBulkEvaluateSuccessResponse, 0),
	}
	tracer := otel.GetTracerProvider().Tracer(config.OtelTracerName)
	_, span := tracer.Start(c.Request().Context(), "AllFlagsState")
	defer span.End()

	allFlagsResp := h.goFF.AllFlagsState(evalCtx)
	for key, val := range allFlagsResp.GetFlags() {
		value := val.Value
		if val.Reason == flag.ReasonError {
			value = nil
		}
		response.Flags = append(response.Flags, model.OFREPFlagBulkEvaluateSuccessResponse{
			OFREPEvaluateSuccessResponse: model.OFREPEvaluateSuccessResponse{
				Key:      key,
				Value:    value,
				Reason:   val.Reason,
				Variant:  val.VariationType,
				Metadata: val.Metadata,
			},
			ErrorCode: val.ErrorCode,
		})
	}

	sort.Slice(response.Flags, func(i, j int) bool {
		return response.Flags[i].Key < response.Flags[j].Key
	})

	span.SetAttributes(
		attribute.Int("AllFlagsState.numberEvaluation", len(response.Flags)),
	)

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	return c.JSON(http.StatusOK, response)
}

func assertOFREPEvaluateRequest(ofrepEvalReq *model.OFREPEvalFlagRequest) *model.OFREPCommonErrorResponse {
	if ofrepEvalReq.Context == nil || ofrepEvalReq.Context["targetingKey"] == "" {
		return NewOFREPCommonError(flag.ErrorCodeTargetingKeyMissing,
			"GO Feature Flag MUST have a targeting key in the request.")
	}
	return nil
}

func evaluationContextFromOFREPRequest(ctx map[string]any) (ffcontext.Context, error) {
	if targetingKey, ok := ctx["targetingKey"].(string); ok {
		delete(ctx, "targetingKey")
		evalCtx := utils.ConvertEvaluationCtxFromRequest(targetingKey, ctx)
		return evalCtx, nil
	}
	return ffcontext.EvaluationContext{}, NewOFREPCommonError(
		flag.ErrorCodeTargetingKeyMissing,
		"GO Feature Flag has received no targetingKey or a none string value that is not a string.")
}
