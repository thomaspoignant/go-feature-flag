package flag_test

import (
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

func TestInternalFlag_Value_Dependencies(t *testing.T) {
	dep := func(parent string, variation, notVariation *string, appliesTo ...string) flag.FlagDependency {
		return flag.FlagDependency{Flag: parent, Variation: variation, NotVariation: notVariation, AppliesTo: appliesTo}
	}
	disabled := true
	expired := time.Now().Add(-time.Hour)

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
			name:  "expired parent is not met",
			child: boolFlag("enabled", dep("parent", strPtr("enabled"), nil)),
			flags: map[string]flag.Flag{"parent": &flag.InternalFlag{
				Variations: boolVariations(), ExpiresAt: &expired,
				DefaultRule: &flag.Rule{VariationResult: strPtr("enabled")},
			}},
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
			wantValue: false, wantVariant: flag.VariationSDKDefault, wantErrorCode: flag.ErrorCodeFlagDependency,
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

func TestInternalFlag_IsExpired_Dependencies(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)
	farFuture := now.Add(24 * time.Hour)
	follows := func(follow bool, parent string) *[]flag.FlagDependency {
		return &[]flag.FlagDependency{{Flag: parent, Variation: strPtr("enabled"), ExpiresWithParent: follow}}
	}

	tests := []struct {
		name  string
		child flag.InternalFlag
		flags map[string]flag.Flag
		want  bool
	}{
		{
			name:  "parent expired, child follows",
			child: flag.InternalFlag{ExpiresAt: &farFuture, DependsOn: follows(true, "parent")},
			flags: map[string]flag.Flag{"parent": &flag.InternalFlag{ExpiresAt: &past}},
			want:  true,
		},
		{
			name:  "parent expired, child not configured to follow",
			child: flag.InternalFlag{ExpiresAt: &farFuture, DependsOn: follows(false, "parent")},
			flags: map[string]flag.Flag{"parent": &flag.InternalFlag{ExpiresAt: &past}},
		},
		{
			name:  "child expired before parent",
			child: flag.InternalFlag{ExpiresAt: &past, DependsOn: follows(true, "parent")},
			flags: map[string]flag.Flag{"parent": &flag.InternalFlag{ExpiresAt: &farFuture}},
			want:  true,
		},
		{
			name:  "parent not expired yet",
			child: flag.InternalFlag{ExpiresAt: &farFuture, DependsOn: follows(true, "parent")},
			flags: map[string]flag.Flag{"parent": &flag.InternalFlag{ExpiresAt: &future}},
		},
		{
			name:  "child without own expiration follows parent",
			child: flag.InternalFlag{DependsOn: follows(true, "parent")},
			flags: map[string]flag.Flag{"parent": &flag.InternalFlag{ExpiresAt: &past}},
			want:  true,
		},
		{
			name:  "grandparent expiration cascades",
			child: flag.InternalFlag{DependsOn: follows(true, "parent")},
			flags: map[string]flag.Flag{
				"parent":      &flag.InternalFlag{DependsOn: follows(true, "grandparent")},
				"grandparent": &flag.InternalFlag{ExpiresAt: &past},
			},
			want: true,
		},
		{
			name:  "cascade stops at a flag not following its parent",
			child: flag.InternalFlag{DependsOn: follows(true, "parent")},
			flags: map[string]flag.Flag{
				"parent":      &flag.InternalFlag{DependsOn: follows(false, "grandparent")},
				"grandparent": &flag.InternalFlag{ExpiresAt: &past},
			},
		},
		{
			name:  "missing parent is ignored",
			child: flag.InternalFlag{DependsOn: follows(true, "parent")},
			flags: map[string]flag.Flag{},
		},
		{
			name:  "circular expiration dependency terminates",
			child: flag.InternalFlag{DependsOn: follows(true, "other")},
			flags: map[string]flag.Flag{
				"other": &flag.InternalFlag{DependsOn: follows(true, "child"), ExpiresAt: &past},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.flags["child"] = &tt.child
			assert.Equal(t, tt.want, tt.child.IsExpired(now, resolverOf(tt.flags)))
		})
	}
}

func TestInternalFlag_Value_ExpiredParentExpiresChild(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	child := boolFlag("enabled")
	child.DependsOn = ptrTo([]flag.FlagDependency{{Flag: "parent", Variation: strPtr("enabled"), ExpiresWithParent: true}})
	flags := map[string]flag.Flag{"parent": &flag.InternalFlag{ExpiresAt: &past}, "child": child}

	got, details := child.Value("child", ffcontext.NewEvaluationContext("user"), resolverOf(flags))
	assert.Equal(t, false, got)
	assert.Equal(t, flag.ErrorCodeFlagExpired, details.ErrorCode)
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
