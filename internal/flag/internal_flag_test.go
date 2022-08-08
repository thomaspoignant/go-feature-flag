package flag_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
)

func TestInternalFlag_Value(t *testing.T) {
	type args struct {
		flagName      string
		user          ffuser.User
		evaluationCtx flag.EvaluationContext
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUser("user-key"),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: false,
				},
			},
			want: true,
			want1: flag.ResolutionDetails{
				Variant: "variation_A",
				Reason:  flag.ReasonDefault,
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUser("user-key"),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: false,
				},
			},
			want: true,
			want1: flag.ResolutionDetails{
				Variant: "variation_A",
				Reason:  flag.ReasonDefault,
			},
		},
		{
			name: "Should return sdk default value when flag is disabled",
			flag: flag.InternalFlag{
				Disable: testconvert.Bool(true),
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUser("user-key"),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "default-sdk",
				},
			},
			want: "default-sdk",
			want1: flag.ResolutionDetails{
				Variant: "SdkDefault",
				Reason:  flag.ReasonDisabled,
			},
		},
		{
			name: "Should return sdk default value when experimentation rollout not started",
			flag: flag.InternalFlag{
				Rollout: &flag.Rollout{
					Experimentation: &flag.ExperimentationRollout{
						Start: testconvert.Time(time.Now().Add(1 * time.Second)),
						End:   testconvert.Time(time.Now().Add(5 * time.Second)),
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUser("user-key"),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "default-sdk",
				},
			},
			want: "default-sdk",
			want1: flag.ResolutionDetails{
				Variant: "SdkDefault",
				Reason:  flag.ReasonDisabled,
			},
		},
		{
			name: "Should return sdk default value when experimentation rollout ended",
			flag: flag.InternalFlag{
				Rollout: &flag.Rollout{
					Experimentation: &flag.ExperimentationRollout{
						Start: testconvert.Time(time.Now().Add(-15 * time.Second)),
						End:   testconvert.Time(time.Now().Add(-5 * time.Second)),
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUser("user-key"),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "default-sdk",
				},
			},
			want: "default-sdk",
			want1: flag.ResolutionDetails{
				Variant: "SdkDefault",
				Reason:  flag.ReasonDisabled,
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUser("user-key"),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: false,
				},
			},
			want: false,
			want1: flag.ResolutionDetails{
				Variant:   "variation_B",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(0),
				RuleName:  testconvert.String("rule1"),
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUser("user-key"),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_C",
			want1: flag.ResolutionDetails{
				Variant:   "variation_C",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(1),
				RuleName:  testconvert.String("rule2"),
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUser("user-key"),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_C",
			want1: flag.ResolutionDetails{
				Variant:   "variation_C",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(1),
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").AddCustom("company", "go-feature-flag").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_D",
			want1: flag.ResolutionDetails{
				Variant:   "variation_D",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(1),
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").AddCustom("company", "go-feature-flag").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_C",
			want1: flag.ResolutionDetails{
				Variant:   "variation_C",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(2),
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
						Query: testconvert.String("key eq \"user-key\""),
						Percentages: &map[string]float64{
							"variation_A": 10,
							"variation_B": 100,
						},
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_default",
			want1: flag.ResolutionDetails{
				Variant:   flag.VariationSDKDefault,
				Reason:    flag.ReasonError,
				ErrorCode: flag.ErrorFlagConfiguration,
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_default",
			want1: flag.ResolutionDetails{
				Variant:   flag.VariationSDKDefault,
				Reason:    flag.ReasonError,
				ErrorCode: flag.ErrorFlagConfiguration,
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
					Percentages: &map[string]float64{
						"variation_A": 10,
						"variation_B": 100,
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_default",
			want1: flag.ResolutionDetails{
				Variant:   flag.VariationSDKDefault,
				Reason:    flag.ReasonError,
				ErrorCode: flag.ErrorFlagConfiguration,
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
				Rollout: &flag.Rollout{
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant: "variation_A",
				Reason:  flag.ReasonDefault,
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
				Rollout: &flag.Rollout{
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_B",
			want1: flag.ResolutionDetails{
				Variant: "variation_B",
				Reason:  flag.ReasonDefault,
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
				Rollout: &flag.Rollout{
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_QWERTY",
			want1: flag.ResolutionDetails{
				Variant: "variation_B",
				Reason:  flag.ReasonDefault,
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
				Rollout: &flag.Rollout{
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant: "variation_A",
				Reason:  flag.ReasonDefault,
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
				Rollout: &flag.Rollout{
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_C",
			want1: flag.ResolutionDetails{
				Variant:   "variation_C",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(0),
				RuleName:  testconvert.String("rule1"),
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
					Percentages: &map[string]float64{
						"variation_A": 10,
						"variation_B": 90,
					},
				},
				Rollout: &flag.Rollout{
					Scheduled: &[]flag.ScheduledStep{
						{
							Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							InternalFlag: flag.InternalFlag{
								DefaultRule: &flag.Rule{
									Percentages: &map[string]float64{
										"variation_B": 20,
										"variation_C": 70,
									},
								},
							},
						},
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key-123").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_C",
			want1: flag.ResolutionDetails{
				Variant: "variation_C",
				Reason:  flag.ReasonDefault,
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
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("variation_A"),
				},
				Rollout: &flag.Rollout{
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key-123").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_C",
			want1: flag.ResolutionDetails{
				Variant:   "variation_C",
				Reason:    flag.ReasonTargetingMatch,
				RuleIndex: testconvert.Int(1),
				RuleName:  testconvert.String("rule2"),
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
				Rollout: &flag.Rollout{
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_QWERTY",
			want1: flag.ResolutionDetails{
				Variant: "variation_B",
				Reason:  flag.ReasonDefault,
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
				Rollout: &flag.Rollout{
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_default",
			want1: flag.ResolutionDetails{
				Variant: flag.VariationSDKDefault,
				Reason:  flag.ReasonDisabled,
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
				Rollout: &flag.Rollout{
					Scheduled: &[]flag.ScheduledStep{
						{
							InternalFlag: flag.InternalFlag{
								Rollout: &flag.Rollout{
									Experimentation: &flag.ExperimentationRollout{
										Start: testconvert.Time(time.Now().Add(-2 * time.Second)),
										End:   testconvert.Time(time.Now().Add(2 * time.Second)),
									},
								},
							},
							Date: testconvert.Time(time.Now().Add(-1 * time.Second)),
						},
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant: "variation_A",
				Reason:  flag.ReasonDefault,
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
				Rollout: &flag.Rollout{
					Scheduled: &[]flag.ScheduledStep{
						{
							InternalFlag: flag.InternalFlag{
								Variations: &map[string]*interface{}{
									"variation_A": testconvert.Interface("value_AB"),
									"variation_B": testconvert.Interface("value_B"),
								},
								Rollout: &flag.Rollout{
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
							},
							Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
						},
					},
				},
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_AB",
			want1: flag.ResolutionDetails{
				Variant: "variation_A",
				Reason:  flag.ReasonDefault,
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
			},
			args: args{
				flagName: "my-flag",
				user:     ffuser.NewUserBuilder("user-key").Build(),
				evaluationCtx: flag.EvaluationContext{
					DefaultSdkValue: "value_default",
				},
			},
			want: "value_A",
			want1: flag.ResolutionDetails{
				Variant: "variation_A",
				Reason:  flag.ReasonDefault,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.flag.Value(tt.args.flagName, tt.args.user, tt.args.evaluationCtx)
			assert.Equalf(t, tt.want, got, "not expected value: %s", cmp.Diff(tt.want, got))
			assert.Equalf(t, tt.want1, got1, "not expected value: %s", cmp.Diff(tt.want1, got1))
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

	user := ffuser.NewAnonymousUser("test")
	flagName := "test-flag"

	// We evaluate the same flag multiple time overtime.
	v, _ := f.Value(flagName, user, flag.EvaluationContext{})
	assert.Equal(t, f.GetVariationValue("variation_A"), v)

	time.Sleep(1 * time.Second)
	v2, _ := f.Value(flagName, user, flag.EvaluationContext{})
	assert.Equal(t, f.GetVariationValue("variation_A"), v2)

	time.Sleep(1 * time.Second)
	v3, _ := f.Value(flagName, user, flag.EvaluationContext{})
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
			assert.Equalf(t, tt.want, tt.flag.GetRuleIndexByName(tt.ruleName), "GetRuleIndexByName(%v)", tt.ruleName)
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
			name: "Should return nil if variation does not exists",
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
			assert.Equalf(t, tt.want, tt.flag.GetVariationValue(tt.variation), "GetVariationValue(%v)", tt.variation)
		})
	}
}
