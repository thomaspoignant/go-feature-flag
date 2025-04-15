package model

import (
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

type EvaluateInput struct {
	FlagKey         string                      `json:"flagKey"`
	Flag            flag.InternalFlag           `json:"flag"`
	EvaluationCtx   ffcontext.EvaluationContext `json:"evaluationContext"`
	FlagContext     flag.Context                `json:"flagContext"`
	SdkDefaultValue interface{}                 `json:"sdkDefaultValue"`
}
