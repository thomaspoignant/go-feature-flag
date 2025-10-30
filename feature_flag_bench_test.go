//go:build bench

package ffclient_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"text/template"
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

var client *ffclient.GoFeatureFlag

// init is creating a flag file for this test with the expected date.
// nolint
func init() {
	content, _ := os.ReadFile("testdata/benchmark/flag-config.yaml")
	t, _ := template.New("example-flag-config").Parse(string(content))

	var buf bytes.Buffer
	_ = t.Execute(&buf, struct {
		DateNow    string
		DateBefore string
		DateAfter  string
	}{
		DateBefore: time.Now().Add(-3 * time.Second).Format(time.RFC3339),
		DateNow:    time.Now().Format(time.RFC3339),
		DateAfter:  time.Now().Add(3 * time.Second).Format(time.RFC3339),
	})

	flagFile, _ := os.CreateTemp("", "")
	_ = os.WriteFile(flagFile.Name(), buf.Bytes(), os.ModePerm)

	client, _ = ffclient.New(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: flagFile.Name()},
	})
}

func BenchmarkAllFlags(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_ = client.AllFlagsState(user)
	}
}

func BenchmarkBoolVar_NoRule100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.BoolVariation("bool-no-rule-100", user, false)
	}
}

func BenchmarkBoolVar_NoRule0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.BoolVariation("bool-no-rule-0", user, false)
	}
}

func BenchmarkBoolVar_NoRule50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.BoolVariation("bool-no-rule-50", user, false)
	}
}

func BenchmarkBoolVar_Rule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.BoolVariation("bool-rule", user, false)
	}
}

func BenchmarkBoolVar_RuleComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.BoolVariation("bool-rule-complex", user, false)
	}
}

func BenchmarkBoolVar_RolloutProgressive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.BoolVariation("bool-rollout-progressive", user, false)
	}
}

func BenchmarkBoolVar_RolloutScheduled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.BoolVariation("bool-rollout-scheduled", user, false)
	}
}

func BenchmarkStringVar_NoRule100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.StringVariation("string-no-rule-100", user, "error")
	}
}

func BenchmarkStringVar_NoRule0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.StringVariation("string-no-rule-0", user, "error")
	}
}

func BenchmarkStringVar_NoRule50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.StringVariation("string-no-rule-50", user, "error")
	}
}

func BenchmarkStringVar_Rule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.StringVariation("string-rule", user, "error")
	}
}

func BenchmarkStringVar_RuleComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.StringVariation("string-rule-complex", user, "error")
	}
}

func BenchmarkStringVar_RolloutProgressive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.StringVariation("string-rollout-progressive", user, "error")
	}
}

func BenchmarkStringVar_RolloutScheduled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.StringVariation("string-rollout-scheduled", user, "error")
	}
}

func BenchmarkIntVar_NoRule100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.IntVariation("int-no-rule-100", user, 4)
	}
}

func BenchmarkIntVar_NoRule0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.IntVariation("int-no-rule-0", user, 4)
	}
}

func BenchmarkIntVar_NoRule50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.IntVariation("int-no-rule-50", user, 4)
	}
}

func BenchmarkIntVar_Rule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.IntVariation("int-rule", user, 4)
	}
}

func BenchmarkIntVar_RuleComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.IntVariation("int-rule-complex", user, 4)
	}
}

func BenchmarkIntVar_RolloutProgressive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.IntVariation("int-rollout-progressive", user, 4)
	}
}

func BenchmarkIntVar_RolloutScheduled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.IntVariation("int-rollout-scheduled", user, 4)
	}
}

func BenchmarkFloat64Var_NoRule100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.Float64Variation("float64-no-rule-100", user, 4.0)
	}
}

func BenchmarkFloat64Var_NoRule0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.Float64Variation("float64-no-rule-0", user, 4.0)
	}
}

func BenchmarkFloat64Var_NoRule50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.Float64Variation("float64-no-rule-50", user, 4.0)
	}
}

func BenchmarkFloat64Var_Rule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.Float64Variation("float64-rule", user, 4.0)
	}
}

func BenchmarkFloat64Var_RuleComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.Float64Variation("float64-rule-complex", user, 4.0)
	}
}

func BenchmarkFloat64Var_RolloutProgressive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.Float64Variation("float64-rollout-progressive", user, 4.0)
	}
}

func BenchmarkFloat64Var_RolloutScheduled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.Float64Variation("float64-rollout-scheduled", user, 4.0)
	}
}

func BenchmarkJSONVar_NoRule100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.JSONVariation("json-no-rule-100", user, map[string]interface{}{"sdkDefault": "default"})
	}
}

func BenchmarkJSONVar_NoRule0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.JSONVariation("json-no-rule-0", user, map[string]interface{}{"sdkDefault": "default"})
	}
}

func BenchmarkJSONVar_NoRule50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.JSONVariation("json-no-rule-50", user, map[string]interface{}{"sdkDefault": "default"})
	}
}

func BenchmarkJSONVar_Rule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.JSONVariation("json-rule", user, map[string]interface{}{"sdkDefault": "default"})
	}
}

func BenchmarkJSONVar_RuleComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.JSONVariation("json-rule-complex", user, map[string]interface{}{"sdkDefault": "default"})
	}
}

func BenchmarkJSONVar_RolloutProgressive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.JSONVariation("json-rollout-progressive", user, map[string]interface{}{"sdkDefault": "default"})
	}
}

func BenchmarkJSONVar_RolloutScheduled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.JSONVariation("json-rollout-scheduled", user, map[string]interface{}{"sdkDefault": "default"})
	}
}

func BenchmarkJSONArrayVar_NoRule100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.JSONArrayVariation("jsonArr-no-rule-100", user, []interface{}{"sdkDefault", "default"})
	}
}

func BenchmarkJSONArrayVar_NoRule0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.JSONArrayVariation("jsonArr-no-rule-0", user, []interface{}{"sdkDefault", "default"})
	}
}

func BenchmarkJSONArrayVar_NoRule50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.JSONArrayVariation("jsonArr-no-rule-50", user, []interface{}{"sdkDefault", "default"})
	}
}

func BenchmarkJSONArrayVar_Rule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.JSONArrayVariation("jsonArr-rule", user, []interface{}{"sdkDefault", "default"})
	}
}

func BenchmarkJSONArrayVar_RuleComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.JSONArrayVariation("jsonArr-rule-complex", user, []interface{}{"sdkDefault", "default"})
	}
}

func BenchmarkJSONArrayVar_RolloutProgressive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.JSONArrayVariation("jsonArr-rollout-progressive", user, []interface{}{"sdkDefault", "default"})
	}
}

func BenchmarkJSONArrayVar_RolloutScheduled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffcontext.NewEvaluationContext(fmt.Sprintf("user-%d", i))
		_, _ = client.JSONArrayVariation("jsonArr-rollout-scheduled", user, []interface{}{"sdkDefault", "default"})
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
