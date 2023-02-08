package model

import (
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

type JSONType interface {
	float64 | int | string | bool | interface{} | map[string]interface{}
}

type GenericVariationResult[T JSONType] struct {
	TrackEvents   bool                  `json:"trackEvents"`
	VariationType string                `json:"variationType"`
	Failed        bool                  `json:"failed"`
	Version       string                `json:"version"`
	Reason        flag.ResolutionReason `json:"reason"`
	ErrorCode     flag.ErrorCode        `json:"errorCode"`
	Value         T                     `json:"value"`
}
