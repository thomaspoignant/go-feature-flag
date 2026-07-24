package main

import (
	"fmt"

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

// nikunjyQueryScan holds complexity metrics collected in a single pass over a
// nikunjy targeting query.
type nikunjyQueryScan struct {
	maxDepth       int
	maxListItems   int
	conditionCount int
}

// unquotedScanner iterates over bytes outside double-quoted string literals.
// It returns the opening quote so callers can treat it as a token boundary.
type unquotedScanner struct {
	input    string
	offset   int
	inString bool
	escaped  bool
}

func (s *unquotedScanner) next() (index int, c byte, ok bool) {
	for s.offset < len(s.input) {
		index = s.offset
		c = s.input[s.offset]
		s.offset++

		if s.inString {
			switch {
			case s.escaped:
				s.escaped = false
			case c == '\\':
				s.escaped = true
			case c == '"':
				s.inString = false
			}
			continue
		}
		if c == '"' {
			s.inString = true
		}
		return index, c, true
	}
	return 0, 0, false
}

// scanJSONDepth returns the maximum {} / [] nesting depth of a JSON document.
func scanJSONDepth(input string) int {
	depth, maxDepth := 0, 0
	scanner := unquotedScanner{input: input}
	for _, c, ok := scanner.next(); ok; _, c, ok = scanner.next() {
		switch c {
		case '{', '[':
			depth++
			if depth > maxDepth {
				maxDepth = depth
			}
		case '}', ']':
			depth--
		}
	}
	return maxDepth
}

// scanNikunjyQuery collects nesting depth, largest list size, and condition
// count in one pass over a nikunjy targeting query.
func scanNikunjyQuery(query string) nikunjyQueryScan {
	var result nikunjyQueryScan
	depth := 0
	var listCounts []int
	largestList := 0
	operators := 0
	wordStart := -1

	foldList := func() {
		n := listCounts[len(listCounts)-1] + 1 // items = commas + 1
		listCounts = listCounts[:len(listCounts)-1]
		if n > largestList {
			largestList = n
		}
	}
	endWord := func(end int) {
		if wordStart >= 0 {
			if w := query[wordStart:end]; w == "and" || w == "or" {
				operators++
			}
			wordStart = -1
		}
	}

	scanner := unquotedScanner{input: query}
	for i, c, ok := scanner.next(); ok; i, c, ok = scanner.next() {
		switch c {
		case '"':
			endWord(i)
		case '{', '[':
			depth++
			if depth > result.maxDepth {
				result.maxDepth = depth
			}
			if c == '[' {
				listCounts = append(listCounts, 0)
			}
		case '}', ']':
			depth--
			if c == ']' && len(listCounts) > 0 {
				foldList()
			}
		case '(':
			depth++
			if depth > result.maxDepth {
				result.maxDepth = depth
			}
		case ')':
			depth--
		case ',':
			if len(listCounts) > 0 {
				listCounts[len(listCounts)-1]++
			}
		default:
			if isQueryWordChar(c) {
				if wordStart < 0 {
					wordStart = i
				}
			} else {
				endWord(i)
			}
		}
	}
	for len(listCounts) > 0 {
		foldList()
	}
	endWord(len(query))
	result.maxListItems = largestList
	result.conditionCount = operators + 1
	return result
}

// isQueryWordChar reports whether c can be part of an identifier-like token
// of a nikunjy query (ATTRNAME chars plus the '.' of attribute paths). Used
// to detect word boundaries around logical operators.
func isQueryWordChar(c byte) bool {
	return c == '-' || c == '_' || c == ':' || c == '.' ||
		(c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
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

// firstQueryViolation reports the first targeting query of the flag that
// exceeds a complexity budget, with a formatted error detail string.
func firstQueryViolation(f *flag.InternalFlag) (detail string, over bool) {
	walkRuleQueries(f, func(q string) bool {
		if isJSONLogicQuery(q) {
			if depth := scanJSONDepth(q); depth > maxJSONLogicQueryNestingDepth {
				detail = fmt.Sprintf(
					"targeting query exceeds the maximum nesting depth (%d > %d)",
					depth, maxJSONLogicQueryNestingDepth)
				over = true
				return true
			}
			return false
		}

		scan := scanNikunjyQuery(q)
		if scan.maxDepth > maxQueryNestingDepth {
			detail = fmt.Sprintf(
				"targeting query exceeds the maximum nesting depth (%d > %d)",
				scan.maxDepth, maxQueryNestingDepth)
			over = true
			return true
		}
		if scan.maxListItems > maxQueryListItems {
			detail = fmt.Sprintf(
				"targeting query list exceeds the maximum item count (%d > %d)",
				scan.maxListItems, maxQueryListItems)
			over = true
			return true
		}
		if scan.conditionCount > maxQueryConditions {
			detail = fmt.Sprintf(
				"targeting query exceeds the maximum condition count (%d > %d)",
				scan.conditionCount, maxQueryConditions)
			over = true
			return true
		}
		return false
	})
	return detail, over
}
