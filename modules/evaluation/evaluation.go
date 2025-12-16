package evaluation

import (
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/model"
)

const errorWrongVariation = "wrong variation used for flag %v"

func Evaluate[T model.JSONType](
	f flag.Flag,
	flagKey string,
	evaluationCtx ffcontext.Context,
	flagCtx flag.Context,
	expectedType string,
	sdkDefaultValue T) (model.VariationResult[T], error) {
	flagValue, resolutionDetails := f.Value(flagKey, evaluationCtx, flagCtx)
	var convertedValue any
	switch value := flagValue.(type) {
	case float64:
		// this part ensures that we convert float64 value into int if we call IntVariation on a float64 value.
		if expectedType == "int" {
			convertedValue = int(value)
		} else {
			convertedValue = value
		}
	default:
		convertedValue = value
	}

	var v T
	switch val := convertedValue.(type) {
	case T:
		v = val
	default:
		if val != nil {
			return model.VariationResult[T]{
				Value:         sdkDefaultValue,
				VariationType: flag.VariationSDKDefault,
				Reason:        flag.ReasonError,
				ErrorCode:     flag.ErrorCodeTypeMismatch,
				Failed:        true,
				TrackEvents:   f.IsTrackEvents(),
				Version:       f.GetVersion(),
				Metadata:      f.GetMetadata(),
			}, fmt.Errorf(errorWrongVariation, flagKey)
		}
	}
	return model.VariationResult[T]{
		Value:         v,
		VariationType: resolutionDetails.Variant,
		Reason:        resolutionDetails.Reason,
		ErrorCode:     resolutionDetails.ErrorCode,
		ErrorDetails:  resolutionDetails.ErrorMessage,
		Failed:        resolutionDetails.ErrorCode != "",
		TrackEvents:   f.IsTrackEvents(),
		Version:       f.GetVersion(),
		Cacheable:     resolutionDetails.Cacheable,
		Metadata:      resolutionDetails.Metadata,
	}, nil
}
