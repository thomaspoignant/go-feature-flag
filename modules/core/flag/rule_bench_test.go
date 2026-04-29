package flag_test

import (
	"testing"

	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

// BenchmarkRuleEvaluate_NikunjyComplex measures the per-call cost of evaluating a
// nikunjy targeting query. Before the parser cache, every call re-parses the query
// via ANTLR; after the cache, parsing happens once per distinct query string.
func BenchmarkRuleEvaluate_NikunjyComplex(b *testing.B) {
	rule := flag.Rule{
		Name:            testconvert.String("complex"),
		VariationResult: testconvert.String("on"),
		Query: testconvert.String(
			`language eq "ar" and isNewUser eq true and clubsTimeSpent gt 600 ` +
				`and clubsTimeSpent le 3600 and concurrencyLocked eq true and segment eq 0`,
		),
	}
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").
		AddCustom("language", "ar").
		AddCustom("isNewUser", true).
		AddCustom("clubsTimeSpent", 1500).
		AddCustom("concurrencyLocked", true).
		AddCustom("segment", 0).
		Build()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rule.Evaluate("user-1", ctx, "test-flag", false)
	}
}

// BenchmarkRuleEvaluate_NikunjySimple is a single-predicate variant for comparison.
func BenchmarkRuleEvaluate_NikunjySimple(b *testing.B) {
	rule := flag.Rule{
		Name:            testconvert.String("simple"),
		VariationResult: testconvert.String("on"),
		Query:           testconvert.String(`language eq "ar"`),
	}
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").
		AddCustom("language", "ar").
		Build()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rule.Evaluate("user-1", ctx, "test-flag", false)
	}
}
