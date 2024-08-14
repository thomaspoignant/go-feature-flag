package flagstate

import (
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"time"
)

// FlagState represents the state of an individual feature flag, with regard to a specific user, when it was called.
type FlagState struct {
	Value         interface{}            `json:"value"`
	Timestamp     int64                  `json:"timestamp"`
	VariationType string                 `json:"variationType"`
	TrackEvents   bool                   `json:"trackEvents"`
	Failed        bool                   `json:"-"`
	ErrorCode     flag.ErrorCode         `json:"errorCode"`
	Reason        flag.ResolutionReason  `json:"reason"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

func FromFlagEvaluation(key string, evaluationCtx ffcontext.Context,
	flagCtx flag.Context, currentFlag flag.Flag) FlagState {
	flagValue, resolutionDetails := currentFlag.Value(key, evaluationCtx, flagCtx)

	// if the flag is disabled, we are ignoring it.
	if resolutionDetails.Reason == flag.ReasonDisabled {
		return FlagState{
			Timestamp:   time.Now().Unix(),
			TrackEvents: currentFlag.IsTrackEvents(),
			Failed:      resolutionDetails.ErrorCode != "",
			ErrorCode:   resolutionDetails.ErrorCode,
			Reason:      resolutionDetails.Reason,
			Metadata:    resolutionDetails.Metadata,
		}
	}

	switch v := flagValue; v.(type) {
	case int, float64, bool, string, []interface{}, map[string]interface{}:
		return FlagState{
			Value:         v,
			Timestamp:     time.Now().Unix(),
			VariationType: resolutionDetails.Variant,
			TrackEvents:   currentFlag.IsTrackEvents(),
			Failed:        resolutionDetails.ErrorCode != "",
			ErrorCode:     resolutionDetails.ErrorCode,
			Reason:        resolutionDetails.Reason,
			Metadata:      resolutionDetails.Metadata,
		}

	default:
		defaultVariationName := flag.VariationSDKDefault
		defaultVariationValue := currentFlag.GetVariationValue(defaultVariationName)
		return FlagState{
			Value:         defaultVariationValue,
			Timestamp:     time.Now().Unix(),
			VariationType: defaultVariationName,
			TrackEvents:   currentFlag.IsTrackEvents(),
			Failed:        true,
			ErrorCode:     flag.ErrorCodeTypeMismatch,
			Reason:        flag.ReasonError,
			Metadata:      resolutionDetails.Metadata,
		}
	}
}
