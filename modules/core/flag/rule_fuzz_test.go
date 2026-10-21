package flag

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

// FuzzEvaluateRule checks that a query accepted by the rule validation never
// panics in the query engines. evaluateRule hides such panics behind a recover
// (the rule silently does not match), so this calls the engines the same way
// evaluateRule does, but without the recover.
func FuzzEvaluateRule(f *testing.F) {
	ctx := `{"key": "abc", "version": "1.0.0", "age": 42, "company": {"name": "goff"}, "tags": ["a", "b"]}`
	for _, query := range []string{
		`key eq "abc"`,
		`key eq "test" and version gt 1.0.0`,
		`version gt 1.0.0-10`,
		`(age gt 10 or company.name eq "goff") and not (key in ["x", "y"])`,
		`tags co "a"`,
		`{"==": [{"var": "key"}, "abc"]}`,
		`{"in": [{"var": "key"}, ["abcdef", "1234"]]}`,
		`{"cat": ["t", "r", "u", "e"]}`,
		`{"and": [{">": [{"var": "age"}, 18]}, {"==": [{"var": "company.name"}, "goff"]}]}`,
	} {
		f.Add(query, ctx)
	}
	f.Fuzz(func(t *testing.T, query, ctxJSON string) {
		var attributes map[string]any
		if json.Unmarshal([]byte(ctxJSON), &attributes) != nil {
			return
		}
		r := Rule{Query: &query, VariationResult: testconvert.String("A")}
		if r.isQueryValid(false) != nil {
			return
		}

		mapCtx := utils.ContextToMap(utils.ConvertEvaluationCtxFromRequest("key", attributes))
		switch r.GetQueryFormat() {
		case JSONLogicQueryFormat:
			strCtx, err := json.Marshal(mapCtx)
			require.NoError(t, err)
			_ = jsonlogic.Apply(strings.NewReader(r.GetTrimmedQuery()), bytes.NewReader(strCtx), io.Discard)
		default:
			ev, err := getNikunjyEvaluator(r.GetTrimmedQuery())
			require.NoError(t, err, "query passed validation but does not parse")
			_, _ = ev.process(mapCtx)
		}
	})
}

// FuzzPercentageRollout checks that a percentage rule accepted by the
// validation always serves one of its variations, and never one with a
// non-positive share.
func FuzzPercentageRollout(f *testing.F) {
	f.Add("user-1", "my-flag", 10.0, 80.0, 10.0)
	f.Add("user-2", "my-flag", 100.0, 0.0, 0.0)
	f.Add("", "", 33.3, 33.3, 33.4)
	f.Add("user-3", "other-flag", 0.5, 0.25, 0.25)
	f.Fuzz(func(t *testing.T, key, flagName string, pA, pB, pC float64) {
		percentages := map[string]float64{"A": pA, "B": pB, "C": pC}
		r := Rule{Percentages: &percentages}
		if hasKnownValidationGap(r) || r.IsValid(true, fuzzVariations("A", "B", "C")) != nil {
			return
		}
		got, err := r.EvaluatePercentageRollout(key, flagName)
		require.NoError(t, err, "percentages: %v", percentages)
		require.Contains(t, percentages, got)
		assert.Positive(t, percentages[got], "served a variation with no share: %v", percentages)
	})
}

// FuzzProgressiveRollout checks that a progressive rollout accepted by the
// validation always serves the initial or the end variation, and that a user
// who received the end variation keeps it as time goes on.
func FuzzProgressiveRollout(f *testing.F) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	f.Add("user-1", 0.0, 100.0, start, int64(3600), int64(-10), int64(1800))
	f.Add("user-2", 10.0, 50.0, start, int64(60), int64(30), int64(30))
	f.Add("user-3", 0.0, 0.0, start, int64(86400), int64(0), int64(90000))
	f.Fuzz(func(t *testing.T, key string, initialPct, endPct float64, startUnix, duration, offset1, offset2 int64) {
		initialDate := time.Unix(startUnix, 0)
		endDate := time.Unix(startUnix+duration, 0)
		r := Rule{ProgressiveRollout: &ProgressiveRollout{
			Initial: &ProgressiveRolloutStep{
				Variation: testconvert.String("A"), Percentage: &initialPct, Date: &initialDate,
			},
			End: &ProgressiveRolloutStep{
				Variation: testconvert.String("B"), Percentage: &endPct, Date: &endDate,
			},
		}}
		if hasKnownValidationGap(r) || r.IsValid(true, fuzzVariations("A", "B")) != nil {
			return
		}
		if offset1 > offset2 {
			offset1, offset2 = offset2, offset1
		}
		first, err := r.EvaluateProgressiveRollout(key, "my-flag", time.Unix(startUnix+offset1, 0))
		require.NoError(t, err)
		later, err := r.EvaluateProgressiveRollout(key, "my-flag", time.Unix(startUnix+offset2, 0))
		require.NoError(t, err)
		require.Contains(t, []string{"A", "B"}, first)
		if first == "B" {
			assert.Equal(t, "B", later, "user went back to the initial variation")
		}
	})
}

// hasKnownValidationGap reports rules that the validation wrongly accepts today,
// and that are then served wrongly or fail for every user:
//   - a percentage rule with a negative or NaN share, or a sum too small to bucket anyone,
//   - a progressive rollout without initial or end step (IsValid panics),
//   - a progressive rollout with a missing date or an end date not after the initial
//     date (every evaluation fails with "missing params"),
//   - a progressive rollout with a negative or NaN percentage (users flip back to
//     the initial variation).
//
// ponytail: skips known gaps so the fuzzers keep looking for new ones, drop it
// once the validation rejects these rules.
func hasKnownValidationGap(r Rule) bool {
	if r.Percentages != nil {
		sum := 0.0
		for _, p := range r.GetPercentages() {
			if p < 0 || math.IsNaN(p) {
				return true
			}
			sum += p
		}
		if sum*PercentageMultiplier < 1 {
			return true
		}
	}
	if p := r.ProgressiveRollout; p != nil {
		if p.Initial == nil || p.End == nil {
			return true
		}
		if p.Initial.Date == nil || p.End.Date == nil || !p.End.Date.After(*p.Initial.Date) {
			return true
		}
		initialPct, endPct := p.Initial.getPercentage(), p.End.getPercentage()
		if initialPct < 0 || math.IsNaN(initialPct) || math.IsNaN(endPct) {
			return true
		}
	}
	return false
}

func fuzzVariations(names ...string) map[string]*any {
	variations := make(map[string]*any, len(names))
	for _, name := range names {
		variations[name] = testconvert.Interface(name)
	}
	return variations
}
