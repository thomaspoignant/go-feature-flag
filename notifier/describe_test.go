package notifier_test

import (
	"fmt"
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
				"Scheduled rollout added: 2 steps on 2026-10-05 09:00 UTC, 2026-10-12 09:00 UTC",
			},
		},
		{
			name: "should fall back to field changes for an edited scheduled rollout",
			setup: func(f *flag.InternalFlag) {
				f.Scheduled = &[]flag.ScheduledStep{
					{Date: &start, InternalFlag: flag.InternalFlag{Disable: testconvert.Bool(true)}},
				}
			},
			update: func(f *flag.InternalFlag) {
				f.Scheduled = &[]flag.ScheduledStep{
					{Date: &end, InternalFlag: flag.InternalFlag{Disable: testconvert.Bool(true)}},
				}
			},
			want: []string{"Scheduled.0.Date: \"2026-10-05T09:00:00Z\" => \"2026-10-12T09:00:00Z\""},
		},
		{
			name:   "should fall back to field changes when no sentence covers the change",
			update: func(f *flag.InternalFlag) { f.DefaultRule.Name = testconvert.String("default") },
			want:   []string{"DefaultRule.Name: null => \"default\""},
		},
		{
			name:   "should return nothing for an unchanged flag",
			update: func(_ *flag.InternalFlag) {},
			want:   []string{},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, notifier.ChangeList(tt.changes, "• ", tt.maxLength))
		})
	}
}
