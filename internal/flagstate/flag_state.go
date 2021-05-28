package flagstate

import (
	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"time"
)

// NewFlagState is creating a state for a flag.
func NewFlagState(
	trackEvents bool,
	value interface{},
	variationType model.VariationType,
	failed bool) FlagState {
	return FlagState{
		Value:         value,
		Timestamp:     time.Now().Unix(),
		VariationType: variationType,
		Failed:        failed,
		TrackEvents:   trackEvents,
	}
}

// FlagState represents the state of an individual feature flag, with regard to a specific user, when it was called.
type FlagState struct {
	Value         interface{}         `json:"value"`
	Timestamp     int64               `json:"timestamp"`
	VariationType model.VariationType `json:"variationType"`
	TrackEvents   bool                `json:"trackEvents"`
	Failed        bool                `json:"-"`
}
