package main

import (
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

type EvaluateInput struct {
	FlagKey string            `json:"flagKey"`
	Flag    flag.InternalFlag `json:"flag"`
	// ParentFlags holds the parents of the flag, the host sends only the current flag
	// and its transitive parents to keep the payload small. The current flag is added
	// to the set by localEvaluation, so a self-dependency is detected as a cycle.
	ParentFlags   map[string]flag.InternalFlag `json:"parentFlags,omitempty"`
	EvaluationCtx map[string]any               `json:"evalContext"`
	FlagContext   flag.Context                 `json:"flagContext"`
}
