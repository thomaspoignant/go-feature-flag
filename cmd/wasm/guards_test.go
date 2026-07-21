package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

func Test_jsonNestingDepth(t *testing.T) {
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
			assert.Equal(t, tt.want, jsonNestingDepth(tt.input))
		})
	}
}

func Test_queryNestingDepth(t *testing.T) {
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
			assert.Equal(t, tt.want, queryNestingDepth(tt.query))
		})
	}
}

func Test_queryNestingLimit(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "nikunjy expression", query: `targetingKey eq "a"`, want: maxQueryNestingDepth},
		{name: "empty query", query: ``, want: maxQueryNestingDepth},
		{name: "jsonlogic object", query: `{"==":[{"var":"a"},1]}`, want: maxJSONLogicQueryNestingDepth},
		{name: "jsonlogic with leading whitespace", query: "\n  {\"==\":[1,1]}", want: maxJSONLogicQueryNestingDepth},
		{name: "jsonlogic array", query: `[{"var":"a"}]`, want: maxJSONLogicQueryNestingDepth},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, queryNestingLimit(tt.query))
		})
	}
}

func Test_maxListItemCount(t *testing.T) {
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
			assert.Equal(t, tt.want, maxListItemCount(tt.query))
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

func Test_conditionCount(t *testing.T) {
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
			assert.Equal(t, tt.want, conditionCount(tt.query))
		})
	}
}

func Test_firstQueryOverConditionCount(t *testing.T) {
	chain := func(n int) *string {
		q := orChainQuery(n)
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
		name     string
		flag     *flag.InternalFlag
		wantOver bool
	}{
		{name: "nil flag", flag: nil, wantOver: false},
		{name: "no rules", flag: &flag.InternalFlag{}, wantOver: false},
		{
			name: "chain at the limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: chain(maxQueryConditions)}},
			},
			wantOver: false,
		},
		{
			name: "chain over the limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: chain(maxQueryConditions + 1)}},
			},
			wantOver: true,
		},
		{
			name: "offending chain on the default rule",
			flag: &flag.InternalFlag{
				DefaultRule: &flag.Rule{Query: chain(maxQueryConditions + 1)},
			},
			wantOver: true,
		},
		{
			name: "jsonlogic operand lists are exempt",
			flag: &flag.InternalFlag{
				// A flat JSONLogic or-list is decoded iteratively; its length
				// does not become recursion depth.
				Rules: &[]flag.Rule{{Query: jsonlogicChain(maxQueryConditions * 2)}},
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
			wantOver: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditions, limit, over := firstQueryOverConditionCount(tt.flag)
			assert.Equal(t, tt.wantOver, over)
			if tt.wantOver {
				assert.Equal(t, maxQueryConditions, limit)
				assert.Greater(t, conditions, limit)
			}
		})
	}
}

func Test_firstQueryOverBreadth(t *testing.T) {
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
	tests := []struct {
		name     string
		flag     *flag.InternalFlag
		wantOver bool
	}{
		{name: "nil flag", flag: nil, wantOver: false},
		{name: "no rules", flag: &flag.InternalFlag{}, wantOver: false},
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
			wantOver: true,
		},
		{
			name: "offending list on the default rule",
			flag: &flag.InternalFlag{
				DefaultRule: &flag.Rule{Query: nikunjyList(maxQueryListItems + 1)},
			},
			wantOver: true,
		},
		{
			name: "jsonlogic arrays are exempt",
			flag: &flag.InternalFlag{
				// encoding/json decodes arrays iteratively: length does not
				// become recursion depth, so no breadth limit applies.
				Rules: &[]flag.Rule{{Query: jsonlogicList(maxQueryListItems * 2)}},
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
			wantOver: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, limit, over := firstQueryOverBreadth(tt.flag)
			assert.Equal(t, tt.wantOver, over)
			if tt.wantOver {
				assert.Equal(t, maxQueryListItems, limit)
				assert.Greater(t, items, limit)
			}
		})
	}
}

func Test_firstQueryOverLimit(t *testing.T) {
	parenQuery := func(depth int) *string {
		q := strings.Repeat("(", depth) + `a eq "b"` + strings.Repeat(")", depth)
		return &q
	}
	jsonlogicQuery := func(depth int) *string {
		q := strings.Repeat(`{"and":[`, depth) + `{"==":[1,1]}` + strings.Repeat(`]}`, depth)
		return &q
	}
	tests := []struct {
		name      string
		flag      *flag.InternalFlag
		wantOver  bool
		wantLimit int
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
			name: "nikunjy query over the limit",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: parenQuery(maxQueryNestingDepth + 1)}},
			},
			wantOver:  true,
			wantLimit: maxQueryNestingDepth,
		},
		{
			name: "jsonlogic gets the larger budget",
			flag: &flag.InternalFlag{
				// 50 nested and-operators = bracket depth 100: over the
				// nikunjy limit but fine for JSONLogic.
				Rules: &[]flag.Rule{{Query: jsonlogicQuery(50)}},
			},
			wantOver: false,
		},
		{
			name: "jsonlogic over its own budget",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: jsonlogicQuery(150)}},
			},
			wantOver:  true,
			wantLimit: maxJSONLogicQueryNestingDepth,
		},
		{
			name: "scheduled step carries the offending query",
			flag: &flag.InternalFlag{
				Rules: &[]flag.Rule{{Query: parenQuery(2)}},
				Scheduled: &[]flag.ScheduledStep{
					{InternalFlag: flag.InternalFlag{
						Rules: &[]flag.Rule{{Query: parenQuery(maxQueryNestingDepth + 5)}},
					}},
				},
			},
			wantOver:  true,
			wantLimit: maxQueryNestingDepth,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depth, limit, over := firstQueryOverLimit(tt.flag)
			assert.Equal(t, tt.wantOver, over)
			if tt.wantOver {
				assert.Equal(t, tt.wantLimit, limit)
				assert.Greater(t, depth, limit)
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
			name:          "large jsonlogic array is exempt from the item limit",
			input:         jsonlogicListQuery(maxQueryListItems * 2),
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
