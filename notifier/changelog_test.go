package notifier_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

func TestFlagChanges(t *testing.T) {
	start := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	end := start.Add(7 * 24 * time.Hour)
	baseFlag := func() *flag.InternalFlag {
		return &flag.InternalFlag{
			Variations: &map[string]*any{
				"on":  testconvert.Interface(true),
				"off": testconvert.Interface(false),
			},
			DefaultRule: &flag.Rule{VariationResult: testconvert.String("off")},
			Metadata:    &map[string]any{"team": "payments"},
		}
	}

	tests := []struct {
		name   string
		update func(f *flag.InternalFlag)
		want   []notifier.FieldChange
	}{
		{
			name:   "should show a disabled flag without pointer syntax",
			update: func(f *flag.InternalFlag) { f.Disable = testconvert.Bool(true) },
			want:   []notifier.FieldChange{{Path: "Disable", From: "null", To: "true"}},
		},
		{
			name: "should show a new rule as JSON",
			update: func(f *flag.InternalFlag) {
				f.Rules = &[]flag.Rule{{
					Name:            testconvert.String("beta"),
					Query:           testconvert.String(`country eq "FR"`),
					VariationResult: testconvert.String("on"),
				}}
			},
			want: []notifier.FieldChange{{
				Path: "Rules",
				From: "null",
				To:   `[{"name":"beta","query":"country eq \"FR\"","variation":"on"}]`,
			}},
		},
		{
			name: "should show a percentage rollout sorted by path",
			update: func(f *flag.InternalFlag) {
				f.DefaultRule = &flag.Rule{Percentages: &map[string]float64{"on": 10, "off": 90}}
			},
			want: []notifier.FieldChange{
				{Path: "DefaultRule.Percentages", From: "null", To: `{"off":90,"on":10}`},
				{Path: "DefaultRule.VariationResult", From: `"off"`, To: "null"},
			},
		},
		{
			name: "should show dates as RFC 3339",
			update: func(f *flag.InternalFlag) {
				f.Experimentation = &flag.ExperimentationRollout{Start: &start, End: &end}
			},
			want: []notifier.FieldChange{{
				Path: "Experimentation",
				From: "null",
				To:   `{"start":"2026-10-05T09:00:00Z","end":"2026-10-12T09:00:00Z"}`,
			}},
		},
		{
			name: "should report a metadata key that was added",
			update: func(f *flag.InternalFlag) {
				f.Metadata = &map[string]any{"team": "payments", "owner": "jane"}
			},
			want: []notifier.FieldChange{{Path: "Metadata.owner", From: "null", To: `"jane"`}},
		},
		{
			name:   "should report a metadata key that was removed",
			update: func(f *flag.InternalFlag) { f.Metadata = &map[string]any{} },
			want:   []notifier.FieldChange{{Path: "Metadata.team", From: `"payments"`, To: "null"}},
		},
		{
			name: "should report a variation that was added",
			update: func(f *flag.InternalFlag) {
				(*f.Variations)["beta"] = testconvert.Interface("blue")
			},
			want: []notifier.FieldChange{{Path: "Variations.beta", From: "null", To: `"blue"`}},
		},
		{
			name:   "should return nothing when the flag did not change",
			update: func(_ *flag.InternalFlag) {},
			want:   []notifier.FieldChange{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			after := baseFlag()
			tt.update(after)
			assert.Equal(t, tt.want, notifier.FlagChanges(baseFlag(), after))
		})
	}
}

func TestReadableValue(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "nil", value: nil, want: "null"},
		{name: "string pointer", value: testconvert.String("on"), want: `"on"`},
		{name: "map", value: map[string]float64{"on": 50, "off": 50}, want: `{"off":50,"on":50}`},
		{name: "falls back to Go syntax when JSON cannot encode the value", value: make(chan int), want: "(chan int)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, notifier.ReadableValue(tt.value), tt.want)
		})
	}
}
