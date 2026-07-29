package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/wasm/helpers"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

func Test_scanJSONDepth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "flat object", input: `{"a": 1, "b": "c"}`, want: 1},
		{name: "nested objects and arrays", input: `{"a": [{"b": {"c": 1}}]}`, want: 4},
		{name: "brackets inside strings are ignored", input: `{"a": "{[{[{["}`, want: 1},
		{name: "escaped quote inside string", input: `{"a": "x\"{[y"}`, want: 1},
		{name: "empty input", input: ``, want: 0},
		{name: "parens do not count as JSON nesting", input: `{"a": "b"} (((`, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, scanJSONDepth(tt.input))
		})
	}
}

func Test_scanJSONMaxArrayLength(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "empty input", input: ``, want: 0},
		{name: "flat object", input: `{"a": 1}`, want: 0},
		{name: "comparison operand array", input: `{"==":[{"var":"a"},1]}`, want: 2},
		{name: "in list", input: `{"in":[{"var":"age"},[1,2,3]]}`, want: 3},
		{name: "or operand array", input: `{"or":[{"==":[1,1]},{"==":[2,2]}]}`, want: 2},
		{name: "largest array wins", input: `{"or":[{"in":[{"var":"a"},[1,2]]},{"in":[{"var":"b"},[1,2,3,4]]}]}`, want: 4},
		{name: "brackets inside strings ignored", input: `{"a": "[1,2,3,4,5]"}`, want: 0},
		{name: "empty array over-approximates to one item", input: `{"in":[{"var":"a"},[]]}`, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, scanJSONMaxArrayLength(tt.input))
		})
	}
}

func Test_scanNikunjyQuery_depth(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "no nesting", query: `targetingKey eq "user-1"`, want: 0},
		{name: "single parens", query: `(a eq "b") and (c eq "d")`, want: 1},
		{name: "deeply nested parens", query: strings.Repeat("(", 10) + `a eq "b"` + strings.Repeat(")", 10), want: 10},
		{name: "parens inside string literal ignored", query: `a eq "((((("`, want: 0},
		{name: "jsonlogic style braces", query: `{"==": [{"var": "a"}, "b"]}`, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, scanNikunjyQuery(tt.query).maxDepth)
		})
	}
}

func Test_isJSONLogicQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{name: "nikunjy expression", query: `targetingKey eq "a"`, want: false},
		{name: "empty query", query: ``, want: false},
		{name: "jsonlogic object", query: `{"==":[{"var":"a"},1]}`, want: true},
		{name: "jsonlogic with leading whitespace", query: "\n  {\"==\":[1,1]}", want: true},
		{name: "jsonlogic array", query: `[{"var":"a"}]`, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isJSONLogicQuery(tt.query))
		})
	}
}

func Test_scanNikunjyQuery_listItems(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "empty query", query: ``, want: 0},
		{name: "no list", query: `targetingKey eq "user-1"`, want: 0},
		{name: "int list", query: `age in [1,2,3]`, want: 3},
		{name: "spaces after commas", query: `age in [1, 2, 3]`, want: 3},
		{name: "commas inside string literals ignored", query: `a in ["x,y","z"]`, want: 2},
		{name: "escaped quote inside item", query: `a in ["x\",y","z"]`, want: 2},
		{name: "unclosed list still counted", query: `age in [1,2`, want: 2},
		{name: "largest of several lists", query: `(a in [1,2]) or (b in [1,2,3,4])`, want: 4},
		{name: "comma outside any list ignored", query: `a eq "b", c`, want: 0},
		{name: "brackets inside strings ignored", query: `a eq "[1,2,3]"`, want: 0},
		{name: "empty brackets over-approximate to one item", query: `a in []`, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, scanNikunjyQuery(tt.query).maxListItems)
		})
	}
}

func intListQuery(n int) string {
	items := make([]string, n)
	for i := range items {
		items[i] = fmt.Sprintf("%d", i+1)
	}
	return "age in [" + strings.Join(items, ",") + "]"
}

func orChainQuery(n int) string {
	conditions := make([]string, n)
	for i := range conditions {
		conditions[i] = fmt.Sprintf("age eq %d", i+1)
	}
	return strings.Join(conditions, " or ")
}

// splitOperatorChainQuery builds an or-chain whose operators are cut in half by
// a line break. StrTrim rejoins the lines with the empty string, so the
// evaluator parses a real n-condition chain out of it.
func splitOperatorChainQuery(n int) string {
	conditions := make([]string, n)
	for i := range conditions {
		conditions[i] = fmt.Sprintf("age eq %d", i+1)
	}
	return strings.Join(conditions, " o\nr ")
}

// attrPathQuery builds a query whose left-hand side is a single attribute path
// of n '.'-separated segments.
func attrPathQuery(n int) string {
	segments := make([]string, n)
	for i := range segments {
		segments[i] = fmt.Sprintf("s%d", i)
	}
	return strings.Join(segments, ".") + ` eq "x"`
}

func Test_scanNikunjyQuery_conditionCount(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "empty query", query: ``, want: 1},
		{name: "single condition", query: `targetingKey eq "user-1"`, want: 1},
		{name: "two anded conditions", query: `a eq "b" and c eq "d"`, want: 2},
		{name: "five or-joined conditions", query: orChainQuery(5), want: 5},
		{name: "mixed and/or chain", query: `a eq 1 and b eq 2 or c eq 3`, want: 3},
		{name: "operators inside string literals ignored", query: `a eq "b and c or d"`, want: 1},
		{name: "escaped quote before operator", query: `a eq "x\" and y" or b eq 1`, want: 2},
		{name: "uppercase is not an operator", query: `a eq 1 AND b eq 2`, want: 1},
		{name: "words containing operators ignored", query: `land eq 1 or orange sw "android"`, want: 2},
		{name: "attribute path with operator segment", query: `a.and.b eq 1`, want: 1},
		{name: "trailing operator of a malformed query still counted", query: `a eq 1 and`, want: 2},
		{name: "parenthesized conditions", query: `(a eq 1) or (b eq 2)`, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, scanNikunjyQuery(tt.query).conditionCount)
		})
	}
}

func Test_scanNikunjyQuery_pathSegments(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "empty query", query: ``, want: 0},
		{name: "single segment", query: `targetingKey eq "user-1"`, want: 1},
		{name: "dotted attribute path", query: `user.profile.country eq "FR"`, want: 3},
		{name: "longest path of several wins", query: `a.b eq 1 and c.d.e.f eq 2`, want: 4},
		{name: "dots inside string literals ignored", query: `a eq "x.y.z.w.v"`, want: 1},
		{name: "float literal counts as two segments", query: `score gt 1.5`, want: 2},
		{name: "trailing dot still counted", query: `a.b. eq 1`, want: 3},
		{name: "generated path", query: attrPathQuery(64), want: 64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, scanNikunjyQuery(tt.query).maxPathSegments)
		})
	}
}

// The guard must scan the same string the evaluator parses. StrTrim splits on
// '\n' and rejoins with the empty string, which glues tokens back together
// across line breaks — scanning the raw query would miss the operators.
func Test_firstQueryViolation_scansTheTrimmedQuery(t *testing.T) {
	raw := splitOperatorChainQuery(maxQueryConditions + 1)
	rule := flag.Rule{Query: &raw}

	require.Equal(t, 1, scanNikunjyQuery(raw).conditionCount,
		"raw query hides the operators, so this test would be vacuous otherwise")
	require.Equal(t, maxQueryConditions+1, scanNikunjyQuery(rule.GetTrimmedQuery()).conditionCount)

	detail, over := firstQueryViolation(&flag.InternalFlag{Rules: &[]flag.Rule{rule}})
	assert.True(t, over)
	assert.Contains(t, detail, "maximum condition count")
}

func Test_firstQueryViolation(t *testing.T) {
	parenQuery := func(depth int) *string {
		q := strings.Repeat("(", depth) + `a eq "b"` + strings.Repeat(")", depth)
		return &q
	}
	jsonlogicQuery := func(depth int) *string {
		q := strings.Repeat(`{"and":[`, depth) + `{"==":[1,1]}` + strings.Repeat(`]}`, depth)
		return &q
	}
	nikunjyList := func(n int) *string {
		q := intListQuery(n)
		return &q
	}
	jsonlogicList := func(n int) *string {
		items := make([]string, n)
		for i := range items {
			items[i] = fmt.Sprintf("%d", i+1)
		}
		q := `{"in":[{"var":"age"},[` + strings.Join(items, ",") + `]]}`
		return &q
	}
	chain := func(n int) *string {
		q := orChainQuery(n)
		return &q
	}
	attrPath := func(n int) *string {
		q := attrPathQuery(n)
		return &q
	}
	jsonlogicVarPath := func(n int) *string {
		segments := make([]string, n)
		for i := range segments {
			segments[i] = fmt.Sprintf("s%d", i)
		}
		q := `{"==":[{"var":"` + strings.Join(segments, ".") + `"},"x"]}`
		return &q
	}
	jsonlogicChain := func(n int) *string {
		clauses := make([]string, n)
		for i := range clauses {
			clauses[i] = fmt.Sprintf(`{"==":[{"var":"age"},%d]}`, i+1)
		}
		q := `{"or":[` + strings.Join(clauses, ",") + `]}`
		return &q
	}

	tests := []struct {
		name             string
		flag             *flag.InternalFlag
		wantOver         bool
		wantDetailsMatch string
	}{
		{name: "nil flag", flag: nil, wantOver: false},
		{name: "no rules", flag: &flag.InternalFlag{}, wantOver: false},
		{
			name: "nikunjy queries under the limit",
			flag: &flag.InternalFlag{
				Rules:       &[]flag.Rule{{Query: parenQuery(3)}, {Query: parenQuery(7)}},
				DefaultRule: &flag.Rule{},
			},
			wantOver: false,
		},
		{
			name: "nikunjy query over the nesting limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: parenQuery(maxQueryNestingDepth + 1)}},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum nesting depth",
		},
		{
			name: "jsonlogic gets the larger nesting budget",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: jsonlogicQuery(50)}},
			},
			wantOver: false,
		},
		{
			name: "jsonlogic over its own nesting budget",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: jsonlogicQuery(150)}},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum nesting depth",
		},
		{
			name: "scheduled step carries the offending nesting query",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: parenQuery(2)}},
				Scheduled: &[]flag.ScheduledStep{
					{InternalFlag: flag.InternalFlag{
						Rules: &[]flag.Rule{{Query: parenQuery(maxQueryNestingDepth + 5)}},
					}},
				},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum nesting depth",
		},
		{
			name: "list under the limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: nikunjyList(maxQueryListItems)}},
			},
			wantOver: false,
		},
		{
			name: "list over the limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: nikunjyList(maxQueryListItems + 1)}},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum item count",
		},
		{
			name: "offending list on the default rule",
			flag: &flag.InternalFlag{
				DefaultRule: &flag.Rule{Query: nikunjyList(maxQueryListItems + 1)},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum item count",
		},
		{
			name: "jsonlogic list over the item limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: jsonlogicList(maxQueryListItems + 1)}},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum item count",
		},
		{
			name: "jsonlogic list under the item limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: jsonlogicList(maxQueryListItems)}},
			},
			wantOver: false,
		},
		{
			name: "scheduled step carries the offending list",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: nikunjyList(3)}},
				Scheduled: &[]flag.ScheduledStep{
					{InternalFlag: flag.InternalFlag{
						Rules: &[]flag.Rule{{Query: nikunjyList(maxQueryListItems + 1)}},
					}},
				},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum item count",
		},
		{
			name: "chain at the condition limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: chain(maxQueryConditions)}},
			},
			wantOver: false,
		},
		{
			name: "chain over the condition limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: chain(maxQueryConditions + 1)}},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum condition count",
		},
		{
			name: "offending chain on the default rule",
			flag: &flag.InternalFlag{
				DefaultRule: &flag.Rule{Query: chain(maxQueryConditions + 1)},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum condition count",
		},
		{
			name: "jsonlogic or-chain over the condition limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: jsonlogicChain(maxQueryConditions + 1)}},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum item count",
		},
		{
			name: "jsonlogic or-chain under the condition limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: jsonlogicChain(maxQueryConditions)}},
			},
			wantOver: false,
		},
		{
			name: "scheduled step carries the offending chain",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: chain(3)}},
				Scheduled: &[]flag.ScheduledStep{
					{InternalFlag: flag.InternalFlag{
						Rules: &[]flag.Rule{{Query: chain(maxQueryConditions + 1)}},
					}},
				},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum condition count",
		},
		{
			name: "realistic attribute path",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: attrPath(6)}},
			},
			wantOver: false,
		},
		{
			name: "attribute path at the segment limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: attrPath(maxQueryAttrPathSegments)}},
			},
			wantOver: false,
		},
		{
			name: "attribute path over the segment limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: attrPath(maxQueryAttrPathSegments + 1)}},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum segment count",
		},
		{
			// The path that traps the module carries no bracket, comma or
			// logical operator, so every other guard reports it as trivial.
			name: "attribute path at the measured trap threshold",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: attrPath(602)}},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum segment count",
		},
		{
			name: "offending attribute path on the default rule",
			flag: &flag.InternalFlag{
				DefaultRule: &flag.Rule{Query: attrPath(maxQueryAttrPathSegments + 1)},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum segment count",
		},
		{
			name: "jsonlogic var paths are exempt from the segment limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: jsonlogicVarPath(maxQueryAttrPathSegments * 2)}},
			},
			wantOver: false,
		},
		{
			name: "scheduled step carries the offending attribute path",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: attrPath(3)}},
				Scheduled: &[]flag.ScheduledStep{
					{InternalFlag: flag.InternalFlag{
						Rules: &[]flag.Rule{{Query: attrPath(maxQueryAttrPathSegments + 1)}},
					}},
				},
			},
			wantOver:         true,
			wantDetailsMatch: "maximum segment count",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detail, over := firstQueryViolation(tt.flag)
			assert.Equal(t, tt.wantOver, over)
			if tt.wantDetailsMatch != "" {
				assert.Contains(t, detail, tt.wantDetailsMatch)
			}
		})
	}
}

func Test_localEvaluation_guards(t *testing.T) {
	deepJSON := func(depth int) string {
		ctx := strings.Repeat(`{"a":`, depth) + `1` + strings.Repeat("}", depth)
		return fmt.Sprintf(
			`{"flagKey":"f","flag":{"variations":{"on":true},"defaultRule":{"variation":"on"}},"evalContext":%s,"flagContext":{"defaultSdkValue":false}}`,
			ctx)
	}
	deepQuery := func(depth int) string {
		q := strings.Repeat("(", depth) + `targetingKey eq \"a\"` + strings.Repeat(")", depth)
		return fmt.Sprintf(
			`{"flagKey":"f","flag":{"variations":{"on":true,"off":false},"targeting":[{"query":"%s","variation":"on"}],"defaultRule":{"variation":"off"}},"evalContext":{"targetingKey":"u"},"flagContext":{"defaultSdkValue":false}}`,
			q)
	}
	queryPayload := func(marshaledQuery string) string {
		return fmt.Sprintf(
			`{"flagKey":"f","flag":{"variations":{"on":true,"off":false},"targeting":[{"query":%s,"variation":"on"}],"defaultRule":{"variation":"off"}},"evalContext":{"targetingKey":"u","age":1},"flagContext":{"defaultSdkValue":false}}`,
			marshaledQuery)
	}
	listQuery := func(n int) string {
		q, _ := json.Marshal(intListQuery(n))
		return queryPayload(string(q))
	}
	jsonlogicListQuery := func(n int) string {
		items := make([]string, n)
		for i := range items {
			items[i] = fmt.Sprintf("%d", i+1)
		}
		q, _ := json.Marshal(`{"in":[{"var":"age"},[` + strings.Join(items, ",") + `]]}`)
		return queryPayload(string(q))
	}
	chainQuery := func(n int) string {
		q, _ := json.Marshal(orChainQuery(n))
		return queryPayload(string(q))
	}
	attrPathQ := func(n int) string {
		q, _ := json.Marshal(attrPathQuery(n))
		return queryPayload(string(q))
	}
	splitOperatorQuery := func(n int) string {
		q, _ := json.Marshal(splitOperatorChainQuery(n))
		return queryPayload(string(q))
	}
	splitListQuery := func(count, chunk int) string {
		parts := make([]string, 0, (count+chunk-1)/chunk)
		for start := 0; start < count; start += chunk {
			end := start + chunk
			if end > count {
				end = count
			}
			items := make([]string, 0, end-start)
			for i := start; i < end; i++ {
				items = append(items, fmt.Sprintf("%d", i+1))
			}
			parts = append(parts, "(age in ["+strings.Join(items, ",")+"])")
		}
		q, _ := json.Marshal(strings.Join(parts, " or "))
		return queryPayload(string(q))
	}
	tests := []struct {
		name             string
		input            string
		wantErrorCode    string
		wantDetailsMatch string
	}{
		{
			name:             "input JSON too deep returns PARSE_ERROR",
			input:            deepJSON(200),
			wantErrorCode:    "PARSE_ERROR",
			wantDetailsMatch: "maximum nesting depth",
		},
		{
			name:             "targeting query too deep returns PARSE_ERROR",
			input:            deepQuery(100),
			wantErrorCode:    "PARSE_ERROR",
			wantDetailsMatch: "targeting query exceeds",
		},
		{
			name:          "input at reasonable depth evaluates normally",
			input:         deepJSON(50),
			wantErrorCode: "",
		},
		{
			name:          "query at reasonable depth evaluates normally",
			input:         deepQuery(10),
			wantErrorCode: "",
		},
		{
			name:             "in list over the item limit returns PARSE_ERROR",
			input:            listQuery(maxQueryListItems + 1),
			wantErrorCode:    "PARSE_ERROR",
			wantDetailsMatch: "maximum item count",
		},
		{
			// The production trigger of issue #5651: a flat allow-list of
			// ~200 integers must evaluate, not be rejected by the guard.
			name:          "reporter-shaped 200-int list evaluates normally",
			input:         listQuery(200),
			wantErrorCode: "",
		},
		{
			name:             "jsonlogic list over the item limit returns PARSE_ERROR",
			input:            jsonlogicListQuery(maxQueryListItems + 1),
			wantErrorCode:    "PARSE_ERROR",
			wantDetailsMatch: "maximum item count",
		},
		{
			name:          "jsonlogic list under the item limit evaluates normally",
			input:         jsonlogicListQuery(200),
			wantErrorCode: "",
		},
		{
			name:             "and/or chain over the condition limit returns PARSE_ERROR",
			input:            chainQuery(maxQueryConditions + 1),
			wantErrorCode:    "PARSE_ERROR",
			wantDetailsMatch: "maximum condition count",
		},
		{
			name:          "moderate or-chain evaluates normally",
			input:         chainQuery(200),
			wantErrorCode: "",
		},
		{
			// The recommended workaround for big allow-lists: or-joined `in`
			// chunks. It must stay under both the list and condition caps.
			name:          "split-list workaround shape evaluates normally",
			input:         splitListQuery(20_000, 50),
			wantErrorCode: "",
		},
		{
			name:             "attribute path over the segment limit returns PARSE_ERROR",
			input:            attrPathQ(maxQueryAttrPathSegments + 1),
			wantErrorCode:    "PARSE_ERROR",
			wantDetailsMatch: "maximum segment count",
		},
		{
			// 602 segments trap the built module; every other guard reads this
			// query as trivial (no bracket, comma or logical operator).
			name:             "attribute path at the measured trap threshold returns PARSE_ERROR",
			input:            attrPathQ(602),
			wantErrorCode:    "PARSE_ERROR",
			wantDetailsMatch: "maximum segment count",
		},
		{
			name:          "realistic dotted attribute path evaluates normally",
			input:         attrPathQ(6),
			wantErrorCode: "",
		},
		{
			// StrTrim glues "o\nr" back into "or", so this parses as a real
			// chain. The guard has to see it the same way.
			name:             "line-split operators are counted after trimming",
			input:            splitOperatorQuery(maxQueryConditions + 1),
			wantErrorCode:    "PARSE_ERROR",
			wantDetailsMatch: "maximum condition count",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := localEvaluation(tt.input)
			var result map[string]any
			assert.NoError(t, json.Unmarshal([]byte(got), &result))
			errorCode, _ := result["errorCode"].(string)
			assert.Equal(t, tt.wantErrorCode, errorCode)
			if tt.wantDetailsMatch != "" {
				details, _ := result["errorDetails"].(string)
				assert.Contains(t, details, tt.wantDetailsMatch)
			}
		})
	}
}

func Test_evaluatePanicOutput_isStructuredError(t *testing.T) {
	assert.NotZero(t, evaluatePanicOutput)

	outputPtr := uint32(evaluatePanicOutput >> 32)
	outputLen := uint32(evaluatePanicOutput & 0xFFFFFFFF)
	assert.NotZero(t, outputPtr)
	assert.Equal(t, uint32(len(evaluatePanicBuffer)), outputLen)

	var result map[string]any
	assert.NoError(t, json.Unmarshal(evaluatePanicBuffer, &result))
	assert.Equal(t, "GENERAL", result["errorCode"])
	assert.Contains(t, result["errorDetails"], "recovered from panic during evaluation")

	// The panic buffer must be rooted by its own variable, not by
	// helpers.lastOutput, which every successful evaluation overwrites.
	// Re-packing it after an intervening output must yield the same pointer.
	helpers.WasmCopyBufferToMemory([]byte(`{"value":true}`))
	assert.Equal(t, evaluatePanicOutput, helpers.WasmCopyBufferToMemory(evaluatePanicBuffer))
}

func Test_safeEvaluation_recovers_from_panic(t *testing.T) {
	original := evaluationFn
	defer func() { evaluationFn = original }()
	evaluationFn = func(string) string { panic("boom during evaluation") }

	got := safeEvaluation((*uint32)(nil), 0)

	var result map[string]any
	assert.NoError(t, json.Unmarshal([]byte(got), &result))
	assert.Equal(t, "GENERAL", result["errorCode"])
	assert.Contains(t, result["errorDetails"], "boom during evaluation")
}
