package flag_test

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

// largeInQuery returns `key in ["user-0000000", ...]` with 300000 ids, a ~4.5 MB query.
func largeInQuery() string {
	return largeInQueryVariant(300000, "")
}

// largeInQueryVariant is largeInQuery with a prefix in its ids, so that variants differ.
func largeInQueryVariant(n int, prefix string) string {
	var sb strings.Builder
	sb.Grow(n*16 + 16)
	sb.WriteString(`key in [`)
	for i := 0; i < n; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, `"%suser-%07d"`, prefix, i)
	}
	sb.WriteString(`]`)
	return sb.String()
}

// newValidatedFlag builds a flag with one targeting rule and validates it, as the flag
// cache does when it loads a configuration.
func newValidatedFlag(b *testing.B, query string) *flag.InternalFlag {
	b.Helper()
	f := &flag.InternalFlag{
		Variations: &map[string]*any{"on": testconvert.Interface(true), "off": testconvert.Interface(false)},
		Rules: &[]flag.Rule{{
			Name:            testconvert.String("targeting"),
			Query:           testconvert.String(query),
			VariationResult: testconvert.String("on"),
		}},
		DefaultRule: &flag.Rule{VariationResult: testconvert.String("off")},
	}
	if err := f.IsValid(); err != nil {
		b.Fatal(err)
	}
	return f
}

var (
	shared100Once  sync.Once
	shared100Flags []*flag.InternalFlag
)

// sharedBy100Flags returns 100 validated flags that carry the same 300000 ids list, each with
// its own copy of the text, as every flag parsed from a configuration file does. They are
// built once per process: loading them parses the list for every flag on some versions.
func sharedBy100Flags(b *testing.B) []*flag.InternalFlag {
	b.Helper()
	shared100Once.Do(func() {
		query := largeInQuery()
		shared100Flags = make([]*flag.InternalFlag, 100)
		for i := range shared100Flags {
			shared100Flags[i] = newValidatedFlag(b, strings.Clone(query))
		}
	})
	return shared100Flags
}

func benchValue(b *testing.B, flags []*flag.InternalFlag, key string) {
	ctx := ffcontext.NewEvaluationContext(key)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flags[i%len(flags)].Value("flag", ctx, flag.Context{})
	}
}

// BenchmarkInternalFlagValue_LargeIn300kMiss evaluates a flag whose single rule lists 300000
// ids (a ~4.5 MB query) against a context that is not in the list.
func BenchmarkInternalFlagValue_LargeIn300kMiss(b *testing.B) {
	benchValue(b, []*flag.InternalFlag{newValidatedFlag(b, largeInQuery())}, "someone-else")
}

// BenchmarkInternalFlagValue_LargeIn300kHit is the same with the last id of the list.
func BenchmarkInternalFlagValue_LargeIn300kHit(b *testing.B) {
	benchValue(b, []*flag.InternalFlag{newValidatedFlag(b, largeInQuery())}, "user-0299999")
}

// BenchmarkInternalFlagValue_LargeIn300kSharedBy100Flags evaluates 100 flags that carry
// the same 300000 ids list, round-robin. Every flag owns its copy of the query text, as
// every flag parsed from a configuration file does.
func BenchmarkInternalFlagValue_LargeIn300kSharedBy100Flags(b *testing.B) {
	benchValue(b, sharedBy100Flags(b), "someone-else")
}

// BenchmarkInternalFlagValue_LargeIn300kDifferentIn10Flags evaluates 10 flags that carry 10
// different 300000 ids lists, round-robin (100 lists need too much memory).
func BenchmarkInternalFlagValue_LargeIn300kDifferentIn10Flags(b *testing.B) {
	flags := make([]*flag.InternalFlag, 10)
	for i := range flags {
		flags[i] = newValidatedFlag(b, largeInQueryVariant(300000, fmt.Sprintf("list%d-", i)))
	}
	benchValue(b, flags, "someone-else")
}

// BenchmarkInternalFlagValue_LargeIn300kParallel evaluates 100 flags sharing a 300000 ids
// list from all the CPUs.
func BenchmarkInternalFlagValue_LargeIn300kParallel(b *testing.B) {
	flags := sharedBy100Flags(b)
	ctx := ffcontext.NewEvaluationContext("someone-else")
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			flags[i%len(flags)].Value("flag", ctx, flag.Context{})
			i++
		}
	})
}

// benchReload is the time to load a configuration of 100 flags that carry the same list of ids
// while the previous configuration is still loaded.
func benchReload(b *testing.B, ids int) {
	query := largeInQueryVariant(ids, "")
	load := func() []*flag.InternalFlag {
		flags := make([]*flag.InternalFlag, 100)
		for i := range flags {
			flags[i] = newValidatedFlag(b, strings.Clone(query))
		}
		return flags
	}
	previous := load()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// The loader builds the flags from the file before validating them.
		b.StopTimer()
		next := make([]*flag.InternalFlag, 100)
		for j := range next {
			next[j] = &flag.InternalFlag{
				Variations: &map[string]*any{"on": testconvert.Interface(true), "off": testconvert.Interface(false)},
				Rules: &[]flag.Rule{{
					Name:            testconvert.String("targeting"),
					Query:           testconvert.String(strings.Clone(query)),
					VariationResult: testconvert.String("on"),
				}},
				DefaultRule: &flag.Rule{VariationResult: testconvert.String("off")},
			}
		}
		b.StartTimer()
		for _, f := range next {
			if err := f.IsValid(); err != nil {
				b.Fatal(err)
			}
		}
		previous = next
	}
	_ = previous
}

func BenchmarkInternalFlagIsValid_Reload100FlagsSharingLargeIn(b *testing.B) { benchReload(b, 300000) }
func BenchmarkInternalFlagIsValid_Reload100FlagsSharing20kIn(b *testing.B)   { benchReload(b, 20000) }

// BenchmarkInternalFlagValue_SmallQuery is the baseline that a query lookup change must
// not regress.
func BenchmarkInternalFlagValue_SmallQuery(b *testing.B) {
	f := newValidatedFlag(b, `key eq "user-1" and plan eq "pro"`)
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").AddCustom("plan", "pro").Build()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Value("flag", ctx, flag.Context{})
	}
}

// BenchmarkInternalFlagValue_LargeIn300kGCEvery10 is BenchmarkInternalFlagValue_LargeIn300kMiss
// with two garbage collections, which drop the evaluators held by a sync.Pool, every 10
// evaluations. The collections are not timed, what follows them is.
func BenchmarkInternalFlagValue_LargeIn300kGCEvery10(b *testing.B) {
	f := newValidatedFlag(b, largeInQuery())
	ctx := ffcontext.NewEvaluationContext("someone-else")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%10 == 0 {
			b.StopTimer()
			runtime.GC()
			runtime.GC()
			b.StartTimer()
		}
		f.Value("flag", ctx, flag.Context{})
	}
}

func newSmallFlag(i int, query string) *flag.InternalFlag {
	f := &flag.InternalFlag{
		Variations:  &map[string]*any{"on": testconvert.Interface(true), "off": testconvert.Interface(false)},
		DefaultRule: &flag.Rule{VariationResult: testconvert.String("off")},
	}
	if query != "" {
		f.Rules = &[]flag.Rule{{
			Name:            testconvert.String(fmt.Sprintf("rule-%d", i)),
			Query:           testconvert.String(query),
			VariationResult: testconvert.String("on"),
		}}
	}
	return f
}

func smallQueries(n int) []string {
	queries := make([]string, n)
	for i := range queries {
		queries[i] = fmt.Sprintf(`key eq "user-%d" and plan eq "pro"`, i)
	}
	return queries
}

// benchLoad measures IsValid of n flags that each carry a distinct short query, the shape of a
// typical configuration file. The flags are built outside the measure.
func benchLoad(b *testing.B, n int) {
	queries := smallQueries(n)
	var keep []*flag.InternalFlag
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		flags := make([]*flag.InternalFlag, n)
		for j := range flags {
			flags[j] = newSmallFlag(j, strings.Clone(queries[j]))
		}
		b.StartTimer()
		for _, f := range flags {
			if err := f.IsValid(); err != nil {
				b.Fatal(err)
			}
		}
		keep = flags
	}
	_ = keep
}

func BenchmarkInternalFlagIsValid_Load1000SmallFlags(b *testing.B)  { benchLoad(b, 1000) }
func BenchmarkInternalFlagIsValid_Load10000SmallFlags(b *testing.B) { benchLoad(b, 10000) }

// BenchmarkInternalFlagValue_Small10000DistinctFlags evaluates 10000 validated flags with
// distinct short queries, round-robin.
func BenchmarkInternalFlagValue_Small10000DistinctFlags(b *testing.B) {
	queries := smallQueries(10000)
	flags := make([]*flag.InternalFlag, len(queries))
	for i := range flags {
		flags[i] = newSmallFlag(i, queries[i])
		if err := flags[i].IsValid(); err != nil {
			b.Fatal(err)
		}
	}
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").AddCustom("plan", "pro").Build()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flags[i%len(flags)].Value("flag", ctx, flag.Context{})
	}
}

func BenchmarkInternalFlagValue_SmallQueryParallel(b *testing.B) {
	f := newSmallFlag(0, `key eq "user-1" and plan eq "pro"`)
	if err := f.IsValid(); err != nil {
		b.Fatal(err)
	}
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").AddCustom("plan", "pro").Build()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f.Value("flag", ctx, flag.Context{})
		}
	})
}

func BenchmarkInternalFlagValue_NoQuery(b *testing.B) {
	f := newSmallFlag(0, "")
	if err := f.IsValid(); err != nil {
		b.Fatal(err)
	}
	ctx := ffcontext.NewEvaluationContext("user-1")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Value("flag", ctx, flag.Context{})
	}
}

func BenchmarkInternalFlagValue_JSONLogic(b *testing.B) {
	f := newSmallFlag(0, `{"and": [{"==": [{"var": "key"}, "user-1"]}, {"==": [{"var": "plan"}, "pro"]}]}`)
	if err := f.IsValid(); err != nil {
		b.Fatal(err)
	}
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").AddCustom("plan", "pro").Build()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Value("flag", ctx, flag.Context{})
	}
}

// BenchmarkRuleEvaluate_NotValidated evaluates a rule that was never validated, as a rule built
// in code is.
func BenchmarkRuleEvaluate_NotValidated(b *testing.B) {
	r := flag.Rule{
		Query:           testconvert.String(`key eq "user-1" and plan eq "pro"`),
		VariationResult: testconvert.String("on"),
	}
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").AddCustom("plan", "pro").Build()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Evaluate("user-1", ctx, "flag", false)
	}
}

// BenchmarkInternalFlagValue_Scheduled evaluates a flag with a scheduled rollout, which copies the
// flag through JSON on every evaluation.
func BenchmarkInternalFlagValue_Scheduled(b *testing.B) {
	f := newSmallFlag(0, `key eq "user-1" and plan eq "pro"`)
	past := time.Now().Add(-time.Hour)
	f.Scheduled = &[]flag.ScheduledStep{{
		InternalFlag: flag.InternalFlag{Rules: &[]flag.Rule{{
			Name:  testconvert.String("rule-0"),
			Query: testconvert.String(`key eq "user-1" and plan eq "pro"`),
		}}},
		Date: &past,
	}}
	if err := f.IsValid(); err != nil {
		b.Fatal(err)
	}
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").AddCustom("plan", "pro").Build()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Value("flag", ctx, flag.Context{})
	}
}

func heapInUse() float64 {
	runtime.GC()
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.HeapAlloc)
}

// BenchmarkHeap_AfterLoadAndReloads reports the heap retained by a loaded configuration of 1000
// flags with distinct short queries, each evaluated once (heap-load-B), and after 30 reloads that
// change every query (heap-reloads-B).
func BenchmarkHeap_AfterLoadAndReloads(b *testing.B) {
	ctx := ffcontext.NewEvaluationContextBuilder("user-1").AddCustom("plan", "pro").Build()
	for i := 0; i < b.N; i++ {
		base := heapInUse()
		load := func(generation int) []*flag.InternalFlag {
			flags := make([]*flag.InternalFlag, 1000)
			for j := range flags {
				flags[j] = newSmallFlag(j, fmt.Sprintf(`key eq "user-%d-%d" and plan eq "pro"`, generation, j))
				if err := flags[j].IsValid(); err != nil {
					b.Fatal(err)
				}
				flags[j].Value("flag", ctx, flag.Context{})
			}
			return flags
		}
		current := load(0)
		afterLoad := heapInUse() - base
		for g := 1; g <= 30; g++ {
			current = load(g)
		}
		afterReloads := heapInUse() - base
		b.ReportMetric(afterLoad, "heap-load-B")
		b.ReportMetric(afterReloads, "heap-reloads-B")
		runtime.KeepAlive(current)
	}
}

// BenchmarkHeap_LargeListReloads reports the heap retained after 8 reloads of 5 flags that carry
// a 20000 ids list, each reload with another list, every flag evaluated once (heap-B).
func BenchmarkHeap_LargeListReloads(b *testing.B) {
	ctx := ffcontext.NewEvaluationContext("someone-else")
	for i := 0; i < b.N; i++ {
		base := heapInUse()
		var current []*flag.InternalFlag
		for g := 0; g < 8; g++ {
			query := largeInQueryVariant(20000, fmt.Sprintf("gen%d-", g))
			current = make([]*flag.InternalFlag, 5)
			for j := range current {
				current[j] = newValidatedFlag(b, strings.Clone(query))
				current[j].Value("flag", ctx, flag.Context{})
			}
		}
		b.ReportMetric(heapInUse()-base, "heap-B")
		runtime.KeepAlive(current)
	}
}
