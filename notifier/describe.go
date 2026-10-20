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
// A change it has no sentence for is shown as "path: before => after", as FlagChanges
// renders it, so a notification never hides a change.
func DescribeChanges(before, after flag.Flag) ([]string, error) {
	b, okBefore := before.(*flag.InternalFlag)
	a, okAfter := after.(*flag.InternalFlag)
	if !okBefore || !okAfter || b == nil || a == nil {
		return describeAsFieldChanges("", before, after)
	}

	d := &describer{}
	d.flagStatus(b, a)
	d.text("Version", b.Version, a.Version)
	d.text("Bucketing key", b.BucketingKey, a.BucketingKey)
	d.variations(b.Variations, a.Variations)
	d.targeting(b.Rules, a.Rules)
	d.defaultRule(b.DefaultRule, a.DefaultRule)
	d.experimentation(b.Experimentation, a.Experimentation)
	d.scheduled(b.Scheduled, a.Scheduled)
	d.metadata(b.Metadata, a.Metadata)
	if d.err != nil {
		return nil, d.err
	}
	if len(d.lines) == 0 {
		// Changes the sentences do not cover, such as a renamed default rule.
		return describeAsFieldChanges("", before, after)
	}
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
	err   error
}

func (d *describer) add(format string, args ...any) {
	d.lines = append(d.lines, fmt.Sprintf(format, args...))
}

func (d *describer) flagStatus(b, a *flag.InternalFlag) {
	if b.IsDisable() != a.IsDisable() {
		if a.IsDisable() {
			d.add("Flag disabled")
		} else {
			d.add("Flag enabled")
		}
	}
	if b.IsTrackEvents() != a.IsTrackEvents() {
		if a.IsTrackEvents() {
			d.add("Event tracking turned on")
		} else {
			d.add("Event tracking turned off")
		}
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

	for _, key := range a.order {
		newRule := a.rules[key]
		oldRule, existed := b.rules[key]
		if !existed {
			d.add("Targeting rule %s added: %s", ruleLabel(key), describeRule(newRule))
			continue
		}
		label := "Targeting rule " + ruleLabel(key)
		if oldRule.GetTrimmedQuery() != newRule.GetTrimmedQuery() {
			d.add("%s now matches `%s` (was `%s`)",
				label, newRule.GetTrimmedQuery(), oldRule.GetTrimmedQuery())
		}
		if oldServe, newServe := describeServe(&oldRule), describeServe(&newRule); oldServe != newServe {
			d.add("%s now serves %s (was %s)", label, newServe, oldServe)
		}
		if oldRule.IsDisable() != newRule.IsDisable() {
			if newRule.IsDisable() {
				d.add("%s disabled", label)
			} else {
				d.add("%s enabled", label)
			}
		}
	}
	for _, key := range b.order {
		if _, kept := a.rules[key]; !kept {
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

func (d *describer) defaultRule(before, after *flag.Rule) {
	switch {
	case before == nil && after == nil:
	case before == nil:
		d.add("Default rule set to serve %s", describeServe(after))
	case after == nil:
		d.add("Default rule removed (was serving %s)", describeServe(before))
	default:
		if oldServe, newServe := describeServe(before), describeServe(after); oldServe != newServe {
			d.add("Default rule now serves %s (was %s)", newServe, oldServe)
		}
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

func (d *describer) scheduled(before, after *[]flag.ScheduledStep) {
	b, a := derefSlice(before), derefSlice(after)
	switch {
	case len(b) == 0 && len(a) == 0:
	case len(b) == 0:
		d.add("Scheduled rollout added: %s", describeSteps(a))
	case len(a) == 0:
		d.add("Scheduled rollout removed (was %s)", describeSteps(b))
	default:
		lines, err := describeAsFieldChanges("Scheduled.", b, a)
		if err != nil {
			d.err = err
			return
		}
		d.lines = append(d.lines, lines...)
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

// describeAsFieldChanges renders changes as "path: before => after" lines.
func describeAsFieldChanges(prefix string, before, after any) ([]string, error) {
	changes, err := FlagChanges(before, after)
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0, len(changes))
	for _, change := range changes {
		lines = append(lines, fmt.Sprintf("%s%s: %s => %s", prefix, change.Path, change.From, change.To))
	}
	return lines, nil
}

// describeRule describes a targeting rule as a whole, for a rule that was just added.
func describeRule(rule flag.Rule) string {
	text := "serves " + describeServe(&rule)
	if query := rule.GetTrimmedQuery(); query != "" {
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

func describeSteps(steps []flag.ScheduledStep) string {
	dates := make([]string, 0, len(steps))
	for _, step := range steps {
		if step.Date != nil {
			dates = append(dates, formatDate(*step.Date))
		}
	}
	text := fmt.Sprintf("%d steps", len(steps))
	if len(steps) == 1 {
		text = "1 step"
	}
	if len(dates) > 0 {
		text += " on " + strings.Join(dates, ", ")
	}
	return text
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
