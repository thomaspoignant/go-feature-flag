package main

import (
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

type EvaluateInput struct {
	FlagKey       string            `json:"flagKey"`
	Flag          flag.InternalFlag `json:"flag"`
	EvaluationCtx map[string]any    `json:"evalContext"`
	FlagContext   flag.Context      `json:"flagContext"`
}
