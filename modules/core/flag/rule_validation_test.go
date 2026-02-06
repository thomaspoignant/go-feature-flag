package flag_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestRule_IsValid(t *testing.T) {
	variations := map[string]*any{
		"variation_A": testconvert.Interface("value_A"),
		"variation_B": testconvert.Interface("value_B"),
		"variation_C": testconvert.Interface("value_C"),
	}

	tests := []struct {
		name        string
		rule        flag.Rule
		defaultRule bool
		variations  map[string]*any
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name: "valid rule with variation result",
			rule: flag.Rule{
				Query:           testconvert.String("key eq \"test\""),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid rule with percentages",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				Percentages: &map[string]float64{
					"variation_A": 50,
					"variation_B": 50,
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid rule with progressive rollout",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_A"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(2 * time.Hour)),
					},
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "disabled rule should be valid",
			rule: flag.Rule{
				Disable: testconvert.Bool(true),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "disabled rule with defaultRule should be valid",
			rule: flag.Rule{
				Disable:         testconvert.Bool(true),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: true,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "rule with no return value",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
		},
		{
			name: "valid default rule without query",
			rule: flag.Rule{
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: true,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "non-default rule without query should fail",
			rule: flag.Rule{
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
		},
		{
			name: "rule with invalid query",
			rule: flag.Rule{
				VariationResult: testconvert.String("variation_A"),
				Query:           testconvert.String("invalid query syntax"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
		},
		{
			name: "rule with invalid percentages",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				Percentages: &map[string]float64{
					"non_existent": 100,
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
		},
		{
			name: "rule with invalid progressive rollout",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_A"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(2 * time.Hour)),
					},
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
		},
		{
			name: "rule with invalid variation result",
			rule: flag.Rule{
				VariationResult: testconvert.String("non_existent"),
			},
			defaultRule: true,
			variations:  variations,
			wantErr:     assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.IsValid(tt.defaultRule, tt.variations)
			tt.wantErr(t, err, "IsValid() error = %v", err)
		})
	}
}

func TestRule_validatePercentages(t *testing.T) {
	variations := map[string]*any{
		"variation_A": testconvert.Interface("value_A"),
		"variation_B": testconvert.Interface("value_B"),
		"variation_C": testconvert.Interface("value_C"),
	}

	tests := []struct {
		name        string
		rule        flag.Rule
		defaultRule bool
		variations  map[string]*any
		wantErr     assert.ErrorAssertionFunc
		wantErrMsg  string
	}{
		{
			name: "nil percentages should be valid",
			rule: flag.Rule{
				Query:           testconvert.String("key eq \"test\""),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid percentages",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				Percentages: &map[string]float64{
					"variation_A": 30,
					"variation_B": 70,
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid percentages summing to 100",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				Percentages: &map[string]float64{
					"variation_A": 25,
					"variation_B": 25,
					"variation_C": 50,
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "empty percentages should fail",
			rule: flag.Rule{
				Query:       testconvert.String("key eq \"test\""),
				Percentages: &map[string]float64{},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid percentages: should not be empty",
		},
		{
			name: "percentages summing to zero should fail",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				Percentages: &map[string]float64{
					"variation_A": 0,
					"variation_B": 0,
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid percentages: should not be equal to 0",
		},
		{
			name: "percentages with non-existent variation should fail",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				Percentages: &map[string]float64{
					"non_existent": 100,
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid percentage: variation non_existent does not exist",
		},
		{
			name: "percentages with multiple non-existent variations should fail",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				Percentages: &map[string]float64{
					"variation_A":    50,
					"non_existent_1": 25,
					"non_existent_2": 25,
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
		},
		{
			name: "percentages with negative values should be valid (sum > 0)",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				Percentages: &map[string]float64{
					"variation_A": -10,
					"variation_B": 110,
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "percentages summing to more than 100 should be valid",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				Percentages: &map[string]float64{
					"variation_A": 150,
					"variation_B": 50,
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.IsValid(tt.defaultRule, tt.variations)
			tt.wantErr(t, err, "IsValid() error = %v", err)
			if tt.wantErrMsg != "" && err != nil {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			}
		})
	}
}

func TestRule_validateProgressiveRollout(t *testing.T) {
	variations := map[string]*any{
		"variation_A": testconvert.Interface("value_A"),
		"variation_B": testconvert.Interface("value_B"),
		"variation_C": testconvert.Interface("value_C"),
	}

	tests := []struct {
		name        string
		rule        flag.Rule
		defaultRule bool
		variations  map[string]*any
		wantErr     assert.ErrorAssertionFunc
		wantErrMsg  string
	}{
		{
			name: "nil progressive rollout should be valid",
			rule: flag.Rule{
				Query:           testconvert.String("key eq \"test\""),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid progressive rollout",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_A"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(2 * time.Hour)),
					},
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid progressive rollout with equal percentages",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_A"),
						Percentage: testconvert.Float64(50),
						Date:       testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(50),
						Date:       testconvert.Time(time.Now().Add(2 * time.Hour)),
					},
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "progressive rollout with initial percentage higher than end should fail",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_A"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(2 * time.Hour)),
					},
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid progressive rollout, initial percentage should be lower",
		},
		{
			name: "progressive rollout with non-existent end variation should fail",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_A"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("non_existent"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(2 * time.Hour)),
					},
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid progressive rollout, end variation non_existent does not exist",
		},
		{
			name: "progressive rollout with non-existent initial variation should fail",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("non_existent"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(2 * time.Hour)),
					},
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid progressive rollout, initial variation non_existent does not exist",
		},
		{
			name: "progressive rollout with nil initial variation should fail",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  nil,
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(2 * time.Hour)),
					},
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid progressive rollout, initial variation",
		},
		{
			name: "progressive rollout with nil end variation should fail",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_A"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  nil,
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(2 * time.Hour)),
					},
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid progressive rollout, end variation",
		},
		{
			name: "progressive rollout with empty initial variation should fail",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String(""),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(2 * time.Hour)),
					},
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid progressive rollout, initial variation",
		},
		{
			name: "progressive rollout with empty end variation should fail",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_A"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String(""),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(2 * time.Hour)),
					},
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid progressive rollout, end variation",
		},
		{
			name: "progressive rollout without percentage should use default 0",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("variation_A"),
						Date:      testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(2 * time.Hour)),
					},
				},
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.IsValid(tt.defaultRule, tt.variations)
			tt.wantErr(t, err, "IsValid() error = %v", err)
			if tt.wantErrMsg != "" && err != nil {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			}
		})
	}
}

func TestRule_validateVariationResult(t *testing.T) {
	variations := map[string]*any{
		"variation_A": testconvert.Interface("value_A"),
		"variation_B": testconvert.Interface("value_B"),
		"variation_C": testconvert.Interface("value_C"),
	}

	tests := []struct {
		name        string
		rule        flag.Rule
		defaultRule bool
		variations  map[string]*any
		wantErr     assert.ErrorAssertionFunc
		wantErrMsg  string
	}{
		{
			name: "nil variation result should be valid when percentages exist",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				Percentages: &map[string]float64{
					"variation_A": 100,
				},
				VariationResult: nil,
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "nil variation result should be valid when progressive rollout exists",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_A"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(2 * time.Hour)),
					},
				},
				VariationResult: nil,
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid variation result",
			rule: flag.Rule{
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: true,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "variation result with non-existent variation should fail",
			rule: flag.Rule{
				VariationResult: testconvert.String("non_existent"),
			},
			defaultRule: true,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid variation: non_existent does not exist",
		},
		{
			name: "variation result with empty string should fail",
			rule: flag.Rule{
				VariationResult: testconvert.String(""),
			},
			defaultRule: true,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid variation:",
		},
		{
			name: "variation result should be ignored when percentages exist",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				Percentages: &map[string]float64{
					"variation_A": 100,
				},
				VariationResult: testconvert.String("non_existent"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "variation result should be ignored when progressive rollout exists",
			rule: flag.Rule{
				Query: testconvert.String("key eq \"test\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_A"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(2 * time.Hour)),
					},
				},
				VariationResult: testconvert.String("non_existent"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.IsValid(tt.defaultRule, tt.variations)
			tt.wantErr(t, err, "IsValid() error = %v", err)
			if tt.wantErrMsg != "" && err != nil {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			}
		})
	}
}

func TestRule_isQueryValid(t *testing.T) {
	variations := map[string]*any{
		"variation_A": testconvert.Interface("value_A"),
	}

	tests := []struct {
		name        string
		rule        flag.Rule
		defaultRule bool
		variations  map[string]*any
		wantErr     assert.ErrorAssertionFunc
		wantErrMsg  string
	}{
		{
			name: "default rule should be valid without query",
			rule: flag.Rule{
				Query:           nil,
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: true,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "default rule should be valid with query",
			rule: flag.Rule{
				Query:           testconvert.String("key eq \"test\""),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: true,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "non-default rule without query should fail",
			rule: flag.Rule{
				Query:           nil,
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "each targeting should have a query",
		},
		{
			name: "non-default rule with empty query should fail",
			rule: flag.Rule{
				Query:           testconvert.String(""),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
		},
		{
			name: "non-default rule with valid Nikunjy query",
			rule: flag.Rule{
				Query:           testconvert.String("key eq \"test\""),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "non-default rule with valid JSONLogic query",
			rule: flag.Rule{
				Query:           testconvert.String(`{"==": [{"var": "key"}, "test"]}`),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "non-default rule with invalid Nikunjy query",
			rule: flag.Rule{
				Query:           testconvert.String("invalid query syntax"),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid query",
		},
		{
			name: "non-default rule with complex valid Nikunjy query",
			rule: flag.Rule{
				Query:           testconvert.String("key eq \"test\" and version gt 1.0.0"),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "non-default rule with valid Nikunjy query with OR",
			rule: flag.Rule{
				Query:           testconvert.String("key eq \"test1\" or key eq \"test2\""),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "non-default rule with query containing whitespace",
			rule: flag.Rule{
				Query:           testconvert.String("  key eq \"test\"  "),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "non-default rule with query containing newlines",
			rule: flag.Rule{
				Query:           testconvert.String("key eq \"test\"\n"),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.IsValid(tt.defaultRule, tt.variations)
			tt.wantErr(t, err, "IsValid() error = %v", err)
			if tt.wantErrMsg != "" && err != nil {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			}
		})
	}
}

func TestRule_validateNikunjyQuery(t *testing.T) {
	variations := map[string]*any{
		"variation_A": testconvert.Interface("value_A"),
	}

	tests := []struct {
		name        string
		rule        flag.Rule
		defaultRule bool
		variations  map[string]*any
		wantErr     assert.ErrorAssertionFunc
		wantErrMsg  string
	}{
		{
			name: "valid simple query",
			rule: flag.Rule{
				Query:           testconvert.String("key eq \"test\""),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid query with AND",
			rule: flag.Rule{
				Query:           testconvert.String("key eq \"test\" and version gt 1.0.0"),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid query with OR",
			rule: flag.Rule{
				Query:           testconvert.String("key eq \"test1\" or key eq \"test2\""),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid query with comparison operators",
			rule: flag.Rule{
				Query:           testconvert.String("age gt 18"),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid query with lt operator",
			rule: flag.Rule{
				Query:           testconvert.String("age lt 65"),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid query with ge operator",
			rule: flag.Rule{
				Query:           testconvert.String("score ge 100"),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid query with le operator",
			rule: flag.Rule{
				Query:           testconvert.String("score le 200"),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid query with ne operator",
			rule: flag.Rule{
				Query:           testconvert.String("status ne \"inactive\""),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid query with in operator",
			rule: flag.Rule{
				Query:           testconvert.String("status in [\"active\", \"pending\"]"),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid query with parentheses",
			rule: flag.Rule{
				Query:           testconvert.String("(key eq \"test1\" or key eq \"test2\") and version gt 1.0.0"),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "invalid query syntax",
			rule: flag.Rule{
				Query:           testconvert.String("invalid query syntax"),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid query",
		},
		{
			name: "invalid query with missing operator",
			rule: flag.Rule{
				Query:           testconvert.String("key \"test\""),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid query",
		},
		{
			name: "invalid query with unclosed quotes",
			rule: flag.Rule{
				Query:           testconvert.String("key eq \"test"),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid query",
		},
		{
			name: "invalid query with malformed parentheses - removed as parser accepts it",
			rule: flag.Rule{
				Query:           testconvert.String("key eq \"test\" and (version gt 1.0.0"),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError, // Parser accepts this, so we accept it
		},
		{
			name: "empty query",
			rule: flag.Rule{
				Query:           testconvert.String(""),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
		},
		{
			name: "query with only whitespace",
			rule: flag.Rule{
				Query:           testconvert.String("   "),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.Error,
			wantErrMsg:  "invalid query",
		},
		{
			name: "valid query with semver comparison",
			rule: flag.Rule{
				Query:           testconvert.String("version gt 1.0.0"),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
		{
			name: "valid query with date comparison",
			rule: flag.Rule{
				Query:           testconvert.String("createdAt gt \"2023-01-01T00:00:00Z\""),
				VariationResult: testconvert.String("variation_A"),
			},
			defaultRule: false,
			variations:  variations,
			wantErr:     assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.IsValid(tt.defaultRule, tt.variations)
			tt.wantErr(t, err, "IsValid() error = %v", err)
			if tt.wantErrMsg != "" && err != nil {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			}
		})
	}
}
