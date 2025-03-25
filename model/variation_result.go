package model

import "github.com/thomaspoignant/go-feature-flag/internal/flag"

// JSONType contains all acceptable flag value types
type JSONType interface {
	float64 | int | string | bool | any | map[string]any
}

// VariationResult contains all the field available in a flag variation result.
type VariationResult[T JSONType] struct {
	TrackEvents   bool                  `json:"trackEvents"`
	VariationType string                `json:"variationType"`
	Failed        bool                  `json:"failed"`
	Version       string                `json:"version"`
	Reason        flag.ResolutionReason `json:"reason"`
	ErrorCode     flag.ErrorCode        `json:"errorCode"`
	ErrorDetails  string                `json:"errrorDetails,omitempty"`
	Value         T                     `json:"value"`
	Cacheable     bool                  `json:"cacheable"`
	Metadata      map[string]any        `json:"metadata,omitempty"`
}

// RawVarResult is the result of the raw variation call.
// This is used by ffclient.RawVariation functions, this should be used only by internal calls.
type RawVarResult struct {
	TrackEvents   bool                  `json:"trackEvents"`
	VariationType string                `json:"variationType"`
	Failed        bool                  `json:"failed"`
	Version       string                `json:"version"`
	Reason        flag.ResolutionReason `json:"reason"`
	ErrorCode     flag.ErrorCode        `json:"errorCode"`
	ErrorDetails  string                `json:"errorDetails,omitempty"`
	Value         any                   `json:"value"`
	Cacheable     bool                  `json:"cacheable"`
	Metadata      map[string]any        `json:"metadata,omitempty"`
}
