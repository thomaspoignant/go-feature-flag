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
// in the order the fields appear in a flag file. Names and values are formatted as
// markdown inline code.
// It describes every field of flag.InternalFlag and returns an error for any other flag type.
func DescribeChanges(before, after flag.Flag) ([]string, error) {
	return describeChanges(before, after, func(text string) string { return "`" + text + "`" })
}

// DescribeChangesPlain is DescribeChanges without inline code, for destinations that cannot display it.
func DescribeChangesPlain(before, after flag.Flag) ([]string, error) {
	return describeChanges(before, after, func(text string) string { return text })
}

func describeChanges(before, after flag.Flag, code func(string) string) ([]string, error) {
	b, okBefore := before.(*flag.InternalFlag)
	a, okAfter := after.(*flag.InternalFlag)
	if !okBefore || !okAfter || b == nil || a == nil {
		return nil, fmt.Errorf("impossible to describe changes between flags of type %T and %T", before, after)
	}
	d := &describer{code: code}
	d.flag(b, a)
	return d.lines, nil
}

// MaxListedChanges is the number of changes ChangeList shows for one flag.
const MaxListedChanges = 50

// ChangeList renders changes one per line, each starting with bullet.
// It shows at most MaxListedChanges changes and, when maxLength is above zero,
// stays within maxLength characters. A last line counts the changes left out,
// and the list is empty when not even that line fits.
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
		if maxLength <= 0 || utf8.RuneCountInString(text) <= maxLength {
			return text
		}
		if shown == 0 {
			return ""
		}
	}
}

// notAppliedInStep marks fields that scheduled rollout steps do not apply.
const notAppliedInStep = " (not applied by scheduled steps)"

type describer struct {
	lines []string
	// code formats a name or a value.
	code func(string) string
	// inStep is true when describing a scheduled rollout step. A step is a partial update:
	// it only changes the fields it sets, and rules are merged field by field.
	inStep bool
}

func (d *describer) addf(format string, args ...any) {
	d.lines = append(d.lines, fmt.Sprintf(format, args...))
}

func (d *describer) flag(b, a *flag.InternalFlag) {
	stepNote := ""
	if d.inStep {
		stepNote = notAppliedInStep
	}
	d.toggle("Flag", "disable", b.Disable, a.Disable, false, "disabled", "enabled")
	d.toggle("Event tracking", "trackEvents", b.TrackEvents, a.TrackEvents, true, "turned on", "turned off")
	d.text("Version", b.Version, a.Version, "")
	d.text("Bucketing key", b.BucketingKey, a.BucketingKey, stepNote)
	d.variations(b.Variations, a.Variations)
	d.targeting(b.Rules, a.Rules)
	d.defaultRule(b.DefaultRule, a.DefaultRule)
	d.experimentation(b.Experimentation, a.Experimentation)
	if d.inStep {
		if ReadableValue(b.Scheduled) != ReadableValue(a.Scheduled) {
			d.addf("Scheduled rollout changed%s", notAppliedInStep)
		}
	} else {
		d.scheduled(b.Scheduled, a.Scheduled)
	}
	d.metadata(b.Metadata, a.Metadata, stepNote)
}

// toggle describes an optional boolean. Setting or removing it without changing its
// effective value is still a change to the flag file, so it gets its own sentence.
func (d *describer) toggle(subject, field string, before, after *bool, defaultValue bool, onTrue, onFalse string) {
	oldValue, newValue := derefOr(before, defaultValue), derefOr(after, defaultValue)
	if d.inStep && before != after && (before == nil || after == nil) {
		// A scheduled step sets the value at its date, whatever the default is,
		// and leaves it alone when the field is absent.
		if after == nil {
			d.addf("%s: %s no longer set", subject, d.code(fmt.Sprintf("%s: %t", field, *before)))
			return
		}
		oldValue = !newValue
	}
	switch {
	case oldValue != newValue && newValue:
		d.addf("%s %s", subject, onTrue)
	case oldValue != newValue:
		d.addf("%s %s", subject, onFalse)
	case before == nil && after != nil:
		d.addf("%s: %s added (already the default)", subject, d.code(fmt.Sprintf("%s: %t", field, *after)))
	case before != nil && after == nil:
		d.addf("%s: %s removed (already the default)", subject, d.code(fmt.Sprintf("%s: %t", field, *before)))
	}
}

func (d *describer) text(label string, before, after *string, note string) {
	switch {
	case equalPtr(before, after):
	case before == nil:
		d.addf("%s set to %s%s", label, d.code(*after), note)
	case after == nil && d.inStep:
		d.addf("%s no longer set%s", label, note)
	case after == nil:
		d.addf("%s removed (was %s)%s", label, d.code(*before), note)
	default:
		d.addf("%s changed from %s to %s%s", label, d.code(*before), d.code(*after), note)
	}
}

func (d *describer) variations(before, after *map[string]*any) {
	b, a := derefMap(before), derefMap(after)
	for _, name := range unionKeys(b, a) {
		oldValue, inBefore := b[name]
		newValue, inAfter := a[name]
		switch {
		case !inBefore && d.inStep:
			d.addf("Variation %s set to %s", d.code(name), d.code(readablePtr(newValue)))
		case !inBefore:
			d.addf("Variation %s added: %s", d.code(name), d.code(readablePtr(newValue)))
		case !inAfter && d.inStep:
			d.addf("Variation %s no longer set", d.code(name))
		case !inAfter:
			d.addf("Variation %s removed (was %s)", d.code(name), d.code(readablePtr(oldValue)))
		case readablePtr(oldValue) != readablePtr(newValue):
			d.addf("Variation %s changed from %s to %s",
				d.code(name), d.code(readablePtr(oldValue)), d.code(readablePtr(newValue)))
		}
	}
}

// targeting pairs rules by name, so inserting a rule does not show every later rule as changed.
// Rules without a name, or sharing a name with an earlier rule, are paired by position.
func (d *describer) targeting(before, after *[]flag.Rule) {
	b, a := indexRules(derefSlice(before)), indexRules(derefSlice(after))
	renamed := map[ruleKey]ruleKey{}
	if !d.inStep {
		renamed = findRenames(b, a)
	}

	for _, key := range a.order {
		newRule := a.rules[key]
		if oldKey, ok := renamed[key]; ok {
			d.addf("Targeting rule %s renamed to %s", d.ruleLabel(oldKey), d.ruleLabel(key))
			continue
		}
		oldRule, existed := b.rules[key]
		label := "Targeting rule " + d.ruleLabel(key)
		switch {
		case d.inStep && existed:
			d.patchChange(label, oldRule, newRule)
		case d.inStep:
			d.addf("%s updated with %s", label, d.describePatch(newRule))
		case existed:
			d.rule(label, oldRule, newRule, false)
		default:
			d.addf("%s added: %s", label, d.describeRule(newRule))
		}
	}
	renamedFrom := make(map[ruleKey]bool, len(renamed))
	for _, oldKey := range renamed {
		renamedFrom[oldKey] = true
	}
	for _, key := range b.order {
		if _, kept := a.rules[key]; kept || renamedFrom[key] {
			continue
		}
		if d.inStep {
			d.addf("Targeting rule %s no longer changed", d.ruleLabel(key))
		} else {
			d.addf("Targeting rule %s removed", d.ruleLabel(key))
		}
	}

	// Rules are evaluated in order, so a new order changes who gets which variation.
	// Steps merge rules by name, so their order does not matter.
	if keptBefore, keptAfter := keptKeys(b.order, a.rules), keptKeys(a.order, b.rules); !d.inStep &&
		!slices.Equal(keptBefore, keptAfter) {
		labels := make([]string, 0, len(a.order))
		for _, key := range a.order {
			labels = append(labels, d.ruleLabel(key))
		}
		d.addf("Targeting rules reordered: %s", strings.Join(labels, ", "))
	}
}

// rule describes the changes to one rule. The default rule has no query and cannot be
// disabled, but both fields can still be written in the flag file.
func (d *describer) rule(label string, before, after flag.Rule, isDefault bool) {
	ignored := ""
	if isDefault {
		ignored = " (ignored for the default rule)"
		d.text(label+" name", before.Name, after.Name, "")
	}
	oldQuery, newQuery := derefOr(before.Query, ""), derefOr(after.Query, "")
	switch {
	case oldQuery == newQuery:
	case isDefault && newQuery == "":
		d.addf("%s query removed (was %s)%s", label, d.code(oldQuery), ignored)
	case isDefault:
		d.addf("%s query set to %s%s", label, d.code(newQuery), ignored)
	case oldQuery == "":
		d.addf("%s now matches %s (was matching everyone)", label, d.code(newQuery))
	case newQuery == "":
		d.addf("%s now matches everyone (was %s)", label, d.code(oldQuery))
	default:
		d.addf("%s now matches %s (was %s)", label, d.code(newQuery), d.code(oldQuery))
	}
	if oldServe, newServe := d.describeServe(&before), d.describeServe(&after); oldServe != newServe {
		d.addf("%s now serves %s (was %s)", label, newServe, oldServe)
	}
	d.toggle(label, "disable", before.Disable, after.Disable, false, "disabled"+ignored, "enabled"+ignored)
}

// patchChange describes a rule a scheduled step already updated, and now updates differently.
func (d *describer) patchChange(label string, before, after flag.Rule) {
	if oldPatch, newPatch := d.describePatch(before), d.describePatch(after); oldPatch != newPatch {
		d.addf("%s now updated with %s (was %s)", label, newPatch, oldPatch)
	}
}

func (d *describer) defaultRule(before, after *flag.Rule) {
	switch {
	case before == nil && after == nil:
	case d.inStep && before == nil:
		d.addf("Default rule updated with %s", d.describePatch(*after))
	case d.inStep && after == nil:
		d.addf("Default rule no longer changed")
	case d.inStep:
		d.patchChange("Default rule", *before, *after)
	case before == nil:
		d.addf("Default rule set to serve %s", d.describeServe(after))
	case after == nil:
		d.addf("Default rule removed (was serving %s)", d.describeServe(before))
	default:
		d.rule("Default rule", *before, *after, true)
	}
}

func (d *describer) experimentation(before, after *flag.ExperimentationRollout) {
	oldWindow, newWindow := describeWindow(before), describeWindow(after)
	switch {
	case oldWindow == newWindow:
	case before == nil:
		d.addf("Experimentation scheduled %s", newWindow)
	case after == nil && d.inStep:
		d.addf("Experimentation no longer set")
	case after == nil:
		d.addf("Experimentation removed (was %s)", oldWindow)
	default:
		d.addf("Experimentation now runs %s (was %s)", newWindow, oldWindow)
	}
}

// scheduled pairs scheduled rollout steps by position and describes what each step sets.
func (d *describer) scheduled(before, after *[]flag.ScheduledStep) {
	b, a := derefSlice(before), derefSlice(after)
	for i := range max(len(b), len(a)) {
		switch {
		case i >= len(b):
			step := &describer{code: d.code, inStep: true}
			step.flag(&flag.InternalFlag{}, &a[i].InternalFlag)
			content := "no changes"
			if len(step.lines) > 0 {
				content = strings.Join(step.lines, "; ")
			}
			d.addf("Scheduled step %s added: %s", stepLabel(a[i], i), content)
		case i >= len(a):
			d.addf("Scheduled step %s removed", stepLabel(b[i], i))
		default:
			if oldDate, newDate := stepDate(b[i], i), stepDate(a[i], i); oldDate != newDate {
				d.addf("Scheduled step moved from %s to %s", oldDate, newDate)
			}
			step := &describer{code: d.code, inStep: true}
			step.flag(&b[i].InternalFlag, &a[i].InternalFlag)
			for _, line := range step.lines {
				d.addf("Scheduled step %s: %s", stepLabel(a[i], i), line)
			}
		}
	}
}

func (d *describer) metadata(before, after *map[string]any, note string) {
	b, a := derefMap(before), derefMap(after)
	for _, key := range unionKeys(b, a) {
		oldValue, inBefore := b[key]
		newValue, inAfter := a[key]
		switch {
		case !inBefore:
			d.addf("Metadata %s set to %s%s", d.code(key), d.code(ReadableValue(newValue)), note)
		case !inAfter:
			d.addf("Metadata %s removed (was %s)%s", d.code(key), d.code(ReadableValue(oldValue)), note)
		case ReadableValue(oldValue) != ReadableValue(newValue):
			d.addf("Metadata %s changed from %s to %s%s",
				d.code(key), d.code(ReadableValue(oldValue)), d.code(ReadableValue(newValue)), note)
		}
	}
}

// describeRule describes a targeting rule as a whole, for a rule that was just added.
func (d *describer) describeRule(rule flag.Rule) string {
	text := "serves " + d.describeServe(&rule)
	if query := derefOr(rule.Query, ""); query != "" {
		text += " when " + d.code(query)
	}
	if rule.IsDisable() {
		text += " (disabled)"
	}
	return text
}

// describeServe says what a rule serves, using the same precedence as Rule.Evaluate.
func (d *describer) describeServe(rule *flag.Rule) string {
	switch {
	case rule.ProgressiveRollout != nil:
		return d.describeProgressive(rule.ProgressiveRollout)
	case len(rule.GetPercentages()) > 0:
		percentages := rule.GetPercentages()
		parts := make([]string, 0, len(percentages))
		for _, name := range slices.Sorted(maps.Keys(percentages)) {
			parts = append(parts, d.code(name)+" "+formatNumber(percentages[name])+"%")
		}
		return strings.Join(parts, ", ")
	case rule.VariationResult != nil:
		return d.code(rule.GetVariationResult())
	default:
		return "no variation"
	}
}

// describePatch lists the fields a scheduled step sets on a rule, as Rule.MergeRules applies them.
func (d *describer) describePatch(rule flag.Rule) string {
	parts := make([]string, 0, 5)
	if rule.Query != nil {
		parts = append(parts, "query "+d.code(*rule.Query))
	}
	if rule.VariationResult != nil {
		parts = append(parts, "variation "+d.code(*rule.VariationResult))
	}
	percentages := rule.GetPercentages()
	for _, name := range slices.Sorted(maps.Keys(percentages)) {
		// A negative percentage removes the variation from the rule's percentages.
		if percentages[name] < 0 {
			parts = append(parts, d.code(name)+" removed from the percentages")
		} else {
			parts = append(parts, d.code(name)+" "+formatNumber(percentages[name])+"%")
		}
	}
	if rule.ProgressiveRollout != nil {
		parts = append(parts, d.describeProgressive(rule.ProgressiveRollout))
	}
	if rule.Disable != nil {
		parts = append(parts, d.code(fmt.Sprintf("disable: %t", *rule.Disable)))
	}
	if len(parts) == 0 {
		return "nothing"
	}
	return strings.Join(parts, ", ")
}

func (d *describer) describeProgressive(rollout *flag.ProgressiveRollout) string {
	step := func(s *flag.ProgressiveRolloutStep) string {
		if s == nil {
			return "an unset step"
		}
		parts := make([]string, 0, 3)
		if s.Variation != nil {
			parts = append(parts, d.code(*s.Variation))
		}
		if s.Percentage != nil {
			parts = append(parts, formatNumber(*s.Percentage)+"%")
		}
		if s.Date != nil {
			parts = append(parts, "on "+formatDate(*s.Date))
		}
		if len(parts) == 0 {
			return "an unset step"
		}
		return strings.Join(parts, " ")
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

// ruleKey identifies a rule by name, or by position when it has no name of its own.
type ruleKey struct {
	name     string
	position int
}

func (d *describer) ruleLabel(key ruleKey) string {
	if key.name == "" {
		return "#" + strconv.Itoa(key.position)
	}
	return d.code(key.name)
}

type indexedRules struct {
	order []ruleKey
	rules map[ruleKey]flag.Rule
}

func indexRules(rules []flag.Rule) indexedRules {
	index := indexedRules{order: make([]ruleKey, 0, len(rules)), rules: make(map[ruleKey]flag.Rule, len(rules))}
	for i, rule := range rules {
		key := ruleKey{name: derefOr(rule.Name, "")}
		if _, taken := index.rules[key]; key.name == "" || taken {
			key = ruleKey{position: i + 1}
		}
		index.order = append(index.order, key)
		index.rules[key] = rule
	}
	return index
}

// findRenames returns, for each new rule key, the old key of a rule that only changed its name:
// same position, same content, and neither key present on the other side.
func findRenames(before, after indexedRules) map[ruleKey]ruleKey {
	renames := map[ruleKey]ruleKey{}
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

// keptKeys returns the keys of order that are also in other, keeping their order.
func keptKeys(order []ruleKey, other map[ruleKey]flag.Rule) []ruleKey {
	kept := make([]ruleKey, 0, len(order))
	for _, key := range order {
		if _, ok := other[key]; ok {
			kept = append(kept, key)
		}
	}
	return kept
}

// formatDate shows a date to the minute, with seconds only when it has them,
// so two different dates never read the same.
func formatDate(t time.Time) string {
	t = t.UTC()
	if t.Second() == 0 && t.Nanosecond() == 0 {
		return t.Format("2006-01-02 15:04 UTC")
	}
	return t.Format("2006-01-02 15:04:05.999999999 UTC")
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
