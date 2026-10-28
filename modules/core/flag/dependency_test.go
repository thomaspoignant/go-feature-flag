package flag_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

func strPtr(s string) *string { return &s }

func ptrTo[T any](v T) *T { return &v }

func boolVariations() *map[string]*any {
	t, f := any(true), any(false)
	return &map[string]*any{"enabled": &t, "disabled": &f}
}

// boolFlag builds a flag always serving the given variation.
func boolFlag(variation string, deps ...flag.FlagDependency) *flag.InternalFlag {
	f := &flag.InternalFlag{
		Variations:  boolVariations(),
		DefaultRule: &flag.Rule{VariationResult: strPtr(variation)},
	}
	if len(deps) > 0 {
		f.DependsOn = &deps
	}
	return f
}

func resolverOf(flags map[string]flag.Flag) flag.Context {
	return flag.Context{
		DefaultSdkValue: false,
		GetFlag: func(name string) (flag.Flag, bool) {
			f, ok := flags[name]
			return f, ok
		},
	}
}

func dep(parent string, variation, notVariation *string, appliesTo ...string) flag.FlagDependency {
	return flag.FlagDependency{Flag: parent, Variation: variation, NotVariation: notVariation, AppliesTo: appliesTo}
}

func TestInternalFlag_Value_Dependencies(t *testing.T) {
	disabled := true
	past := time.Now().Add(-time.Hour)

	// scheduledDisabledParent returns a parent flag disabled by a due scheduled rollout step.
	scheduledDisabledParent := func() *flag.InternalFlag {
		return &flag.InternalFlag{
			Variations:  boolVariations(),
			DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
			Scheduled: &[]flag.ScheduledStep{{
				InternalFlag: flag.InternalFlag{Disable: &disabled},
				Date:         &past,
			}},
		}
	}

	tests := []struct {
		name          string
		child         *flag.InternalFlag
		flags         map[string]flag.Flag
		noResolver    bool
		wantValue     any
		wantVariant   string
		wantErrorCode flag.ErrorCode
		wantCacheable bool
	}{
		{
			name:  "no dependency",
			child: boolFlag("enabled"), flags: map[string]flag.Flag{},
			wantValue: true, wantVariant: "enabled", wantCacheable: true,
		},
		{
			name:      "variation met",
			child:     boolFlag("enabled", dep("parent", strPtr("enabled"), nil)),
			flags:     map[string]flag.Flag{"parent": boolFlag("enabled")},
			wantValue: true, wantVariant: "enabled",
		},
		{
			name:      "variation not met",
			child:     boolFlag("enabled", dep("parent", strPtr("enabled"), nil)),
			flags:     map[string]flag.Flag{"parent": boolFlag("disabled")},
			wantValue: false, wantVariant: flag.VariationSDKDefault, wantErrorCode: flag.ErrorCodeFlagDependency,
		},
		{
			name:      "notVariation met",
			child:     boolFlag("enabled", dep("parent", nil, strPtr("disabled"))),
			flags:     map[string]flag.Flag{"parent": boolFlag("enabled")},
			wantValue: true, wantVariant: "enabled",
		},
		{
			name:      "notVariation not met",
			child:     boolFlag("enabled", dep("parent", nil, strPtr("disabled"))),
			flags:     map[string]flag.Flag{"parent": boolFlag("disabled")},
			wantValue: false, wantVariant: flag.VariationSDKDefault, wantErrorCode: flag.ErrorCodeFlagDependency,
		},
		{
			name:      "unmet dependency ignored when it applies to another variation",
			child:     boolFlag("enabled", dep("parent", strPtr("enabled"), nil, "disabled")),
			flags:     map[string]flag.Flag{"parent": boolFlag("disabled")},
			wantValue: true, wantVariant: "enabled",
		},
		{
			name:      "unmet dependency applies to the selected variation",
			child:     boolFlag("enabled", dep("parent", strPtr("enabled"), nil, "enabled")),
			flags:     map[string]flag.Flag{"parent": boolFlag("disabled")},
			wantValue: false, wantVariant: flag.VariationSDKDefault, wantErrorCode: flag.ErrorCodeFlagDependency,
		},
		{
			name: "all dependencies must be met",
			child: boolFlag("enabled",
				dep("p1", strPtr("enabled"), nil), dep("p2", strPtr("enabled"), nil)),
			flags:     map[string]flag.Flag{"p1": boolFlag("enabled"), "p2": boolFlag("disabled")},
			wantValue: false, wantVariant: flag.VariationSDKDefault, wantErrorCode: flag.ErrorCodeFlagDependency,
		},
		{
			name:      "missing parent",
			child:     boolFlag("enabled", dep("parent", strPtr("enabled"), nil)),
			flags:     map[string]flag.Flag{},
			wantValue: false, wantVariant: flag.VariationSDKDefault, wantErrorCode: flag.ErrorCodeFlagDependency,
		},
		{
			name:       "no resolver",
			child:      boolFlag("enabled", dep("parent", strPtr("enabled"), nil)),
			noResolver: true,
			wantValue:  false, wantVariant: flag.VariationSDKDefault, wantErrorCode: flag.ErrorCodeFlagDependency,
		},
		{
			name:  "disabled parent is not met",
			child: boolFlag("enabled", dep("parent", nil, strPtr("disabled"))),
			flags: map[string]flag.Flag{"parent": &flag.InternalFlag{
				Variations: boolVariations(), Disable: &disabled,
				DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
			}},
			wantValue: false, wantVariant: flag.VariationSDKDefault, wantErrorCode: flag.ErrorCodeFlagDependency,
		},
		{
			name:      "parent disabled by a scheduled rollout step is not met",
			child:     boolFlag("enabled", dep("parent", strPtr("enabled"), nil)),
			flags:     map[string]flag.Flag{"parent": scheduledDisabledParent()},
			wantValue: false, wantVariant: flag.VariationSDKDefault, wantErrorCode: flag.ErrorCodeFlagDependency,
		},
		{
			name:  "chained dependencies met",
			child: boolFlag("enabled", dep("parent", strPtr("enabled"), nil)),
			flags: map[string]flag.Flag{
				"parent":      boolFlag("enabled", dep("grandparent", strPtr("enabled"), nil)),
				"grandparent": boolFlag("enabled"),
			},
			wantValue: true, wantVariant: "enabled",
		},
		{
			name:  "chained dependencies broken at the top",
			child: boolFlag("enabled", dep("parent", strPtr("enabled"), nil)),
			flags: map[string]flag.Flag{
				"parent":      boolFlag("enabled", dep("grandparent", strPtr("enabled"), nil)),
				"grandparent": boolFlag("disabled"),
			},
			wantValue: false, wantVariant: flag.VariationSDKDefault, wantErrorCode: flag.ErrorCodeFlagDependency,
		},
		{
			name:  "circular dependency",
			child: boolFlag("enabled", dep("other", strPtr("enabled"), nil)),
			flags: map[string]flag.Flag{
				"other": boolFlag("enabled", dep("child", strPtr("enabled"), nil)),
			},
			wantValue: false, wantVariant: flag.VariationSDKDefault, wantErrorCode: flag.ErrorFlagConfiguration,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagCtx := resolverOf(tt.flags)
			if tt.flags != nil {
				tt.flags["child"] = tt.child
			}
			if tt.noResolver {
				flagCtx.GetFlag = nil
			}
			got, details := tt.child.Value("child", ffcontext.NewEvaluationContext("user"), flagCtx)
			assert.Equal(t, tt.wantValue, got)
			assert.Equal(t, tt.wantVariant, details.Variant)
			assert.Equal(t, tt.wantErrorCode, details.ErrorCode)
			assert.Equal(t, tt.wantCacheable, details.Cacheable)
			if tt.wantErrorCode != "" {
				assert.Equal(t, flag.ReasonError, details.Reason)
				assert.NotEmpty(t, details.ErrorMessage)
			}
		})
	}
}

func TestInternalFlag_Value_SelfDependencyIsAConfigError(t *testing.T) {
	flags := map[string]flag.Flag{
		"a": boolFlag("enabled", flag.FlagDependency{Flag: "a", Variation: strPtr("enabled")}),
	}
	got, details := flags["a"].Value("a", ffcontext.NewEvaluationContext("user"), resolverOf(flags))
	assert.Equal(t, false, got)
	assert.Equal(t, flag.ErrorFlagConfiguration, details.ErrorCode)
}

// evaluationCtxAt returns an evaluation context evaluated at the given date.
func evaluationCtxAt(date time.Time) ffcontext.Context {
	ctx := ffcontext.NewEvaluationContext("user")
	ctx.AddCustomAttribute("gofeatureflag", map[string]any{"currentDateTime": date})
	return ctx
}

func TestInternalFlag_Value_ParentScheduleDisablesChild(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)
	disabled := true
	follows := func(follow bool, parent string) *[]flag.FlagDependency {
		return &[]flag.FlagDependency{{Flag: parent, Variation: strPtr("enabled"), ExpiresWithParent: follow}}
	}
	disableStepAt := func(date time.Time) *[]flag.ScheduledStep {
		return &[]flag.ScheduledStep{{
			InternalFlag: flag.InternalFlag{Disable: &disabled},
			Date:         &date,
		}}
	}

	tests := []struct {
		name  string
		child *flag.InternalFlag
		flags map[string]flag.Flag
		want  flag.ResolutionReason
	}{
		{
			name:  "parent disabled by a due step without expiresWithParent is a dependency error",
			child: boolFlag("enabled", flag.FlagDependency{Flag: "parent", Variation: strPtr("enabled")}),
			flags: map[string]flag.Flag{"parent": &flag.InternalFlag{
				Variations: boolVariations(), Scheduled: disableStepAt(past),
				DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
			}},
			want: flag.ReasonError,
		},
		{
			name:  "parent disabled by a due step, child follows",
			child: boolFlag("enabled", flag.FlagDependency{Flag: "parent", Variation: strPtr("enabled"), ExpiresWithParent: true}),
			flags: map[string]flag.Flag{"parent": &flag.InternalFlag{
				Variations: boolVariations(), Scheduled: disableStepAt(past),
				DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
			}},
			want: flag.ReasonDisabled,
		},
		{
			name:  "parent disabled by its disable field, child follows",
			child: boolFlag("enabled", flag.FlagDependency{Flag: "parent", Variation: strPtr("enabled"), ExpiresWithParent: true}),
			flags: map[string]flag.Flag{"parent": &flag.InternalFlag{
				Variations: boolVariations(), Disable: &disabled,
				DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
			}},
			want: flag.ReasonDisabled,
		},
		{
			name:  "parent not disabled yet",
			child: boolFlag("enabled", flag.FlagDependency{Flag: "parent", Variation: strPtr("enabled"), ExpiresWithParent: true}),
			flags: map[string]flag.Flag{"parent": &flag.InternalFlag{
				Variations: boolVariations(), Scheduled: disableStepAt(future),
				DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
			}},
			want: flag.ReasonStatic,
		},
		{
			name:  "child not following still gets the dependency error",
			child: boolFlag("enabled", flag.FlagDependency{Flag: "parent", Variation: strPtr("enabled"), ExpiresWithParent: false}),
			flags: map[string]flag.Flag{"parent": &flag.InternalFlag{
				Variations: boolVariations(), Scheduled: disableStepAt(past),
				DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
			}},
			want: flag.ReasonError,
		},
		{
			name:  "grandparent schedule cascades",
			child: boolFlag("enabled", flag.FlagDependency{Flag: "parent", Variation: strPtr("enabled"), ExpiresWithParent: true}),
			flags: map[string]flag.Flag{
				"parent": &flag.InternalFlag{
					Variations:  boolVariations(),
					DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
					DependsOn:   follows(true, "grandparent"),
				},
				"grandparent": &flag.InternalFlag{
					Variations: boolVariations(), Scheduled: disableStepAt(past),
					DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
				},
			},
			want: flag.ReasonDisabled,
		},
		{
			name:  "a disabled grandparent breaks the chain even without expiresWithParent",
			child: boolFlag("enabled", flag.FlagDependency{Flag: "parent", Variation: strPtr("enabled"), ExpiresWithParent: true}),
			flags: map[string]flag.Flag{
				"parent": &flag.InternalFlag{
					Variations:  boolVariations(),
					DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
					DependsOn:   follows(false, "grandparent"),
				},
				"grandparent": &flag.InternalFlag{
					Variations: boolVariations(), Scheduled: disableStepAt(past),
					DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
				},
			},
			want: flag.ReasonError,
		},
		{
			name:  "missing parent is a dependency error",
			child: boolFlag("enabled", flag.FlagDependency{Flag: "parent", Variation: strPtr("enabled"), ExpiresWithParent: true}),
			flags: map[string]flag.Flag{},
			want:  flag.ReasonError,
		},
		{
			name:  "circular schedule dependency terminates",
			child: boolFlag("enabled", flag.FlagDependency{Flag: "other", Variation: strPtr("enabled"), ExpiresWithParent: true}),
			flags: map[string]flag.Flag{
				"other": &flag.InternalFlag{
					Variations: boolVariations(), Scheduled: disableStepAt(past),
					DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
					DependsOn:   follows(true, "child"),
				},
			},
			want: flag.ReasonDisabled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.flags["child"] = tt.child
			got, details := tt.child.Value("child", evaluationCtxAt(now), resolverOf(tt.flags))
			assert.Equal(t, tt.want, details.Reason)
			if tt.want == flag.ReasonDisabled {
				assert.Equal(t, false, got)
				assert.Equal(t, flag.VariationSDKDefault, details.Variant)
			}
		})
	}
}

func TestInternalFlag_Value_ParentScheduleDisablesUngatedVariations(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	disabled := true
	child := boolFlag("enabled", flag.FlagDependency{
		Flag: "parent", Variation: strPtr("enabled"), AppliesTo: []string{"disabled"}, ExpiresWithParent: true,
	})
	flags := map[string]flag.Flag{
		"parent": &flag.InternalFlag{
			Variations: boolVariations(), Scheduled: &[]flag.ScheduledStep{{
				InternalFlag: flag.InternalFlag{Disable: &disabled}, Date: &past,
			}},
			DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
		},
		"child": child,
	}

	// the variation "enabled" of the child is not gated by the dependency, but the
	// parent is disabled, the whole child is disabled.
	got, details := child.Value("child", evaluationCtxAt(now), resolverOf(flags))
	assert.Equal(t, false, got)
	assert.Equal(t, flag.VariationSDKDefault, details.Variant)
	assert.Equal(t, flag.ReasonDisabled, details.Reason)
	assert.Empty(t, details.ErrorCode)
}

func TestInternalFlag_IsValid_Dependencies(t *testing.T) {
	tests := []struct {
		name    string
		dep     flag.FlagDependency
		wantErr bool
	}{
		{name: "variation", dep: flag.FlagDependency{Flag: "p", Variation: strPtr("enabled")}},
		{name: "notVariation", dep: flag.FlagDependency{Flag: "p", NotVariation: strPtr("enabled")}},
		{name: "appliesTo known variation", dep: flag.FlagDependency{Flag: "p", Variation: strPtr("x"), AppliesTo: []string{"enabled"}}},
		{name: "missing flag name", dep: flag.FlagDependency{Variation: strPtr("enabled")}, wantErr: true},
		{name: "no condition", dep: flag.FlagDependency{Flag: "p"}, wantErr: true},
		{name: "both conditions", dep: flag.FlagDependency{Flag: "p", Variation: strPtr("a"), NotVariation: strPtr("b")}, wantErr: true},
		{name: "appliesTo unknown variation", dep: flag.FlagDependency{Flag: "p", Variation: strPtr("a"), AppliesTo: []string{"nope"}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := boolFlag("enabled", tt.dep)
			if tt.wantErr {
				assert.Error(t, f.IsValid())
				return
			}
			assert.NoError(t, f.IsValid())
		})
	}
}

func TestInternalFlag_Value_CycleErrorIsKeptThroughTheChain(t *testing.T) {
	flags := map[string]flag.Flag{
		"a": boolFlag("enabled", dep("b", strPtr("enabled"), nil)),
		"b": boolFlag("enabled", dep("c", strPtr("enabled"), nil)),
		"c": boolFlag("enabled", dep("a", strPtr("enabled"), nil)),
	}
	got, details := flags["a"].Value("a", ffcontext.NewEvaluationContext("user"), resolverOf(flags))
	assert.Equal(t, false, got)
	assert.Equal(t, flag.ErrorFlagConfiguration, details.ErrorCode)
	assert.Contains(t, details.ErrorMessage, "circular dependency")
}

func TestInternalFlag_Value_SharedParentsAreResolvedOnce(t *testing.T) {
	const layers = 12
	flags := map[string]flag.Flag{}
	name := func(layer, i int) string { return fmt.Sprintf("flag-%d-%d", layer, i) }
	for layer := 0; layer < layers; layer++ {
		for i := 0; i < 2; i++ {
			if layer == layers-1 {
				flags[name(layer, i)] = boolFlag("enabled")
				continue
			}
			flags[name(layer, i)] = boolFlag("enabled",
				dep(name(layer+1, 0), strPtr("enabled"), nil),
				dep(name(layer+1, 1), strPtr("enabled"), nil))
		}
	}
	lookups := 0
	flagCtx := flag.Context{
		DefaultSdkValue: false,
		GetFlag: func(n string) (flag.Flag, bool) {
			lookups++
			f, ok := flags[n]
			return f, ok
		},
	}
	got, details := flags[name(0, 0)].Value(name(0, 0), ffcontext.NewEvaluationContext("user"), flagCtx)
	assert.Equal(t, true, got)
	assert.Empty(t, details.ErrorCode)
	assert.Less(t, lookups, 4*layers)
}

func TestInternalFlag_Value_SharedExpiringParentsAreResolvedOnce(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	disabled := true
	follows := ptrTo([]flag.FlagDependency{
		{Flag: "left", Variation: strPtr("enabled"), ExpiresWithParent: true},
		{Flag: "right", Variation: strPtr("enabled"), ExpiresWithParent: true},
	})
	scheduledDisabled := &flag.InternalFlag{
		Variations: boolVariations(), Scheduled: &[]flag.ScheduledStep{{
			InternalFlag: flag.InternalFlag{Disable: &disabled}, Date: &past,
		}},
		DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
	}
	flags := map[string]flag.Flag{
		"left":   &flag.InternalFlag{Variations: boolVariations(), DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")}, DependsOn: ptrTo([]flag.FlagDependency{{Flag: "root", ExpiresWithParent: true}})},
		"right":  &flag.InternalFlag{Variations: boolVariations(), DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")}, DependsOn: ptrTo([]flag.FlagDependency{{Flag: "root", ExpiresWithParent: true}})},
		"root":   scheduledDisabled,
		"parent": &flag.InternalFlag{Variations: boolVariations(), DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")}, DependsOn: follows},
	}
	rootLookups := 0
	flagCtx := flag.Context{GetFlag: func(n string) (flag.Flag, bool) {
		if n == "root" {
			rootLookups++
		}
		f, ok := flags[n]
		return f, ok
	}}
	_, details := flags["parent"].(*flag.InternalFlag).Value("parent", evaluationCtxAt(now), flagCtx)
	assert.Equal(t, flag.ReasonDisabled, details.Reason)
	assert.LessOrEqual(t, rootLookups, 2)
}
