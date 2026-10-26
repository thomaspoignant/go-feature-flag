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
	return describeChanges(before, after, inlineCode)
}

// inlineCode wraps text in a code span whose fence is longer than any run of backticks in the text,
// padding with spaces when the text starts or ends with a backtick, as CommonMark specifies.
func inlineCode(text string) string {
	longest, run := 0, 0
	for _, r := range text {
		if r == '`' {
			run++
			longest = max(longest, run)
		} else {
			run = 0
		}
	}
	if longest == 0 {
		return "`" + text + "`"
	}
	fence := strings.Repeat("`", longest+1)
	if strings.HasPrefix(text, "`") || strings.HasSuffix(text, "`") {
		text = " " + text + " "
	}
	return fence + text + fence
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
			noun := "changes"
			if hidden == 1 {
				noun = "change"
			}
			lines = append(lines, fmt.Sprintf("…and %d more %s", hidden, noun))
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
	mapChanges(d, "Variation", b.GetVariations(), a.GetVariations(), readablePtr, "")
	d.targeting(b.GetRules(), a.GetRules())
	d.defaultRule(b.DefaultRule, a.DefaultRule)
	d.experimentation(b.Experimentation, a.Experimentation)
	if d.inStep {
		if ReadableValue(b.Scheduled) != ReadableValue(a.Scheduled) {
			d.addf("Scheduled rollout changed%s", notAppliedInStep)
		}
	} else {
		d.scheduled(b.Scheduled, a.Scheduled)
	}
	mapChanges(d, "Metadata", b.GetMetadata(), a.GetMetadata(), ReadableValue, stepNote)
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

// mapChanges describes the entries added, removed or changed in a map field such as variations or metadata.
func mapChanges[V any](d *describer, subject string, before, after map[string]V, value func(V) string, note string) {
	for _, key := range unionKeys(before, after) {
		oldValue, inBefore := before[key]
		newValue, inAfter := after[key]
		name := subject + " " + d.code(key)
		switch {
		case !inBefore && d.inStep:
			d.addf("%s set to %s%s", name, d.code(value(newValue)), note)
		case !inBefore:
			d.addf("%s added: %s%s", name, d.code(value(newValue)), note)
		case !inAfter && d.inStep:
			d.addf("%s no longer set%s", name, note)
		case !inAfter:
			d.addf("%s removed (was %s)%s", name, d.code(value(oldValue)), note)
		case value(oldValue) != value(newValue):
			d.addf("%s changed from %s to %s%s", name, d.code(value(oldValue)), d.code(value(newValue)), note)
		}
	}
}

// targeting pairs rules by name, so inserting a rule does not show every later rule as changed.
// Rules without a name, or sharing a name with an earlier rule, are paired by position.
func (d *describer) targeting(before, after []flag.Rule) {
	b, a := indexRules(before), indexRules(after)
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

	// Rules are evaluated in order, and steps append new rules in order, so a new order can change the result.
	keptBefore, keptAfter := keptKeys(b.order, a.rules), keptKeys(a.order, b.rules)
	if !slices.Equal(keptBefore, keptAfter) {
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
		stillDisabled := ""
		if !isDefault && before.IsDisable() && after.IsDisable() {
			stillDisabled = " (the rule is disabled)"
		}
		d.addf("%s now serves %s (was %s)%s", label, newServe, oldServe, stillDisabled)
	}
	d.unusedFields(label, before, after)
	d.toggle(label, "disable", before.Disable, after.Disable, false, "disabled"+ignored, "enabled"+ignored)
}

// unusedFields describes changes to fields a rule does not serve from, before or after the change.
// Rule.Evaluate uses the first of progressive rollout, percentages and variation that is set.
func (d *describer) unusedFields(label string, before, after flag.Rule) {
	percentagesUnused := func(r flag.Rule) bool { return r.ProgressiveRollout != nil || len(r.GetPercentages()) == 0 }
	variationUnused := func(r flag.Rule) bool { return r.ProgressiveRollout != nil || len(r.GetPercentages()) > 0 }

	oldPercentages, newPercentages := ReadableValue(before.Percentages), ReadableValue(after.Percentages)
	if oldPercentages != newPercentages && percentagesUnused(before) && percentagesUnused(after) {
		d.addf("%s percentages changed from %s to %s (not used: %s)",
			label, d.code(oldPercentages), d.code(newPercentages), unusedReason(after, "percentages"))
	}
	oldVariation, newVariation := ReadableValue(before.VariationResult), ReadableValue(after.VariationResult)
	if oldVariation != newVariation && variationUnused(before) && variationUnused(after) {
		d.addf("%s variation changed from %s to %s (not used: %s)",
			label, d.code(oldVariation), d.code(newVariation), unusedReason(after, "variation"))
	}
}

func unusedReason(rule flag.Rule, field string) string {
	switch {
	case rule.ProgressiveRollout != nil:
		return "the rule serves a progressive rollout"
	case field == "variation":
		return "the rule serves percentages"
	default:
		return "empty percentages are ignored"
	}
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
	if d.inStep {
		// Steps set the start and the end separately, keeping the flag's other date.
		var oldStart, oldEnd, newStart, newEnd *time.Time
		if before != nil {
			oldStart, oldEnd = before.Start, before.End
		}
		if after != nil {
			newStart, newEnd = after.Start, after.End
		}
		d.stepDate("Experimentation start", oldStart, newStart)
		d.stepDate("Experimentation end", oldEnd, newEnd)
		return
	}
	oldWindow, newWindow := describeWindow(before), describeWindow(after)
	switch {
	case oldWindow == newWindow:
	case before == nil:
		d.addf("Experimentation scheduled %s", newWindow)
	case after == nil:
		d.addf("Experimentation removed (was %s)", oldWindow)
	default:
		d.addf("Experimentation now runs %s (was %s)", newWindow, oldWindow)
	}
}

func (d *describer) stepDate(label string, before, after *time.Time) {
	switch {
	case before == nil && after == nil:
	case before == nil:
		d.addf("%s set to %s", label, formatDate(*after))
	case after == nil:
		d.addf("%s no longer set", label)
	case formatDate(*before) != formatDate(*after):
		d.addf("%s changed from %s to %s", label, formatDate(*before), formatDate(*after))
	}
}

// scheduled pairs scheduled rollout steps by date, the order they are applied in, so adding a step
// does not show every later step as changed. Steps left unpaired on both sides are paired by
// position and read as moved; steps without a date are paired by position.
func (d *describer) scheduled(before, after *[]flag.ScheduledStep) {
	b, a := derefSlice(before), derefSlice(after)
	pairs := pairSteps(b, a)
	for j := range a {
		i, paired := pairs[j]
		if !paired {
			content := "no changes"
			if lines := d.stepLines(&flag.InternalFlag{}, &a[j].InternalFlag); len(lines) > 0 {
				content = strings.Join(lines, "; ")
			}
			d.addf("Scheduled step %s added: %s", stepLabel(a[j], j), content)
			continue
		}
		if oldDate, newDate := stepDate(b[i], i), stepDate(a[j], j); oldDate != newDate {
			d.addf("Scheduled step moved from %s to %s", oldDate, newDate)
		}
		for _, line := range d.stepLines(&b[i].InternalFlag, &a[j].InternalFlag) {
			d.addf("Scheduled step %s: %s", stepLabel(a[j], j), line)
		}
	}
	kept := make(map[int]bool, len(pairs))
	for _, i := range pairs {
		kept[i] = true
	}
	for i := range b {
		if !kept[i] {
			d.addf("Scheduled step %s removed", stepLabel(b[i], i))
		}
	}
}

// pairSteps maps each step of after to the step of before it replaces.
func pairSteps(before, after []flag.ScheduledStep) map[int]int {
	key := func(steps []flag.ScheduledStep, index int) string {
		if steps[index].Date == nil {
			return "#" + strconv.Itoa(index)
		}
		return steps[index].Date.UTC().Format(time.RFC3339Nano)
	}
	byKey := map[string][]int{}
	for i := range before {
		byKey[key(before, i)] = append(byKey[key(before, i)], i)
	}
	pairs := map[int]int{}
	var unpairedAfter []int
	for j := range after {
		k := key(after, j)
		if candidates := byKey[k]; len(candidates) > 0 {
			pairs[j], byKey[k] = candidates[0], candidates[1:]
			continue
		}
		unpairedAfter = append(unpairedAfter, j)
	}
	kept := make(map[int]bool, len(pairs))
	for _, i := range pairs {
		kept[i] = true
	}
	var unpairedBefore []int
	for i := range before {
		if !kept[i] {
			unpairedBefore = append(unpairedBefore, i)
		}
	}
	for n := range min(len(unpairedBefore), len(unpairedAfter)) {
		pairs[unpairedAfter[n]] = unpairedBefore[n]
	}
	return pairs
}

// stepLines describes the changes a scheduled step makes, as a partial update.
func (d *describer) stepLines(before, after *flag.InternalFlag) []string {
	step := &describer{code: d.code, inStep: true}
	step.flag(before, after)
	return step.lines
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
		return strings.Join(d.percentageParts(rule.GetPercentages()), ", ")
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
	parts = append(parts, d.percentageParts(rule.GetPercentages())...)
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

// percentageParts describes each variation's percentage, sorted by name.
// A negative percentage, which a scheduled step uses to remove a variation, reads as removed.
func (d *describer) percentageParts(percentages map[string]float64) []string {
	parts := make([]string, 0, len(percentages))
	for _, name := range slices.Sorted(maps.Keys(percentages)) {
		if percentages[name] < 0 {
			parts = append(parts, d.code(name)+" removed from the percentages")
		} else {
			parts = append(parts, d.code(name)+" "+formatNumber(percentages[name])+"%")
		}
	}
	return parts
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

func derefSlice[T any](s *[]T) []T {
	if s == nil {
		return nil
	}
	return *s
}
