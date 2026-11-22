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
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
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

type namedBool bool

type namedString string

func TestIsJSONLogicTruthy(t *testing.T) {
	f, tr, empty := false, true, ""
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{name: "infinity", value: math.Inf(1)},
		{name: "NaN", value: math.NaN()},
		{name: "infinity in an array", value: []any{true, math.Inf(-1)}},
		{name: "NaN in an object", value: map[string]any{"a": map[string]any{"b": math.NaN()}}},
		{name: "nil slice", value: []any(nil)},
		{name: "nil map", value: map[string]any(nil)},
		{name: "*bool false", value: &f},
		{name: "named bool false", value: namedBool(false)},
		{name: "*string empty", value: &empty},
		{name: "named string empty", value: namedString("")},
		{name: "raw false", value: json.RawMessage("false")},
		{name: "*bool true", value: &tr, want: true},
		{name: "named bool true", value: namedBool(true), want: true},
		{name: "named string", value: namedString("x"), want: true},
		{name: "raw 1", value: json.RawMessage("1"), want: true},
		{name: "non-empty slice", value: []any{"a"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isJSONLogicTruthy(tt.value))
		})
	}
}

func TestGetCompiledQuery_OnlyEvaluationStores(t *testing.T) {
	t.Cleanup(compiledQueryCache.Clear)
	compiledQueryCache.Clear()
	query := "{\n  \"==\": [{\"var\": \"key\"}, \"abc\"]\n}"
	r := Rule{Query: &query, VariationResult: testconvert.String("A")}

	assert.NoError(t, r.isQueryValid(false))
	assert.Equal(t, JSONLogicQueryFormat, r.GetQueryFormat())
	assert.Equal(t, `{"==": [{"var": "key"}, "abc"]}`, r.GetTrimmedQuery())
	assert.Zero(t, compiledQueryCacheLen(), "validation must not store queries")

	ctx := ffcontext.NewEvaluationContext("abc")
	_, err := r.Evaluate("abc", ctx, "flag", false)
	assert.NoError(t, err)
	assert.Equal(t, 1, compiledQueryCacheLen(), "a multi-line query is stored once")
}

func compiledQueryCacheLen() int {
	n := 0
	compiledQueryCache.Range(func(_, _ any) bool {
		n++
		return true
	})
	return n
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
				assert.Equal(t, applyReference(t, q, ctx), evaluateRule(getCompiledQuery(q), &queryContext{ctx: ctx}))
			})
		}
	}
}

func TestEvaluateRule_JSONLogicCustomOperatorMatchesApply(t *testing.T) {
	f, tr, empty := false, true, ""
	results := map[string]any{
		"nil_slice":    []any(nil),
		"nil_map":      map[string]any(nil),
		"ptr_false":    &f,
		"named_false":  namedBool(false),
		"ptr_empty":    &empty,
		"named_empty":  namedString(""),
		"raw_false":    json.RawMessage("false"),
		"ptr_true":     &tr,
		"named_true":   namedBool(true),
		"named_string": namedString("x"),
	}
	ctx := ffcontext.NewEvaluationContext("abc")
	for name, result := range results {
		op := "goff_test_" + name
		jsonlogic.AddOperator(op, func(_, _ any) any { return result })
		q := `{"` + op + `": []}`
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, applyReference(t, q, ctx), evaluateRule(getCompiledQuery(q), &queryContext{ctx: ctx}))
		})
	}
}

func FuzzEvaluateRule_JSONLogicMatchesApply(f *testing.F) {
	ctx := `{"dsp": "ttd", "floor": 1.2, "tags": ["a", "b"], "orgs": [{"plan": "enterprise"}]}`
	for _, q := range []string{
		`{"==": [{"var": "key"}, "abc"]}`,
		`{"and": [{"in": [{"var": "dsp"}, ["ttd", "dv360"]]}, {">": [{"var": "floor"}, 0.5]}]}`,
		`{"some": [{"var": "orgs"}, {"==": [{"var": "plan"}, "enterprise"]}]}`,
		`{"/": [1, {"var": "floor"}]}`,
		`{"merge": [{"var": "tags"}, ["x"]]}`,
	} {
		f.Add(q, ctx)
	}
	f.Fuzz(func(t *testing.T, query, ctxJSON string) {
		t.Cleanup(compiledQueryCache.Clear)
		var attributes map[string]any
		if json.Unmarshal([]byte(ctxJSON), &attributes) != nil {
			return
		}
		compiled := getCompiledQuery(query)
		if compiled.format != JSONLogicQueryFormat {
			return
		}
		evalCtx := utils.ConvertEvaluationCtxFromRequest("key", attributes)
		assert.Equal(t, applyReference(t, compiled.trimmed, evalCtx), evaluateRule(compiled, &queryContext{ctx: evalCtx}))
	})
}

func BenchmarkEvaluateRule_JSONLogic(b *testing.B) {
	query := `{"and": [{"in": [{"var": "dsp"}, ["ttd", "dv360"]]}, {"==": [{"var": "inventory"}, "web"]}]}`
	ctx := ffcontext.NewEvaluationContextBuilder("request-1").AddCustom("dsp", "ttd").AddCustom("inventory", "web").Build()
	b.Run("jsonlogic.Apply", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			strCtx, _ := json.Marshal(utils.ContextToMap(ctx))
			var result bytes.Buffer
			_ = jsonlogic.Apply(strings.NewReader(query), bytes.NewReader(strCtx), &result)
		}
	})
	b.Run("cached", func(b *testing.B) {
		compiled := getCompiledQuery(query)
		b.ReportAllocs()
		for b.Loop() {
			evaluateRule(compiled, &queryContext{ctx: ctx})
		}
	})
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
				assert.Equal(t, dsp == "ttd", evaluateRule(getCompiledQuery(q), &queryContext{ctx: ctx}))
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
