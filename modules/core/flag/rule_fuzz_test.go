package flag

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/diegoholiveira/jsonlogic/v3"
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
