package ofrep

import (
	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/ofrep/customerr"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	http "net/http"
)

type ofrepBulkEvaluateCtrl struct {
	goFF    *ffclient.GoFeatureFlag
	metrics metric.Metrics
}

func NewOFREPBulkEvaluate(goFF *ffclient.GoFeatureFlag, metrics metric.Metrics) Controller {
	return &ofrepBulkEvaluateCtrl{
		goFF:    goFF,
		metrics: metrics,
	}
}

// OFREPHandler is the entry point to evaluate in bulk flags using the OpenFeature Remote Evaluation Protocol
// @Summary     Evaluate in bulk feature flags using the OpenFeature Remote Evaluation Protocol
// @Description Making a **POST** request to the URL `/ofrep/v1/evaluate` will give you the value of the list of feature
// @Description flags for this evaluation context.
// @Description
// @Description If no flags are provided, the API will evaluate all available flags in the configuration.
// @Security     ApiKeyAuth
// @Produce      json
// @Accept	 	 json
// @Param 		 data body model.OFREPBulkEvalFlagRequest true "Evaluation Context and list of flag for this API call"
// @Success      200  {object} model.OFREPBulkEvaluateSuccessResponse "Success"
// @Failure      400 {object}  model.OFREPErrorResponse "Bad Request"
// @Failure      401 {object}  modeldocs.HTTPErrorDoc "Unauthorized"
// @Failure      500 {object}  modeldocs.HTTPErrorDoc "Internal server error"
// @Router       /ofrep/v1/evaluate [post]
func (h *ofrepBulkEvaluateCtrl) OFREPHandler(c echo.Context) error {
	// h.metrics.IncFlagEvaluation(flagKey)

	request := new(model.OFREPBulkEvalFlagRequest)
	if err := c.Bind(request); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			customerr.NewOFREPGenericError("INVALID_CONTEXT", err.Error()).ToOFRErrorResponse())
	}
	if err := assertOFREPBulkEvaluateRequest(request); err != nil {
		return c.JSON(http.StatusBadRequest, err.ToOFRErrorResponse())
	}
	evalCtx, err := evaluationContextFromOFREPRequest(request.Context)
	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			err)
	}

	flags, _ := h.goFF.GetFlagsFromCache()

	if request.Flags != nil && len(request.Flags) > 0 {
		// if a list of flag is provided we evaluate all flags from the list
		response := model.OFREPBulkEvaluateSuccessResponse{}
		for _, key := range request.Flags {
			// TODO: check to change this
			defaultValue := "thisisadefaultvaluethatItest1233%%"

			evalResp, _ := h.goFF.RawVariation(key, evalCtx, defaultValue)

			value := evalResp.Value
			if evalResp.Reason == flag.ReasonError {
				value = nil
			}

			response = append(response, model.OFREPFlagBulkEvaluateSuccessResponse{
				OFREPEvaluateSuccessResponse: model.OFREPEvaluateSuccessResponse{
					Key:      key,
					Value:    value,
					Reason:   evalResp.Reason,
					Variant:  evalResp.VariationType,
					Metadata: evalResp.Metadata,
				},
				ErrorCode: evalResp.ErrorCode,
				ETag:      flagCheckSum(flags[key]),
			})
		}
		return c.JSON(http.StatusOK, response)
	}

	// if no flag is not provided, we evaluate all available flags
	response := model.OFREPBulkEvaluateSuccessResponse{}
	allFlagsResp := h.goFF.AllFlagsState(evalCtx)
	for key, val := range allFlagsResp.GetFlags() {
		value := val.Value
		if val.Reason == flag.ReasonError {
			value = nil
		}
		response = append(response, model.OFREPFlagBulkEvaluateSuccessResponse{
			OFREPEvaluateSuccessResponse: model.OFREPEvaluateSuccessResponse{
				Key:      key,
				Value:    value,
				Reason:   val.Reason,
				Variant:  val.VariationType,
				Metadata: val.Metadata,
			},
			ErrorCode: val.ErrorCode,
			ETag:      flagCheckSum(flags[key]),
		})
	}
	return c.JSON(http.StatusOK, response)
}

// tracer := otel.GetTracerProvider().Tracer(config.OtelTracerName)
// _, span := tracer.Start(c.Request().Context(), "flagEvaluation")
// defer span.End()
// defaultValue := "thisisadefaultvaluethatItest1233%%"
// flagValue, _ := h.goFF.RawVariation(flagKey, evalCtx, defaultValue)
//
// if flagValue.Reason == flag.ReasonError {
//	httpStatus := http.StatusBadRequest
//	if flagValue.ErrorCode == flag.ErrorCodeFlagNotFound {
//		httpStatus = http.StatusNotFound
//	}
//	return c.JSON(
//		httpStatus,
//		customErr.NewOFREPEvaluateError(flagKey, flagValue.ErrorCode,
//			fmt.Sprintf("Error while evaluating the flag: %s", flagValue.ErrorCode)).ToOFRErrorResponse())
//}
//
// span.SetAttributes(
//	attribute.String("flagEvaluation.flagName", flagKey),
//	attribute.Bool("flagEvaluation.trackEvents", flagValue.TrackEvents),
//	attribute.String("flagEvaluation.variant", flagValue.VariationType),
//	attribute.Bool("flagEvaluation.failed", flagValue.Failed),
//	attribute.String("flagEvaluation.version", flagValue.Version),
//	attribute.String("flagEvaluation.reason", flagValue.Reason),
//	attribute.String("flagEvaluation.errorCode", flagValue.ErrorCode),
//	attribute.Bool("flagEvaluation.cacheable", flagValue.Cacheable),
//	// we convert to string because there is no attribute for interface{}
//	attribute.String("flagEvaluation.value", fmt.Sprintf("%v", flagValue.Value)),

func assertOFREPBulkEvaluateRequest(ofrepEvalReq *model.OFREPBulkEvalFlagRequest) *customerr.OfrepGenericError {
	if ofrepEvalReq.Context == nil || ofrepEvalReq.Context["targetingKey"] == "" {
		return customerr.NewOFREPGenericError(
			"TARGETING_KEY_MISSING", "GO Feature Flag MUST have a targeting key in the request.")
	}

	return nil
}
