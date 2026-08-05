// The scanners below are hand-rolled on purpose. No maintained Go module
// implements non-recursive JSON depth limiting: encoding/json's decode stage is
// itself recursive so it cannot guard itself, json/v2's WithDepthLimit is still
// unimplemented (golang/go#56733) and unreachable from TinyGo anyway, and
// fastjson's MaxDepth is hard-coded inside a recursive-descent parser. The ANTLR
// lexer that nikunjy/rules exports would tokenize a query iteratively, but it
// would run on every evaluation while the parse it guards is memoized per unique
// query — the guard would cost more than what it protects.

package main

import (
	"fmt"
	"strings"

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

	// maxQueryAttrPathSegments is the maximum number of '.'-separated segments
	// accepted in a single attribute path (`a.b.c...`) of a nikunjy targeting
	// query. The grammar is indirectly right-recursive — attrPath : ATTRNAME
	// subAttr? and subAttr : '.' attrPath — and ANTLR only rewrites *direct
	// left* recursion, so the generated parser recurses once per segment. A
	// dotted path carries no brackets, no commas and no logical operator, so
	// it is a single word to every other guard here: depth 0, 0 list items,
	// 1 condition. Measured on the built 1MB-stack .wasi: 601 segments
	// evaluate, 602 trap with "memory access out of bounds" — a lower
	// threshold than any other budget above. 128 keeps ~4.7x safety margin,
	// matches the input nesting budget, and is ~20x the deepest attribute
	// path used anywhere in this repo.
	maxQueryAttrPathSegments = 128
)

// nikunjyQueryScan holds complexity metrics collected in a single pass over a
// nikunjy targeting query.
type nikunjyQueryScan struct {
	maxDepth        int
	maxListItems    int
	conditionCount  int
	maxPathSegments int
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

// scanJSONMaxArrayLength returns the largest number of items in any [...]
// array of a JSON document (commas + 1, with empty arrays counting as 1).
// JSONLogic `in` lists and `and`/`or` operand arrays are flat JSON arrays,
// so this single metric covers both the list-item and condition-count budgets.
func scanJSONMaxArrayLength(input string) int {
	listCounts := []int{}
	largestList := 0
	closeList := func() {
		if len(listCounts) == 0 {
			return
		}
		n := listCounts[len(listCounts)-1] + 1
		listCounts = listCounts[:len(listCounts)-1]
		if n > largestList {
			largestList = n
		}
	}
	scanner := unquotedScanner{input: input}
	for _, c, ok := scanner.next(); ok; _, c, ok = scanner.next() {
		switch c {
		case '[':
			listCounts = append(listCounts, 0)
		case ']':
			closeList()
		case ',':
			if len(listCounts) > 0 {
				listCounts[len(listCounts)-1]++
			}
		}
	}
	for len(listCounts) > 0 {
		closeList()
	}
	return largestList
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

// nikunjyScanState accumulates the metrics of a nikunjy query while it is
// walked byte by byte. The per-byte work lives in methods so that the scan
// loop itself stays flat and each rule is testable in isolation.
type nikunjyScanState struct {
	query       string
	depth       int
	maxDepth    int
	listCounts  []int
	largestList int
	operators   int
	wordStart   int
	longestPath int
}

// pushDepth records entering a ( / [ / { nesting level.
func (s *nikunjyScanState) pushDepth() {
	s.depth++
	if s.depth > s.maxDepth {
		s.maxDepth = s.depth
	}
}

// popDepth records leaving a ) / ] / } nesting level.
func (s *nikunjyScanState) popDepth() {
	s.depth--
}

// openList starts counting the items of a new [...] list.
func (s *nikunjyScanState) openList() {
	s.listCounts = append(s.listCounts, 0)
}

// closeList folds the innermost open list into the running maximum. It is a
// no-op when no list is open, so a stray ']' cannot underflow the stack.
func (s *nikunjyScanState) closeList() {
	if len(s.listCounts) == 0 {
		return
	}
	n := s.listCounts[len(s.listCounts)-1] + 1 // items = commas + 1
	s.listCounts = s.listCounts[:len(s.listCounts)-1]
	if n > s.largestList {
		s.largestList = n
	}
}

// countItem attributes a comma to the innermost open list.
func (s *nikunjyScanState) countItem() {
	if len(s.listCounts) > 0 {
		s.listCounts[len(s.listCounts)-1]++
	}
}

// startWord marks the beginning of an identifier-like token.
func (s *nikunjyScanState) startWord(i int) {
	if s.wordStart < 0 {
		s.wordStart = i
	}
}

// endWord closes the token opened by startWord, counts it when it is a logical
// operator, and folds its attribute-path length into the running maximum.
func (s *nikunjyScanState) endWord(end int) {
	if s.wordStart < 0 {
		return
	}
	w := s.query[s.wordStart:end]
	if w == "and" || w == "or" {
		s.operators++
	}
	// Every word is measured, not just attribute paths: telling a path apart
	// from a dotted literal would need the grammar, and over-counting only
	// errs towards rejecting, which is the safe direction here.
	if n := strings.Count(w, ".") + 1; n > s.longestPath {
		s.longestPath = n
	}
	s.wordStart = -1
}

// scanNikunjyQuery collects nesting depth, largest list size, condition count
// and longest attribute path in one pass over a nikunjy targeting query.
func scanNikunjyQuery(query string) nikunjyQueryScan {
	state := nikunjyScanState{query: query, wordStart: -1}
	scanner := unquotedScanner{input: query}
	for i, c, ok := scanner.next(); ok; i, c, ok = scanner.next() {
		switch c {
		case '"':
			state.endWord(i)
		case '{', '(':
			state.pushDepth()
		case '[':
			state.pushDepth()
			state.openList()
		case '}', ')':
			state.popDepth()
		case ']':
			state.popDepth()
			state.closeList()
		case ',':
			state.countItem()
		default:
			if isQueryWordChar(c) {
				state.startWord(i)
			} else {
				state.endWord(i)
			}
		}
	}
	// Fold any list left open by an unbalanced query so its width still counts.
	for len(state.listCounts) > 0 {
		state.closeList()
	}
	state.endWord(len(query))
	return nikunjyQueryScan{
		maxDepth:        state.maxDepth,
		maxListItems:    state.largestList,
		conditionCount:  state.operators + 1,
		maxPathSegments: state.longestPath,
	}
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
//
// Queries are handed to visit through GetTrimmedQuery, which is the exact
// string the evaluator parses. Scanning the raw query instead would leave a
// bypass: StrTrim splits on '\n' and rejoins with the empty string, so it
// merges tokens across line breaks — "a eq 1 o\nr b eq 2" scans as the two
// words "o" and "r" but parses as a real or-chain.
func walkRuleQueries(f *flag.InternalFlag, visit func(query string) bool) bool {
	if f == nil {
		return false
	}
	checkRule := func(r *flag.Rule) bool {
		return r != nil && r.Query != nil && visit(r.GetTrimmedQuery())
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
			detail, over = jsonLogicQueryViolation(q)
		} else {
			detail, over = nikunjyQueryViolation(q)
		}
		return over
	})
	return detail, over
}

// jsonLogicQueryViolation checks a JSONLogic targeting query against its
// complexity budget, returning a formatted error detail when exceeded.
func jsonLogicQueryViolation(q string) (detail string, over bool) {
	if depth := scanJSONDepth(q); depth > maxJSONLogicQueryNestingDepth {
		return fmt.Sprintf(
			"targeting query exceeds the maximum nesting depth (%d > %d)",
			depth, maxJSONLogicQueryNestingDepth), true
	}
	if n := scanJSONMaxArrayLength(q); n > maxQueryListItems {
		return fmt.Sprintf(
			"targeting query list exceeds the maximum item count (%d > %d)",
			n, maxQueryListItems), true
	}
	return "", false
}

// nikunjyQueryViolation checks a nikunjy targeting query against its
// complexity budget, returning a formatted error detail when exceeded.
func nikunjyQueryViolation(q string) (detail string, over bool) {
	scan := scanNikunjyQuery(q)
	if scan.maxDepth > maxQueryNestingDepth {
		return fmt.Sprintf(
			"targeting query exceeds the maximum nesting depth (%d > %d)",
			scan.maxDepth, maxQueryNestingDepth), true
	}
	if scan.maxListItems > maxQueryListItems {
		return fmt.Sprintf(
			"targeting query list exceeds the maximum item count (%d > %d)",
			scan.maxListItems, maxQueryListItems), true
	}
	if scan.conditionCount > maxQueryConditions {
		return fmt.Sprintf(
			"targeting query exceeds the maximum condition count (%d > %d)",
			scan.conditionCount, maxQueryConditions), true
	}
	if scan.maxPathSegments > maxQueryAttrPathSegments {
		return fmt.Sprintf(
			"targeting query attribute path exceeds the maximum segment count (%d > %d)",
			scan.maxPathSegments, maxQueryAttrPathSegments), true
	}
	return "", false
}
