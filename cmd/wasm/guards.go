package main

import (
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

const (
	// maxInputNestingDepth is the maximum {} / [] nesting depth accepted for
	// the input JSON document. encoding/json decodes recursively, and the
	// module runs on a fixed-size shadow stack: unbounded nesting overflows
	// it and traps the instance, which permanently poisons it (the trap does
	// not unwind the stack pointer). See README.md.
	maxInputNestingDepth = 128

	// maxQueryNestingDepth is the maximum ( ) / [ ] / { } nesting depth
	// accepted inside a nikunjy targeting query. The query parser is
	// recursive and consumes far more stack per nesting level than the JSON
	// decoder: on a 64KB stack ~30 nested parentheses were enough to trap
	// the module.
	maxQueryNestingDepth = 64

	// maxJSONLogicQueryNestingDepth is the budget for JSONLogic queries,
	// which are bracket-heavy JSON documents: a single comparison like
	// {"==":[{"var":"a"},1]} already costs ~5 bracket levels, so ~13 nested
	// logical operators would hit the nikunjy limit while being nowhere near
	// stack-overflow territory. JSONLogic decoding costs roughly as much
	// stack per level as the JSON input decoder, so this budget stays far
	// below the input guard's safety margin.
	maxJSONLogicQueryNestingDepth = 256

	// maxQueryListItems is the maximum number of items accepted in a single
	// [...] list of a nikunjy targeting query. List parsing is right-recursive
	// (subListOfInts : INT COMMA subListOfInts), so each item costs a parser
	// stack frame (~356 bytes) regardless of bracket nesting — a flat list is
	// invisible to the nesting guards above. Measured: 154 items overflow a
	// 64KB stack (the ~200-item allow-list of issue #5651 trapped production)
	// and 2,947 items overflow the 1MB stack, identically for int, double and
	// string lists. 1000 keeps ~3x safety margin.
	maxQueryListItems = 1000

	// maxQueryConditions is the maximum number of and/or-joined conditions
	// accepted in a single nikunjy targeting query. Logical expressions are
	// binary and recursive (query SP LOGICAL_OPERATOR SP query), so a flat
	// bracket-less chain (`a eq 1 or b eq 2 or ...`) costs parser stack per
	// operator while having neither the brackets nor the list commas the
	// guards above look at. Measured: 341 conditions overflow a 64KB stack
	// and 3,266 the 1MB stack. 1000 keeps ~3x safety margin.
	maxQueryConditions = 1000
)

// maxBracketDepth returns the maximum nesting depth of brackets in s,
// ignoring characters inside double-quoted string literals.
// It counts {} and []; parentheses are included when countParens is true.
func maxBracketDepth(s string, countParens bool) int {
	depth, deepest := 0, 0
	inString, escaped := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{', '[':
			depth++
			if depth > deepest {
				deepest = depth
			}
		case '}', ']':
			depth--
		case '(':
			if countParens {
				depth++
				if depth > deepest {
					deepest = depth
				}
			}
		case ')':
			if countParens {
				depth--
			}
		}
	}
	return deepest
}

// jsonNestingDepth returns the maximum {} / [] nesting depth of a JSON document.
func jsonNestingDepth(input string) int {
	return maxBracketDepth(input, false)
}

// queryNestingDepth returns the maximum bracket nesting depth of a targeting
// query (nikunjy expression or JSONLogic document).
func queryNestingDepth(query string) int {
	return maxBracketDepth(query, true)
}

// isJSONLogicQuery reports whether the query is a JSONLogic document (first
// non-space byte '{' or '['). A nikunjy expression can never start with one
// of those: its grammar starts with an attribute path, NOT or '('.
func isJSONLogicQuery(query string) bool {
	for i := 0; i < len(query); i++ {
		switch query[i] {
		case ' ', '\t', '\n', '\r':
			continue
		case '{', '[':
			return true
		default:
			return false
		}
	}
	return false
}

// queryNestingLimit returns the nesting budget for one targeting query.
// JSONLogic queries get a larger budget than nikunjy expressions; see the
// constants above.
func queryNestingLimit(query string) int {
	if isJSONLogicQuery(query) {
		return maxJSONLogicQueryNestingDepth
	}
	return maxQueryNestingDepth
}

// maxListItemCount returns the item count of the largest [...] list in a
// query, ignoring characters inside double-quoted string literals. Unclosed
// groups (malformed queries) are still counted so a truncated giant list is
// not waved through.
func maxListItemCount(s string) int {
	var counts []int // one comma counter per open [ group
	largest := 0
	fold := func() {
		n := counts[len(counts)-1] + 1 // items = commas + 1
		counts = counts[:len(counts)-1]
		if n > largest {
			largest = n
		}
	}
	inString, escaped := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '[':
			counts = append(counts, 0)
		case ',':
			if len(counts) > 0 {
				counts[len(counts)-1]++
			}
		case ']':
			if len(counts) > 0 {
				fold()
			}
		}
	}
	for len(counts) > 0 {
		fold()
	}
	return largest
}

// isQueryWordChar reports whether c can be part of an identifier-like token
// of a nikunjy query (ATTRNAME chars plus the '.' of attribute paths). Used
// to detect word boundaries around logical operators.
func isQueryWordChar(c byte) bool {
	return c == '-' || c == '_' || c == ':' || c == '.' ||
		(c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// conditionCount returns the number of and/or-joined conditions of a nikunjy
// query: one more than the number of word-boundary `and` / `or` tokens
// outside double-quoted string literals. The grammar only accepts the
// lowercase forms, always space-delimited, and an attribute can never be
// named `and`/`or` (the lexer claims those tokens first), so every such word
// is a logical operator.
func conditionCount(s string) int {
	operators := 0
	inString, escaped := false, false
	wordStart := -1 // start of the current identifier-like run, -1 when outside one
	endWord := func(end int) {
		if wordStart >= 0 {
			if w := s[wordStart:end]; w == "and" || w == "or" {
				operators++
			}
			wordStart = -1
		}
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}
		switch {
		case c == '"':
			endWord(i)
			inString = true
		case isQueryWordChar(c):
			if wordStart < 0 {
				wordStart = i
			}
		default:
			endWord(i)
		}
	}
	endWord(len(s))
	return operators + 1
}

// walkRuleQueries calls visit for every targeting query of the flag
// (targeting rules, default rule, scheduled rollout steps) and stops at the
// first visit returning true. The recursion over scheduled steps is bounded:
// the input already passed the maxInputNestingDepth guard, and each scheduled
// level costs several JSON nesting levels.
func walkRuleQueries(f *flag.InternalFlag, visit func(query string) bool) bool {
	if f == nil {
		return false
	}
	checkRule := func(r *flag.Rule) bool {
		return r != nil && r.Query != nil && visit(*r.Query)
	}
	if f.Rules != nil {
		for i := range *f.Rules {
			if checkRule(&(*f.Rules)[i]) {
				return true
			}
		}
	}
	if checkRule(f.DefaultRule) {
		return true
	}
	if f.Scheduled != nil {
		for i := range *f.Scheduled {
			if walkRuleQueries(&(*f.Scheduled)[i].InternalFlag, visit) {
				return true
			}
		}
	}
	return false
}

// firstQueryOverLimit reports the first query of the flag whose bracket
// nesting exceeds its format-specific budget.
func firstQueryOverLimit(f *flag.InternalFlag) (depth int, limit int, over bool) {
	walkRuleQueries(f, func(q string) bool {
		if d := queryNestingDepth(q); d > queryNestingLimit(q) {
			depth, limit, over = d, queryNestingLimit(q), true
			return true
		}
		return false
	})
	return depth, limit, over
}

// firstQueryOverBreadth reports the first nikunjy query of the flag with a
// [...] list of more than maxQueryListItems items. JSONLogic queries are
// exempt: encoding/json decodes arrays iteratively, so their length does not
// translate into recursion depth, and JSON documents are comma-heavy by
// nature.
func firstQueryOverBreadth(f *flag.InternalFlag) (items int, limit int, over bool) {
	walkRuleQueries(f, func(q string) bool {
		if isJSONLogicQuery(q) {
			return false
		}
		if n := maxListItemCount(q); n > maxQueryListItems {
			items, limit, over = n, maxQueryListItems, true
			return true
		}
		return false
	})
	return items, limit, over
}

// firstQueryOverConditionCount reports the first nikunjy query of the flag
// with more than maxQueryConditions and/or-joined conditions. JSONLogic
// queries are exempt: their operator names live inside JSON string literals
// and their operands in arrays, which encoding/json decodes iteratively.
func firstQueryOverConditionCount(f *flag.InternalFlag) (conditions int, limit int, over bool) {
	walkRuleQueries(f, func(q string) bool {
		if isJSONLogicQuery(q) {
			return false
		}
		if n := conditionCount(q); n > maxQueryConditions {
			conditions, limit, over = n, maxQueryConditions, true
			return true
		}
		return false
	})
	return conditions, limit, over
}
