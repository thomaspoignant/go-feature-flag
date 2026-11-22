package flag

import (
	"encoding/json"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

// compiledQueryCache memoizes the trimmed form, format and decoded JsonLogic rule per raw query.
// Entries are not evicted: queries come from flag config, so the set is bounded in practice.
var compiledQueryCache sync.Map // map[string]*compiledQuery

type compiledQuery struct {
	trimmed   string
	format    QueryFormat
	jsonLogic any
}

func getCompiledQuery(raw string) *compiledQuery {
	if v, ok := compiledQueryCache.Load(raw); ok {
		return v.(*compiledQuery)
	}
	c := &compiledQuery{trimmed: utils.StrTrim(raw), format: NikunjyQueryFormat}
	if utils.IsJSONObject(c.trimmed) {
		c.format = JSONLogicQueryFormat
		var rule any
		if json.Unmarshal([]byte(c.trimmed), &rule) == nil {
			c.jsonLogic = rule
		}
	}
	actual, _ := compiledQueryCache.LoadOrStore(raw, c)
	return actual.(*compiledQuery)
}

// isJSONLogicTruthy treats false, null, "", 0 and [] as falsy, as per the JsonLogic spec.
func isJSONLogicTruthy(v any) bool {
	switch val := v.(type) {
	case nil:
		return false
	case bool:
		return val
	case string:
		return val != ""
	}
	// Other results are checked through JSON like jsonlogic.Apply did, so NaN and ±Inf are falsy.
	b, err := json.Marshal(v)
	return err == nil && string(b) != "0" && string(b) != "[]"
}
