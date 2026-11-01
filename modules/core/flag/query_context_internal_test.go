package flag

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestQueryContext_ConvertsTheContextOnce(t *testing.T) {
	queryCtx := &queryContext{ctx: ffcontext.NewEvaluationContextBuilder("user-1").AddCustom("plan", "pro").Build()}

	first := queryCtx.asMap()
	second := queryCtx.asMap()

	assert.Equal(t, "pro", first["plan"])
	assert.Equal(t, "user-1", first["key"])
	assert.Equal(t, reflect.ValueOf(first).Pointer(), reflect.ValueOf(second).Pointer(), "the map must be built once")
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
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").AddCustom("plan", "pro").Build()

	value, details := f.Value("mixed-formats", ctx, Context{DefaultSdkValue: "sdk-default"})

	assert.Equal(t, "jsonlogic", value)
	assert.Equal(t, "jsonlogic", details.Variant)
}
