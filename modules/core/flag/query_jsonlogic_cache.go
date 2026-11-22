package flag

import (
	"encoding/json"
	"math"
	"sync"
	"unicode/utf8"

	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

// compiledQueryCache memoizes the trimmed form, format and decoded JsonLogic rule per raw query.
// Only evaluation stores entries, so the set is bounded by the evaluated flag config.
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
	trimmed := utils.StrTrim(raw)
	c := &compiledQuery{trimmed: trimmed, format: queryFormatOf(trimmed)}
	if c.format == JSONLogicQueryFormat {
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
	if err != nil {
		return false
	}
	switch string(b) {
	case "false", "null", `""`, "0", "[]":
		return false
	}
	return true
}

// asJSONData returns v as a JSON round trip would, or false when only a real round trip can tell.
func asJSONData(v any) (any, bool) {
	switch val := v.(type) {
	case nil, bool:
		return val, true
	case string:
		return val, utf8.ValidString(val)
	case float64:
		return val, !math.IsNaN(val) && !math.IsInf(val, 0)
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case map[string]any:
		if val == nil {
			return nil, true
		}
		out := make(map[string]any, len(val))
		for k, e := range val {
			if !utf8.ValidString(k) {
				return nil, false
			}
			d, ok := asJSONData(e)
			if !ok {
				return nil, false
			}
			out[k] = d
		}
		return out, true
	case []any:
		if val == nil {
			return nil, true
		}
		out := make([]any, len(val))
		for i, e := range val {
			d, ok := asJSONData(e)
			if !ok {
				return nil, false
			}
			out[i] = d
		}
		return out, true
	}
	return nil, false
}
