package flag_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

// newDependencyResolver builds a resolver backed by an in-memory map of flags, mimicking what
// the cache provides to the evaluation engine in the real consumers (root module, relay proxy,
// WASM).
func newDependencyResolver(flags map[string]*flag.InternalFlag) func(string) (flag.Flag, bool) {
	return func(name string) (flag.Flag, bool) {
		f, ok := flags[name]
		if !ok {
			return nil, false
		}
		return f, true
	}
}

// servingFlag returns a flag that always serves the provided value through its default rule.
func servingFlag(value any) *flag.InternalFlag {
	return &flag.InternalFlag{
		Variations:  &map[string]*any{"v": testconvert.Interface(value)},
		DefaultRule: &flag.Rule{VariationResult: testconvert.String("v")},
	}
}

// dependentFlag returns a boolean flag (serving true) that declares the given `needs`.
func dependentFlag(needs []flag.NeedsDependency) flag.InternalFlag {
	return flag.InternalFlag{
		Variations: &map[string]*any{
			"enabled":  testconvert.Interface(true),
			"disabled": testconvert.Interface(false),
		},
		DefaultRule: &flag.Rule{VariationResult: testconvert.String("enabled")},
		Needs:       &needs,
	}
}

func TestInternalFlag_Value_Needs(t *testing.T) {
	tests := []struct {
		name string
		flag flag.InternalFlag
		// siblings are the flags reachable through the dependency resolver.
		siblings map[string]*flag.InternalFlag
		// noResolver leaves the DependencyFlagResolver nil to test the fail-closed behavior.
		noResolver bool
		want       any
		want1      flag.ResolutionDetails
	}{
		{
			name: "dependency met with implicit true",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("payments-enabled")},
			}),
			siblings: map[string]*flag.InternalFlag{
				"payments-enabled": servingFlag(true),
			},
			want: true,
			want1: flag.ResolutionDetails{
				Variant:   "enabled",
				Reason:    flag.ReasonStatic,
				Cacheable: true,
			},
		},
		{
			name: "dependency met with explicit value true",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("payments-enabled"), Value: testconvert.Interface(true)},
			}),
			siblings: map[string]*flag.InternalFlag{
				"payments-enabled": servingFlag(true),
			},
			want: true,
			want1: flag.ResolutionDetails{
				Variant:   "enabled",
				Reason:    flag.ReasonStatic,
				Cacheable: true,
			},
		},
		{
			name: "dependency unmet because value differs",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("payments-enabled"), Value: testconvert.Interface(true)},
			}),
			siblings: map[string]*flag.InternalFlag{
				"payments-enabled": servingFlag(false),
			},
			want: nil,
			want1: flag.ResolutionDetails{
				Variant:   "SdkDefault",
				Reason:    flag.ReasonDisabled,
				Cacheable: true,
				Metadata:  map[string]any{"unsatisfiedDependency": "payments-enabled"},
			},
		},
		{
			name: "dependency unmet because the flag is missing",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("does-not-exist")},
			}),
			siblings: map[string]*flag.InternalFlag{},
			want:     nil,
			want1: flag.ResolutionDetails{
				Variant:   "SdkDefault",
				Reason:    flag.ReasonDisabled,
				Cacheable: true,
				Metadata:  map[string]any{"unsatisfiedDependency": "does-not-exist"},
			},
		},
		{
			name: "dependency met with a string value",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("user-plan"), Value: testconvert.Interface("enterprise")},
			}),
			siblings: map[string]*flag.InternalFlag{
				"user-plan": servingFlag("enterprise"),
			},
			want: true,
			want1: flag.ResolutionDetails{
				Variant:   "enabled",
				Reason:    flag.ReasonStatic,
				Cacheable: true,
			},
		},
		{
			name: "dependency unmet with a string value",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("user-plan"), Value: testconvert.Interface("enterprise")},
			}),
			siblings: map[string]*flag.InternalFlag{
				"user-plan": servingFlag("free"),
			},
			want: nil,
			want1: flag.ResolutionDetails{
				Variant:   "SdkDefault",
				Reason:    flag.ReasonDisabled,
				Cacheable: true,
				Metadata:  map[string]any{"unsatisfiedDependency": "user-plan"},
			},
		},
		{
			name: "multiple dependencies all met (AND)",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("beta-program"), Value: testconvert.Interface(true)},
				{Flag: testconvert.String("user-plan"), Value: testconvert.Interface("enterprise")},
			}),
			siblings: map[string]*flag.InternalFlag{
				"beta-program": servingFlag(true),
				"user-plan":    servingFlag("enterprise"),
			},
			want: true,
			want1: flag.ResolutionDetails{
				Variant:   "enabled",
				Reason:    flag.ReasonStatic,
				Cacheable: true,
			},
		},
		{
			name: "multiple dependencies with one unmet (AND)",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("beta-program"), Value: testconvert.Interface(true)},
				{Flag: testconvert.String("user-plan"), Value: testconvert.Interface("enterprise")},
			}),
			siblings: map[string]*flag.InternalFlag{
				"beta-program": servingFlag(true),
				"user-plan":    servingFlag("free"),
			},
			want: nil,
			want1: flag.ResolutionDetails{
				Variant:   "SdkDefault",
				Reason:    flag.ReasonDisabled,
				Cacheable: true,
				Metadata:  map[string]any{"unsatisfiedDependency": "user-plan"},
			},
		},
		{
			name: "no resolver available fails closed",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("payments-enabled")},
			}),
			noResolver: true,
			want:       nil,
			want1: flag.ResolutionDetails{
				Variant:   "SdkDefault",
				Reason:    flag.ReasonDisabled,
				Cacheable: true,
				Metadata:  map[string]any{"unsatisfiedDependency": "payments-enabled"},
			},
		},
		{
			name: "disabled dependency is treated as unmet",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("payments-enabled")},
			}),
			siblings: map[string]*flag.InternalFlag{
				"payments-enabled": {
					Disable:     testconvert.Bool(true),
					Variations:  &map[string]*any{"v": testconvert.Interface(true)},
					DefaultRule: &flag.Rule{VariationResult: testconvert.String("v")},
				},
			},
			want: nil,
			want1: flag.ResolutionDetails{
				Variant:   "SdkDefault",
				Reason:    flag.ReasonDisabled,
				Cacheable: true,
				Metadata:  map[string]any{"unsatisfiedDependency": "payments-enabled"},
			},
		},
		{
			name: "dependency whose experimentation is over is treated as unmet",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("payments-enabled")},
			}),
			siblings: map[string]*flag.InternalFlag{
				"payments-enabled": {
					Variations:  &map[string]*any{"v": testconvert.Interface(true)},
					DefaultRule: &flag.Rule{VariationResult: testconvert.String("v")},
					// The experimentation window is entirely in the past, so the dependency serves
					// its (disabled) default value instead of true, which makes the need unmet.
					Experimentation: &flag.ExperimentationRollout{
						Start: testconvert.Time(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
						End:   testconvert.Time(time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
			// The dependency is time-based (experimentation), so the disabled result must not be
			// cacheable even though the dependent flag itself is static.
			want: nil,
			want1: flag.ResolutionDetails{
				Variant:   "SdkDefault",
				Reason:    flag.ReasonDisabled,
				Cacheable: false,
				Metadata:  map[string]any{"unsatisfiedDependency": "payments-enabled"},
			},
		},
		{
			name: "met dependency that is time-based makes the enabled result non-cacheable",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("payments-enabled")},
			}),
			siblings: map[string]*flag.InternalFlag{
				"payments-enabled": {
					Variations:  &map[string]*any{"v": testconvert.Interface(true)},
					DefaultRule: &flag.Rule{VariationResult: testconvert.String("v")},
					// Experimentation is currently running (past start, future end), so the
					// dependency serves true but its result is not cacheable; the dependent flag's
					// enabled result must therefore also be non-cacheable.
					Experimentation: &flag.ExperimentationRollout{
						Start: testconvert.Time(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
						End:   testconvert.Time(time.Date(2999, 12, 31, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
			want: true,
			want1: flag.ResolutionDetails{
				Variant:   "enabled",
				Reason:    flag.ReasonStatic,
				Cacheable: false,
			},
		},
		{
			name: "disabled dependency does not satisfy an explicit value:null need (fail closed)",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("payments-enabled"), Value: testconvert.Interface(nil)},
			}),
			siblings: map[string]*flag.InternalFlag{
				"payments-enabled": {
					Disable:     testconvert.Bool(true),
					Variations:  &map[string]*any{"v": testconvert.Interface(true)},
					DefaultRule: &flag.Rule{VariationResult: testconvert.String("v")},
				},
			},
			want: nil,
			want1: flag.ResolutionDetails{
				Variant:   "SdkDefault",
				Reason:    flag.ReasonDisabled,
				Cacheable: true,
				Metadata:  map[string]any{"unsatisfiedDependency": "payments-enabled"},
			},
		},
		{
			name: "numeric value is compared regardless of int/float type",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("level"), Value: testconvert.Interface(1)},
			}),
			siblings: map[string]*flag.InternalFlag{
				"level": servingFlag(float64(1)),
			},
			want: true,
			want1: flag.ResolutionDetails{
				Variant:   "enabled",
				Reason:    flag.ReasonStatic,
				Cacheable: true,
			},
		},
		{
			name: "dependency own needs are ignored (one level only / cycle safe)",
			flag: dependentFlag([]flag.NeedsDependency{
				{Flag: testconvert.String("flag-b"), Value: testconvert.Interface(true)},
			}),
			siblings: map[string]*flag.InternalFlag{
				// flag-b serves true but declares a needs on the flag under test. Because
				// dependencies are resolved one level deep, flag-b's own needs is ignored and
				// it resolves to true, so the dependent flag is evaluated normally.
				"flag-b": {
					Variations:  &map[string]*any{"v": testconvert.Interface(true)},
					DefaultRule: &flag.Rule{VariationResult: testconvert.String("v")},
					Needs: &[]flag.NeedsDependency{
						{Flag: testconvert.String("flag-a"), Value: testconvert.Interface(true)},
					},
				},
			},
			want: true,
			want1: flag.ResolutionDetails{
				Variant:   "enabled",
				Reason:    flag.ReasonStatic,
				Cacheable: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagContext := flag.Context{DefaultSdkValue: nil}
			if !tt.noResolver {
				flagContext.DependencyFlagResolver = newDependencyResolver(tt.siblings)
			}
			got, got1 := tt.flag.Value("flag-under-test", ffcontext.NewEvaluationContext("user-key"), flagContext)
			assert.Equalf(t, tt.want, got, "not expected value: %s", cmp.Diff(tt.want, got))
			assert.Equalf(t, tt.want1, got1, "not expected details: %s", cmp.Diff(tt.want1, got1))
		})
	}
}
