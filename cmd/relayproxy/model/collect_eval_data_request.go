package model

import (
	"github.com/thomaspoignant/go-feature-flag/exporter"
)

// CollectEvalDataRequest is the request to collect data in
type CollectEvalDataRequest struct {
	// Meta are the extra information added during the configuration
	Meta map[string]string `json:"meta"`

	// Events is the list of the event we send in the payload
	Events []exporter.FeatureEvent `json:"events"`
}
