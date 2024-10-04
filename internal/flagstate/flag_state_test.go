package flagstate_test

import (
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/flagstate"
)

func TestFromFlagEvaluation(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		evaluationCtx ffcontext.Context
		flagCtx       flag.Context
		currentFlag   flag.Flag
		expected      flagstate.FlagState
	}{
		{
			name:          "Flag is disabled",
			key:           "test-key",
			evaluationCtx: ffcontext.NewEvaluationContext("user-key"),
			flagCtx:       flag.Context{},
			currentFlag: &flag.InternalFlag{
				Disable: testconvert.Bool(true),
			},
			expected: flagstate.FlagState{
				Timestamp:    time.Now().Unix(),
				TrackEvents:  true,
				Failed:       false,
				ErrorCode:    "",
				ErrorDetails: "",
				Reason:       flag.ReasonDisabled,
				Metadata:     nil,
			},
		},
		{
			name:          "Flag evaluation error",
			key:           "test-key",
			evaluationCtx: ffcontext.NewEvaluationContext(""),
			flagCtx:       flag.Context{},
			currentFlag: &flag.InternalFlag{
				Disable: testconvert.Bool(false),
				Variations: &map[string]*interface{}{
					"var1": testconvert.Interface(1),
					"var2": testconvert.Interface(2),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("var1"),
				},
			},
			expected: flagstate.FlagState{
				Timestamp:    time.Now().Unix(),
				TrackEvents:  true,
				Failed:       true,
				ErrorCode:    flag.ErrorCodeTargetingKeyMissing,
				ErrorDetails: "Error: Empty bucketing key",
				Reason:       flag.ReasonError,
				Metadata:     nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flagstate.FromFlagEvaluation(tt.key, tt.evaluationCtx, tt.flagCtx, tt.currentFlag)
			assert.Equal(t, tt.expected, got)
		})
	}
}
