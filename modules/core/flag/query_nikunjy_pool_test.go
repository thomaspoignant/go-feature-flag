package flag_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/internalerror"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestNikunjyPool_Evaluate(t *testing.T) {
	type args struct {
		key       string
		ctx       ffcontext.Context
		flagName  string
		isDefault bool
	}
	tests := []struct {
		name    string
		rule    flag.Rule
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "simple match",
			rule: flag.Rule{
				Name:            testconvert.String("pool-simple-match"),
				VariationResult: testconvert.String("on"),
				Query:           testconvert.String(`key eq "abc"`),
			},
			args: args{
				key:      "abc",
				ctx:      ffcontext.NewEvaluationContext("abc"),
				flagName: "test-pool",
			},
			want:    "on",
			wantErr: assert.NoError,
		},
		{
			name: "simple no-match",
			rule: flag.Rule{
				Name:            testconvert.String("pool-simple-nomatch"),
				VariationResult: testconvert.String("on"),
				Query:           testconvert.String(`key eq "abc"`),
			},
			args: args{
				key:      "xyz",
				ctx:      ffcontext.NewEvaluationContext("xyz"),
				flagName: "test-pool",
			},
			wantErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				var target *internalerror.RuleNotApplyError
				return assert.ErrorAs(t, err, &target)
			},
		},
		{
			name: "multi-predicate match",
			rule: flag.Rule{
				Name:            testconvert.String("pool-multi-match"),
				VariationResult: testconvert.String("variant-b"),
				Query:           testconvert.String(`language eq "ar" and premium eq true`),
			},
			args: args{
				key: "user-42",
				ctx: ffcontext.NewEvaluationContextBuilder("user-42").
					AddCustom("language", "ar").
					AddCustom("premium", true).
					Build(),
				flagName: "test-pool-multi",
			},
			want:    "variant-b",
			wantErr: assert.NoError,
		},
		{
			name: "multi-predicate no-match (one attribute fails)",
			rule: flag.Rule{
				Name:            testconvert.String("pool-multi-nomatch"),
				VariationResult: testconvert.String("variant-b"),
				Query:           testconvert.String(`language eq "ar" and premium eq true`),
			},
			args: args{
				key: "user-42",
				ctx: ffcontext.NewEvaluationContextBuilder("user-42").
					AddCustom("language", "ar").
					AddCustom("premium", false).
					Build(),
				flagName: "test-pool-multi",
			},
			wantErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				var target *internalerror.RuleNotApplyError
				return assert.ErrorAs(t, err, &target)
			},
		},
		{
			name: "repeated evaluation hits cache",
			rule: flag.Rule{
				Name:            testconvert.String("pool-cache-hit"),
				VariationResult: testconvert.String("cached"),
				Query:           testconvert.String(`key eq "repeat"`),
			},
			args: args{
				key:      "repeat",
				ctx:      ffcontext.NewEvaluationContext("repeat"),
				flagName: "test-pool-cache",
			},
			want:    "cached",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.rule.Evaluate(tt.args.key, tt.args.ctx, tt.args.flagName, tt.args.isDefault)
			if tt.wantErr != nil {
				tt.wantErr(t, err)
			}
			if tt.want != "" {
				assert.Equal(t, tt.want, got)
			}

			if tt.name == "repeated evaluation hits cache" {
				got2, err2 := tt.rule.Evaluate(tt.args.key, tt.args.ctx, tt.args.flagName, tt.args.isDefault)
				assert.NoError(t, err2)
				assert.Equal(t, got, got2)
			}
		})
	}
}

func TestNikunjyPool_ConcurrentEvaluations(t *testing.T) {
	rule := flag.Rule{
		Name:            testconvert.String("pool-concurrent"),
		VariationResult: testconvert.String("on"),
		Query:           testconvert.String(`key eq "concurrent-user"`),
	}
	ctx := ffcontext.NewEvaluationContext("concurrent-user")

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	results := make([]string, goroutines)
	errs := make([]error, goroutines)

	for i := range goroutines {
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = rule.Evaluate("concurrent-user", ctx, "test-concurrent", false)
		}(i)
	}
	wg.Wait()

	for i := range goroutines {
		assert.NoError(t, errs[i], "goroutine %d returned an error", i)
		assert.Equal(t, "on", results[i], "goroutine %d returned wrong variation", i)
	}
}

func TestNikunjyPool_ConcurrentDifferentQueries(t *testing.T) {
	const numQueries = 10
	const goroutinesPerQuery = 5

	type queryCase struct {
		rule flag.Rule
		ctx  ffcontext.Context
		key  string
	}

	cases := make([]queryCase, numQueries)
	for i := range numQueries {
		attr := fmt.Sprintf("attr_%d", i)
		cases[i] = queryCase{
			rule: flag.Rule{
				Name:            testconvert.String(fmt.Sprintf("pool-diffquery-%d", i)),
				VariationResult: testconvert.String(fmt.Sprintf("var-%d", i)),
				Query:           testconvert.String(fmt.Sprintf(`%s eq "val"`, attr)),
			},
			ctx: ffcontext.NewEvaluationContextBuilder(fmt.Sprintf("user-%d", i)).
				AddCustom(attr, "val").
				Build(),
			key: fmt.Sprintf("user-%d", i),
		}
	}

	var wg sync.WaitGroup
	total := numQueries * goroutinesPerQuery
	wg.Add(total)

	type result struct {
		variation string
		err       error
	}
	results := make([]result, total)

	for i := range cases {
		for j := range goroutinesPerQuery {
			idx := i*goroutinesPerQuery + j
			qc := cases[i]
			go func() {
				defer wg.Done()
				v, err := qc.rule.Evaluate(qc.key, qc.ctx, "test-diffquery", false)
				results[idx] = result{variation: v, err: err}
			}()
		}
	}
	wg.Wait()

	for i := range cases {
		expected := fmt.Sprintf("var-%d", i)
		for j := range goroutinesPerQuery {
			idx := i*goroutinesPerQuery + j
			assert.NoError(t, results[idx].err, "query %d goroutine %d returned an error", i, j)
			assert.Equal(t, expected, results[idx].variation, "query %d goroutine %d returned wrong variation", i, j)
		}
	}
}
