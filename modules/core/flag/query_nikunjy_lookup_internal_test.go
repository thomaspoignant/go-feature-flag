package flag

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/google/go-cmp/cmp"
	"github.com/nikunjy/rules/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

// useIndexingNikunjyEvaluators is useFreshNikunjyEvaluators with keys for the short queries of
// the tests too.
func useIndexingNikunjyEvaluators(tb testing.TB) *nikunjyEvaluatorCache {
	c := useFreshNikunjyEvaluators(tb)
	previous := nikunjyKeysMinQueryLen
	nikunjyKeysMinQueryLen = 0
	tb.Cleanup(func() { nikunjyKeysMinQueryLen = previous })
	return c
}

// idList returns `key in ["user-0000000", ...]` with n ids, where the id at position
// replaced (if it is not negative) is changed to replacement.
func idList(n, replaced int, replacement string) string {
	var sb strings.Builder
	sb.Grow(n*16 + 16)
	sb.WriteString(`key in [`)
	for i := 0; i < n; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		if i == replaced {
			fmt.Fprintf(&sb, `"%s"`, replacement)
			continue
		}
		fmt.Fprintf(&sb, `"user-%07d"`, i)
	}
	sb.WriteString(`]`)
	return sb.String()
}

// applies reports whether the rule applies to the evaluation context of key.
func applies(r *Rule, key string) bool {
	_, err := r.Evaluate(key, ffcontext.NewEvaluationContext(key), "test-flag", false)
	return err == nil
}

func validatedFlag(t testing.TB, queries ...string) *InternalFlag {
	t.Helper()
	f := newFlagWithQueries(queries...)
	require.NoError(t, f.IsValid())
	return f
}

func indexedEvaluator(c *nikunjyEvaluatorCache, r *Rule) *parser.Evaluator {
	return c.lookup(r.QueryHash, r.GetQuery())
}

func indexSize(c *nikunjyEvaluatorCache) int {
	n := 0
	c.byKey.Range(func(_, _ any) bool { n++; return true })
	return n
}

// assertSharedEvaluators checks that every rule of flags is found by its hash, that rules with
// one text share its bytes, and that wantEvaluators distinct evaluators are behind the rules.
func assertSharedEvaluators(t *testing.T, c *nikunjyEvaluatorCache, flags []*InternalFlag, wantEvaluators int) {
	t.Helper()
	distinct := map[*parser.Evaluator]string{}
	bytesOf := map[string]*byte{}
	for _, f := range flags {
		for i := range *f.Rules {
			r := &(*f.Rules)[i]
			ev := indexedEvaluator(c, r)
			require.NotNil(t, ev, "a validated rule must be found by its hash")
			distinct[ev] = r.GetQuery()
			if shared, ok := bytesOf[r.GetQuery()]; ok {
				assert.Same(t, shared, unsafe.StringData(r.GetQuery()), "rules with one text share its bytes")
			}
			bytesOf[r.GetQuery()] = unsafe.StringData(r.GetQuery())
			want, err := c.get(r.GetTrimmedQuery())
			require.NoError(t, err)
			assert.Same(t, want, ev, "the hash lookup must give the evaluator of the query text")
		}
	}
	assert.Len(t, distinct, wantEvaluators)
}

func TestNikunjyEvaluatorCache_Lookup(t *testing.T) {
	const query = `key eq "shared"`
	tests := []struct {
		name  string
		flags int
		rules int
		// queryOf returns the query of rule r of flag f.
		queryOf func(f, r int) string
		// wantEvaluators is the number of distinct evaluators behind all the rules.
		wantEvaluators int
	}{
		{
			name:           "the same text in many rules of many flags shares one evaluator",
			flags:          20,
			rules:          5,
			queryOf:        func(_, _ int) string { return strings.Clone(query) },
			wantEvaluators: 1,
		},
		{
			name:           "different texts have different evaluators",
			flags:          4,
			rules:          3,
			queryOf:        func(f, r int) string { return fmt.Sprintf(`key eq "shared-%d-%d"`, f, r) },
			wantEvaluators: 12,
		},
		{
			name:  "texts that differ by whitespace or case are different queries",
			flags: 1,
			rules: 4,
			queryOf: func(_, r int) string {
				return []string{`key eq "shared"`, `key  eq "shared"`, `KEY eq "shared"`, `key eq "shared" `}[r]
			},
			wantEvaluators: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := useIndexingNikunjyEvaluators(t)
			flags := make([]*InternalFlag, tt.flags)
			for f := range flags {
				queries := make([]string, tt.rules)
				for r := range queries {
					queries[r] = tt.queryOf(f, r)
				}
				flags[f] = validatedFlag(t, queries...)
			}
			assertSharedEvaluators(t, c, flags, tt.wantEvaluators)
			runtime.KeepAlive(flags)
		})
	}
}

// splitRulesByQuery returns the rules of flags whose query is listA, and the other ones.
func splitRulesByQuery(flags []*InternalFlag, listA string) (rulesA, rulesB []*Rule) {
	for _, f := range flags {
		for i := range *f.Rules {
			r := &(*f.Rules)[i]
			if r.GetQuery() == listA {
				rulesA = append(rulesA, r)
			} else {
				rulesB = append(rulesB, r)
			}
		}
	}
	return rulesA, rulesB
}

// assertListsHaveOwnEvaluators checks that the rules of a list share an evaluator, and that the
// two lists never share one.
func assertListsHaveOwnEvaluators(t *testing.T, c *nikunjyEvaluatorCache, rulesA, rulesB []*Rule) {
	t.Helper()
	evA, evB := indexedEvaluator(c, rulesA[0]), indexedEvaluator(c, rulesB[0])
	require.NotNil(t, evA)
	require.NotNil(t, evB)
	assert.NotSame(t, evA, evB, "two different queries must never share an evaluator")
	for _, r := range rulesA {
		assert.Same(t, evA, indexedEvaluator(c, r))
	}
	for _, r := range rulesB {
		assert.Same(t, evB, indexedEvaluator(c, r))
	}
}

// assertListsApply checks which ids apply to the rules of list A and to the rules of list B.
func assertListsApply(t assert.TestingT, rulesA, rulesB []*Rule, idA, idB string) {
	for _, r := range rulesA {
		assert.True(t, applies(r, idA), "the id of list A must match list A")
		assert.False(t, applies(r, idB), "the id of list B must not match list A")
		assert.True(t, applies(r, "user-0000001"))
		assert.False(t, applies(r, "nobody"))
	}
	for _, r := range rulesB {
		assert.True(t, applies(r, idB), "the id of list B must match list B")
		assert.False(t, applies(r, idA), "the id of list A must not match list B")
		assert.True(t, applies(r, "user-0000001"))
		assert.False(t, applies(r, "nobody"))
	}
}

// runConcurrently runs check from 8 goroutines, 3 times each.
func runConcurrently(t *testing.T, check func(t assert.TestingT)) {
	t.Helper()
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 3; i++ {
				check(t)
			}
		}()
	}
	wg.Wait()
}

func TestNikunjyEvaluatorCache_NearIdenticalLargeQueries(t *testing.T) {
	tests := []struct {
		name string
		n    int
		// listA and listB differ in exactly one id (or in one more id).
		listB func(n int) string
		idA   string // only in A
		idB   string // only in B
	}{
		{
			name:  "300k ids, one id differs, same length",
			n:     300000,
			listB: func(n int) string { return idList(n, n/2, "usex-0150000") },
			idA:   "user-0150000",
			idB:   "usex-0150000",
		},
		{
			name:  "20k ids, the last id differs, same length",
			n:     20000,
			listB: func(n int) string { return idList(n, n-1, "usex-0019999") },
			idA:   "user-0019999",
			idB:   "usex-0019999",
		},
		{
			name:  "ids, one id is longer, different length",
			n:     20000,
			listB: func(n int) string { return idList(n, n/2, "user-0010000-longer") },
			idA:   "user-0010000",
			idB:   "user-0010000-longer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := useIndexingNikunjyEvaluators(t)
			listA, listB := idList(tt.n, -1, ""), tt.listB(tt.n)
			require.NotEqual(t, listA, listB)

			// Both lists are in several flags, in different rule positions, with their
			// own copy of the text.
			flags := []*InternalFlag{
				validatedFlag(t, strings.Clone(listA), strings.Clone(listB)),
				validatedFlag(t, strings.Clone(listB), strings.Clone(listA)),
				validatedFlag(t, strings.Clone(listA)),
			}
			rulesA, rulesB := splitRulesByQuery(flags, listA)
			require.Len(t, rulesA, 3)
			require.Len(t, rulesB, 2)
			assertListsHaveOwnEvaluators(t, c, rulesA, rulesB)

			check := func(t assert.TestingT) {
				assertListsApply(t, rulesA, rulesB, tt.idA, tt.idB)
			}
			check(t)
			t.Run("interleaved and concurrent evaluations", func(t *testing.T) {
				runConcurrently(t, check)
			})
			runtime.KeepAlive(flags)
		})
	}
}

func TestNikunjyEvaluatorCache_LookupIsExact(t *testing.T) {
	tests := []struct {
		name string
		// mutate changes a validated rule, or returns another rule to evaluate.
		mutate func(r *Rule) *Rule
		// wantIndexed is whether the rule can still be found by its hash.
		wantIndexed bool
		// matches and doesNotMatch are keys that must (not) apply to the rule afterwards.
		matches, doesNotMatch string
	}{
		{
			name:        "unchanged rule",
			mutate:      func(r *Rule) *Rule { return r },
			wantIndexed: true,
			matches:     "a", doesNotMatch: "b",
		},
		{
			name:        "copy of the rule keeps the hash",
			mutate:      func(r *Rule) *Rule { c := *r; return &c },
			wantIndexed: true,
			matches:     "a", doesNotMatch: "b",
		},
		{
			name: "query rewritten through its pointer is never evaluated with the old evaluator",
			mutate: func(r *Rule) *Rule {
				*r.Query = `key eq "b"`
				return r
			},
			wantIndexed: false,
			matches:     "b", doesNotMatch: "a",
		},
		{
			name: "another query with the hash of the old one",
			mutate: func(r *Rule) *Rule {
				q := `key eq "b"`
				c := *r
				c.Query = &q
				return &c
			},
			wantIndexed: false,
			matches:     "b", doesNotMatch: "a",
		},
		{
			name: "same text in another string keeps the evaluator",
			mutate: func(r *Rule) *Rule {
				c := *r
				q := strings.Clone(*r.Query)
				c.Query = &q
				return &c
			},
			wantIndexed: true,
			matches:     "a", doesNotMatch: "b",
		},
		{
			name: "merging a query resets the hash",
			mutate: func(r *Rule) *Rule {
				q := `key eq "b"`
				r.MergeRules(Rule{Query: &q})
				return r
			},
			wantIndexed: false,
			matches:     "b", doesNotMatch: "a",
		},
		{
			name: "hash that is not the one of any query",
			mutate: func(r *Rule) *Rule {
				c := *r
				c.QueryHash = 12345
				return &c
			},
			wantIndexed: false,
			matches:     "a", doesNotMatch: "b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := useIndexingNikunjyEvaluators(t)
			f := validatedFlag(t, `key eq "a"`)
			require.NotZero(t, (*f.Rules)[0].QueryHash)
			r := tt.mutate(&(*f.Rules)[0])
			assert.Equal(t, tt.wantIndexed, indexedEvaluator(c, r) != nil)
			assert.True(t, applies(r, tt.matches))
			assert.False(t, applies(r, tt.doesNotMatch))
			runtime.KeepAlive(f)
		})
	}
}

// assertOneTextPerKeyAndEvaluator checks that every rule of flags is found by its hash, that no
// key and no evaluator belongs to two texts, and that there are n of each.
func assertOneTextPerKeyAndEvaluator(t *testing.T, c *nikunjyEvaluatorCache, flags []*InternalFlag, n int) {
	t.Helper()
	keys := map[uint64]string{}
	evaluators := map[*parser.Evaluator]string{}
	for _, f := range flags {
		for i := range *f.Rules {
			r := &(*f.Rules)[i]
			ev := indexedEvaluator(c, r)
			require.NotNil(t, ev)
			if other, ok := keys[r.QueryHash]; ok {
				require.Equal(t, other, r.GetQuery(), "a key must belong to one text")
			}
			keys[r.QueryHash] = r.GetQuery()
			if other, ok := evaluators[ev]; ok {
				require.Equal(t, other, r.GetQuery(), "an evaluator must belong to one text")
			}
			evaluators[ev] = r.GetQuery()
		}
	}
	assert.Len(t, keys, n)
	assert.Len(t, evaluators, n)
	assert.NotContains(t, keys, uint64(0))
}

// assertCollidingQueriesApply checks that each rule of flags applies to the key of its own query
// only, among the n colliding queries.
func assertCollidingQueriesApply(t *testing.T, flags []*InternalFlag, n int) {
	t.Helper()
	for _, f := range flags {
		for i := range *f.Rules {
			r := &(*f.Rules)[i]
			for j := 0; j < n; j++ {
				want := fmt.Sprintf("collide-%d", j)
				assert.Equal(t, r.GetQuery() == fmt.Sprintf(`key eq "%s"`, want), applies(r, want))
			}
		}
	}
}

// TestNikunjyEvaluatorCache_ForcedCollisions makes every query hash to the same value, which
// is the worst case of a 64-bit hash: no two queries may share an evaluator.
func TestNikunjyEvaluatorCache_ForcedCollisions(t *testing.T) {
	tests := []struct {
		name string
		hash func(string) uint64
	}{
		{name: "every query has the same hash", hash: func(string) uint64 { return 42 }},
		{name: "every query hashes to zero, the key of rules without a hash", hash: func(string) uint64 { return 0 }},
		{name: "every query hashes to the last key", hash: func(string) uint64 { return ^uint64(0) }},
		{name: "queries collide in pairs", hash: func(q string) uint64 { return uint64(len(q) % 2) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := useIndexingNikunjyEvaluators(t)
			c.hash = tt.hash
			queries := make([]string, 0, 10)
			for i := 0; i < 10; i++ {
				queries = append(queries, fmt.Sprintf(`key eq "collide-%d"`, i))
			}
			// A second flag with the same texts, in another order and in other strings.
			reversed := make([]string, len(queries))
			for i, q := range queries {
				reversed[len(queries)-1-i] = strings.Clone(q)
			}
			flags := []*InternalFlag{validatedFlag(t, queries...), validatedFlag(t, reversed...)}

			assertOneTextPerKeyAndEvaluator(t, c, flags, len(queries))
			assertCollidingQueriesApply(t, flags, len(queries))

			// Releasing one flag keeps the keys of the other.
			flags[0] = nil
			gcUntil(t, func() bool { return c.keyRefsTotal() == len(queries) }, "the collected flag must release its keys")
			for i := range *flags[1].Rules {
				r := &(*flags[1].Rules)[i]
				assert.NotNil(t, indexedEvaluator(c, r))
			}
			runtime.KeepAlive(flags[1])
		})
	}
}

func (c *nikunjyEvaluatorCache) keyRefsTotal() int {
	c.keysMu.Lock()
	defer c.keysMu.Unlock()
	n := 0
	for _, refs := range c.keyRefs {
		n += refs
	}
	return n
}

func TestRule_QueryHashIsNotSerialized(t *testing.T) {
	useIndexingNikunjyEvaluators(t)
	f := validatedFlag(t, `key eq "a"`)
	require.NotZero(t, (*f.Rules)[0].QueryHash)

	asJSON, err := json.Marshal(f)
	require.NoError(t, err)
	assert.NotContains(t, strings.ToLower(string(asJSON)), "hash")

	var decoded InternalFlag
	require.NoError(t, json.Unmarshal([]byte(`{"targeting":[{"query":"key eq \"a\"","variation":"on","queryHash":7}]}`), &decoded))
	assert.Zero(t, (*decoded.Rules)[0].QueryHash)
}

func TestRule_QueryHashDoesNotMakeFlagsDiffer(t *testing.T) {
	useIndexingNikunjyEvaluators(t)
	loaded := validatedFlag(t, `key eq "a"`)
	reloaded := validatedFlag(t, strings.Clone(`key eq "a"`))
	assert.True(t, cmp.Equal(loaded, reloaded), "two loads of the same flag are the same flag")
	assert.False(t, cmp.Equal(loaded, validatedFlag(t, `key eq "b"`)))
}

func TestNikunjyEvaluatorCache_NotIndexed(t *testing.T) {
	tests := []struct {
		name  string
		query string
		// valid is whether the flag is valid.
		valid        bool
		matches      string
		doesNotMatch string
	}{
		{
			name:         "jsonlogic query",
			query:        `{"==": [{"var": "key"}, "a"]}`,
			valid:        true,
			matches:      "a",
			doesNotMatch: "b",
		},
		{
			name:         "multi-line nikunjy query is indexed by its text before trimming",
			query:        "key eq \n    \"a\"",
			valid:        true,
			matches:      "a",
			doesNotMatch: "b",
		},
		{
			name:  "invalid query",
			query: `key eq`,
			valid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := useIndexingNikunjyEvaluators(t)
			f := newFlagWithQueries(tt.query)
			err := f.IsValid()
			if !tt.valid {
				require.Error(t, err)
				assert.Zero(t, indexSize(c), "an invalid flag must not index its queries")
				return
			}
			require.NoError(t, err)
			r := &(*f.Rules)[0]
			assert.Equal(t, strings.Contains(tt.query, "eq"), indexedEvaluator(c, r) != nil)
			assert.True(t, applies(r, tt.matches))
			assert.False(t, applies(r, tt.doesNotMatch))
			runtime.KeepAlive(f)
		})
	}
}

func TestNikunjyEvaluatorCache_UnvalidatedRuleEvaluates(t *testing.T) {
	useIndexingNikunjyEvaluators(t)
	q := `key eq "a"`
	r := &Rule{Query: &q, VariationResult: &q}
	assert.True(t, applies(r, "a"))
	assert.False(t, applies(r, "b"))
}

func TestNikunjyEvaluatorCache_Reload(t *testing.T) {
	c := useIndexingNikunjyEvaluators(t)
	const flags = 10
	load := func(query func(f int) string) []*InternalFlag {
		set := make([]*InternalFlag, flags)
		for f := range set {
			set[f] = validatedFlag(t, query(f))
		}
		return set
	}
	current := load(func(f int) string { return fmt.Sprintf(`key eq "gen-0-%d"`, f%3) })
	check := func(set []*InternalFlag, gen int) {
		for f, flg := range set {
			r := &(*flg.Rules)[0]
			assert.NotNil(t, indexedEvaluator(c, r))
			assert.True(t, applies(r, fmt.Sprintf("gen-%d-%d", gen, f%3)))
			assert.False(t, applies(r, fmt.Sprintf("gen-%d-%d", gen-1, f%3)), "stale evaluator of the previous load")
		}
	}
	check(current, 0)

	for gen := 1; gen <= 30; gen++ {
		// Even generations reload the same queries, the others change them.
		changed := gen%2 == 1
		next := load(func(f int) string {
			if changed {
				return fmt.Sprintf(`key eq "gen-%d-%d"`, gen, f%3)
			}
			return fmt.Sprintf(`key eq "gen-%d-%d"`, gen-1, f%3)
		})
		if changed {
			check(next, gen)
		} else {
			check(next, gen-1)
		}
		runtime.KeepAlive(current)
		current = next
		runtime.GC()
	}
	gcUntil(t, func() bool { return indexSize(c) <= flags }, "the rules of the old loads must be unindexed")
	assert.LessOrEqual(t, indexSize(c), flags)
	runtime.KeepAlive(current)

	clear(current) // drop the last references to the flags
	gcUntil(t, func() bool { return indexSize(c) == 0 }, "no rule left to index")
	cached := 0
	c.entries.Range(func(_, _ any) bool { cached++; return true })
	assert.Zero(t, cached)
	runtime.KeepAlive(current)
}

func TestNikunjyEvaluatorCache_ConcurrentEvaluationsDuringReloads(t *testing.T) {
	useIndexingNikunjyEvaluators(t)
	listA, listB := idList(2000, -1, ""), idList(2000, 1000, "usex-0001000")
	var current struct {
		sync.RWMutex
		flags []*InternalFlag
	}
	newSet := func(i int) []*InternalFlag {
		a, b := listA, listB
		if i%2 == 1 {
			a, b = listB, listA
		}
		return []*InternalFlag{validatedFlag(t, strings.Clone(a)), validatedFlag(t, strings.Clone(b))}
	}
	current.flags = newSet(0)

	stop := make(chan struct{})
	var wg sync.WaitGroup
	for g := 0; g < 6; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				current.RLock()
				set := current.flags
				current.RUnlock()
				// Whatever the set, one flag has list A and the other list B.
				for _, f := range set {
					r := &(*f.Rules)[0]
					isA := strings.Contains(r.GetQuery(), "user-0001000")
					assert.Equal(t, isA, applies(r, "user-0001000"))
					assert.Equal(t, !isA, applies(r, "usex-0001000"))
				}
			}
		}()
	}
	for i := 1; i <= 20; i++ {
		next := newSet(i)
		current.Lock()
		current.flags = next
		current.Unlock()
		runtime.GC()
	}
	close(stop)
	wg.Wait()
}

func TestHashNikunjyQuery(t *testing.T) {
	long := strings.Repeat("0123456789abcdef", 8)
	tests := []struct {
		name string
		a, b string
		// wantEqual is whether the two queries must have the same hash.
		wantEqual bool
	}{
		{name: "same text in two strings", a: strings.Clone(long), b: strings.Clone(long), wantEqual: true},
		{name: "empty and a zero byte", a: "", b: "\x00"},
		{name: "a text and the text with a zero byte", a: "abc", b: "abc\x00"},
		{name: "one byte of a short text", a: `key eq "a"`, b: `key eq "b"`},
		{name: "one byte of a long text, first block", a: long, b: "1" + long[1:]},
		{name: "one byte of a long text, middle", a: long, b: long[:50] + "X" + long[51:]},
		{name: "one byte of a long text, tail", a: long, b: long[:len(long)-1] + "X"},
		{name: "swapped blocks of 8 bytes", a: long[:8] + "ABCDEFGH" + long[16:], b: "ABCDEFGH" + long[8:16] + long[16:]},
		{name: "case", a: `key eq "a"`, b: `KEY eq "a"`},
		{name: "whitespace", a: `key eq "a"`, b: `key  eq "a"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantEqual, hashNikunjyQuery(tt.a) == hashNikunjyQuery(tt.b))
		})
	}

	t.Run("every length from 0 to 100 has a different hash and is stable", func(t *testing.T) {
		seen := map[uint64]int{}
		for n := 0; n <= 100; n++ {
			h := hashNikunjyQuery(strings.Repeat("a", n))
			assert.Equal(t, h, hashNikunjyQuery(strings.Repeat("a", n)))
			if other, ok := seen[h]; ok {
				t.Errorf("lengths %d and %d have the same hash", other, n)
			}
			seen[h] = n
		}
	})
}

func TestNikunjyEvaluatorCache_ValidatingAgain(t *testing.T) {
	tests := []struct {
		name        string
		validations int
	}{
		{name: "validated once", validations: 1},
		{name: "validated again, as the linter or a copy of a cache does", validations: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := useIndexingNikunjyEvaluators(t)
			f := newFlagWithQueries(`key eq "a"`, `key eq "b"`)
			for i := 0; i < tt.validations; i++ {
				require.NoError(t, f.IsValid())
			}
			hashes := []uint64{(*f.Rules)[0].QueryHash, (*f.Rules)[1].QueryHash}
			assert.NotZero(t, hashes[0])
			assert.NotEqual(t, hashes[0], hashes[1])
			assert.Equal(t, 2*tt.validations, c.keyRefsTotal(), "one reference per rule and validation")
			assert.Equal(t, 2, indexSize(c))
			assert.True(t, applies(&(*f.Rules)[0], "a"))
			assert.False(t, applies(&(*f.Rules)[1], "a"))

			runtime.KeepAlive(f)
			*f = InternalFlag{} // drop the rules
			gcUntil(t, func() bool { return indexSize(c) == 0 }, "the keys of a collected flag must be released")
		})
	}
}

func TestRule_validateReturnsTheEvaluator(t *testing.T) {
	tests := []struct {
		name        string
		rule        Rule
		isDefault   bool
		wantEval    bool
		wantInvalid bool
	}{
		{name: "nikunjy query", rule: Rule{Query: testconvert.String(`key eq "a"`), VariationResult: testconvert.String("on")}, wantEval: true},
		{name: "jsonlogic query", rule: Rule{Query: testconvert.String(`{"==": [1, 1]}`), VariationResult: testconvert.String("on")}},
		{name: "default rule", rule: Rule{VariationResult: testconvert.String("on")}, isDefault: true},
		{
			name: "disabled rule",
			rule: Rule{Query: testconvert.String(`key eq `), Disable: testconvert.Bool(true), VariationResult: testconvert.String("on")},
		},
		{name: "invalid query", rule: Rule{Query: testconvert.String(`key eq `), VariationResult: testconvert.String("on")}, wantInvalid: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, err := tt.rule.validate(tt.isDefault, map[string]*any{"on": testconvert.Interface(true)}, nil)
			if tt.wantInvalid {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantEval, ev != nil)
		})
	}
}

func TestNikunjyEvaluatorCache_Disabled(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		wantIndexed bool
	}{
		{name: "keys are saved when the cleanups that release them run", enabled: true, wantIndexed: true},
		{name: "keys are not saved when the cleanups never run, as with TinyGo", enabled: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := useIndexingNikunjyEvaluators(t)
			previous := nikunjyKeysEnabled
			nikunjyKeysEnabled = tt.enabled
			t.Cleanup(func() { nikunjyKeysEnabled = previous })

			f := validatedFlag(t, `key eq "a"`)
			r := &(*f.Rules)[0]
			assert.Equal(t, tt.wantIndexed, indexedEvaluator(c, r) != nil)
			assert.Equal(t, tt.wantIndexed, indexSize(c) == 1)
			assert.True(t, applies(r, "a"), "the rule evaluates through the evaluator cache")
			assert.False(t, applies(r, "b"))
			runtime.KeepAlive(f)
		})
	}
}

func TestNikunjyEvaluatorCache_ShortQueriesHaveNoKey(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantIndexed bool
	}{
		{name: "short query", query: `key eq "a"`},
		{name: "query just under the length", query: `key eq "` + strings.Repeat("a", nikunjyKeysMinQueryLen-len(`key eq ""`)-1) + `"`},
		{name: "query of the length", query: `key eq "` + strings.Repeat("a", nikunjyKeysMinQueryLen-len(`key eq ""`)) + `"`, wantIndexed: true},
		{name: "long query", query: idList(200, -1, ""), wantIndexed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := useFreshNikunjyEvaluators(t)
			f := newFlagWithQueries(tt.query)
			require.NoError(t, f.IsValid())
			r := &(*f.Rules)[0]
			assert.Equal(t, tt.wantIndexed, indexedEvaluator(c, r) != nil)
			assert.Equal(t, tt.wantIndexed, r.QueryHash != 0)
			assert.True(t, applies(r, "user-0000001") || !strings.Contains(tt.query, "user-"), "the rule evaluates either way")
			runtime.KeepAlive(f)
		})
	}
}

func TestNikunjyEvaluatorCache_KnownQueryIsNotLookedUpByText(t *testing.T) {
	query := idList(200, -1, "")
	tests := []struct {
		name string
		// loads is the number of flags that carry the query, loaded one after the other.
		loads int
		// wantTextRefs is the references of the evaluator cache to the query after the loads: only the
		// first load looks the query up by its text, the others find its key.
		wantTextRefs int
		wantKeyRefs  int
	}{
		{name: "first load: the query is looked up by its text", loads: 1, wantTextRefs: 1, wantKeyRefs: 1},
		{name: "second load: found by its key", loads: 2, wantTextRefs: 1, wantKeyRefs: 2},
		{name: "many loads: found by their key", loads: 20, wantTextRefs: 1, wantKeyRefs: 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := useIndexingNikunjyEvaluators(t)
			flags := make([]*InternalFlag, tt.loads)
			for i := range flags {
				flags[i] = validatedFlag(t, strings.Clone(query))
			}
			trimmed := (&(*flags[0].Rules)[0]).GetTrimmedQuery()
			assert.Equal(t, tt.wantTextRefs, refCount(c, trimmed))
			assert.Equal(t, tt.wantKeyRefs, c.keyRefsTotal())

			first := indexedEvaluator(c, &(*flags[0].Rules)[0])
			require.NotNil(t, first)
			for _, f := range flags {
				r := &(*f.Rules)[0]
				assert.Same(t, first, indexedEvaluator(c, r), "one evaluator for the query")
				assert.True(t, applies(r, "user-0000199"))
				assert.False(t, applies(r, "nobody"))
			}

			runtime.KeepAlive(flags)
			clear(flags)
			gcUntil(t, func() bool { return c.keyRefsTotal() == 0 && refCount(c, trimmed) == 0 },
				"every reference must be released once the flags are collected")
			assert.Zero(t, indexSize(c))
		})
	}
}

func TestNikunjyEvaluatorCache_KnownQueryOfAnInvalidFlag(t *testing.T) {
	query := idList(200, -1, "")
	tests := []struct {
		name string
		// invalid makes the flag that carries the known query fail its validation.
		invalid func(f *InternalFlag)
	}{
		{
			name:    "a rule of the flag without variation",
			invalid: func(f *InternalFlag) { *f.Rules = append(*f.Rules, Rule{Query: testconvert.String(`key eq "x"`)}) },
		},
		{
			name:    "a rule of the flag with a query that does not parse",
			invalid: func(f *InternalFlag) { *f.Rules = append(*f.Rules, *newFlagWithQueries(`key eq`).Rules...) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := useIndexingNikunjyEvaluators(t)
			valid := validatedFlag(t, strings.Clone(query))
			require.Equal(t, 1, c.keyRefsTotal())

			invalid := newFlagWithQueries(strings.Clone(query))
			tt.invalid(invalid)
			require.Error(t, invalid.IsValid())
			assert.Equal(t, 2, c.keyRefsTotal(), "the reference taken by the invalid flag is released with it")

			runtime.KeepAlive(invalid)
			*invalid = InternalFlag{} // drop the rules
			gcUntil(t, func() bool { return c.keyRefsTotal() == 1 }, "the invalid flag must release its reference once")
			assert.True(t, applies(&(*valid.Rules)[0], "user-0000199"), "the query is still evaluated")
			runtime.KeepAlive(valid)
		})
	}
}
