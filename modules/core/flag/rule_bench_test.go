package flag_test

import (
	"fmt"
	"runtime"
	"strings"
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

// BenchmarkRuleEvaluate_NikunjyColdAfterGC evaluates many distinct queries, each
// once per round, with garbage collections between rounds. It models a large flag
// configuration where most flags are evaluated less often than the GC runs.
func BenchmarkRuleEvaluate_NikunjyColdAfterGC(b *testing.B) {
	const queries = 1000
	rules := make([]flag.Rule, queries)
	for i := range rules {
		rules[i] = flag.Rule{
			Name:            testconvert.String(fmt.Sprintf("cold-%d", i)),
			VariationResult: testconvert.String("on"),
			Query:           testconvert.String(fmt.Sprintf(`group eq "group-%d"`, i)),
		}
		_, _ = rules[i].Evaluate("user-1", ffcontext.NewEvaluationContext("user-1"), "warm-up", false)
	}
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").
		AddCustom("group", "group-0").
		Build()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%queries == 0 {
			b.StopTimer()
			runtime.GC()
			runtime.GC()
			b.StartTimer()
		}
		_, _ = rules[i%queries].Evaluate("user-1", ctx, "test-flag", false)
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

// BenchmarkRuleEvaluate_NikunjyLargeIn evaluates a rule that targets 400 UUID-length
// user IDs, the shape of a flag that lists individual targets, with a context
// that is not in the list.
func BenchmarkRuleEvaluate_NikunjyLargeIn(b *testing.B) {
	users := make([]string, 400)
	for i := range users {
		users[i] = fmt.Sprintf(`"%08d-0000-4000-8000-000000000000"`, i)
	}
	rule := flag.Rule{
		Name:            testconvert.String("large-in"),
		VariationResult: testconvert.String("on"),
		Query:           testconvert.String(`userId in [` + strings.Join(users, ", ") + `]`),
	}
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").
		AddCustom("userId", "ffffffff-0000-4000-8000-000000000000").
		Build()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rule.Evaluate("user-1", ctx, "test-flag", false)
	}
}

// BenchmarkInternalFlagValue_MultiRule evaluates a flag whose targeting rules all
// carry a query, the last one matching, so every rule reads the evaluation context.
func BenchmarkInternalFlagValue_MultiRule(b *testing.B) {
	rules := make([]flag.Rule, 0, 5)
	for _, group := range []string{"group-a", "group-b", "group-c", "group-d", "group-e"} {
		rules = append(rules, flag.Rule{
			Query:           testconvert.String(`group eq "` + group + `" and plan eq "pro"`),
			VariationResult: testconvert.String("on"),
		})
	}
	f := flag.InternalFlag{
		Variations: &map[string]*any{
			"on":  testconvert.Interface(true),
			"off": testconvert.Interface(false),
		},
		Rules:       &rules,
		DefaultRule: &flag.Rule{VariationResult: testconvert.String("off")},
	}
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").
		AddCustom("group", "group-e").
		AddCustom("plan", "pro").
		AddCustom("region", "eu").
		AddCustom("beta", true).
		Build()
	flagCtx := flag.Context{DefaultSdkValue: false}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = f.Value("multi-rule", ctx, flagCtx)
	}
}

// BenchmarkInternalFlagValue_MultiRuleJSONLogic is BenchmarkInternalFlagValue_MultiRule
// with JsonLogic queries, which read the evaluation context as JSON.
func BenchmarkInternalFlagValue_MultiRuleJSONLogic(b *testing.B) {
	rules := make([]flag.Rule, 0, 5)
	for _, group := range []string{"group-a", "group-b", "group-c", "group-d", "group-e"} {
		rules = append(rules, flag.Rule{
			Query: testconvert.String(
				`{"and": [{"==": [{"var": "group"}, "` + group + `"]}, {"==": [{"var": "plan"}, "pro"]}]}`,
			),
			VariationResult: testconvert.String("on"),
		})
	}
	f := flag.InternalFlag{
		Variations: &map[string]*any{
			"on":  testconvert.Interface(true),
			"off": testconvert.Interface(false),
		},
		Rules:       &rules,
		DefaultRule: &flag.Rule{VariationResult: testconvert.String("off")},
	}
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").
		AddCustom("group", "group-e").
		AddCustom("plan", "pro").
		AddCustom("region", "eu").
		AddCustom("beta", true).
		Build()
	flagCtx := flag.Context{DefaultSdkValue: false}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = f.Value("multi-rule-jsonlogic", ctx, flagCtx)
	}
}

// BenchmarkRuleEvaluate_NikunjySimpleParallel evaluates one query from every P at
// once, the case where concurrent evaluations compete for the cached evaluators.
func BenchmarkRuleEvaluate_NikunjySimpleParallel(b *testing.B) {
	rule := flag.Rule{
		Name:            testconvert.String("simple-parallel"),
		VariationResult: testconvert.String("on"),
		Query:           testconvert.String(`language eq "ar"`),
	}
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").
		AddCustom("language", "ar").
		Build()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = rule.Evaluate("user-1", ctx, "test-flag", false)
		}
	})
}
