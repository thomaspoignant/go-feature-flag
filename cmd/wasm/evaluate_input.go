package main

import (
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

type EvaluateInput struct {
	FlagKey       string            `json:"flagKey"`
	Flag          flag.InternalFlag `json:"flag"`
	EvaluationCtx map[string]any    `json:"evalContext"`
	FlagContext   flag.Context      `json:"flagContext"`
}
