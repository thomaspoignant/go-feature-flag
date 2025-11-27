package ofrep

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/helper"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/internal/flagstate"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type EvaluateCtrl struct {
	flagsetManager service.FlagsetManager
	metrics        metric.Metrics
}

func NewOFREPEvaluate(flagsetManager service.FlagsetManager, metrics metric.Metrics) EvaluateCtrl {
	return EvaluateCtrl{
		flagsetManager: flagsetManager,
		metrics:        metrics,
	}
}

// Evaluate is the entry point to evaluate a flag using the OpenFeature Remote Evaluation Protocol
// @Summary     Evaluate a feature flag using the OpenFeature Remote Evaluation Protocol
// @Tags OpenFeature Remote Evaluation Protocol (OFREP)
// @Description Making a **POST** request to the URL `/ofrep/v1/evaluate/flags/{your_flag_name}` will give you the
// @Description value of the flag for this evaluation context
// @Description
// @Security     ApiKeyAuth
// @Security     XApiKeyAuth
// @Produce      json
// @Accept	 	 json
// @Param 		 data body model.OFREPEvalFlagRequest true "Evaluation Context for this API call"
// @Param        flag_key path string true "Name of your feature flag"
// @Success      200  {object} model.OFREPEvaluateSuccessResponse "Success"
// @Failure      400 {object}  model.OFREPEvaluateResponseError "Bad Request"
// @Failure      401 {object}  modeldocs.HTTPErrorDoc "Unauthorized"
// @Failure      404 {object}  model.OFREPEvaluateResponseError "Flag Not Found"
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
		return c.JSON(http.StatusBadRequest, model.OFREPEvaluateResponseError{
			OFREPCommonResponseError: *err,
			Key:                      flagKey,
		})
	}
	evalCtx, err := evaluationContextFromOFREPRequest(reqBody)
	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			err)
	}

	tracer := otel.GetTracerProvider().Tracer(config.OtelTracerName)
	_, span := tracer.Start(c.Request().Context(), "flagEvaluation")
	defer span.End()

	flagset, httpErr := helper.GetFlagSet(h.flagsetManager, helper.GetAPIKey(c))
	if httpErr != nil {
		return httpErr
	}

	defaultValue := "thisisadefaultvaluethatItest1233%%"
	flagValue, _ := flagset.RawVariation(flagKey, evalCtx, defaultValue)

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

	metadata := flagValue.Metadata
	if flagValue.Cacheable {
		if metadata == nil {
			metadata = make(map[string]interface{})
		}
		metadata["gofeatureflag_cacheable"] = true
	}

	return c.JSON(http.StatusOK, model.OFREPEvaluateSuccessResponse{
		Key:      flagKey,
		Value:    flagValue.Value,
		Reason:   flagValue.Reason,
		Variant:  flagValue.VariationType,
		Metadata: metadata,
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
// @Security    XApiKeyAuth
// @Produce     json
// @Accept	 	json
// @Param       If-None-Match header string false "The request will be processed only if ETag doesn't match."
// @Param 		data body model.OFREPEvalFlagRequest true "Evaluation Context and list of flag for this API call"
// @Success     200 {object} model.OFREPBulkEvaluateSuccessResponse "OFREP successful evaluation response"
// @Success     304 {string} string "Etag: \"117-0193435c612c50d93b798619d9464856263dbf9f\""
// @Failure     400 {object}  model.OFREPCommonResponseError "Bad evaluation request"
// @Failure     401 {object}  modeldocs.HTTPErrorDoc "Unauthorized - You need credentials to access the API"
// @Failure     403 {object}  modeldocs.HTTPErrorDoc "Forbidden - You are not authorized to access the API"
// @Failure     500 {object}  modeldocs.HTTPErrorDoc "Internal server error"
// @Router      /ofrep/v1/evaluate/flags [post]
func (h *EvaluateCtrl) BulkEvaluate(c echo.Context) error {
	request := new(model.OFREPEvalFlagRequest)
	if err := c.Bind(request); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			NewOFREPCommonError("INVALID_CONTEXT", err.Error()))
	}
	if err := assertOFREPEvaluateRequest(request); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	evalCtx, err := evaluationContextFromOFREPRequest(request)
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

	flagset, httpErr := helper.GetFlagSet(h.flagsetManager, helper.GetAPIKey(c))
	if httpErr != nil {
		return httpErr
	}

	var allFlagsResp flagstate.AllFlags
	if len(evalCtx.ExtractGOFFProtectedFields().FlagList) > 0 {
		// if we have a list of flags to evaluate in the evaluation context, we evaluate only those flags.
		allFlagsResp = flagset.GetFlagStates(evalCtx, evalCtx.ExtractGOFFProtectedFields().FlagList)
	} else {
		allFlagsResp = flagset.AllFlagsState(evalCtx)
	}
	flagNames := make([]string, 0, len(allFlagsResp.GetFlags()))
	for key, val := range allFlagsResp.GetFlags() {
		flagNames = append(flagNames, key)
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
			ErrorCode:    val.ErrorCode,
			ErrorDetails: val.ErrorDetails,
		})
	}
	h.metrics.IncAllFlag(flagNames...)

	sort.Slice(response.Flags, func(i, j int) bool {
		return response.Flags[i].Key < response.Flags[j].Key
	})

	span.SetAttributes(
		attribute.Int("AllFlagsState.numberEvaluation", len(response.Flags)),
	)

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.JSON(http.StatusOK, response)
}

func assertOFREPEvaluateRequest(
	ofrepEvalReq *model.OFREPEvalFlagRequest,
) *model.OFREPCommonResponseError {
	if ofrepEvalReq.Context == nil {
		return NewOFREPCommonError(flag.ErrorCodeInvalidContext,
			"GO Feature Flag requires an evaluation context in the request.")
	}

	// An empty context object is allowed since the evaluation context is optional.
	// If the context does not have any targetingKey, this is fine since the core
	// evaluation logic will handle if it is required or not.

	return nil
}

func evaluationContextFromOFREPRequest(req *model.OFREPEvalFlagRequest) (ffcontext.Context, error) {
	if req == nil || req.Context == nil {
		return ffcontext.EvaluationContext{}, NewOFREPCommonError(
			flag.ErrorCodeInvalidContext,
			"GO Feature Flag has received an invalid context.")
	}

	ctx := req.Context

	// targetingKey is optional, it is only required if the flag needs bucketing and
	// the check is done in the core evaluation logic.
	// If we don't have a targetingKey, we return an empty string.
	targetingKey := ""
	if key, ok := ctx["targetingKey"].(string); ok {
		targetingKey = key
	}

	// Create evaluation context (empty targeting key is allowed)
	evalCtx := utils.ConvertEvaluationCtxFromRequest(targetingKey, ctx)
	return evalCtx, nil
}
