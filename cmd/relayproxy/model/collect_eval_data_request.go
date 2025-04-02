package model

import (
	"github.com/thomaspoignant/go-feature-flag/exporter"
)

// CollectEvalDataRequest is the request to collect data in
type CollectEvalDataRequest struct {
	// Meta are the extra information added during the configuration
	Meta exporter.FeatureEventMetadata `json:"meta"`

	// Events is the list of the event we send in the payload
	// here the type is any because we will unmarshal later in the different event types
	Events []map[string]any `json:"events"`
}
