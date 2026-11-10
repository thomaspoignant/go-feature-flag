package flag

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/nikunjy/rules/parser"
)

// nikunjyEvaluatorCache memoizes parsed *parser.Evaluator values per query string.
// Parsing the query via ANTLR is the dominant cost of a nikunjy rule evaluation
// (see https://github.com/nikunjy/rules/blob/master/parser/evaluate.go); the parsed
// Evaluator is reusable, but parser.Evaluator.Process writes to an internal field, so
// concurrent calls require one evaluator per call in flight.
//
// Memory note: cache entries are not evicted, since a feature flag's query strings
// are part of static config and the set is bounded in practice.
var nikunjyEvaluatorCache sync.Map // map[string]*pooledNikunjyEvaluator

// The pool's New is wired in newPooledNikunjyEvaluator and is expected to return non-nil
// values since the query was already parsed once during pool initialization.
var errNilNikunjyEvaluator = errors.New("nikunjy evaluator pool returned nil")

// pooledNikunjyEvaluator hands out evaluators for one query. From the first evaluation
// on, one evaluator is pinned: the garbage collector never drops it, so a query evaluated
// less often than the GC runs is not parsed again on every evaluation, as it would be if
// every evaluator lived in the sync.Pool. Concurrent evaluations of the same query take
// more evaluators from the sync.Pool, which keeps that path free of shared locks.
//
// Validation never pins. Every configuration load validates every rule, including the
// rules of flags a process never evaluates, and pinning those would keep a parse of every
// query in memory. Validation uses the sync.Pool, where the first evaluation can pick its
// parse up, and the entry remembers that the query is valid, so later loads do not parse
// it again.
type pooledNikunjyEvaluator struct {
	pinned atomic.Pointer[parser.Evaluator]
	valid  atomic.Bool
	pool   sync.Pool
}

func newPooledNikunjyEvaluator(query string) (*pooledNikunjyEvaluator, error) {
	// Validate the query parses now so callers get the parse error instead of having
	// it surface inside pool.New (where errors cannot be returned).
	first, err := parser.NewEvaluator(query)
	if err != nil {
		return nil, err
	}
	p := &pooledNikunjyEvaluator{}
	p.pool.New = func() interface{} {
		ev, _ := parser.NewEvaluator(query)
		return ev
	}
	p.pool.Put(first)
	return p, nil
}

func (p *pooledNikunjyEvaluator) process(items map[string]interface{}) (bool, error) {
	ev := p.get()
	if ev == nil {
		return false, errNilNikunjyEvaluator
	}
	defer p.put(ev)
	return ev.Process(items)
}

// validate checks that the query evaluates on an empty context. The result never changes
// for a query, so it only runs until the query passes once.
func (p *pooledNikunjyEvaluator) validate() error {
	if p.valid.Load() {
		return nil
	}
	ev, _ := p.pool.Get().(*parser.Evaluator)
	if ev == nil {
		return errNilNikunjyEvaluator
	}
	defer p.pool.Put(ev)
	if _, err := ev.Process(map[string]interface{}{}); err != nil {
		return err
	}
	p.valid.Store(true)
	return nil
}

// get takes the pinned evaluator, or one from the pool when none is pinned or the pinned
// one is in use.
func (p *pooledNikunjyEvaluator) get() *parser.Evaluator {
	if ev := p.pinned.Swap(nil); ev != nil {
		return ev
	}
	ev, _ := p.pool.Get().(*parser.Evaluator)
	return ev
}

// put pins ev when no evaluator is pinned, and returns it to the pool otherwise.
func (p *pooledNikunjyEvaluator) put(ev *parser.Evaluator) {
	if !p.pinned.CompareAndSwap(nil, ev) {
		p.pool.Put(ev)
	}
}

func getNikunjyEvaluator(query string) (*pooledNikunjyEvaluator, error) {
	if v, ok := nikunjyEvaluatorCache.Load(query); ok {
		return v.(*pooledNikunjyEvaluator), nil
	}
	p, err := newPooledNikunjyEvaluator(query)
	if err != nil {
		return nil, err
	}
	actual, _ := nikunjyEvaluatorCache.LoadOrStore(query, p)
	return actual.(*pooledNikunjyEvaluator), nil
}
