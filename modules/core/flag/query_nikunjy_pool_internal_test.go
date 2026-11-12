package flag

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestNikunjyEvaluatorCache_KeepsEvaluatorAcrossGC(t *testing.T) {
	p, err := newNikunjyEvaluatorCache().get(`key eq "survives-gc"`)
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
	c := useFreshNikunjyEvaluators(t)
	const query = `key eq "seeded-by-validation"`

	require.NoError(t, validateNikunjyQuery(query))

	cached, ok := c.entries.Load(query)
	require.True(t, ok, "validation must leave the parsed query in the cache")
	p, err := c.get(query)
	require.NoError(t, err)
	assert.Same(t, cached, p, "evaluation must reuse the cache entry created during validation")
}

func TestValidateNikunjyQuery_DoesNotPin(t *testing.T) {
	c := useFreshNikunjyEvaluators(t)
	const query = `key eq "validated-only"`

	require.NoError(t, validateNikunjyQuery(query))

	p, err := c.get(query)
	require.NoError(t, err)
	assert.Nil(t, p.pinned.Load(), "the parse of a query that is only validated must stay collectable")
}

func TestValidateNikunjyQuery_RemembersAValidQuery(t *testing.T) {
	c := useFreshNikunjyEvaluators(t)
	const query = `key eq "validated-twice"`
	require.NoError(t, validateNikunjyQuery(query))
	p, err := c.get(query)
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
	c := useFreshNikunjyEvaluators(t)
	const query = `invalid query syntax`

	require.Error(t, validateNikunjyQuery(query))
	_, ok := c.entries.Load(query)
	require.True(t, ok, "the query parses, so validation leaves a cache entry")

	assert.Error(t, validateNikunjyQuery(query), "an invalid query must fail every validation")
}

func TestValidateNikunjyQuery_KeepsAnEvaluatedQueryPinned(t *testing.T) {
	c := useFreshNikunjyEvaluators(t)
	const query = `key eq "evaluated-then-validated"`
	p, err := c.get(query)
	require.NoError(t, err)
	_, err = p.process(map[string]any{})
	require.NoError(t, err)
	pinned := p.pinned.Load()
	require.NotNil(t, pinned)

	require.NoError(t, validateNikunjyQuery(query))

	assert.Same(t, pinned, p.pinned.Load(), "validating an evaluated query must leave its evaluator pinned")
}

func TestNikunjyEvaluatorCache_ConcurrentUseTakesAnotherEvaluator(t *testing.T) {
	p, err := newNikunjyEvaluatorCache().get(`key eq "overflow"`)
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
	p, err := newNikunjyEvaluatorCache().get(`tenant eq "match"`)
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
	useFreshNikunjyEvaluators(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%1000 == 0 {
			b.StopTimer()
			nikunjyEvaluators = newNikunjyEvaluatorCache()
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

func newFlagWithQueries(queries ...string) *InternalFlag {
	rules := make([]Rule, len(queries))
	for i, q := range queries {
		rules[i] = Rule{
			Name:            testconvert.String(fmt.Sprintf("rule-%d", i)),
			Query:           testconvert.String(q),
			VariationResult: testconvert.String("on"),
		}
	}
	return &InternalFlag{
		Variations:  &map[string]*any{"on": testconvert.Interface(true)},
		Rules:       &rules,
		DefaultRule: &Rule{VariationResult: testconvert.String("on")},
	}
}

// useFreshNikunjyEvaluators gives the test its own package evaluator cache. Cleanups
// registered by earlier tests release into the cache they retained in, so they cannot
// evict entries of this test.
func useFreshNikunjyEvaluators(tb testing.TB) *nikunjyEvaluatorCache {
	previous := nikunjyEvaluators
	nikunjyEvaluators = newNikunjyEvaluatorCache()
	tb.Cleanup(func() { nikunjyEvaluators = previous })
	return nikunjyEvaluators
}

func refCount(c *nikunjyEvaluatorCache, query string) int {
	c.refsMu.Lock()
	defer c.refsMu.Unlock()
	return c.refs[query]
}

func gcUntil(t *testing.T, condition func() bool, msg string) {
	t.Helper()
	require.Eventually(t, func() bool {
		runtime.GC()
		return condition()
	}, 5*time.Second, 10*time.Millisecond, msg)
}

func TestNikunjyEvaluatorCache_EvictsQueriesRemovedByAReload(t *testing.T) {
	c := useFreshNikunjyEvaluators(t)
	const (
		removed   = `key eq "removed-by-reload"`
		kept      = `key eq "kept-by-reload"`
		added     = `key eq "added-by-reload"`
		jsonLogic = `{"==": [{"var": "key"}, "json-logic"]}`
	)
	oldFlag := newFlagWithQueries(removed, kept, jsonLogic)
	require.NoError(t, oldFlag.IsValid())
	newFlag := newFlagWithQueries(kept, added)
	require.NoError(t, newFlag.IsValid())
	require.Equal(t, 2, refCount(c, kept))
	require.Zero(t, refCount(c, jsonLogic), "a jsonlogic query is not a nikunjy query")

	oldFlag = nil
	gcUntil(t, func() bool { return refCount(c, kept) == 1 },
		"the rules of the previous configuration must release their queries once collected")

	_, ok := c.entries.Load(removed)
	assert.False(t, ok, "a query no rule uses anymore must be evicted")
	_, ok = c.entries.Load(kept)
	assert.True(t, ok, "a query the new configuration still uses must stay cached")
	_, ok = c.entries.Load(added)
	assert.True(t, ok, "a query added by the new configuration must stay cached")
	runtime.KeepAlive(newFlag)
}

func TestNikunjyEvaluatorCache_EvictsAQueryOnceEveryRuleListReleasedIt(t *testing.T) {
	c := useFreshNikunjyEvaluators(t)
	const query = `key eq "shared-by-two-clients"`
	first := newFlagWithQueries(query)
	require.NoError(t, first.IsValid())
	second := newFlagWithQueries(query)
	require.NoError(t, second.IsValid())

	first = nil
	gcUntil(t, func() bool { return refCount(c, query) == 1 }, "the first rule list must release its query")
	_, ok := c.entries.Load(query)
	require.True(t, ok, "a query another rule list still uses must stay cached")

	runtime.KeepAlive(second)
	second = nil
	gcUntil(t, func() bool {
		_, ok := c.entries.Load(query)
		return !ok
	}, "a query must be evicted once the last rule list using it is collected")
	assert.Zero(t, refCount(c, query))
}

func TestNikunjyEvaluatorCache_RetainIgnoresNilRules(t *testing.T) {
	assert.NotPanics(t, func() { newNikunjyEvaluatorCache().retain(nil) })
}
