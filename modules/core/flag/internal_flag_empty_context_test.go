package flag_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
)

func TestInternalFlag_RequiresBucketing(t *testing.T) {
	tests := []struct {
		name     string
		flag     flag.InternalFlag
		expected bool
	}{
		{
			name: "Should not require bucketing - static variation only",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("disabled"),
				},
			},
			expected: false,
		},
		{
			name: "Should require bucketing - default rule has percentages",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"enabled":  20,
						"disabled": 80,
					},
				},
			},
			expected: true,
		},
		{
			name: "Should require bucketing - targeting rule has percentages",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				Rules: &[]flag.Rule{
					{
						Query: testconvert.String("key eq \"admin\""),
						Percentages: &map[string]float64{
							"enabled":  50,
							"disabled": 50,
						},
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("disabled"),
				},
			},
			expected: true,
		},
		{
			name: "Should require bucketing - has progressive rollout",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					ProgressiveRollout: &flag.ProgressiveRollout{
						Initial: &flag.ProgressiveRolloutStep{
							Variation:  testconvert.String("disabled"),
							Percentage: testconvert.Float64(0),
						},
						End: &flag.ProgressiveRolloutStep{
							Variation:  testconvert.String("enabled"),
							Percentage: testconvert.Float64(100),
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "Should not require bucketing - only targeting queries without percentages",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				Rules: &[]flag.Rule{
					{
						Query:           testconvert.String("key eq \"admin\""),
						VariationResult: testconvert.String("enabled"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("disabled"),
				},
			},
			expected: false,
		},
		{
			name: "Should require bucketing - scheduled rollout with percentages",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				Rules: &[]flag.Rule{
					{
						Query:           testconvert.String("key eq \"admin\""),
						VariationResult: testconvert.String("enabled"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("disabled"),
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						InternalFlag: flag.InternalFlag{
							DefaultRule: &flag.Rule{
								Percentages: &map[string]float64{
									"enabled":  50,
									"disabled": 50,
								},
							},
						},
						Date: testconvert.Time(time.Now()),
					},
				},
			},
			expected: true,
		},
		{
			name: "Should require bucketing - scheduled rollout with progressive rollout",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("disabled"),
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						InternalFlag: flag.InternalFlag{
							DefaultRule: &flag.Rule{
								ProgressiveRollout: &flag.ProgressiveRollout{
									Initial: &flag.ProgressiveRolloutStep{
										Variation:  testconvert.String("disabled"),
										Percentage: testconvert.Float64(0),
										Date:       testconvert.Time(time.Now()),
									},
									End: &flag.ProgressiveRolloutStep{
										Variation:  testconvert.String("enabled"),
										Percentage: testconvert.Float64(100),
										Date:       testconvert.Time(time.Now().Add(time.Hour)),
									},
								},
							},
						},
						Date: testconvert.Time(time.Now()),
					},
				},
			},
			expected: true,
		},
		{
			name: "Should require bucketing - scheduled rollout with targeting rule percentages",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("disabled"),
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						InternalFlag: flag.InternalFlag{
							Rules: &[]flag.Rule{
								{
									Query: testconvert.String("beta eq true"),
									Percentages: &map[string]float64{
										"enabled":  25,
										"disabled": 75,
									},
								},
							},
						},
						Date: testconvert.Time(time.Now()),
					},
				},
			},
			expected: true,
		},
		{
			name: "Should not require bucketing - scheduled rollout with only static variations",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("disabled"),
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						InternalFlag: flag.InternalFlag{
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("enabled"),
							},
						},
						Date: testconvert.Time(time.Now()),
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.flag.RequiresBucketing()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInternalFlag_Value_EmptyContext(t *testing.T) {
	tests := []struct {
		name              string
		flag              flag.InternalFlag
		evaluationCtx     ffcontext.Context
		expectedValue     interface{}
		expectedErrorCode string
		expectedReason    string
		shouldSucceed     bool
		description       string
	}{
		{
			name: "Should succeed with empty context for static flag",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("disabled"),
				},
			},
			evaluationCtx:  ffcontext.NewEvaluationContext(""),
			expectedValue:  false,
			expectedReason: flag.ReasonStatic,
			shouldSucceed:  true,
			description:    "Flag without bucketing requirements should work with empty context",
		},
		{
			name: "Should fail with empty context for percentage-based flag",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"enabled":  20,
						"disabled": 80,
					},
				},
			},
			evaluationCtx:     ffcontext.NewEvaluationContext(""),
			expectedErrorCode: flag.ErrorCodeTargetingKeyMissing,
			expectedReason:    flag.ReasonError,
			shouldSucceed:     false,
			description:       "Flag with percentage-based rollout should fail with empty context",
		},
		{
			name: "Should succeed with targeting key for percentage-based flag",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"enabled":  20,
						"disabled": 80,
					},
				},
			},
			evaluationCtx:  ffcontext.NewEvaluationContext("user-123"),
			expectedReason: flag.ReasonSplit,
			shouldSucceed:  true,
			description:    "Flag with percentage-based rollout should work with targeting key",
		},
		{
			name: "Should succeed with empty context for targeting rule without percentages",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				Rules: &[]flag.Rule{
					{
						Query:           testconvert.String("anonymous eq true"),
						VariationResult: testconvert.String("enabled"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("disabled"),
				},
			},
			evaluationCtx: func() ffcontext.Context {
				ctx := ffcontext.NewEvaluationContext("")
				ctx.AddCustomAttribute("anonymous", true)
				return ctx
			}(),
			expectedValue:  true,
			expectedReason: flag.ReasonTargetingMatch,
			shouldSucceed:  true,
			description:    "Flag with targeting rule but no bucketing should work with empty context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagContext := flag.Context{DefaultSdkValue: false}
			value, resolutionDetails := tt.flag.Value("test-flag", tt.evaluationCtx, flagContext)

			if tt.shouldSucceed {
				assert.Equal(t, "", resolutionDetails.ErrorCode, tt.description)
				assert.Equal(t, tt.expectedReason, resolutionDetails.Reason, tt.description)
				if tt.expectedValue != nil {
					assert.Equal(t, tt.expectedValue, value, tt.description)
				}
			} else {
				assert.Equal(t, tt.expectedErrorCode, resolutionDetails.ErrorCode, tt.description)
				assert.Equal(t, tt.expectedReason, resolutionDetails.Reason, tt.description)
			}
		})
	}
}

func TestInternalFlag_Value_NilContext(t *testing.T) {
	tests := []struct {
		name              string
		flag              flag.InternalFlag
		evaluationCtx     ffcontext.Context
		expectedValue     interface{}
		expectedErrorCode string
		expectedReason    string
		shouldSucceed     bool
		description       string
	}{
		{
			name: "Should succeed with nil context for static flag",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("disabled"),
				},
			},
			evaluationCtx:  nil,
			expectedValue:  false,
			expectedReason: flag.ReasonStatic,
			shouldSucceed:  true,
			description:    "Flag without bucketing requirements should work with nil context",
		},
		{
			name: "Should fail with nil context for percentage-based flag",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"enabled":  20,
						"disabled": 80,
					},
				},
			},
			evaluationCtx:     nil,
			expectedErrorCode: flag.ErrorCodeTargetingKeyMissing,
			expectedReason:    flag.ReasonError,
			shouldSucceed:     false,
			description:       "Flag with percentage-based rollout should fail with nil context",
		},
		{
			name: "Should succeed with nil context for flag with targeting rule without percentages",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				Rules: &[]flag.Rule{
					{
						Query:           testconvert.String("key eq \"admin\""),
						VariationResult: testconvert.String("enabled"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("disabled"),
				},
			},
			evaluationCtx:  nil,
			expectedValue:  false,
			expectedReason: flag.ReasonDefault,
			shouldSucceed:  true,
			description:    "Flag with targeting rule but no bucketing should work with nil context (falls back to default)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagContext := flag.Context{DefaultSdkValue: false}
			value, resolutionDetails := tt.flag.Value("test-flag", tt.evaluationCtx, flagContext)

			if tt.shouldSucceed {
				assert.Equal(t, "", resolutionDetails.ErrorCode, tt.description)
				assert.Equal(t, tt.expectedReason, resolutionDetails.Reason, tt.description)
				if tt.expectedValue != nil {
					assert.Equal(t, tt.expectedValue, value, tt.description)
				}
			} else {
				assert.Equal(t, tt.expectedErrorCode, resolutionDetails.ErrorCode, tt.description)
				assert.Equal(t, tt.expectedReason, resolutionDetails.Reason, tt.description)
			}
		})
	}
}

func TestInternalFlag_GetBucketingKeyValue_EmptyContext(t *testing.T) {
	tests := []struct {
		name          string
		flag          flag.InternalFlag
		evaluationCtx ffcontext.Context
		expectedKey   string
		expectedError bool
		description   string
	}{
		{
			name: "Should allow empty key for non-bucketing flag",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("disabled"),
				},
			},
			evaluationCtx: ffcontext.NewEvaluationContext(""),
			expectedKey:   "",
			expectedError: false,
			description:   "Non-bucketing flag should allow empty targeting key",
		},
		{
			name: "Should require key for bucketing flag",
			flag: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"enabled":  50,
						"disabled": 50,
					},
				},
			},
			evaluationCtx: ffcontext.NewEvaluationContext(""),
			expectedKey:   "",
			expectedError: true,
			description:   "Bucketing flag should require targeting key",
		},
		{
			name: "Should use custom bucketing key when provided",
			flag: flag.InternalFlag{
				BucketingKey: testconvert.String("teamId"),
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"enabled":  50,
						"disabled": 50,
					},
				},
			},
			evaluationCtx: func() ffcontext.Context {
				ctx := ffcontext.NewEvaluationContext("")
				ctx.AddCustomAttribute("teamId", "team-123")
				return ctx
			}(),
			expectedKey:   "team-123",
			expectedError: false,
			description:   "Should use custom bucketing key when available",
		},
		{
			name: "Should allow empty custom bucketing key for non-bucketing flag",
			flag: flag.InternalFlag{
				BucketingKey: testconvert.String("teamId"),
				Variations: &map[string]*interface{}{
					"enabled":  testconvert.Interface(true),
					"disabled": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("disabled"),
				},
			},
			evaluationCtx: func() ffcontext.Context {
				ctx := ffcontext.NewEvaluationContext("")
				ctx.AddCustomAttribute("teamId", "")
				return ctx
			}(),
			expectedKey:   "",
			expectedError: false,
			description:   "Non-bucketing flag should allow empty custom bucketing key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := tt.flag.GetBucketingKeyValue(tt.evaluationCtx)

			if tt.expectedError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.expectedKey, key, tt.description)
			}
		})
	}
}
