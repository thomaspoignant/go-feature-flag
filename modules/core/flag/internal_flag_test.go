package flag_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestInternalFlag_Value(t *testing.T) {
	type args struct {
		flagName    string
		user        ffcontext.Context
		flagContext flag.Context
	}
	tests := []struct {
		name  string
		flag  flag.InternalFlag
		args  args
		want  interface{}
		want1 flag.ResolutionDetails
	}{
		{
			name: "Should use default value if no targeting",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface(true),
					"variation_B": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContext("user-key"),
				flagContext: flag.Context{
					DefaultSdkValue: false,
				},
			},
			want: true,
			want1: flag.ResolutionDetails{
				Variant:   "variation_A",
				Reason:    flag.ReasonStatic,
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should use default value if percentages are empty",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface(true),
					"variation_B": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
					Percentages:     &map[string]float64{},
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContext("user-key"),
				flagContext: flag.Context{
					DefaultSdkValue: false,
				},
			},
			want: true,
			want1: flag.ResolutionDetails{
				Variant:   "variation_A",
				Reason:    flag.ReasonStatic,
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should return sdk default value when flag is disabled",
			flag: flag.InternalFlag{
				Disable: testconvert.Bool(true),
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContext("user-key"),
				flagContext: flag.Context{
					DefaultSdkValue: "default-sdk",
				},
			},
			want: "default-sdk",
			want1: flag.ResolutionDetails{
				Variant:   "SdkDefault",
				Reason:    flag.ReasonDisabled,
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should return sdk default value when experimentation rollout not started",
			flag: flag.InternalFlag{
				Experimentation: &flag.ExperimentationRollout{
					Start: testconvert.Time(time.Now().Add(1 * time.Second)),
					End:   testconvert.Time(time.Now().Add(5 * time.Second)),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContext("user-key"),
				flagContext: flag.Context{
					DefaultSdkValue: "default-sdk",
				},
			},
			want: "default-sdk",
			want1: flag.ResolutionDetails{
				Variant:   "SdkDefault",
				Reason:    flag.ReasonDisabled,
				Cacheable: false,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should return sdk default value when experimentation rollout is now, but context date override",
			flag: flag.InternalFlag{
				Experimentation: &flag.ExperimentationRollout{
					Start: testconvert.Time(time.Now().Add(-5 * time.Second)),
					End:   testconvert.Time(time.Now().Add(5 * time.Second)),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user: ffcontext.NewEvaluationContextBuilder("user-key").
					AddCustom("gofeatureflag", map[string]interface{}{
						"currentDateTime": time.Now().Add(6 * time.Second).Format(time.RFC3339),
					}).
					Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "default-sdk",
				},
			},
			want: "default-sdk",
			want1: flag.ResolutionDetails{
				Variant:   "SdkDefault",
				Reason:    flag.ReasonDisabled,
				Cacheable: false,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should return sdk default value when experimentation rollout ended",
			flag: flag.InternalFlag{
				Experimentation: &flag.ExperimentationRollout{
					Start: testconvert.Time(time.Now().Add(-15 * time.Second)),
					End:   testconvert.Time(time.Now().Add(-5 * time.Second)),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContext("user-key"),
				flagContext: flag.Context{
					DefaultSdkValue: "default-sdk",
				},
			},
			want: "default-sdk",
			want1: flag.ResolutionDetails{
				Variant:   "SdkDefault",
				Reason:    flag.ReasonDisabled,
				Cacheable: false,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should return the variation specified in the rule if rule match",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface(true),
					"variation_B": testconvert.Interface(false),
				},
				Rules: &[]flag.Rule{
					{
						Name:            testconvert.String("rule1"),
						Query:           testconvert.String("key eq \"user-key\""),
						VariationResult: testconvert.String("variation_B"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContext("user-key"),
				flagContext: flag.Context{
					DefaultSdkValue: false,
				},
			},
			want: false,
			want1: flag.ResolutionDetails{
				Variant:   "variation_B",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(0),
				RuleName:  testconvert.String("rule1"),
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description":       "this is a flag",
					"issue-link":        "https://issue.link/GOFF-1",
					"evaluatedRuleName": "rule1",
				},
			},
		},
		{
			name: "Should return the variation specified in the rule if rule match (jsonLogic)",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface(true),
					"variation_B": testconvert.Interface(false),
				},
				Rules: &[]flag.Rule{
					{
						Name:            testconvert.String("rule1"),
						Query:           testconvert.String(`{"==":[{"var":"key"},"user-key"]}`),
						VariationResult: testconvert.String("variation_B"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContext("user-key"),
				flagContext: flag.Context{
					DefaultSdkValue: false,
				},
			},
			want: false,
			want1: flag.ResolutionDetails{
				Variant:   "variation_B",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(0),
				RuleName:  testconvert.String("rule1"),
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description":       "this is a flag",
					"issue-link":        "https://issue.link/GOFF-1",
					"evaluatedRuleName": "rule1",
				},
			},
		},
		{
			name: "Should match the 2nd rule",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
				},
				Rules: &[]flag.Rule{
					{
						Name:            testconvert.String("rule1"),
						Query:           testconvert.String("key eq \"not-user-key\""),
						VariationResult: testconvert.String("variation_C"),
					},
					{
						Name:            testconvert.String("rule2"),
						Query:           testconvert.String("key eq \"user-key\""),
						VariationResult: testconvert.String("variation_C"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContext("user-key"),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_C",
			want1: flag.ResolutionDetails{
				Variant:   "variation_C",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(1),
				RuleName:  testconvert.String("rule2"),
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description":       "this is a flag",
					"issue-link":        "https://issue.link/GOFF-1",
					"evaluatedRuleName": "rule2",
				},
			},
		},
		{
			name: "Should match the 2nd rule (jsonLogic)",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
				},
				Rules: &[]flag.Rule{
					{
						Name: testconvert.String("rule1"),
						Query: testconvert.String(
							`{"==":[{"var":"key"},"not-user-key"]}`,
						),
						VariationResult: testconvert.String("variation_C"),
					},
					{
						Name:            testconvert.String("rule2"),
						Query:           testconvert.String(`{"==":[{"var":"key"},"user-key"]}`),
						VariationResult: testconvert.String("variation_C"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContext("user-key"),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_C",
			want1: flag.ResolutionDetails{
				Variant:   "variation_C",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(1),
				RuleName:  testconvert.String("rule2"),
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description":       "this is a flag",
					"issue-link":        "https://issue.link/GOFF-1",
					"evaluatedRuleName": "rule2",
				},
			},
		},
		{
			name: "Should match a rule with no name",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
				},
				Rules: &[]flag.Rule{
					{
						Query:           testconvert.String("key eq \"not-user-key\""),
						VariationResult: testconvert.String("variation_C"),
					},
					{
						Query:           testconvert.String("key eq \"user-key\""),
						VariationResult: testconvert.String("variation_C"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContext("user-key"),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_C",
			want1: flag.ResolutionDetails{
				Variant:   "variation_C",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(1),
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should match a rule with no name (jsonLogic)",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
				},
				Rules: &[]flag.Rule{
					{
						Query: testconvert.String(
							`{"==":[{"var":"key"},"not-user-key"]}`,
						),
						VariationResult: testconvert.String("variation_C"),
					},
					{
						Query:           testconvert.String(`{"==":[{"var":"key"},"user-key"]}`),
						VariationResult: testconvert.String("variation_C"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContext("user-key"),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_C",
			want1: flag.ResolutionDetails{
				Variant:   "variation_C",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(1),
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should match only the first rule that apply (even if more than one can apply)",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
					"variation_D": testconvert.Interface("value_D"),
				},
				Rules: &[]flag.Rule{
					{
						Query:           testconvert.String("key eq \"not-user-key\""),
						VariationResult: testconvert.String("variation_C"),
					},
					{
						Query:           testconvert.String("company eq \"go-feature-flag\""),
						VariationResult: testconvert.String("variation_D"),
					},
					{
						Query:           testconvert.String("key eq \"user-key\""),
						VariationResult: testconvert.String("variation_C"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user: ffcontext.NewEvaluationContextBuilder("user-key").
					AddCustom("company", "go-feature-flag").
					Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_D",
			want1: flag.ResolutionDetails{
				Variant:   "variation_D",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(1),
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should match only the first rule that apply (even if more than one can apply) (jsonLogic)",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
					"variation_D": testconvert.Interface("value_D"),
				},
				Rules: &[]flag.Rule{
					{
						Query: testconvert.String(
							`{"==":[{"var":"key"},"not-user-key"]}`,
						),
						VariationResult: testconvert.String("variation_C"),
					},
					{
						Query: testconvert.String(
							`{"==":[{"var":"company"},"go-feature-flag"]}`,
						),
						VariationResult: testconvert.String("variation_D"),
					},
					{
						Query:           testconvert.String(`{"==":[{"var":"key"},"user-key"]}`),
						VariationResult: testconvert.String("variation_C"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user: ffcontext.NewEvaluationContextBuilder("user-key").
					AddCustom("company", "go-feature-flag").
					Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_D",
			want1: flag.ResolutionDetails{
				Variant:   "variation_D",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(1),
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should ignore a rule that is disabled",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
					"variation_D": testconvert.Interface("value_D"),
				},
				Rules: &[]flag.Rule{
					{
						Query:           testconvert.String("key eq \"not-user-key\""),
						VariationResult: testconvert.String("variation_C"),
					},
					{
						Query:           testconvert.String("company eq \"go-feature-flag\""),
						VariationResult: testconvert.String("variation_D"),
						Disable:         testconvert.Bool(true),
					},
					{
						Query:           testconvert.String("key eq \"user-key\""),
						VariationResult: testconvert.String("variation_C"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user: ffcontext.NewEvaluationContextBuilder("user-key").
					AddCustom("company", "go-feature-flag").
					Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_C",
			want1: flag.ResolutionDetails{
				Variant:   "variation_C",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(2),
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should return an error if rule is invalid",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
					"variation_D": testconvert.Interface("value_D"),
				},
				Rules: &[]flag.Rule{
					{
						Query:       testconvert.String("key eq \"user-key\""),
						Percentages: &map[string]float64{},
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_default",
			want1: flag.ResolutionDetails{
				Variant:   flag.VariationSDKDefault,
				Reason:    flag.ReasonError,
				ErrorCode: flag.ErrorFlagConfiguration,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should return an error if rule is invalid (jsonLogic)",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
					"variation_D": testconvert.Interface("value_D"),
				},
				Rules: &[]flag.Rule{
					{
						Query:       testconvert.String(`{"==":[{"var":"key"},"user-key"]}`),
						Percentages: &map[string]float64{},
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_default",
			want1: flag.ResolutionDetails{
				Variant:   flag.VariationSDKDefault,
				Reason:    flag.ReasonError,
				ErrorCode: flag.ErrorFlagConfiguration,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should return an error if no targeting match and we have no default rule",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
					"variation_D": testconvert.Interface("value_D"),
				},
				Rules: &[]flag.Rule{
					{
						Query: testconvert.String("key eq \"not-user-key\""),
						Percentages: &map[string]float64{
							"variation_A": 10,
							"variation_B": 100,
						},
					},
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_default",
			want1: flag.ResolutionDetails{
				Variant:   flag.VariationSDKDefault,
				Reason:    flag.ReasonError,
				ErrorCode: flag.ErrorFlagConfiguration,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should return an error if default rule is invalid",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
					"variation_D": testconvert.Interface("value_D"),
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{},
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_default",
			want1: flag.ResolutionDetails{
				Variant:   flag.VariationSDKDefault,
				Reason:    flag.ReasonError,
				ErrorCode: flag.ErrorFlagConfiguration,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should not have any change if all scheduled steps are in the future",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						InternalFlag: flag.InternalFlag{
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("variation_B"),
							},
						},
						Date: testconvert.Time(time.Now().Add(1 * time.Second)),
					},
					{
						InternalFlag: flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"variation_A": testconvert.Interface("value_QWERTY"),
							},
						},
						Date: testconvert.Time(time.Now().Add(2 * time.Second)),
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant: "variation_A",
				Reason:  flag.ReasonStatic,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should change if all scheduled steps are in the future BUT context override is in the future too",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						InternalFlag: flag.InternalFlag{
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("variation_B"),
							},
						},
						Date: testconvert.Time(time.Now().Add(1 * time.Second)),
					},
					{
						InternalFlag: flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"variation_A": testconvert.Interface("value_QWERTY"),
							},
						},
						Date: testconvert.Time(time.Now().Add(2 * time.Second)),
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user: ffcontext.NewEvaluationContextBuilder("user-key").
					AddCustom("gofeatureflag", map[string]string{
						"currentDateTime": time.Now().Add(1 * time.Minute).Format(time.RFC3339),
					}).
					Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_B",
			want1: flag.ResolutionDetails{
				Variant: "variation_B",
				Reason:  flag.ReasonStatic,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should only apply 1 scheduled step",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						InternalFlag: flag.InternalFlag{
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("variation_B"),
							},
						},
						Date: testconvert.Time(time.Now().Add(-1 * time.Second)),
					},
					{
						InternalFlag: flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"variation_B": testconvert.Interface("value_QWERTY"),
							},
						},
						Date: testconvert.Time(time.Now().Add(2 * time.Second)),
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_B",
			want1: flag.ResolutionDetails{
				Variant: "variation_B",
				Reason:  flag.ReasonStatic,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should only apply 1 scheduled step if step at the exact same time",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						InternalFlag: flag.InternalFlag{
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("variation_B"),
							},
						},
						Date: testconvert.Time(time.Date(2022, 1, 1, 12, 12, 12, 12, time.UTC)),
					},
					{
						InternalFlag: flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"variation_B": testconvert.Interface("value_QWERTY"),
							},
						},
						Date: testconvert.Time(time.Now().Add(2 * time.Second)),
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user: ffcontext.NewEvaluationContextBuilder("user-key").
					AddCustom("gofeatureflag", ffcontext.GoffContextSpecifics{
						CurrentDateTime: testconvert.Time(
							time.Date(2022, 1, 1, 12, 12, 12, 12, time.UTC),
						),
					}).
					Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_B",
			want1: flag.ResolutionDetails{
				Variant: "variation_B",
				Reason:  flag.ReasonStatic,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should apply all scheduled steps in the past",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						InternalFlag: flag.InternalFlag{
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("variation_B"),
							},
						},
						Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
					},
					{
						InternalFlag: flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"variation_B": testconvert.Interface("value_QWERTY"),
							},
						},
						Date: testconvert.Time(time.Now().Add(-1 * time.Second)),
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_QWERTY",
			want1: flag.ResolutionDetails{
				Variant: "variation_B",
				Reason:  flag.ReasonStatic,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should ignore scheduled steps with no dates",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						InternalFlag: flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"variation_A": testconvert.Interface("value_QWERTY"),
							},
						},
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant: "variation_A",
				Reason:  flag.ReasonStatic,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should update a rule that exists with a scheduled step",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				Rules: &[]flag.Rule{
					{
						Name:            testconvert.String("rule1"),
						Query:           testconvert.String("key eq \"user-key\""),
						VariationResult: testconvert.String("variation_B"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						Date: testconvert.Time(time.Now().Add(-1 * time.Second)),
						InternalFlag: flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"variation_C": testconvert.Interface("value_C"),
							},
							Rules: &[]flag.Rule{
								{
									Name:            testconvert.String("rule1"),
									VariationResult: testconvert.String("variation_C"),
								},
							},
						},
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_C",
			want1: flag.ResolutionDetails{
				Variant:   "variation_C",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(0),
				RuleName:  testconvert.String("rule1"),
				Metadata: map[string]interface{}{
					"description":       "this is a flag",
					"issue-link":        "https://issue.link/GOFF-1",
					"evaluatedRuleName": "rule1",
				},
			},
		},
		{
			name: "Should update default rule with a scheduled step",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_C"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
						InternalFlag: flag.InternalFlag{
							DefaultRule: &flag.Rule{
								Percentages: &map[string]float64{
									"variation_B": 30,
									"variation_C": 70,
								},
							},
						},
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key-123").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_C",
			want1: flag.ResolutionDetails{
				Variant: "variation_C",
				Reason:  flag.ReasonSplit,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should add a new rule with a scheduled step",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				Rules: &[]flag.Rule{
					{
						Name:            testconvert.String("rule1"),
						Query:           testconvert.String("key eq \"user-key\""),
						VariationResult: testconvert.String("variation_B"),
					},
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						Date: testconvert.Time(time.Now().Add(-1 * time.Second)),
						InternalFlag: flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"variation_C": testconvert.Interface("value_C"),
							},
							Rules: &[]flag.Rule{
								{
									Name:            testconvert.String("rule2"),
									Query:           testconvert.String("key eq \"user-key-123\""),
									VariationResult: testconvert.String("variation_C"),
								},
							},
						},
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key-123").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_C",
			want1: flag.ResolutionDetails{
				Variant:   "variation_C",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(1),
				RuleName:  testconvert.String("rule2"),
				Metadata: map[string]interface{}{
					"description":       "this is a flag",
					"issue-link":        "https://issue.link/GOFF-1",
					"evaluatedRuleName": "rule2",
				},
			},
		},
		{
			name: "Should apply all the changes if all scheduled steps are in the past",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						InternalFlag: flag.InternalFlag{
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("variation_B"),
							},
						},
						Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
					},
					{
						InternalFlag: flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"variation_B": testconvert.Interface("value_QWERTY"),
							},
						},
						Date: testconvert.Time(time.Now().Add(-1 * time.Second)),
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_QWERTY",
			want1: flag.ResolutionDetails{
				Variant: "variation_B",
				Reason:  flag.ReasonStatic,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should disable the flag with a scheduled step",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						InternalFlag: flag.InternalFlag{
							Disable:     testconvert.Bool(true),
							TrackEvents: testconvert.Bool(false),
							Version:     testconvert.String("1.0.0"),
						},
						Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_default",
			want1: flag.ResolutionDetails{
				Variant: flag.VariationSDKDefault,
				Reason:  flag.ReasonDisabled,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should create an experimentation for a dedicated time",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						InternalFlag: flag.InternalFlag{
							Experimentation: &flag.ExperimentationRollout{
								Start: testconvert.Time(time.Now().Add(-2 * time.Second)),
								End:   testconvert.Time(time.Now().Add(2 * time.Second)),
							},
						},
						Date: testconvert.Time(time.Now().Add(-1 * time.Second)),
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant: "variation_A",
				Reason:  flag.ReasonStatic,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should not apply a scheduled step inside another scheduled step",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						InternalFlag: flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"variation_A": testconvert.Interface("value_AB"),
								"variation_B": testconvert.Interface("value_B"),
							},
							Scheduled: &[]flag.ScheduledStep{
								{
									InternalFlag: flag.InternalFlag{
										Variations: &map[string]*interface{}{
											"variation_A": testconvert.Interface("value_ABC"),
										},
									},
									Date: testconvert.Time(time.Now().Add(-3 * time.Second)),
								},
							},
						},
						Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_AB",
			want1: flag.ResolutionDetails{
				Variant: "variation_A",
				Reason:  flag.ReasonStatic,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should return the false value if not in between initial and end percentage",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				DefaultRule: &flag.Rule{
					ProgressiveRollout: &flag.ProgressiveRollout{
						Initial: &flag.ProgressiveRolloutStep{
							Variation:  testconvert.String("variation_A"),
							Percentage: testconvert.Float64(0),
							Date:       testconvert.Time(time.Now().Add(-10 * time.Second)),
						},
						End: &flag.ProgressiveRolloutStep{
							Variation:  testconvert.String("variation_B"),
							Percentage: testconvert.Float64(5),
							Date:       testconvert.Time(time.Now().Add(-1 * time.Second)),
						},
					},
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant:   "variation_A",
				Reason:    flag.ReasonSplit,
				Cacheable: false,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should return the reason TARGETING_MATCH if rule apply and return a simple variation",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				Rules: &[]flag.Rule{
					{
						Name:            testconvert.String("test-rule"),
						Query:           testconvert.String("key eq \"user-key\""),
						VariationResult: testconvert.String("variation_A"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant:   "variation_A",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(0),
				RuleName:  testconvert.String("test-rule"),
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description":       "this is a flag",
					"issue-link":        "https://issue.link/GOFF-1",
					"evaluatedRuleName": "test-rule",
				},
			},
		},
		{
			name: "Should return the reason TARGETING_MATCH_SPLIT if rule apply and has percentage",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				Rules: &[]flag.Rule{
					{
						Name:  testconvert.String("test-rule"),
						Query: testconvert.String("key eq \"user-key\""),
						Percentages: &map[string]float64{
							"variation_A": 50,
							"variation_B": 50,
						},
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant:   "variation_A",
				Reason:    flag.ReasonTargetingMatchSplit,
				RuleIndex: testconvert.Int(0),
				RuleName:  testconvert.String("test-rule"),
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description":       "this is a flag",
					"issue-link":        "https://issue.link/GOFF-1",
					"evaluatedRuleName": "test-rule",
				},
			},
		},
		{
			name: "Should return the reason TARGETING_MATCH_SPLIT if rule apply and has progressive rollout",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				Rules: &[]flag.Rule{
					{
						Name:  testconvert.String("test-rule"),
						Query: testconvert.String("key eq \"user-key\""),
						ProgressiveRollout: &flag.ProgressiveRollout{
							Initial: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("variation_A"),
								Percentage: testconvert.Float64(0),
								Date:       testconvert.Time(time.Now()),
							},
							End: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("variation_B"),
								Percentage: testconvert.Float64(100),
								Date:       testconvert.Time(time.Now().Add(1 * time.Minute)),
							},
						},
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant:   "variation_A",
				Reason:    flag.ReasonTargetingMatchSplit,
				RuleIndex: testconvert.Int(0),
				RuleName:  testconvert.String("test-rule"),
				Cacheable: false,
				Metadata: map[string]interface{}{
					"description":       "this is a flag",
					"issue-link":        "https://issue.link/GOFF-1",
					"evaluatedRuleName": "test-rule",
				},
			},
		},
		{
			name: "Should return the reason SPLIT if rule not apply and has percentage",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				Rules: &[]flag.Rule{
					{
						Name:            testconvert.String("test-rule"),
						Query:           testconvert.String("key eq \"user-key2\""),
						VariationResult: testconvert.String("variation_B"),
					},
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"variation_A": 50,
						"variation_B": 50,
					},
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant:   "variation_A",
				Reason:    flag.ReasonSplit,
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should return the reason DEFAULT if rule not apply and has default variation",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				Rules: &[]flag.Rule{
					{
						Name:            testconvert.String("test-rule"),
						Query:           testconvert.String("key eq \"user-key2\""),
						VariationResult: testconvert.String("variation_B"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant:   "variation_A",
				Reason:    flag.ReasonDefault,
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should return the reason STATIC if no rule and has default variation",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant:   "variation_A",
				Reason:    flag.ReasonStatic,
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should return the reason STATIC if no rule and has default percentage to 100%",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"variation_A": 100,
						"variation_B": 0,
					},
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant:   "variation_A",
				Reason:    flag.ReasonStatic,
				Cacheable: true,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should use the environment as a rule criteria",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface(false),
					"B": testconvert.Interface(true),
				},
				Rules: &[]flag.Rule{
					{
						Query:           testconvert.String("env == \"development\""),
						VariationResult: testconvert.String("A"),
					},
					{
						Query: testconvert.String("(env == \"production\") " +
							"or (env == \"staging\") or (env == \"qa\")"),
						VariationResult: testconvert.String("B"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("B"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "feature",
				user:     ffcontext.NewEvaluationContextBuilder("1").Build(),
				flagContext: flag.Context{
					DefaultSdkValue: true,
					EvaluationContextEnrichment: map[string]interface{}{
						"env": "development",
					},
				},
			},
			want: false,
			want1: flag.ResolutionDetails{
				Variant:   "A",
				Reason:    flag.ReasonTargetingMatch,
				Cacheable: true,
				RuleIndex: testconvert.Int(0),
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Should have a targeting match if common evaluation context match",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				Rules: &[]flag.Rule{
					{
						Query:           testconvert.String("environment eq \"prod\""),
						VariationResult: testconvert.String("A"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("B"),
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContext("user-key"),
				flagContext: flag.Context{
					DefaultSdkValue: "default-sdk",
					EvaluationContextEnrichment: map[string]interface{}{
						"environment": "prod",
					},
				},
			},
			want: "A",
			want1: flag.ResolutionDetails{
				Variant:   "A",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(0),
				Cacheable: true,
			},
		},
		{
			name: "EvaluationContextEnrichment should override request evaluation context",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				Rules: &[]flag.Rule{
					{
						Query:           testconvert.String("environment eq \"prod\""),
						VariationResult: testconvert.String("A"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("B"),
				},
			},
			args: args{
				flagName: "my-flag",
				user: ffcontext.NewEvaluationContextBuilder("key1").
					AddCustom("environment", "dev").
					Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "default-sdk",
					EvaluationContextEnrichment: map[string]interface{}{
						"environment": "prod",
					},
				},
			},
			want: "A",
			want1: flag.ResolutionDetails{
				Variant:   "A",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(0),
				Cacheable: true,
			},
		},
		{
			name: "Should return sdk default value when we have an error in the deep copy",
			flag: flag.InternalFlag{
				Experimentation: &flag.ExperimentationRollout{
					Start: testconvert.Time(time.Now().Add(-15 * time.Second)),
					End:   testconvert.Time(time.Now().Add(-5 * time.Second)),
				},
				Metadata: &map[string]interface{}{
					"description": make(chan int),
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						Date: testconvert.Time(time.Now().Add(-10 * time.Second)),
						InternalFlag: flag.InternalFlag{
							Experimentation: &flag.ExperimentationRollout{
								Start: testconvert.Time(time.Now().Add(-5 * time.Second)),
								End:   testconvert.Time(time.Now().Add(5 * time.Second)),
							},
						},
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContext("user-key"),
				flagContext: flag.Context{
					DefaultSdkValue: "default-sdk",
				},
			},
			want: "default-sdk",
			want1: flag.ResolutionDetails{
				Variant:   "SdkDefault",
				Reason:    flag.ReasonError,
				ErrorCode: flag.ErrorCodeGeneral,
			},
		},
		{
			name: "Empty targeting key",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface(true),
					"variation_B": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffcontext.NewEvaluationContext(""),
				flagContext: flag.Context{
					DefaultSdkValue: false,
				},
			},
			want: false,
			want1: flag.ResolutionDetails{
				Variant:      "SdkDefault",
				Reason:       flag.ReasonError,
				ErrorCode:    flag.ErrorCodeTargetingKeyMissing,
				ErrorMessage: "Error: Empty targeting key",
				Cacheable:    false,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
		{
			name: "Empty bucketing key",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface(true),
					"variation_B": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				BucketingKey: testconvert.String("teamId"),
			},
			args: args{
				flagName: "my-flag",
				user: ffcontext.NewEvaluationContextBuilder("toto").
					AddCustom("teamId", "").
					Build(),
				flagContext: flag.Context{
					DefaultSdkValue: false,
				},
			},
			want: false,
			want1: flag.ResolutionDetails{
				Variant:      "SdkDefault",
				Reason:       flag.ReasonError,
				ErrorCode:    flag.ErrorCodeTargetingKeyMissing,
				ErrorMessage: "Error: Empty bucketing key",
				Cacheable:    false,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.flag.Value(tt.args.flagName, tt.args.user, tt.args.flagContext)
			assert.Equalf(t, tt.want, got, "not expected value: %s", cmp.Diff(tt.want, got))
			assert.Equalf(t, tt.want1, got1, "not expected value: %s", cmp.Diff(tt.want1, got1))
		})
	}
}

func TestInternalFlag_ValueWithBucketingKey(t *testing.T) {
	type args struct {
		flagName    string
		user        ffcontext.Context
		flagContext flag.Context
	}
	tests := []struct {
		name                string
		flag                flag.InternalFlag
		args                args
		wantForTargetingKey string
		wantForTeamID       string
	}{
		{
			name: "Should use custom bucketing key when set",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
				},
				BucketingKey: testconvert.String("teamId"),
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"variation_A": 33,
						"variation_B": 67,
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user: ffcontext.NewEvaluationContextBuilder("user-key").
					AddCustom("teamId", "team-123").
					Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			wantForTargetingKey: "variation_A",
			wantForTeamID:       "variation_B",
		},
		{
			name: "Should use nested bucketing key when set",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
				},
				BucketingKey: testconvert.String("company.id"),
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"variation_A": 33,
						"variation_B": 67,
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user: ffcontext.NewEvaluationContextBuilder("user-key").
					AddCustom("teamId", "team-123").
					AddCustom("company", map[string]any{
						"id": "company-456",
					}).
					Build(),
				flagContext: flag.Context{
					DefaultSdkValue: "value_default",
				},
			},
			wantForTargetingKey: "variation_A",
			wantForTeamID:       "variation_B",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagWithBucketingKey := tt.flag

			_, got := flagWithBucketingKey.Value(
				tt.args.flagName,
				tt.args.user,
				tt.args.flagContext,
			)
			assert.Equal(t, tt.wantForTeamID, got.Variant)

			flagWithoutBucketingKey := tt.flag
			flagWithoutBucketingKey.BucketingKey = testconvert.String("")
			_, got = flagWithoutBucketingKey.Value(
				tt.args.flagName,
				tt.args.user,
				tt.args.flagContext,
			)

			assert.Equal(t, tt.wantForTargetingKey, got.Variant)
		})
	}
}

func TestInternalFlag_ValueWithInvalidBucketingKey(t *testing.T) {
	type args struct {
		flagName    string
		user        ffcontext.Context
		flagContext flag.Context
	}
	tests := []struct {
		name string
		flag flag.InternalFlag
		args args
		want flag.ResolutionDetails
	}{
		{
			name: "Should return error when bucketing key not found",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
				},
				BucketingKey: testconvert.String("teamId"),
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"variation_A": 33,
						"variation_B": 67,
					},
				},
			},
			args: args{
				flagName:    "my-flag",
				user:        ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flagContext: flag.Context{},
			},
			want: flag.ResolutionDetails{
				Variant:      "SdkDefault",
				Reason:       flag.ReasonError,
				ErrorCode:    flag.ErrorCodeTargetingKeyMissing,
				ErrorMessage: "impossible to find bucketingKey in context: nested key not found: teamId",
			},
		},
		{
			name: "Should return error when nested bucketing key not found",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
				},
				BucketingKey: testconvert.String("company.id"),
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"variation_A": 33,
						"variation_B": 67,
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user: ffcontext.NewEvaluationContextBuilder("user-key").
					AddCustom("company", map[string]any{
						"name": "acme-corp",
					}).
					Build(),
				flagContext: flag.Context{},
			},
			want: flag.ResolutionDetails{
				Variant:      "SdkDefault",
				Reason:       flag.ReasonError,
				ErrorCode:    flag.ErrorCodeTargetingKeyMissing,
				ErrorMessage: "impossible to find bucketingKey in context: nested key not found: company.id",
			},
		},
		{
			name: "Should return error when nested bucketing key is not a string",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"variation_A": testconvert.Interface("value_A"),
					"variation_B": testconvert.Interface("value_B"),
					"variation_C": testconvert.Interface("value_C"),
				},
				BucketingKey: testconvert.String("company.id"),
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"variation_A": 33,
						"variation_B": 67,
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user: ffcontext.NewEvaluationContextBuilder("user-key").
					AddCustom("company", map[string]any{
						"id": 12345,
					}).
					Build(),
				flagContext: flag.Context{},
			},
			want: flag.ResolutionDetails{
				Variant:      "SdkDefault",
				Reason:       flag.ReasonError,
				ErrorCode:    flag.ErrorCodeTargetingKeyMissing,
				ErrorMessage: "invalid bucketing key",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := tt.flag.Value(tt.args.flagName, tt.args.user, tt.args.flagContext)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFlag_ProgressiveRollout(t *testing.T) {
	f := &flag.InternalFlag{
		Variations: &map[string]*interface{}{
			"variation_A": testconvert.Interface("value_A"),
			"variation_B": testconvert.Interface("value_B"),
		},
		DefaultRule: &flag.Rule{
			ProgressiveRollout: &flag.ProgressiveRollout{
				Initial: &flag.ProgressiveRolloutStep{
					Variation:  testconvert.String("variation_A"),
					Percentage: testconvert.Float64(0),
					Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
				},
				End: &flag.ProgressiveRolloutStep{
					Variation:  testconvert.String("variation_B"),
					Percentage: testconvert.Float64(100),
					Date:       testconvert.Time(time.Now().Add(2 * time.Second)),
				},
			},
		},
	}

	user := ffcontext.NewEvaluationContextBuilder("test").AddCustom("anonymous", true).Build()
	flagName := "test-flag"

	// We evaluate the same flag multiple time overtime.
	v, _ := f.Value(flagName, user, flag.Context{})
	assert.Equal(t, f.GetVariationValue("variation_A"), v)

	time.Sleep(1 * time.Second)
	v2, _ := f.Value(flagName, user, flag.Context{})
	assert.Equal(t, f.GetVariationValue("variation_A"), v2)

	time.Sleep(1 * time.Second)
	v3, _ := f.Value(flagName, user, flag.Context{})
	assert.Equal(t, f.GetVariationValue("variation_B"), v3)
}

func TestInternalFlag_GetVariations(t *testing.T) {
	tests := []struct {
		name string
		flag flag.InternalFlag
		want map[string]*interface{}
	}{
		{
			name: "Should return empty map if variations nil",
			flag: flag.InternalFlag{Variations: nil},
			want: map[string]*interface{}{},
		},
		{
			name: "Should return empty map if variations empty map",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{},
			},
			want: map[string]*interface{}{},
		},
		{
			name: "Should return variations if map is not empty",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"varA": testconvert.Interface("valueA"),
					"varB": testconvert.Interface("valueB"),
				},
			},
			want: map[string]*interface{}{
				"varA": testconvert.Interface("valueA"),
				"varB": testconvert.Interface("valueB"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.flag.GetVariations(), "GetVariations()")
		})
	}
}

func TestInternalFlag_GetRuleIndexByName(t *testing.T) {
	tests := []struct {
		name     string
		flag     flag.InternalFlag
		ruleName string
		want     *int
	}{
		{
			name: "Should return nil if no rules in flag",
			flag: flag.InternalFlag{
				Rules: nil,
			},
			ruleName: "rule1",
			want:     nil,
		},
		{
			name: "Should return nil if empty slide of rule",
			flag: flag.InternalFlag{
				Rules: &[]flag.Rule{},
			},
			ruleName: "rule1",
			want:     nil,
		},
		{
			name: "Should return nil if empty slide of rule",
			flag: flag.InternalFlag{
				Rules: &[]flag.Rule{
					{
						Name: testconvert.String("rule0"),
					},
					{
						Name: testconvert.String("rule1"),
					},
				},
			},
			ruleName: "rule1",
			want:     testconvert.Int(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(
				t,
				tt.want,
				tt.flag.GetRuleIndexByName(tt.ruleName),
				"GetRuleIndexByName(%v)",
				tt.ruleName,
			)
		})
	}
}

func TestInternalFlag_GetVariationValue(t *testing.T) {
	tests := []struct {
		name      string
		flag      flag.InternalFlag
		variation string
		want      interface{}
	}{
		{
			name: "Should return nil if variation does not exist",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"varA": testconvert.Interface("valueA"),
					"varB": testconvert.Interface("valueB"),
				},
			},
			variation: "varC",
			want:      nil,
		},
		{
			name: "Should return variation value if exists",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"varA": testconvert.Interface("valueA"),
					"varB": testconvert.Interface("valueB"),
				},
			},
			variation: "varA",
			want:      "valueA",
		},
		{
			name: "Should return nil if variation value is nil",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"varA": testconvert.Interface(nil),
					"varB": testconvert.Interface("valueB"),
				},
			},
			variation: "varA",
			want:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(
				t,
				tt.want,
				tt.flag.GetVariationValue(tt.variation),
				"GetVariationValue(%v)",
				tt.variation,
			)
		})
	}
}

func TestInternalFlag_IsValid(t *testing.T) {
	type fields struct {
		Variations      *map[string]*interface{}
		Rules           *[]flag.Rule
		DefaultRule     *flag.Rule
		Rollout         *flag.Rollout
		TrackEvents     *bool
		Disable         *bool
		Version         *string
		Experimentation *flag.ExperimentationRollout
		Scheduled       *[]flag.ScheduledStep
		Metadata        *map[string]interface{}
	}
	tests := []struct {
		name     string
		fields   fields
		wantErr  assert.ErrorAssertionFunc
		errorMsg string
	}{
		{
			name: "no variation",
			fields: fields{
				Variations: &map[string]*interface{}{},
			},
			wantErr:  assert.Error,
			errorMsg: "no variation available",
		},
		{
			name: "different types in variation",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"C": testconvert.Interface("C"),
					"B": testconvert.Interface(120.1),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			errorMsg: "all variations should have the same type",
			wantErr:  assert.Error,
		},
		{
			name: "different types in variation int/float should be ok",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface(120),
					"B": testconvert.Interface(120.1),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "no default rule",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
				},
				Rules: &[]flag.Rule{
					{
						Name:            testconvert.String("Rule1"),
						Query:           testconvert.String("key eq 1"),
						VariationResult: testconvert.String("A"),
					},
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			errorMsg: "missing default rule",
			wantErr:  assert.Error,
		},
		{
			name: "multiple rule with same name",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("A"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Rules: &[]flag.Rule{
					{
						Name:            testconvert.String("Rule1"),
						Query:           testconvert.String("key eq 1"),
						VariationResult: testconvert.String("A"),
					},
					{
						Name:            testconvert.String(""),
						Query:           testconvert.String("key eq 3"),
						VariationResult: testconvert.String("A"),
					},
					{
						Name:            testconvert.String("Rule2"),
						Query:           testconvert.String("key eq 2"),
						VariationResult: testconvert.String("A"),
					},
					{
						Name:            testconvert.String(""),
						Query:           testconvert.String("key eq 3"),
						VariationResult: testconvert.String("A"),
					},
					{
						Name:            testconvert.String("Rule1"),
						Query:           testconvert.String("key eq 4"),
						VariationResult: testconvert.String("A"),
					},
					{
						Name:            testconvert.String(""),
						Query:           testconvert.String("key eq 5"),
						VariationResult: testconvert.String("A"),
					},
				},
			},
			errorMsg: "duplicated rule name: Rule1",
			wantErr:  assert.Error,
		},
		{
			name: "wrong percentages for default rule",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"A": 00,
						"B": 0,
					},
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			errorMsg: "invalid percentages: should not be equal to 0",
			wantErr:  assert.Error,
		},
		{
			name: "empty percentages for default rule",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{},
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			errorMsg: "invalid percentages: should not be empty",
			wantErr:  assert.Error,
		},
		{
			name: "wrong percentages for targeting",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"A": 00,
						"B": 00,
					},
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Rules: &[]flag.Rule{
					{
						Name:  testconvert.String("Rule1"),
						Query: testconvert.String("key eq 5"),
						Percentages: &map[string]float64{
							"A": 00,
							"B": 0,
						},
					},
				},
			},
			errorMsg: "invalid percentages: should not be equal to 0",
			wantErr:  assert.Error,
		},
		{
			name: "empty percentages for targeting",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{},
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Rules: &[]flag.Rule{
					{
						Name:        testconvert.String("Rule1"),
						Query:       testconvert.String("key eq 5"),
						Percentages: &map[string]float64{},
					},
				},
			},
			errorMsg: "invalid percentages: should not be empty",
			wantErr:  assert.Error,
		},
		{
			name: "targeting without query",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"A": 90,
						"B": 10,
					},
				},
				Rules: &[]flag.Rule{
					{
						Name: testconvert.String("Rule1"),
						Percentages: &map[string]float64{
							"A": 90,
							"B": 10,
						},
					},
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			},
			errorMsg: "each targeting should have a query",
			wantErr:  assert.Error,
		},
		{
			name: "Nothing to return in rule",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				DefaultRule: &flag.Rule{
					Name: testconvert.String("nothing to return"),
				},
			},
			errorMsg: "impossible to return value",
			wantErr:  assert.Error,
		},
		{
			name: "progressive rollout percentage initial > end",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				DefaultRule: &flag.Rule{
					Name: testconvert.String("nothing to return"),
					ProgressiveRollout: &flag.ProgressiveRollout{
						Initial: &flag.ProgressiveRolloutStep{
							Variation:  testconvert.String("A"),
							Percentage: testconvert.Float64(30),
							Date:       testconvert.Time(time.Now().Add(-2 * time.Second)),
						},
						End: &flag.ProgressiveRolloutStep{
							Variation:  testconvert.String("A"),
							Percentage: testconvert.Float64(20),
							Date:       testconvert.Time(time.Now().Add(2 * time.Second)),
						},
					},
				},
			},
			errorMsg: "invalid progressive rollout, initial percentage should be lower than end percentage: 30/20",
			wantErr:  assert.Error,
		},
		{
			name: "ignore invalid rule if disabled",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				Metadata: &map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
				Rules: &[]flag.Rule{
					{
						Name:    testconvert.String("Rule1"),
						Query:   testconvert.String("key eq 5"),
						Disable: testconvert.Bool(true),
						Percentages: &map[string]float64{
							"A": 90,
							"B": 20,
						},
					},
				},
				DefaultRule: &flag.Rule{
					Name:            testconvert.String("nothing to return"),
					VariationResult: testconvert.String("A"),
				},
			},
			errorMsg: "",
			wantErr:  assert.NoError,
		},
		{
			name: "should error if default rule referencing a variation that does not exist",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("C"),
				},
			},
			errorMsg: "invalid variation: C does not exist",
			wantErr:  assert.Error,
		},
		{
			name: "should error if default percentage rule referencing a variation that does not exist",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"A": 90,
						"B": 5,
						"C": 5,
					},
				},
			},
			errorMsg: "invalid percentage: variation C does not exist",
			wantErr:  assert.Error,
		},
		{
			name: "should error if default progressive rule end rollout step referencing a variation that does not exist",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				DefaultRule: &flag.Rule{
					ProgressiveRollout: &flag.ProgressiveRollout{
						Initial: &flag.ProgressiveRolloutStep{
							Variation:  testconvert.String("A"),
							Percentage: testconvert.Float64(0),
							Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
						},
						End: &flag.ProgressiveRolloutStep{
							Variation:  testconvert.String("C"),
							Percentage: testconvert.Float64(100),
							Date:       testconvert.Time(time.Now().Add(2 * time.Second)),
						},
					},
				},
			},
			errorMsg: "invalid progressive rollout, end variation C does not exist",
			wantErr:  assert.Error,
		},
		{
			name: "should error if default progressive rule initial rollout step referencing a variation that does not exist",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				DefaultRule: &flag.Rule{
					ProgressiveRollout: &flag.ProgressiveRollout{
						Initial: &flag.ProgressiveRolloutStep{
							Variation:  testconvert.String("C"),
							Percentage: testconvert.Float64(0),
							Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
						},
						End: &flag.ProgressiveRolloutStep{
							Variation:  testconvert.String("A"),
							Percentage: testconvert.Float64(100),
							Date:       testconvert.Time(time.Now().Add(2 * time.Second)),
						},
					},
				},
			},
			errorMsg: "invalid progressive rollout, initial variation C does not exist",
			wantErr:  assert.Error,
		},
		{
			name: "should error if targeting rule referencing a variation that does not exist",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				Rules: &[]flag.Rule{
					{
						Query:           testconvert.String("targetingKey eq 1"),
						VariationResult: testconvert.String("C"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("A"),
				},
			},
			errorMsg: "invalid variation: C does not exist",
			wantErr:  assert.Error,
		},
		{
			name: "should error if percentage in targeting rule referencing a variation that does not exist",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				Rules: &[]flag.Rule{
					{
						Query: testconvert.String("targetingKey eq 1"),
						Percentages: &map[string]float64{
							"A": 90,
							"B": 5,
							"C": 5,
						},
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("A"),
				},
			},
			errorMsg: "invalid percentage: variation C does not exist",
			wantErr:  assert.Error,
		},
		{
			name: "should error if progressive rollout in targeting rule referencing an initial variation that does not exist",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				Rules: &[]flag.Rule{
					{
						Query: testconvert.String("targetingKey eq 1"),
						ProgressiveRollout: &flag.ProgressiveRollout{
							Initial: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("C"),
								Percentage: testconvert.Float64(0),
								Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
							},
							End: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("A"),
								Percentage: testconvert.Float64(100),
								Date:       testconvert.Time(time.Now().Add(2 * time.Second)),
							},
						},
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("A"),
				},
			},
			errorMsg: "invalid progressive rollout, initial variation C does not exist",
			wantErr:  assert.Error,
		},
		{
			name: "should error if progressive rollout in targeting rule referencing an end variation that does not exist",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				Rules: &[]flag.Rule{
					{
						Query: testconvert.String("targetingKey eq 1"),
						ProgressiveRollout: &flag.ProgressiveRollout{
							Initial: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("A"),
								Percentage: testconvert.Float64(0),
								Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
							},
							End: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("C"),
								Percentage: testconvert.Float64(100),
								Date:       testconvert.Time(time.Now().Add(2 * time.Second)),
							},
						},
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("A"),
				},
			},
			errorMsg: "invalid progressive rollout, end variation C does not exist",
			wantErr:  assert.Error,
		},
		{
			name: "should error if rule query is not valid for the parser",
			fields: fields{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("A"),
					"B": testconvert.Interface("B"),
				},
				Rules: &[]flag.Rule{
					{
						Query:           testconvert.String("invalid"),
						VariationResult: testconvert.String("A"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("A"),
				},
			},
			errorMsg: "invalid query: Invalid rule",
			wantErr:  assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &flag.InternalFlag{
				Variations:      tt.fields.Variations,
				Rules:           tt.fields.Rules,
				DefaultRule:     tt.fields.DefaultRule,
				TrackEvents:     tt.fields.TrackEvents,
				Disable:         tt.fields.Disable,
				Version:         tt.fields.Version,
				Scheduled:       tt.fields.Scheduled,
				Experimentation: tt.fields.Experimentation,
			}
			err := f.IsValid()
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}
			tt.wantErr(t, err, fmt.Sprintf("IsValid(): %s", err))
			assert.Equal(t, tt.errorMsg, errMsg)
		})
	}
}
func TestInternalFlag_ApplySheduledRollout(t *testing.T) {
	cases := 1000
	internalFlag := flag.InternalFlag{
		Variations: &map[string]*interface{}{
			"variation_A": testconvert.Interface("value_A"),
			"variation_B": testconvert.Interface("value_B"),
		},
		DefaultRule: &flag.Rule{
			VariationResult: testconvert.String("variation_A"),
		},
		Metadata: &map[string]interface{}{
			"description": "this is a flag",
			"issue-link":  "https://issue.link/GOFF-1",
		},
		Scheduled: &[]flag.ScheduledStep{
			{
				InternalFlag: flag.InternalFlag{
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("variation_B"),
					},
				},
				Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
			},
			{
				InternalFlag: flag.InternalFlag{
					Variations: &map[string]*interface{}{
						"variation_B": testconvert.Interface("value_QWERTY"),
					},
				},
				Date: testconvert.Time(time.Now().Add(-1 * time.Second)),
			},
		},
	}
	for i := 0; i < cases; i++ {
		t.Run("scheduledRollout", func(t *testing.T) {
			t.Parallel()

			wantResult := "value_QWERTY"
			wantDetails := flag.ResolutionDetails{
				Variant: "variation_B",
				Reason:  flag.ReasonStatic,
				Metadata: map[string]interface{}{
					"description": "this is a flag",
					"issue-link":  "https://issue.link/GOFF-1",
				},
			}

			got, got1 := internalFlag.Value(
				"flag",
				ffcontext.NewEvaluationContextBuilder("user-key").Build(),
				flag.Context{
					DefaultSdkValue: "value_default",
				},
			)
			assert.Equalf(t, wantResult, got, "not expected value: %s", cmp.Diff(wantResult, got))
			assert.Equalf(
				t,
				wantDetails,
				got1,
				"not expected value: %s",
				cmp.Diff(wantDetails, got1),
			)
		})
	}
}

func TestInternalFlag_GetVersion(t *testing.T) {
	tests := []struct {
		name string
		flag flag.InternalFlag
		want string
	}{
		{
			name: "Should return empty string if version is nil",
			flag: flag.InternalFlag{
				Version: nil,
			},
			want: "",
		},
		{
			name: "Should return version if not nil",
			flag: flag.InternalFlag{
				Version: testconvert.String("1.0.0"),
			},
			want: "1.0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.flag.GetVersion(), "GetVersion()")
		})
	}
}
