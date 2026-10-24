//go:build bench

package ffclient_test

import (
	"fmt"
	"testing"
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

var client *ffclient.GoFeatureFlag

// init starts the client on the benchmark fixture.
// It panics on error so the benchmarks never run against an empty flag set.
func init() {
	var err error
	client, err = ffclient.New(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: "testdata/benchmark/flag-config.yaml"},
	})
	if err != nil {
		panic(err)
	}
}

func BenchmarkAllFlags(b *testing.B) {
	state := client.AllFlagsState(ffcontext.NewEvaluationContext("user-0"))
	if !state.IsValid() || len(state.GetFlags()) != 42 {
		b.Fatalf("expected 42 valid flags, got %d (valid=%t)", len(state.GetFlags()), state.IsValid())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_ = client.AllFlagsState(user)
	}
}

func BenchmarkBoolVar_NoRule100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.BoolVariation("bool-no-rule-100", user, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBoolVar_NoRule0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.BoolVariation("bool-no-rule-0", user, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBoolVar_NoRule50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.BoolVariation("bool-no-rule-50", user, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBoolVar_Rule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.BoolVariation("bool-rule", user, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBoolVar_RuleComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.BoolVariation("bool-rule-complex", user, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBoolVar_RolloutProgressive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.BoolVariation("bool-rollout-progressive", user, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBoolVar_RolloutScheduled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.BoolVariation("bool-rollout-scheduled", user, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringVar_NoRule100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.StringVariation("string-no-rule-100", user, "error"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringVar_NoRule0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.StringVariation("string-no-rule-0", user, "error"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringVar_NoRule50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.StringVariation("string-no-rule-50", user, "error"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringVar_Rule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.StringVariation("string-rule", user, "error"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringVar_RuleComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.StringVariation("string-rule-complex", user, "error"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringVar_RolloutProgressive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.StringVariation("string-rollout-progressive", user, "error"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringVar_RolloutScheduled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.StringVariation("string-rollout-scheduled", user, "error"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIntVar_NoRule100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.IntVariation("int-no-rule-100", user, 4); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIntVar_NoRule0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.IntVariation("int-no-rule-0", user, 4); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIntVar_NoRule50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.IntVariation("int-no-rule-50", user, 4); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIntVar_Rule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.IntVariation("int-rule", user, 4); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIntVar_RuleComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.IntVariation("int-rule-complex", user, 4); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIntVar_RolloutProgressive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.IntVariation("int-rollout-progressive", user, 4); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIntVar_RolloutScheduled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.IntVariation("int-rollout-scheduled", user, 4); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFloat64Var_NoRule100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.Float64Variation("float64-no-rule-100", user, 4.0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFloat64Var_NoRule0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.Float64Variation("float64-no-rule-0", user, 4.0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFloat64Var_NoRule50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.Float64Variation("float64-no-rule-50", user, 4.0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFloat64Var_Rule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.Float64Variation("float64-rule", user, 4.0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFloat64Var_RuleComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.Float64Variation("float64-rule-complex", user, 4.0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFloat64Var_RolloutProgressive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.Float64Variation("float64-rollout-progressive", user, 4.0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFloat64Var_RolloutScheduled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.Float64Variation("float64-rollout-scheduled", user, 4.0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONVar_NoRule100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.JSONVariation("json-no-rule-100", user, map[string]any{"sdkDefault": "default"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONVar_NoRule0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.JSONVariation("json-no-rule-0", user, map[string]any{"sdkDefault": "default"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONVar_NoRule50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.JSONVariation("json-no-rule-50", user, map[string]any{"sdkDefault": "default"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONVar_Rule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.JSONVariation("json-rule", user, map[string]any{"sdkDefault": "default"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONVar_RuleComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.JSONVariation("json-rule-complex", user, map[string]any{"sdkDefault": "default"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONVar_RolloutProgressive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.JSONVariation("json-rollout-progressive", user, map[string]any{"sdkDefault": "default"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONVar_RolloutScheduled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.JSONVariation("json-rollout-scheduled", user, map[string]any{"sdkDefault": "default"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONArrayVar_NoRule100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.JSONArrayVariation("jsonArr-no-rule-100", user, []any{"sdkDefault", "default"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONArrayVar_NoRule0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.JSONArrayVariation("jsonArr-no-rule-0", user, []any{"sdkDefault", "default"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONArrayVar_NoRule50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.JSONArrayVariation("jsonArr-no-rule-50", user, []any{"sdkDefault", "default"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONArrayVar_Rule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.JSONArrayVariation("jsonArr-rule", user, []any{"sdkDefault", "default"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONArrayVar_RuleComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.JSONArrayVariation("jsonArr-rule-complex", user, []any{"sdkDefault", "default"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONArrayVar_RolloutProgressive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.JSONArrayVariation("jsonArr-rollout-progressive", user, []any{"sdkDefault", "default"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONArrayVar_RolloutScheduled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		if _, err := client.JSONArrayVariation("jsonArr-rollout-scheduled", user, []any{"sdkDefault", "default"}); err != nil {
			b.Fatal(err)
		}
	}
}

/* Benchmark list:

Generate a dynamic flag file in the init method
for all tests.

- boolvariation classic
- boolvariation schedule
- boolvariation progressive
- boolvariation experimentation
- boolvariation 0%
- boolvariation 50%

- idem for all type of variations
- all flag with a lot of flags and different types

*/
