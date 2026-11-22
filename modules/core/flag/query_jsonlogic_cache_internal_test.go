package flag

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"sync"
	"testing"

	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

func TestGetCompiledQuery(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		wantTrimmed string
		wantFormat  QueryFormat
		wantDecoded bool
	}{
		{name: "nikunjy", raw: `key eq "abc"`, wantTrimmed: `key eq "abc"`, wantFormat: NikunjyQueryFormat},
		{name: "jsonlogic", raw: `{"==": [{"var": "key"}, "abc"]}`, wantTrimmed: `{"==": [{"var": "key"}, "abc"]}`, wantFormat: JSONLogicQueryFormat, wantDecoded: true},
		{name: "multi-line jsonlogic", raw: "{\n  \"==\": [{\"var\": \"key\"}, \"abc\"]\n}", wantTrimmed: `{"==": [{"var": "key"}, "abc"]}`, wantFormat: JSONLogicQueryFormat, wantDecoded: true},
		{name: "invalid json", raw: `{"==": [{"var": "key"}, "abc"]`, wantTrimmed: `{"==": [{"var": "key"}, "abc"]`, wantFormat: NikunjyQueryFormat},
		{name: "empty", raw: "", wantTrimmed: "", wantFormat: NikunjyQueryFormat},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCompiledQuery(tt.raw)
			assert.Equal(t, tt.wantTrimmed, c.trimmed)
			assert.Equal(t, tt.wantFormat, c.format)
			assert.Equal(t, tt.wantDecoded, c.jsonLogic != nil)
			assert.Same(t, c, getCompiledQuery(tt.raw), "the query must be compiled once")
		})
	}
}

func TestIsJSONLogicTruthy_NonFinite(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "infinity", value: math.Inf(1)},
		{name: "NaN", value: math.NaN()},
		{name: "infinity in an array", value: []any{true, math.Inf(-1)}},
		{name: "NaN in an object", value: map[string]any{"a": map[string]any{"b": math.NaN()}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.False(t, isJSONLogicTruthy(tt.value))
		})
	}
}

// TestEvaluateRule_JSONLogicMatchesApply checks the cached path gives the same answer as jsonlogic.Apply.
func TestEvaluateRule_JSONLogicMatchesApply(t *testing.T) {
	queries := []string{
		`{"==": [{"var": "key"}, "abc"]}`,
		`{"and": [{"in": [{"var": "dsp"}, ["ttd", "dv360"]]}, {">": [{"var": "floor"}, 0.5]}]}`,
		`{"or": [{"==": [{"var": "anonymous"}, true]}, {"!": {"var": "missing"}}]}`,
		`{"cat": ["t", "r", "u", "e"]}`,
		`{"var": "floor"}`,
		`{"var": "plan"}`,
		`{"-": [{"var": "floor"}, {"var": "floor"}]}`,
		`{"filter": [{"var": "tags"}, {"==": [{"var": ""}, "b"]}]}`,
		`{"filter": [{"var": "tags"}, {"==": [{"var": ""}, "z"]}]}`,
		`{"some": [{"var": "orgs"}, {"==": [{"var": "plan"}, "enterprise"]}]}`,
		`{"all": [{"var": "tags"}, {"!=": [{"var": ""}, ""]}]}`,
		`{"merge": [{"var": "tags"}, ["x"]]}`,
		`{"if": [{"<": [{"var": "floor"}, 1]}, "", "yes"]}`,
		`{"substr": [{"var": "key"}, 0, 0]}`,
		`{"xxx": [{"var": "key"}, "abc"]}`,
		`{"/": [1, {"var": "floor"}]}`,
		`{"/": [0, {"var": "floor"}]}`,
		`{"merge": [[{"/": [1, {"var": "floor"}]}]]}`,
		`{"if": [true, {"nested": {"/": [1, {"var": "floor"}]}}, false]}`,
		`{"map": [{"var": "tags"}, {"/": [1, 0]}]}`,
		`{"*": [-1, 0]}`,
		`{"+": [0.5, 0.5]}`,
	}
	ctxs := []ffcontext.Context{
		ffcontext.NewEvaluationContext("abc"),
		ffcontext.NewEvaluationContextBuilder("def").
			AddCustom("dsp", "ttd").AddCustom("floor", 1.2).
			AddCustom("tags", []any{"a", "b"}).
			AddCustom("orgs", []any{map[string]any{"id": 1}}).
			AddCustom("plan", map[string]any{"cat": []any{"enter", "prise"}}).
			Build(),
		ffcontext.NewEvaluationContextBuilder("ghi").AddCustom("floor", 0).AddCustom("tags", []any{}).Build(),
	}
	for _, q := range queries {
		for _, ctx := range ctxs {
			t.Run(q+"/"+ctx.GetKey(), func(t *testing.T) {
				assert.Equal(t, applyReference(t, q, ctx), evaluateRule(q, JSONLogicQueryFormat, &queryContext{ctx: ctx}))
			})
		}
	}
}

func TestEvaluateRule_JSONLogicConcurrent(t *testing.T) {
	q := `{"and": [{"in": [{"var": "dsp"}, ["ttd", "dv360"]]}, {"some": [{"var": "tags"}, {"==": [{"var": ""}, "b"]}]}]}`
	var wg sync.WaitGroup
	for i := range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dsp := []string{"ttd", "xandr"}[i%2]
			ctx := ffcontext.NewEvaluationContextBuilder("k").AddCustom("dsp", dsp).AddCustom("tags", []any{"a", "b"}).Build()
			for range 200 {
				assert.Equal(t, dsp == "ttd", evaluateRule(q, JSONLogicQueryFormat, &queryContext{ctx: ctx}))
			}
		}()
	}
	wg.Wait()
}

func applyReference(t *testing.T, query string, ctx ffcontext.Context) bool {
	t.Helper()
	strCtx, err := json.Marshal(utils.ContextToMap(ctx))
	require.NoError(t, err)
	var result bytes.Buffer
	if jsonlogic.Apply(strings.NewReader(query), bytes.NewReader(strCtx), &result) != nil {
		return false
	}
	switch strings.TrimSpace(result.String()) {
	case "false", "null", `""`, "0", "[]":
		return false
	default:
		return true
	}
}
