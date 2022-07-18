package flagstate

import (
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

// FlagState represents the state of an individual feature flag, with regard to a specific user, when it was called.
type FlagState struct {
	Value         interface{}           `json:"value"`
	Timestamp     int64                 `json:"timestamp"`
	VariationType string                `json:"variationType"`
	TrackEvents   bool                  `json:"trackEvents"`
	Failed        bool                  `json:"-"`
	ErrorCode     flag.ErrorCode        `json:"errorCode"`
	Reason        flag.ResolutionReason `json:"reason"`
}
