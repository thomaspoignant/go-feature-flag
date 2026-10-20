package notifier

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

// DescribeChanges returns one sentence per change between two versions of a flag,
// in the order the fields appear in a flag file.
// It describes every field of flag.InternalFlag and returns an error for any other flag type.
func DescribeChanges(before, after flag.Flag) ([]string, error) {
	b, okBefore := before.(*flag.InternalFlag)
	a, okAfter := after.(*flag.InternalFlag)
	if !okBefore || !okAfter || b == nil || a == nil {
		return nil, fmt.Errorf("impossible to describe changes between flags of type %T and %T", before, after)
	}
	d := &describer{}
	d.flag(b, a)
	return d.lines, nil
}

// MaxListedChanges is the number of changes ChangeList shows for one flag.
const MaxListedChanges = 50

// ChangeList renders changes one per line, each starting with bullet.
// It shows at most MaxListedChanges changes and, when maxLength is above zero,
// stays within maxLength characters. A last line counts the changes left out.
func ChangeList(changes []string, bullet string, maxLength int) string {
	for shown := min(len(changes), MaxListedChanges); ; shown-- {
		lines := make([]string, 0, shown+1)
		for _, change := range changes[:shown] {
			lines = append(lines, bullet+change)
		}
		if hidden := len(changes) - shown; hidden > 0 {
			lines = append(lines, fmt.Sprintf("…and %d more changes", hidden))
		}
		text := strings.Join(lines, "\n")
		if maxLength <= 0 || utf8.RuneCountInString(text) <= maxLength || shown == 0 {
			return text
		}
	}
}

type describer struct {
	lines []string
	// inStep is true when describing what a scheduled rollout step sets, where rules
	// and variations update the flag rather than being added to it.
	inStep bool
}

func (d *describer) add(format string, args ...any) {
	d.lines = append(d.lines, fmt.Sprintf(format, args...))
}

func (d *describer) flag(b, a *flag.InternalFlag) {
	d.toggle("Flag", "disable", b.Disable, a.Disable, false, "disabled", "enabled")
	d.toggle("Event tracking", "trackEvents", b.TrackEvents, a.TrackEvents, true, "turned on", "turned off")
	d.text("Version", b.Version, a.Version)
	d.text("Bucketing key", b.BucketingKey, a.BucketingKey)
	d.variations(b.Variations, a.Variations)
	d.targeting(b.Rules, a.Rules)
	d.defaultRule(b.DefaultRule, a.DefaultRule)
	d.experimentation(b.Experimentation, a.Experimentation)
	d.scheduled(b.Scheduled, a.Scheduled)
	d.metadata(b.Metadata, a.Metadata)
}

// toggle describes an optional boolean. Setting or removing it without changing its
// effective value is still a change to the flag file, so it gets its own sentence.
func (d *describer) toggle(subject, field string, before, after *bool, defaultValue bool, onTrue, onFalse string) {
	oldValue, newValue := derefOr(before, defaultValue), derefOr(after, defaultValue)
	if d.inStep && before != after && (before == nil || after == nil) {
		// A scheduled step sets the value at its date, whatever the default is,
		// and leaves it alone when the field is absent.
		if after == nil {
			d.add("%s: `%s: %t` no longer set", subject, field, *before)
			return
		}
		oldValue = !newValue
	}
	switch {
	case oldValue != newValue && newValue:
		d.add("%s %s", subject, onTrue)
	case oldValue != newValue:
		d.add("%s %s", subject, onFalse)
	case before == nil && after != nil:
		d.add("%s: `%s: %t` added (already the default)", subject, field, *after)
	case before != nil && after == nil:
		d.add("%s: `%s: %t` removed (already the default)", subject, field, *before)
	}
}

func (d *describer) text(label string, before, after *string) {
	switch {
	case equalPtr(before, after):
	case before == nil:
		d.add("%s set to `%s`", label, *after)
	case after == nil:
		d.add("%s removed (was `%s`)", label, *before)
	default:
		d.add("%s changed from `%s` to `%s`", label, *before, *after)
	}
}

func (d *describer) variations(before, after *map[string]*any) {
	b, a := derefMap(before), derefMap(after)
	for _, name := range unionKeys(b, a) {
		oldValue, inBefore := b[name]
		newValue, inAfter := a[name]
		switch {
		case !inBefore && d.inStep:
			d.add("Variation `%s` set to `%s`", name, readablePtr(newValue))
		case !inBefore:
			d.add("Variation `%s` added: `%s`", name, readablePtr(newValue))
		case !inAfter:
			d.add("Variation `%s` removed (was `%s`)", name, readablePtr(oldValue))
		case readablePtr(oldValue) != readablePtr(newValue):
			d.add("Variation `%s` changed from `%s` to `%s`",
				name, readablePtr(oldValue), readablePtr(newValue))
		}
	}
}

// targeting pairs rules by name, so inserting a rule does not show every later rule as changed.
// Rules without a name are paired by position.
func (d *describer) targeting(before, after *[]flag.Rule) {
	b, a := indexRules(derefSlice(before)), indexRules(derefSlice(after))
	renamed := findRenames(b, a)

	for _, key := range a.order {
		newRule := a.rules[key]
		if oldKey, ok := renamed[key]; ok {
			d.add("Targeting rule %s renamed to %s", ruleLabel(oldKey), ruleLabel(key))
			continue
		}
		oldRule, existed := b.rules[key]
		switch {
		case existed:
			d.rule("Targeting rule "+ruleLabel(key), oldRule, newRule, false)
		case d.inStep:
			d.add("Targeting rule %s set: %s", ruleLabel(key), describeRule(newRule))
		default:
			d.add("Targeting rule %s added: %s", ruleLabel(key), describeRule(newRule))
		}
	}
	renamedFrom := make(map[string]bool, len(renamed))
	for _, oldKey := range renamed {
		renamedFrom[oldKey] = true
	}
	for _, key := range b.order {
		if _, kept := a.rules[key]; !kept && !renamedFrom[key] {
			d.add("Targeting rule %s removed", ruleLabel(key))
		}
	}

	// Rules are evaluated in order, so a new order changes who gets which variation.
	if keptBefore, keptAfter := keptKeys(b.order, a.rules), keptKeys(a.order, b.rules); !slices.Equal(
		keptBefore, keptAfter) {
		labels := make([]string, 0, len(a.order))
		for _, key := range a.order {
			labels = append(labels, ruleLabel(key))
		}
		d.add("Targeting rules reordered: %s", strings.Join(labels, ", "))
	}
}

// rule describes the changes to one rule. The default rule has no query and cannot be
// disabled, but both fields can still be written in the flag file.
func (d *describer) rule(label string, before, after flag.Rule, isDefault bool) {
	ignored := ""
	if isDefault {
		ignored = " (ignored for the default rule)"
		d.text(label+" name", before.Name, after.Name)
	}
	oldQuery, newQuery := derefOr(before.Query, ""), derefOr(after.Query, "")
	switch {
	case oldQuery == newQuery:
	case isDefault && newQuery == "":
		d.add("%s query removed (was `%s`)%s", label, oldQuery, ignored)
	case isDefault:
		d.add("%s query set to `%s`%s", label, newQuery, ignored)
	case oldQuery == "":
		d.add("%s now matches `%s` (was matching everyone)", label, newQuery)
	case newQuery == "":
		d.add("%s now matches everyone (was `%s`)", label, oldQuery)
	default:
		d.add("%s now matches `%s` (was `%s`)", label, newQuery, oldQuery)
	}
	if oldServe, newServe := describeServe(&before), describeServe(&after); oldServe != newServe {
		d.add("%s now serves %s (was %s)", label, newServe, oldServe)
	}
	d.toggle(label, "disable", before.Disable, after.Disable, false, "disabled"+ignored, "enabled"+ignored)
}

func (d *describer) defaultRule(before, after *flag.Rule) {
	switch {
	case before == nil && after == nil:
	case before == nil:
		d.add("Default rule set to serve %s", describeServe(after))
	case after == nil:
		d.add("Default rule removed (was serving %s)", describeServe(before))
	default:
		d.rule("Default rule", *before, *after, true)
	}
}

func (d *describer) experimentation(before, after *flag.ExperimentationRollout) {
	oldWindow, newWindow := describeWindow(before), describeWindow(after)
	switch {
	case oldWindow == newWindow:
	case before == nil:
		d.add("Experimentation scheduled %s", newWindow)
	case after == nil:
		d.add("Experimentation removed (was %s)", oldWindow)
	default:
		d.add("Experimentation now runs %s (was %s)", newWindow, oldWindow)
	}
}

// scheduled pairs scheduled rollout steps by position and describes what each step sets.
func (d *describer) scheduled(before, after *[]flag.ScheduledStep) {
	b, a := derefSlice(before), derefSlice(after)
	for i := range max(len(b), len(a)) {
		switch {
		case i >= len(b):
			d.add("Scheduled step %s added: %s", stepLabel(a[i], i), describeStepContent(a[i]))
		case i >= len(a):
			d.add("Scheduled step %s removed", stepLabel(b[i], i))
		default:
			label := stepLabel(a[i], i)
			if oldDate, newDate := stepDate(b[i], i), stepDate(a[i], i); oldDate != newDate {
				d.add("Scheduled step moved from %s to %s", oldDate, newDate)
			}
			step := &describer{inStep: true}
			step.flag(&b[i].InternalFlag, &a[i].InternalFlag)
			for _, line := range step.lines {
				d.add("Scheduled step %s: %s", label, line)
			}
		}
	}
}

func (d *describer) metadata(before, after *map[string]any) {
	b, a := derefMap(before), derefMap(after)
	for _, key := range unionKeys(b, a) {
		oldValue, inBefore := b[key]
		newValue, inAfter := a[key]
		switch {
		case !inBefore:
			d.add("Metadata `%s` set to `%s`", key, ReadableValue(newValue))
		case !inAfter:
			d.add("Metadata `%s` removed (was `%s`)", key, ReadableValue(oldValue))
		case ReadableValue(oldValue) != ReadableValue(newValue):
			d.add("Metadata `%s` changed from `%s` to `%s`",
				key, ReadableValue(oldValue), ReadableValue(newValue))
		}
	}
}

// describeStepContent lists what a new scheduled step sets, as one line.
func describeStepContent(newStep flag.ScheduledStep) string {
	step := &describer{inStep: true}
	step.flag(&flag.InternalFlag{}, &newStep.InternalFlag)
	if len(step.lines) == 0 {
		return "no changes"
	}
	return strings.Join(step.lines, "; ")
}

// stepLabel names a scheduled step by its date, or by its position when it has none.
func stepLabel(step flag.ScheduledStep, index int) string {
	if step.Date == nil {
		return stepDate(step, index)
	}
	return "on " + stepDate(step, index)
}

func stepDate(step flag.ScheduledStep, index int) string {
	if step.Date == nil {
		return "#" + strconv.Itoa(index+1)
	}
	return formatDate(*step.Date)
}

// describeRule describes a targeting rule as a whole, for a rule that was just added.
func describeRule(rule flag.Rule) string {
	text := "serves " + describeServe(&rule)
	if query := derefOr(rule.Query, ""); query != "" {
		text += fmt.Sprintf(" when `%s`", query)
	}
	if rule.IsDisable() {
		text += " (disabled)"
	}
	return text
}

// describeServe says what a rule serves, using the same precedence as Rule.Evaluate.
func describeServe(rule *flag.Rule) string {
	switch {
	case rule.ProgressiveRollout != nil:
		return describeProgressive(rule.ProgressiveRollout)
	case len(rule.GetPercentages()) > 0:
		percentages := rule.GetPercentages()
		parts := make([]string, 0, len(percentages))
		for _, name := range slices.Sorted(maps.Keys(percentages)) {
			parts = append(parts, fmt.Sprintf("`%s` %s%%", name, formatNumber(percentages[name])))
		}
		return strings.Join(parts, ", ")
	case rule.VariationResult != nil:
		return fmt.Sprintf("`%s`", rule.GetVariationResult())
	default:
		return "no variation"
	}
}

func describeProgressive(rollout *flag.ProgressiveRollout) string {
	step := func(s *flag.ProgressiveRolloutStep) string {
		if s == nil {
			return "an unset step"
		}
		text := "`" + derefOr(s.Variation, "?") + "`"
		if s.Percentage != nil {
			text += " " + formatNumber(*s.Percentage) + "%"
		}
		if s.Date != nil {
			text += " on " + formatDate(*s.Date)
		}
		return text
	}
	return fmt.Sprintf("a progressive rollout from %s to %s", step(rollout.Initial), step(rollout.End))
}

func describeWindow(window *flag.ExperimentationRollout) string {
	if window == nil {
		return ""
	}
	switch {
	case window.Start != nil && window.End != nil:
		return "from " + formatDate(*window.Start) + " to " + formatDate(*window.End)
	case window.Start != nil:
		return "from " + formatDate(*window.Start)
	case window.End != nil:
		return "until " + formatDate(*window.End)
	default:
		return "with no dates"
	}
}

type indexedRules struct {
	order []string
	rules map[string]flag.Rule
}

func indexRules(rules []flag.Rule) indexedRules {
	index := indexedRules{order: make([]string, 0, len(rules)), rules: make(map[string]flag.Rule, len(rules))}
	for i, rule := range rules {
		key := "#" + strconv.Itoa(i+1)
		if name := derefOr(rule.Name, ""); name != "" {
			key = name
		}
		index.order = append(index.order, key)
		index.rules[key] = rule
	}
	return index
}

// findRenames returns, for each new rule key, the old key of a rule that only changed its name:
// same position, same content, and neither key present on the other side.
func findRenames(before, after indexedRules) map[string]string {
	renames := map[string]string{}
	for i, newKey := range after.order {
		if i >= len(before.order) {
			break
		}
		oldKey := before.order[i]
		if _, ok := before.rules[newKey]; ok {
			continue
		}
		if _, ok := after.rules[oldKey]; ok {
			continue
		}
		if withoutName(before.rules[oldKey]) == withoutName(after.rules[newKey]) {
			renames[newKey] = oldKey
		}
	}
	return renames
}

func withoutName(rule flag.Rule) string {
	rule.Name = nil
	return ReadableValue(rule)
}

// ruleLabel formats a rule key: "`name`" for named rules, "#2" for unnamed ones.
func ruleLabel(key string) string {
	if strings.HasPrefix(key, "#") {
		return key
	}
	return "`" + key + "`"
}

// keptKeys returns the keys of order that are also in other, keeping their order.
func keptKeys(order []string, other map[string]flag.Rule) []string {
	kept := make([]string, 0, len(order))
	for _, key := range order {
		if _, ok := other[key]; ok {
			kept = append(kept, key)
		}
	}
	return kept
}

func formatDate(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04 UTC")
}

func formatNumber(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}

func readablePtr(value *any) string {
	if value == nil {
		return ReadableValue(nil)
	}
	return ReadableValue(*value)
}

func unionKeys[V any](a, b map[string]V) []string {
	keys := slices.Collect(maps.Keys(a))
	for key := range b {
		if _, ok := a[key]; !ok {
			keys = append(keys, key)
		}
	}
	slices.Sort(keys)
	return keys
}

func equalPtr[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func derefOr[T any](p *T, fallback T) T {
	if p == nil {
		return fallback
	}
	return *p
}

func derefMap[K comparable, V any](m *map[K]V) map[K]V {
	if m == nil {
		return nil
	}
	return *m
}

func derefSlice[T any](s *[]T) []T {
	if s == nil {
		return nil
	}
	return *s
}
