package flag

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestQueryContext_MapOnce(t *testing.T) {
	tests := []struct {
		name string
		ctx  ffcontext.Context
		want map[string]any
	}{
		{
			name: "context with custom attributes",
			ctx:  ffcontext.NewEvaluationContextBuilder("user-1").AddCustom("plan", "pro").Build(),
			want: map[string]any{"plan": "pro", "anonymous": false, "key": "user-1", "targetingKey": "user-1"},
		},
		{
			name: "context without custom attributes",
			ctx:  ffcontext.NewEvaluationContext("user-1"),
			want: map[string]any{"anonymous": false, "key": "user-1", "targetingKey": "user-1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryCtx := &queryContext{ctx: tt.ctx}

			first := queryCtx.mapOnce()
			second := queryCtx.mapOnce()

			assert.Equal(t, tt.want, first)
			assert.Equal(t, reflect.ValueOf(first).Pointer(), reflect.ValueOf(second).Pointer(), "the map must be built once")
		})
	}
}

func TestRule_EvaluateBuildsTheMapOnlyForAQuery(t *testing.T) {
	tests := []struct {
		name       string
		rule       Rule
		isDefault  bool
		wantMapped bool
	}{
		{
			name:       "rule with a query",
			rule:       Rule{Query: testconvert.String(`key eq "user-1"`), VariationResult: testconvert.String("A")},
			wantMapped: true,
		},
		{
			name: "rule without a query",
			rule: Rule{VariationResult: testconvert.String("A")},
		},
		{
			name: "percentage rollout without a query",
			rule: Rule{Percentages: &map[string]float64{"A": 50, "B": 50}},
		},
		{
			name:      "default rule with a query",
			rule:      Rule{Query: testconvert.String(`key eq "user-1"`), VariationResult: testconvert.String("A")},
			isDefault: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryCtx := &queryContext{ctx: ffcontext.NewEvaluationContext("user-1")}

			_, err := tt.rule.evaluate("user-1", queryCtx, "my-flag", tt.isDefault)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantMapped, queryCtx.mapped != nil)
		})
	}
}

func TestInternalFlag_RulesOfBothQueryFormatsShareTheQueryContext(t *testing.T) {
	f := InternalFlag{
		Variations: &map[string]*any{
			"nikunjy":   testconvert.Interface("nikunjy"),
			"jsonlogic": testconvert.Interface("jsonlogic"),
			"default":   testconvert.Interface("default"),
		},
		Rules: &[]Rule{
			{Query: testconvert.String(`plan eq "free"`), VariationResult: testconvert.String("nikunjy")},
			{Query: testconvert.String(`{"==": [{"var": "plan"}, "pro"]}`), VariationResult: testconvert.String("jsonlogic")},
		},
		DefaultRule: &Rule{VariationResult: testconvert.String("default")},
	}
	tests := []struct {
		name string
		plan string
		want string
	}{
		{name: "nikunjy rule matches", plan: "free", want: "nikunjy"},
		{name: "nikunjy rule misses and jsonlogic rule matches", plan: "pro", want: "jsonlogic"},
		{name: "no rule matches", plan: "enterprise", want: "default"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ffcontext.NewEvaluationContextBuilder("user-1").AddCustom("plan", tt.plan).Build()

			value, details := f.Value("mixed-formats", ctx, Context{DefaultSdkValue: "sdk-default"})

			assert.Equal(t, tt.want, value)
			assert.Equal(t, tt.want, details.Variant)
		})
	}
}
