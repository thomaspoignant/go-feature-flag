package flag

import (
	"fmt"
	"sync"

	"github.com/nikunjy/rules/parser"
)

// nikunjyEvaluatorCache memoizes a sync.Pool of *parser.Evaluator per query string.
// Parsing the query via ANTLR is the dominant cost of a nikunjy rule evaluation
// (see https://github.com/nikunjy/rules/blob/master/parser/evaluate.go); the parsed
// Evaluator is reusable, but parser.Evaluator.Process writes to an internal field, so
// concurrent calls require either serialization or a pool of evaluators per query.
// We use sync.Pool because evaluations are typically high-throughput and short-lived.
//
// Memory note: cache entries are not evicted, since a feature flag's query strings
// are part of static config and the set is bounded in practice.
var nikunjyEvaluatorCache sync.Map // map[string]*pooledNikunjyEvaluator

type pooledNikunjyEvaluator struct {
	pool sync.Pool
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
	ev, _ := p.pool.Get().(*parser.Evaluator)
	if ev == nil {
		// The pool's New is wired in newPooledNikunjyEvaluator and is expected to return
		// non-nil values since the query was already validated during pool initialization.
		return false, fmt.Errorf("nikunjy evaluator pool returned nil")
	}
	defer p.pool.Put(ev)
	return ev.Process(items)
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
