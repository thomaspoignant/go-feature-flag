package flag

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/nikunjy/rules/parser"
)

// nikunjyEvaluatorCache keeps one parsed *parser.Evaluator per query string. Parsing the
// query via ANTLR is the dominant cost of a nikunjy rule evaluation. An evaluator compiles
// its rule on the first Process and is then safe for concurrent use, so one evaluator per
// query serves every evaluation, and the garbage collector never drops it while the query
// is cached.
//
// An entry is evicted once no validated rule list contains its query. refs counts those
// rule lists per query: the cache is shared by every client of the process (e.g. every
// flag set of the relay proxy), so one client's reload must not evict a query another
// client still uses.
type nikunjyEvaluatorCache struct {
	entries sync.Map // map[string]*parser.Evaluator
	refsMu  sync.Mutex
	refs    map[string]int
}

func newNikunjyEvaluatorCache() *nikunjyEvaluatorCache {
	return &nikunjyEvaluatorCache{refs: map[string]int{}}
}

var nikunjyEvaluators = newNikunjyEvaluatorCache()

func (c *nikunjyEvaluatorCache) get(query string) (*parser.Evaluator, error) {
	if v, ok := c.entries.Load(query); ok {
		return v.(*parser.Evaluator), nil
	}
	ev, err := parser.NewEvaluator(query)
	if err != nil {
		return nil, err
	}
	actual, _ := c.entries.LoadOrStore(query, ev)
	return actual.(*parser.Evaluator), nil
}

// validate parses the query through the cache and evaluates it on an empty context. That
// first Process also compiles the rule, so its first evaluation does not pay for it.
func (c *nikunjyEvaluatorCache) validate(query string) error {
	ev, err := c.get(query)
	if err != nil {
		return fmt.Errorf("invalid query: %w", err)
	}
	if _, err := ev.Process(map[string]interface{}{}); err != nil {
		return fmt.Errorf("invalid query: %w", err)
	}
	return nil
}

// retain keeps the entries of the nikunjy queries in rules until rules is garbage
// collected. After a configuration reload, the rules of the previous configuration
// become unreachable, and the queries that no remaining rule list uses are evicted.
func (c *nikunjyEvaluatorCache) retain(rules *[]Rule) {
	if rules == nil {
		return
	}
	queries := make([]string, 0, len(*rules))
	for i := range *rules {
		if r := &(*rules)[i]; r.GetQueryFormat() == NikunjyQueryFormat {
			queries = append(queries, r.GetTrimmedQuery())
		}
	}
	if len(queries) == 0 {
		return
	}
	c.refsMu.Lock()
	for _, q := range queries {
		c.refs[q]++
	}
	c.refsMu.Unlock()
	runtime.AddCleanup(rules, c.release, queries)
}

func (c *nikunjyEvaluatorCache) release(queries []string) {
	c.refsMu.Lock()
	defer c.refsMu.Unlock()
	for _, q := range queries {
		c.refs[q]--
		if c.refs[q] == 0 {
			delete(c.refs, q)
			c.entries.Delete(q)
		}
	}
}
