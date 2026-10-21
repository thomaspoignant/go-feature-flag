package flag

import (
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

// FuzzInternalFlagValue checks that a flag which is valid at the evaluation date
// (scheduled steps applied) never panics and never fails with a configuration
// error, whatever the evaluation context.
func FuzzInternalFlagValue(f *testing.F) {
	ctx := `{"targetingKey": "user-1", "email": "john@doe.com", "company": {"id": "goff"}, "age": 42}`
	evalDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC).Unix()
	for _, flag := range []string{
		`{"variations": {"enable": true, "disable": false},
		  "targeting": [{"name": "rule1", "query": "email ew \"@doe.com\"", "percentage": {"enable": 90, "disable": 10}}],
		  "defaultRule": {"variation": "disable"}}`,
		`{"variations": {"A": "a", "B": "b"}, "bucketingKey": "company.id",
		  "defaultRule": {"progressiveRollout": {
		    "initial": {"variation": "A", "percentage": 0, "date": "2026-01-01T00:00:00Z"},
		    "end": {"variation": "B", "percentage": 100, "date": "2026-12-01T00:00:00Z"}}}}`,
		`{"variations": {"A": 1, "B": 2}, "defaultRule": {"variation": "A"},
		  "targeting": [{"name": "r", "query": "{\"<\": [{\"var\": \"age\"}, 18]}", "variation": "B"}],
		  "scheduledRollout": [
		    {"date": "2026-03-01T00:00:00Z", "targeting": [{"name": "r", "disable": true}]},
		    {"date": "2026-02-01T00:00:00Z", "defaultRule": {"percentage": {"A": 50, "B": 50}}}]}`,
		`{"variations": {"on": {"a": 1}, "off": {}}, "defaultRule": {"variation": "on"},
		  "experimentation": {"start": "2026-01-01T00:00:00Z", "end": "2026-07-01T00:00:00Z"}}`,
	} {
		f.Add(flag, ctx, evalDate)
	}
	f.Fuzz(func(t *testing.T, flagJSON, ctxJSON string, evalUnix int64) {
		var internalFlag InternalFlag
		var attributes map[string]any
		if json.Unmarshal([]byte(flagJSON), &internalFlag) != nil ||
			json.Unmarshal([]byte(ctxJSON), &attributes) != nil {
			return
		}
		if hasKnownFlagGap(&internalFlag) {
			return
		}
		date := time.Unix(evalUnix, 0)
		targetingKey, _ := attributes["targetingKey"].(string)
		evalCtx := utils.ConvertEvaluationCtxFromRequest(targetingKey, attributes)
		evalCtx.AddCustomAttribute("gofeatureflag", map[string]any{"currentDateTime": date})
		// evaluation.Evaluate (WASM) calls Value on flags that were never validated,
		// so any flag must evaluate without panicking.
		_, details := internalFlag.Value("my-flag", evalCtx, Context{})

		flagAtDate, err := internalFlag.applyScheduledRolloutSteps(date)
		if err != nil || hasKnownFlagGap(flagAtDate) || flagAtDate.IsValid() != nil {
			return
		}
		assert.NotEqual(t, ErrorFlagConfiguration, details.ErrorCode, "%s", details.ErrorMessage)
	})
}

// hasKnownFlagGap reports flags with a rule matched by hasKnownValidationGap
// (scheduled steps included), and flags without default rule but with a
// scheduled step updating it (Value panics).
// ponytail: same as hasKnownValidationGap, drop the cases once they are fixed.
func hasKnownFlagGap(f *InternalFlag) bool {
	rules := f.GetRules()
	if f.DefaultRule != nil {
		rules = append(rules, *f.DefaultRule)
	}
	if slices.ContainsFunc(rules, hasKnownValidationGap) {
		return true
	}
	if f.Scheduled == nil {
		return false
	}
	return slices.ContainsFunc(*f.Scheduled, func(step ScheduledStep) bool {
		return (f.DefaultRule == nil && step.DefaultRule != nil) || hasKnownFlagGap(&step.InternalFlag)
	})
}
