package flag

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestNikunjyEvaluatorCache_KeepsEvaluatorAcrossGC(t *testing.T) {
	t.Cleanup(nikunjyEvaluatorCache.Clear)
	p, err := getNikunjyEvaluator(`key eq "survives-gc"`)
	require.NoError(t, err)
	_, err = p.process(map[string]any{"key": "survives-gc"})
	require.NoError(t, err)
	parsed := p.pinned.Load()
	require.NotNil(t, parsed, "the first evaluation pins an evaluator")

	runtime.GC()
	runtime.GC()
	ok, err := p.process(map[string]any{"key": "survives-gc"})

	require.NoError(t, err)
	assert.True(t, ok)
	assert.Same(t, parsed, p.pinned.Load(), "evaluation after GC must reuse the evaluator parsed before it")
}

func TestValidateNikunjyQuery_SeedsTheEvaluatorCache(t *testing.T) {
	t.Cleanup(nikunjyEvaluatorCache.Clear)
	const query = `key eq "seeded-by-validation"`

	require.NoError(t, validateNikunjyQuery(query))

	cached, ok := nikunjyEvaluatorCache.Load(query)
	require.True(t, ok, "validation must leave the parsed query in the cache")
	p, err := getNikunjyEvaluator(query)
	require.NoError(t, err)
	assert.Same(t, cached, p, "evaluation must reuse the cache entry created during validation")
}

func TestValidateNikunjyQuery_DoesNotPin(t *testing.T) {
	t.Cleanup(nikunjyEvaluatorCache.Clear)
	const query = `key eq "validated-only"`

	require.NoError(t, validateNikunjyQuery(query))

	p, err := getNikunjyEvaluator(query)
	require.NoError(t, err)
	assert.Nil(t, p.pinned.Load(), "the parse of a query that is only validated must stay collectable")
}

func TestValidateNikunjyQuery_RemembersAValidQuery(t *testing.T) {
	t.Cleanup(nikunjyEvaluatorCache.Clear)
	const query = `key eq "validated-twice"`
	require.NoError(t, validateNikunjyQuery(query))
	p, err := getNikunjyEvaluator(query)
	require.NoError(t, err)
	p.pool.New = func() any {
		t.Error("validating a valid query again must not parse it")
		return nil
	}
	runtime.GC()
	runtime.GC()

	assert.NoError(t, validateNikunjyQuery(query))
}

func TestValidateNikunjyQuery_DoesNotRememberAnInvalidQuery(t *testing.T) {
	t.Cleanup(nikunjyEvaluatorCache.Clear)
	const query = `invalid query syntax`

	require.Error(t, validateNikunjyQuery(query))
	_, ok := nikunjyEvaluatorCache.Load(query)
	require.True(t, ok, "the query parses, so validation leaves a cache entry")

	assert.Error(t, validateNikunjyQuery(query), "an invalid query must fail every validation")
}

func TestValidateNikunjyQuery_KeepsAnEvaluatedQueryPinned(t *testing.T) {
	t.Cleanup(nikunjyEvaluatorCache.Clear)
	const query = `key eq "evaluated-then-validated"`
	p, err := getNikunjyEvaluator(query)
	require.NoError(t, err)
	_, err = p.process(map[string]any{})
	require.NoError(t, err)
	pinned := p.pinned.Load()
	require.NotNil(t, pinned)

	require.NoError(t, validateNikunjyQuery(query))

	assert.Same(t, pinned, p.pinned.Load(), "validating an evaluated query must leave its evaluator pinned")
}

func TestNikunjyEvaluatorCache_ConcurrentUseTakesAnotherEvaluator(t *testing.T) {
	t.Cleanup(nikunjyEvaluatorCache.Clear)
	p, err := getNikunjyEvaluator(`key eq "overflow"`)
	require.NoError(t, err)
	_, err = p.process(map[string]any{})
	require.NoError(t, err)
	pinned := p.pinned.Load()
	require.NotNil(t, pinned)

	first := p.get()
	second := p.get()

	assert.Same(t, pinned, first, "the first evaluation takes the pinned evaluator")
	assert.NotSame(t, first, second, "an evaluation running at the same time gets its own evaluator")
	p.put(second)
	p.put(first)
	assert.Same(t, second, p.pinned.Load(), "the first evaluator returned is pinned again")
}

func TestNikunjyEvaluatorCache_ConcurrentEvaluationsDoNotShareEvaluators(t *testing.T) {
	t.Cleanup(nikunjyEvaluatorCache.Clear)
	p, err := getNikunjyEvaluator(`tenant eq "match"`)
	require.NoError(t, err)

	goroutines := 4 * runtime.GOMAXPROCS(0)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func() {
			defer wg.Done()
			for i := range 200 {
				wantMatch := (g+i)%2 == 0
				tenant := "other"
				if wantMatch {
					tenant = "match"
				}
				ok, err := p.process(map[string]any{"tenant": tenant})
				assert.NoError(t, err)
				assert.Equal(t, wantMatch, ok, "goroutine %d iteration %d", g, i)
			}
		}()
	}
	wg.Wait()

	assert.NotNil(t, p.pinned.Load(), "an evaluator is pinned again once every evaluation returned")
}

// BenchmarkRule_ValidateThenEvaluate measures loading a rule with a query never
// seen before: validating it, as the flag cache does on every configuration
// load, then evaluating it once.
func BenchmarkRule_ValidateThenEvaluate(b *testing.B) {
	variations := map[string]*any{"on": testconvert.Interface(true)}
	ctx := ffcontext.NewEvaluationContext("user-1")
	b.Cleanup(nikunjyEvaluatorCache.Clear)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%1000 == 0 {
			b.StopTimer()
			nikunjyEvaluatorCache.Clear()
			b.StartTimer()
		}
		r := Rule{
			VariationResult: testconvert.String("on"),
			Query:           testconvert.String(fmt.Sprintf(`key eq "user-%d"`, i)),
		}
		if err := r.IsValid(false, variations); err != nil {
			b.Fatal(err)
		}
		_, _ = r.Evaluate("user-1", ctx, "test-flag", false)
	}
}
