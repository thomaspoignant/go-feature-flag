package model

import (
	"github.com/thomaspoignant/go-feature-flag/exporter"
)

// CollectEvalDataRequest is the request to collect data in
type CollectEvalDataRequest struct {
	// Data contains the list of feature event that we want to store
	Data []exporter.FeatureEvent `json:"data"`
}
