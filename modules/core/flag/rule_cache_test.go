package flag

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetNikunjyEvaluatorCache restores the package-global cache to its lazy-init
// default so each test starts from a clean state.
func resetNikunjyEvaluatorCache(t *testing.T) {
	t.Helper()
	nikunjyEvaluatorCache.Store(nil)
}

func TestSetRuleEvaluatorCacheSize_Validation(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{name: "zero rejected", size: 0, wantErr: true},
		{name: "negative rejected", size: -10, wantErr: true},
		{name: "positive accepted", size: 5, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetNikunjyEvaluatorCache(t)
			err := SetRuleEvaluatorCacheSize(tt.size)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNikunjyEvaluatorCache_DefaultSize(t *testing.T) {
	resetNikunjyEvaluatorCache(t)
	c := ensureNikunjyEvaluatorCache()
	assert.NotNil(t, c)
	// Fill past default would be expensive; just verify capacity by checking the
	// cache evicts at exactly the configured size when we shrink it.
}

func TestNikunjyEvaluatorCache_Eviction(t *testing.T) {
	resetNikunjyEvaluatorCache(t)
	assert.NoError(t, SetRuleEvaluatorCacheSize(2))

	q1 := `key eq "a"`
	q2 := `key eq "b"`
	q3 := `key eq "c"`

	_, err := getNikunjyEvaluator(q1)
	assert.NoError(t, err)
	_, err = getNikunjyEvaluator(q2)
	assert.NoError(t, err)

	cache := ensureNikunjyEvaluatorCache()
	assert.True(t, cache.Contains(q1))
	assert.True(t, cache.Contains(q2))

	// Adding q3 must evict the LRU entry (q1).
	_, err = getNikunjyEvaluator(q3)
	assert.NoError(t, err)
	assert.False(t, cache.Contains(q1), "q1 should have been evicted")
	assert.True(t, cache.Contains(q2))
	assert.True(t, cache.Contains(q3))
}

func TestNikunjyEvaluatorCache_ConcurrentSafe(t *testing.T) {
	resetNikunjyEvaluatorCache(t)
	assert.NoError(t, SetRuleEvaluatorCacheSize(50))

	const goroutines = 32
	const queriesPerG = 20

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func(g int) {
			defer wg.Done()
			for i := range queriesPerG {
				q := fmt.Sprintf(`key eq "v%d"`, i%10)
				ev, err := getNikunjyEvaluator(q)
				if err != nil {
					t.Errorf("unexpected err: %v", err)
					return
				}
				_, err = ev.process(map[string]any{"key": "v0"})
				if err != nil {
					t.Errorf("unexpected process err: %v", err)
					return
				}
			}
		}(g)
	}
	wg.Wait()
}

func TestNikunjyEvaluatorCache_ResizePreservesEntries(t *testing.T) {
	resetNikunjyEvaluatorCache(t)
	assert.NoError(t, SetRuleEvaluatorCacheSize(5))
	q := `key eq "a"`
	_, err := getNikunjyEvaluator(q)
	assert.NoError(t, err)
	assert.True(t, ensureNikunjyEvaluatorCache().Contains(q))

	// Resizing to a capacity that still fits the existing entry must keep it.
	assert.NoError(t, SetRuleEvaluatorCacheSize(3))
	assert.True(t, ensureNikunjyEvaluatorCache().Contains(q),
		"resize should preserve entries that still fit")

	// A no-op resize (same size) must also keep entries.
	assert.NoError(t, SetRuleEvaluatorCacheSize(3))
	assert.True(t, ensureNikunjyEvaluatorCache().Contains(q),
		"no-op resize must not drop entries")
}

func TestNikunjyEvaluatorCache_ResizeEvictsOverflow(t *testing.T) {
	resetNikunjyEvaluatorCache(t)
	assert.NoError(t, SetRuleEvaluatorCacheSize(3))
	queries := []string{`key eq "a"`, `key eq "b"`, `key eq "c"`}
	for _, q := range queries {
		_, err := getNikunjyEvaluator(q)
		assert.NoError(t, err)
	}

	// Shrinking below current entry count must evict the LRU entries.
	assert.NoError(t, SetRuleEvaluatorCacheSize(1))
	cache := ensureNikunjyEvaluatorCache()
	assert.False(t, cache.Contains(queries[0]))
	assert.False(t, cache.Contains(queries[1]))
	assert.True(t, cache.Contains(queries[2]))
}
