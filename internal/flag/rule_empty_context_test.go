package flag_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
)

func TestRule_RequiresBucketing(t *testing.T) {
	tests := []struct {
		name     string
		rule     flag.Rule
		expected bool
	}{
		{
			name: "Should not require bucketing - static variation",
			rule: flag.Rule{
				VariationResult: testconvert.String("enabled"),
			},
			expected: false,
		},
		{
			name: "Should require bucketing - has percentages",
			rule: flag.Rule{
				Percentages: &map[string]float64{
					"enabled":  50,
					"disabled": 50,
				},
			},
			expected: true,
		},
		{
			name: "Should not require bucketing - empty percentages",
			rule: flag.Rule{
				Percentages: &map[string]float64{},
			},
			expected: false,
		},
		{
			name: "Should require bucketing - has progressive rollout",
			rule: flag.Rule{
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
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rule.RequiresBucketing()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRule_Evaluate_EmptyKey(t *testing.T) {
	tests := []struct {
		name          string
		rule          flag.Rule
		key           string
		ctx           ffcontext.Context
		expectedError bool
		expectedValue string
		description   string
	}{
		{
			name: "Should succeed with empty key for static variation",
			rule: flag.Rule{
				VariationResult: testconvert.String("enabled"),
			},
			key:           "",
			ctx:           ffcontext.NewEvaluationContext(""),
			expectedError: false,
			expectedValue: "enabled",
			description:   "Static variation should work without key",
		},
		{
			name: "Should fail with empty key for percentage rule",
			rule: flag.Rule{
				Percentages: &map[string]float64{
					"enabled":  50,
					"disabled": 50,
				},
			},
			key:           "",
			ctx:           ffcontext.NewEvaluationContext(""),
			expectedError: true,
			description:   "Percentage rule should fail without key",
		},
		{
			name: "Should succeed with key for percentage rule",
			rule: flag.Rule{
				Percentages: &map[string]float64{
					"enabled":  50,
					"disabled": 50,
				},
			},
			key:           "user-123",
			ctx:           ffcontext.NewEvaluationContext("user-123"),
			expectedError: false,
			description:   "Percentage rule should work with key",
		},
		{
			name: "Should fail with empty key for progressive rollout",
			rule: flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("disabled"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(-time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("enabled"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(time.Hour)),
					},
				},
			},
			key:           "",
			ctx:           ffcontext.NewEvaluationContext(""),
			expectedError: true,
			description:   "Progressive rollout should fail without key",
		},
		{
			name: "Should succeed with empty key for query-only rule",
			rule: flag.Rule{
				Query:           testconvert.String("anonymous eq true"),
				VariationResult: testconvert.String("enabled"),
			},
			key: "",
			ctx: func() ffcontext.Context {
				ctx := ffcontext.NewEvaluationContext("")
				ctx.AddCustomAttribute("anonymous", true)
				return ctx
			}(),
			expectedError: false,
			expectedValue: "enabled",
			description:   "Query-only rule should work without key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.rule.Evaluate(tt.key, tt.ctx, "test-flag", false)

			if tt.expectedError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				if tt.expectedValue != "" {
					assert.Equal(t, tt.expectedValue, result, tt.description)
				}
			}
		})
	}
}
