package notifier_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

func TestDescribeChanges(t *testing.T) {
	start := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	end := start.Add(7 * 24 * time.Hour)
	baseFlag := func() *flag.InternalFlag {
		return &flag.InternalFlag{
			Variations: &map[string]*any{
				"on":  testconvert.Interface(true),
				"off": testconvert.Interface(false),
			},
			Rules: &[]flag.Rule{
				{
					Name:            testconvert.String("internal"),
					Query:           testconvert.String(`email ew "@example.com"`),
					VariationResult: testconvert.String("on"),
				},
				{
					Name:            testconvert.String("beta"),
					Query:           testconvert.String(`beta eq true`),
					VariationResult: testconvert.String("on"),
				},
			},
			DefaultRule: &flag.Rule{
				Percentages: &map[string]float64{"on": 20, "off": 80},
			},
			Version:  testconvert.String("1.0"),
			Metadata: &map[string]any{"owner": "payments"},
		}
	}

	tests := []struct {
		name   string
		setup  func(f *flag.InternalFlag)
		update func(f *flag.InternalFlag)
		want   []string
	}{
		{
			name:   "should describe a disabled flag",
			update: func(f *flag.InternalFlag) { f.Disable = testconvert.Bool(true) },
			want:   []string{"Flag disabled"},
		},
		{
			name:   "should describe turning off event tracking",
			update: func(f *flag.InternalFlag) { f.TrackEvents = testconvert.Bool(false) },
			want:   []string{"Event tracking turned off"},
		},
		{
			name: "should describe a percentage change on the default rule",
			update: func(f *flag.InternalFlag) {
				f.DefaultRule.Percentages = &map[string]float64{"on": 50, "off": 50}
			},
			want: []string{"Default rule now serves `off` 50%, `on` 50% (was `off` 80%, `on` 20%)"},
		},
		{
			name: "should describe a rule added first without reporting the other rules as changed",
			update: func(f *flag.InternalFlag) {
				*f.Rules = append([]flag.Rule{{
					Name:            testconvert.String("canada"),
					Query:           testconvert.String(`country eq "CA"`),
					VariationResult: testconvert.String("off"),
				}}, *f.Rules...)
			},
			want: []string{"Targeting rule `canada` added: serves `off` when `country eq \"CA\"`"},
		},
		{
			name:   "should describe a removed rule",
			update: func(f *flag.InternalFlag) { *f.Rules = (*f.Rules)[:1] },
			want:   []string{"Targeting rule `beta` removed"},
		},
		{
			name: "should describe an edited rule query and result",
			update: func(f *flag.InternalFlag) {
				(*f.Rules)[1].Query = testconvert.String(`beta eq true and country eq "US"`)
				(*f.Rules)[1].VariationResult = nil
				(*f.Rules)[1].Percentages = &map[string]float64{"on": 10, "off": 90}
			},
			want: []string{
				"Targeting rule `beta` now matches `beta eq true and country eq \"US\"` (was `beta eq true`)",
				"Targeting rule `beta` now serves `off` 90%, `on` 10% (was `on`)",
			},
		},
		{
			name:   "should describe a disabled rule",
			update: func(f *flag.InternalFlag) { (*f.Rules)[0].Disable = testconvert.Bool(true) },
			want:   []string{"Targeting rule `internal` disabled"},
		},
		{
			name: "should describe reordered rules",
			update: func(f *flag.InternalFlag) {
				*f.Rules = []flag.Rule{(*f.Rules)[1], (*f.Rules)[0]}
			},
			want: []string{"Targeting rules reordered: `beta`, `internal`"},
		},
		{
			name: "should pair unnamed rules by position",
			update: func(f *flag.InternalFlag) {
				f.Rules = &[]flag.Rule{{
					Query:           testconvert.String(`plan eq "pro"`),
					VariationResult: testconvert.String("on"),
				}}
			},
			want: []string{
				"Targeting rule #1 added: serves `on` when `plan eq \"pro\"`",
				"Targeting rule `internal` removed",
				"Targeting rule `beta` removed",
			},
		},
		{
			name: "should describe variation changes",
			update: func(f *flag.InternalFlag) {
				(*f.Variations)["on"] = testconvert.Interface("enabled")
				(*f.Variations)["legacy"] = testconvert.Interface(map[string]any{"color": "blue"})
				delete(*f.Variations, "off")
			},
			want: []string{
				"Variation `legacy` added: `{\"color\":\"blue\"}`",
				"Variation `off` removed (was `false`)",
				"Variation `on` changed from `true` to `\"enabled\"`",
			},
		},
		{
			name: "should describe a progressive rollout on the default rule",
			update: func(f *flag.InternalFlag) {
				f.DefaultRule = &flag.Rule{ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("off"),
						Percentage: testconvert.Float64(0),
						Date:       &start,
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("on"),
						Percentage: testconvert.Float64(100),
						Date:       &end,
					},
				}}
			},
			want: []string{
				"Default rule now serves a progressive rollout from `off` 0% on 2026-10-05 09:00 UTC " +
					"to `on` 100% on 2026-10-12 09:00 UTC (was `off` 80%, `on` 20%)",
			},
		},
		{
			name: "should describe a new experimentation window",
			update: func(f *flag.InternalFlag) {
				f.Experimentation = &flag.ExperimentationRollout{Start: &start, End: &end}
			},
			want: []string{"Experimentation scheduled from 2026-10-05 09:00 UTC to 2026-10-12 09:00 UTC"},
		},
		{
			name: "should describe version, bucketing key and metadata changes",
			update: func(f *flag.InternalFlag) {
				f.Version = testconvert.String("1.1")
				f.BucketingKey = testconvert.String("teamId")
				f.Metadata = &map[string]any{"owner": "growth", "jira": "PAY-123"}
			},
			want: []string{
				"Version changed from `1.0` to `1.1`",
				"Bucketing key set to `teamId`",
				"Metadata `jira` set to `\"PAY-123\"`",
				"Metadata `owner` changed from `\"payments\"` to `\"growth\"`",
			},
		},
		{
			name: "should describe a new scheduled rollout",
			update: func(f *flag.InternalFlag) {
				f.Scheduled = &[]flag.ScheduledStep{
					{Date: &start, InternalFlag: flag.InternalFlag{Disable: testconvert.Bool(false)}},
					{Date: &end, InternalFlag: flag.InternalFlag{Disable: testconvert.Bool(true)}},
				}
			},
			want: []string{
				"Scheduled step on 2026-10-05 09:00 UTC added: Flag enabled",
				"Scheduled step on 2026-10-12 09:00 UTC added: Flag disabled",
			},
		},
		{
			name: "should describe a moved scheduled step",
			setup: func(f *flag.InternalFlag) {
				f.Scheduled = &[]flag.ScheduledStep{
					{Date: &start, InternalFlag: flag.InternalFlag{Disable: testconvert.Bool(true)}},
				}
			},
			update: func(f *flag.InternalFlag) { (*f.Scheduled)[0].Date = &end },
			want:   []string{"Scheduled step moved from 2026-10-05 09:00 UTC to 2026-10-12 09:00 UTC"},
		},
		{
			name: "should describe what an edited scheduled step sets",
			setup: func(f *flag.InternalFlag) {
				f.Scheduled = &[]flag.ScheduledStep{
					{Date: &start, InternalFlag: flag.InternalFlag{Disable: testconvert.Bool(true)}},
				}
			},
			update: func(f *flag.InternalFlag) {
				(*f.Scheduled)[0].Disable = nil
				(*f.Scheduled)[0].Rules = &[]flag.Rule{{
					Name:        testconvert.String("beta"),
					Percentages: &map[string]float64{"on": 100},
				}}
			},
			want: []string{
				"Scheduled step on 2026-10-05 09:00 UTC: Flag: `disable: true` no longer set",
				"Scheduled step on 2026-10-05 09:00 UTC: Targeting rule `beta` updated with `on` 100%",
			},
		},
		{
			name: "should describe a removed scheduled step",
			setup: func(f *flag.InternalFlag) {
				f.Scheduled = &[]flag.ScheduledStep{
					{Date: &start, InternalFlag: flag.InternalFlag{Disable: testconvert.Bool(true)}},
					{Date: &end, InternalFlag: flag.InternalFlag{Disable: testconvert.Bool(false)}},
				}
			},
			update: func(f *flag.InternalFlag) { *f.Scheduled = (*f.Scheduled)[:1] },
			want:   []string{"Scheduled step on 2026-10-12 09:00 UTC removed"},
		},
		{
			name: "should describe a renamed rule",
			update: func(f *flag.InternalFlag) {
				(*f.Rules)[1].Name = testconvert.String("beta-users")
			},
			want: []string{"Targeting rule `beta` renamed to `beta-users`"},
		},
		{
			name: "should describe naming an unnamed rule",
			setup: func(f *flag.InternalFlag) {
				(*f.Rules)[1].Name = nil
			},
			update: func(f *flag.InternalFlag) {
				(*f.Rules)[1].Name = testconvert.String("beta")
			},
			want: []string{"Targeting rule #2 renamed to `beta`"},
		},
		{
			name:   "should describe a value written out that is already the default",
			update: func(f *flag.InternalFlag) { f.Disable = testconvert.Bool(false) },
			want:   []string{"Flag: `disable: false` added (already the default)"},
		},
		{
			name: "should describe a default value removed from a rule",
			setup: func(f *flag.InternalFlag) {
				(*f.Rules)[0].Disable = testconvert.Bool(false)
			},
			update: func(f *flag.InternalFlag) { (*f.Rules)[0].Disable = nil },
			want:   []string{"Targeting rule `internal`: `disable: false` removed (already the default)"},
		},
		{
			name: "should describe default rule fields that evaluation ignores",
			update: func(f *flag.InternalFlag) {
				f.DefaultRule.Query = testconvert.String(`country eq "FR"`)
				f.DefaultRule.Disable = testconvert.Bool(true)
			},
			want: []string{
				"Default rule query set to `country eq \"FR\"` (ignored for the default rule)",
				"Default rule disabled (ignored for the default rule)",
			},
		},
		{
			name: "should describe every change when several fields change together",
			update: func(f *flag.InternalFlag) {
				f.Disable = testconvert.Bool(true)
				f.DefaultRule.Name = testconvert.String("default")
			},
			want: []string{"Flag disabled", "Default rule name set to `default`"},
		},
		{
			name:   "should describe a renamed default rule",
			update: func(f *flag.InternalFlag) { f.DefaultRule.Name = testconvert.String("default") },
			want:   []string{"Default rule name set to `default`"},
		},
		{
			name: "should show seconds so close dates do not read the same",
			update: func(f *flag.InternalFlag) {
				f.Version = testconvert.String("1.1")
				f.Experimentation = &flag.ExperimentationRollout{End: testconvert.Time(start.Add(50 * time.Second))}
			},
			setup: func(f *flag.InternalFlag) {
				f.Experimentation = &flag.ExperimentationRollout{End: testconvert.Time(start.Add(10 * time.Second))}
			},
			want: []string{
				"Version changed from `1.0` to `1.1`",
				"Experimentation now runs until 2026-10-05 09:00:50 UTC (was until 2026-10-05 09:00:10 UTC)",
			},
		},
		{
			name: "should keep a rule named like a position apart from the unnamed rule at that position",
			setup: func(f *flag.InternalFlag) {
				f.Rules = &[]flag.Rule{
					{Name: testconvert.String("#2"), Query: testconvert.String("a eq 1"), VariationResult: testconvert.String("on")},
					{Query: testconvert.String("b eq 1"), VariationResult: testconvert.String("on")},
				}
			},
			update: func(f *flag.InternalFlag) {
				(*f.Rules)[0].VariationResult = testconvert.String("off")
			},
			want: []string{"Targeting rule `#2` now serves `off` (was `on`)"},
		},
		{
			name: "should pair rules sharing a name by position",
			setup: func(f *flag.InternalFlag) {
				f.Rules = &[]flag.Rule{
					{Name: testconvert.String("beta"), Query: testconvert.String("a eq 1"), VariationResult: testconvert.String("on")},
					{Name: testconvert.String("beta"), Query: testconvert.String("b eq 1"), VariationResult: testconvert.String("on")},
				}
			},
			update: func(f *flag.InternalFlag) {
				(*f.Rules)[1].VariationResult = testconvert.String("off")
			},
			want: []string{"Targeting rule #2 now serves `off` (was `on`)"},
		},
		{
			name: "should describe a scheduled step as a partial update",
			update: func(f *flag.InternalFlag) {
				f.Scheduled = &[]flag.ScheduledStep{{
					Date: &start,
					InternalFlag: flag.InternalFlag{
						Rules: &[]flag.Rule{{
							Name:  testconvert.String("beta"),
							Query: testconvert.String(`beta eq true and plan eq "pro"`),
						}},
						DefaultRule: &flag.Rule{Percentages: &map[string]float64{"on": 50, "off": -1}},
						Metadata:    &map[string]any{"owner": "growth"},
					},
				}}
			},
			want: []string{
				"Scheduled step on 2026-10-05 09:00 UTC added: " +
					"Targeting rule `beta` updated with query `beta eq true and plan eq \"pro\"`; " +
					"Default rule updated with `off` removed from the percentages, `on` 50%; " +
					"Metadata `owner` set to `\"growth\"` (not applied by scheduled steps)",
			},
		},
		{
			name: "should describe a step that stops updating a rule",
			setup: func(f *flag.InternalFlag) {
				f.Scheduled = &[]flag.ScheduledStep{{
					Date:         &start,
					InternalFlag: flag.InternalFlag{Rules: &[]flag.Rule{{Name: testconvert.String("beta"), VariationResult: testconvert.String("off")}}},
				}}
			},
			update: func(f *flag.InternalFlag) { (*f.Scheduled)[0].Rules = nil },
			want:   []string{"Scheduled step on 2026-10-05 09:00 UTC: Targeting rule `beta` no longer changed"},
		},
		{
			name: "should describe a variation the rule does not serve from",
			setup: func(f *flag.InternalFlag) {
				f.DefaultRule.VariationResult = testconvert.String("on")
			},
			update: func(f *flag.InternalFlag) { f.DefaultRule.VariationResult = testconvert.String("off") },
			want: []string{
				"Default rule variation changed from `\"on\"` to `\"off\"` (not used: the rule serves percentages)",
			},
		},
		{
			name:   "should describe empty percentages",
			setup:  func(f *flag.InternalFlag) { (*f.Rules)[0].Percentages = nil },
			update: func(f *flag.InternalFlag) { (*f.Rules)[0].Percentages = &map[string]float64{} },
			want: []string{
				"Targeting rule `internal` percentages changed from `null` to `{}` (not used: empty percentages are ignored)",
			},
		},
		{
			name: "should describe a step that sets only the experimentation end",
			update: func(f *flag.InternalFlag) {
				f.Scheduled = &[]flag.ScheduledStep{{
					Date:         &start,
					InternalFlag: flag.InternalFlag{Experimentation: &flag.ExperimentationRollout{End: &end}},
				}}
			},
			want: []string{"Scheduled step on 2026-10-05 09:00 UTC added: Experimentation end set to 2026-10-12 09:00 UTC"},
		},
		{
			name:   "should return nothing for an unchanged flag",
			update: func(_ *flag.InternalFlag) {},
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before, after := baseFlag(), baseFlag()
			if tt.setup != nil {
				tt.setup(before)
				tt.setup(after)
			}
			tt.update(after)
			got, err := notifier.DescribeChanges(before, after)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChangeList(t *testing.T) {
	changes := func(n int) []string {
		list := make([]string, n)
		for i := range list {
			list[i] = fmt.Sprintf("change %d", i+1)
		}
		return list
	}
	tests := []struct {
		name      string
		changes   []string
		maxLength int
		want      string
	}{
		{name: "should list changes with the bullet", changes: changes(2), want: "• change 1\n• change 2"},
		{name: "should list every change at the limit", changes: changes(50), want: notifier.ChangeList(changes(50)[:50], "• ", 0)},
		{
			name:    "should count changes over the limit",
			changes: changes(51),
			want:    notifier.ChangeList(changes(50), "• ", 0) + "\n…and 1 more changes",
		},
		{
			name:      "should drop changes to stay within the maximum length",
			changes:   changes(3),
			maxLength: 31,
			want:      "• change 1\n…and 2 more changes",
		},
		{name: "should return nothing when there is no change", changes: []string{}, want: ""},
		{
			name:      "should return nothing when not even the summary fits",
			changes:   changes(1),
			maxLength: 1,
			want:      "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, notifier.ChangeList(tt.changes, "• ", tt.maxLength))
		})
	}
}

// otherFlag is a flag.Flag that is not a *flag.InternalFlag.
type otherFlag struct{ flag.Flag }

func TestDescribeChanges_unsupportedFlag(t *testing.T) {
	tests := []struct {
		name          string
		before, after flag.Flag
	}{
		{name: "should refuse another flag type before", before: otherFlag{}, after: &flag.InternalFlag{}},
		{name: "should refuse another flag type after", before: &flag.InternalFlag{}, after: otherFlag{}},
		{name: "should refuse a nil flag", before: (*flag.InternalFlag)(nil), after: &flag.InternalFlag{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := notifier.DescribeChanges(tt.before, tt.after)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, "impossible to describe changes")
		})
	}
}

// DescribeChanges has no generic fallback, so a field added to the flag format must get
// its own sentence. This test fails until the new field is handled and listed here.
func TestDescribeChanges_coversEveryField(t *testing.T) {
	tests := []struct {
		name    string
		typ     reflect.Type
		handled []string
	}{
		{
			name: "flag",
			typ:  reflect.TypeFor[flag.InternalFlag](),
			handled: []string{"BucketingKey", "DefaultRule", "Disable", "Experimentation", "Metadata",
				"Rules", "Scheduled", "TrackEvents", "Variations", "Version"},
		},
		{
			name:    "rule",
			typ:     reflect.TypeFor[flag.Rule](),
			handled: []string{"Disable", "Name", "Percentages", "ProgressiveRollout", "Query", "VariationResult"},
		},
		{name: "progressive rollout", typ: reflect.TypeFor[flag.ProgressiveRollout](), handled: []string{"End", "Initial"}},
		{
			name:    "progressive rollout step",
			typ:     reflect.TypeFor[flag.ProgressiveRolloutStep](),
			handled: []string{"Date", "Percentage", "Variation"},
		},
		{name: "experimentation", typ: reflect.TypeFor[flag.ExperimentationRollout](), handled: []string{"End", "Start"}},
		{name: "scheduled step", typ: reflect.TypeFor[flag.ScheduledStep](), handled: []string{"Date", "InternalFlag"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := make([]string, 0, tt.typ.NumField())
			for i := range tt.typ.NumField() {
				if tt.typ.Field(i).IsExported() {
					fields = append(fields, tt.typ.Field(i).Name)
				}
			}
			assert.ElementsMatch(t, tt.handled, fields)
		})
	}
}

func TestDescribeChangesPlain(t *testing.T) {
	tests := []struct {
		name   string
		update func(f *flag.InternalFlag)
		want   []string
	}{
		{
			name: "should describe changes without inline code",
			update: func(f *flag.InternalFlag) {
				f.DefaultRule = &flag.Rule{VariationResult: testconvert.String("on")}
			},
			want: []string{"Default rule now serves on (was off)"},
		},
		{
			name:   "should keep backticks that are part of a value",
			update: func(f *flag.InternalFlag) { f.Metadata = &map[string]any{"owner": "a`"} },
			want:   []string{"Metadata owner changed from \"a\" to \"a`\""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := &flag.InternalFlag{
				DefaultRule: &flag.Rule{VariationResult: testconvert.String("off")},
				Metadata:    &map[string]any{"owner": "a"},
			}
			after := &flag.InternalFlag{
				DefaultRule: &flag.Rule{VariationResult: testconvert.String("off")},
				Metadata:    &map[string]any{"owner": "a"},
			}
			tt.update(after)
			got, err := notifier.DescribeChangesPlain(before, after)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
