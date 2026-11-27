package flag

import (
	"fmt"
	"maps"
	"runtime"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestNikunjyEvaluatorCache_Validate(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantErr    assert.ErrorAssertionFunc
		wantCached bool
	}{
		{
			name:       "valid query is cached for evaluation",
			query:      `key eq "validated"`,
			wantErr:    assert.NoError,
			wantCached: true,
		},
		{
			name:    "query that does not parse fails every validation",
			query:   `invalid query syntax`,
			wantErr: assert.Error,
		},
		{
			name:       "query that parses but fails on an empty context fails every validation",
			query:      `key co 1`,
			wantErr:    assert.Error,
			wantCached: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newNikunjyEvaluatorCache()

			tt.wantErr(t, c.validate(tt.query))
			cached, ok := c.entries.Load(tt.query)
			require.Equal(t, tt.wantCached, ok, "is the query cached after validation")

			runtime.GC()
			runtime.GC()
			tt.wantErr(t, c.validate(tt.query), "validating a query again must give the same result")
			if tt.wantCached {
				ev, err := c.get(tt.query)
				require.NoError(t, err)
				assert.Same(t, cached, ev, "evaluation after a GC must reuse the evaluator parsed by validation")
			}
		})
	}
}

func TestNikunjyEvaluatorCache_ConcurrentEvaluationsShareOneEvaluator(t *testing.T) {
	c := newNikunjyEvaluatorCache()
	const query = `group eq "match"`
	ev, err := c.get(query)
	require.NoError(t, err)

	goroutines := 4 * runtime.GOMAXPROCS(0)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func() {
			defer wg.Done()
			for i := range 200 {
				shared, err := c.get(query)
				if !assert.NoError(t, err) {
					return
				}
				assert.Same(t, ev, shared, "every evaluation must share the cached evaluator")
				wantMatch := (g+i)%2 == 0
				group := "other"
				if wantMatch {
					group = "match"
				}
				ok, err := shared.Process(map[string]any{"group": group})
				assert.NoError(t, err)
				assert.Equal(t, wantMatch, ok, "goroutine %d iteration %d", g, i)
			}
		}()
	}
	wg.Wait()
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
// evict entries of this test. It swaps a package variable: not for parallel tests.
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

func refCounts(c *nikunjyEvaluatorCache, queries []string) map[string]int {
	counts := make(map[string]int, len(queries))
	for _, query := range queries {
		counts[query] = refCount(c, query)
	}
	return counts
}

// haveRefs reports whether every query has n references.
func haveRefs(c *nikunjyEvaluatorCache, queries []string, n int) bool {
	for _, query := range queries {
		if refCount(c, query) != n {
			return false
		}
	}
	return true
}

// cachedQueries returns the queries that have a cache entry, in their order.
func cachedQueries(c *nikunjyEvaluatorCache, queries []string) []string {
	var cached []string
	for _, query := range queries {
		if _, ok := c.entries.Load(query); ok {
			cached = append(cached, query)
		}
	}
	return cached
}

func gcUntil(t *testing.T, condition func() bool, msg string) {
	t.Helper()
	require.Eventually(t, func() bool {
		runtime.GC()
		return condition()
	}, 5*time.Second, 10*time.Millisecond, msg)
}

func TestNikunjyEvaluatorCache_Eviction(t *testing.T) {
	const (
		removed   = `key eq "removed-by-reload"`
		kept      = `key eq "kept-by-reload"`
		added     = `key eq "added-by-reload"`
		jsonLogic = `{"==": [{"var": "key"}, "json-logic"]}`
	)
	tests := []struct {
		name       string
		oldQueries []string
		newQueries []string
		// wantRefs: the references per query once both rule lists are validated.
		wantRefs map[string]int
		// wantEvicted and wantCached: the cache once the old rule list is collected.
		wantEvicted []string
		wantCached  []string
	}{
		{
			name:        "a reload evicts the removed query and keeps the others",
			oldQueries:  []string{removed, kept, jsonLogic},
			newQueries:  []string{kept, added},
			wantRefs:    map[string]int{removed: 1, kept: 2, added: 1, jsonLogic: 0},
			wantEvicted: []string{removed},
			wantCached:  []string{kept, added},
		},
		{
			name:       "a query shared by two rule lists stays until both are collected",
			oldQueries: []string{kept},
			newQueries: []string{kept},
			wantRefs:   map[string]int{kept: 2},
			wantCached: []string{kept},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := useFreshNikunjyEvaluators(t)
			queries := slices.Sorted(maps.Keys(tt.wantRefs))
			oldFlag := newFlagWithQueries(tt.oldQueries...)
			require.NoError(t, oldFlag.IsValid())
			newFlag := newFlagWithQueries(tt.newQueries...)
			require.NoError(t, newFlag.IsValid())
			require.Equal(t, tt.wantRefs, refCounts(c, queries), "references once both rule lists are validated")

			runtime.KeepAlive(oldFlag)
			gcUntil(t, func() bool { return haveRefs(c, tt.newQueries, 1) },
				"the old rule list must release its queries once collected")
			assert.Empty(t, cachedQueries(c, tt.wantEvicted), "queries that no rule uses anymore must be evicted")
			assert.Equal(t, tt.wantCached, cachedQueries(c, tt.wantCached),
				"queries that the new rule list uses must stay cached")

			runtime.KeepAlive(newFlag)
			gcUntil(t, func() bool { return len(cachedQueries(c, tt.wantCached)) == 0 },
				"the last rule list using a query must evict it once collected")
			assert.True(t, haveRefs(c, queries, 0), "no query may keep a reference once both rule lists are collected")
		})
	}
}
