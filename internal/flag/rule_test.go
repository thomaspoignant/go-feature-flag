package flag_test

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
)

func TestRule_Evaluate(t *testing.T) {
	type args struct {
		user      ffcontext.Context
		hashID    uint32
		isDefault bool
	}
	tests := []struct {
		name    string
		args    args
		rule    flag.Rule
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "No query, default variation result",
			rule: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
			},
			args:    args{},
			want:    "variation_A",
			wantErr: assert.NoError,
		},
		{
			name: "Ignore query if default variation result",
			rule: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
				Query:           testconvert.String("key eq \"def\""),
			},
			args: args{
				isDefault: true,
				user:      ffcontext.NewEvaluationContext("abc"),
			},
			want:    "variation_A",
			wantErr: assert.NoError,
		},
		{
			name: "User does not match the query",
			rule: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
				Query:           testconvert.String("key eq \"def\""),
			},
			args: args{
				isDefault: false,
				user:      ffcontext.NewEvaluationContext("abc"),
			},
			wantErr: assert.Error,
		},
		{
			name: "User match the query",
			rule: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
				Query:           testconvert.String("key eq \"abc\""),
			},
			args: args{
				isDefault: false,
				user:      ffcontext.NewEvaluationContext("abc"),
			},
			want:    "variation_A",
			wantErr: assert.NoError,
		},
		{
			name: "No match and no default variation",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
			},
			args:    args{},
			wantErr: assert.Error,
		},
		{
			name: "Percentage ignore variation result",
			rule: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
				Percentages: &map[string]float64{
					"variation_B": 0,
					"variation_C": 100,
				},
			},
			args: args{
				user:   ffcontext.NewEvaluationContext("userkey"),
				hashID: utils.Hash("flagname+userkey") % flag.MaxPercentage,
			},
			wantErr: assert.NoError,
			want:    "variation_C",
		},
		{
			name: "All percentage does not fit all traffic",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				Percentages: &map[string]float64{
					"variation_B": 0,
					"variation_C": 99,
				},
			},
			args:    args{},
			wantErr: assert.Error,
		},
		{
			name: "All percentage are more than 100%",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				Percentages: &map[string]float64{
					"variation_B": 10,
					"variation_C": 100,
				},
			},
			args:    args{},
			wantErr: assert.Error,
		},
		{
			name: "Percentage in 1st bucket",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				Percentages: &map[string]float64{
					"variation_C": 9,
					"variation_B": 91,
				},
			},
			args: args{
				user:   ffcontext.NewEvaluationContext("userkey"),
				hashID: utils.Hash("flagname+userkey") % flag.MaxPercentage,
			},
			wantErr: assert.NoError,
			want:    "variation_C",
		},
		{
			name: "Percentage in 2nd bucket",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				Percentages: &map[string]float64{
					"variation_D": 70,
					"variation_C": 10,
					"variation_B": 20,
				},
			},
			args: args{
				user:   ffcontext.NewEvaluationContext("randomUserID"),
				hashID: utils.Hash("flagname+randomUserID") % flag.MaxPercentage,
			},
			wantErr: assert.NoError,
			want:    "variation_C",
		},
		{
			name: "Percentage in 3rd bucket",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				Percentages: &map[string]float64{
					"variation_D": 70,
					"variation_C": 10,
					"variation_B": 20,
				},
			},
			args: args{
				user:   ffcontext.NewEvaluationContext("userkey"),
				hashID: utils.Hash("flagname+96ac59e6-7492-436b-b15a-ba1d797d2423") % flag.MaxPercentage,
			},
			wantErr: assert.NoError,
			want:    "variation_B",
		},
		{
			name: "Hash more than max (not supposed to happen)",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				Percentages: &map[string]float64{
					"variation_B": 100,
				},
			},
			args: args{
				user:   ffcontext.NewEvaluationContext("userkey"),
				hashID: flag.MaxPercentage + 1,
			},
			wantErr: assert.Error,
		},
		{
			name: "Percentage + user match query",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				Percentages: &map[string]float64{
					"variation_D": 70,
					"variation_C": 10,
					"variation_B": 20,
				},
				Query: testconvert.String("key eq \"userkey\""),
			},
			args: args{
				user:   ffcontext.NewEvaluationContext("userkey"),
				hashID: utils.Hash("flagname+96ac59e6-7492-436b-b15a-ba1d797d2423") % flag.MaxPercentage,
			},
			wantErr: assert.NoError,
			want:    "variation_B",
		},
		{
			name: "Progressive rollout ignore percentage",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				Percentages: &map[string]float64{
					"variation_A": 100,
				},
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(0 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_C"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(3 * time.Second)),
					},
				},
			},
			args: args{
				user:   ffcontext.NewEvaluationContext("userkey"),
				hashID: utils.Hash("flagname+userKey") % flag.MaxPercentage,
			},
			wantErr: assert.NoError,
			want:    "variation_B",
		},
		{
			name: "Progressive rollout before ramp",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_C"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(3 * time.Second)),
					},
				},
			},
			args: args{
				user:   ffcontext.NewEvaluationContext("userkey"),
				hashID: utils.Hash("flagname+userKey") % flag.MaxPercentage,
			},
			wantErr: assert.NoError,
			want:    "variation_B",
		},
		{
			name: "Progressive rollout after ramp",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(-6 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_C"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(-3 * time.Second)),
					},
				},
			},
			args: args{
				user:   ffcontext.NewEvaluationContext("userkey"),
				hashID: utils.Hash("flagname+userKey") % flag.MaxPercentage,
			},
			wantErr: assert.NoError,
			want:    "variation_C",
		},
		{
			name: "Progressive rollout in the ramp serve initial variation",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(-1 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_C"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(3 * time.Second)),
					},
				},
			},
			args: args{
				user:   ffcontext.NewEvaluationContext("userkey"),
				hashID: utils.Hash("flagname+userKey") % flag.MaxPercentage,
			},
			wantErr: assert.NoError,
			want:    "variation_B",
		},
		{
			name: "Progressive rollout in the ramp serve end variation",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(-3 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_C"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
					},
				},
			},
			args: args{
				user:   ffcontext.NewEvaluationContext("userkey"),
				hashID: utils.Hash("flagname+userKey") % flag.MaxPercentage,
			},
			wantErr: assert.NoError,
			want:    "variation_C",
		},
		{
			name: "Progressive rollout in the ramp serve end variation no percentage specified",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("variation_B"),
						Date:      testconvert.Time(time.Now().Add(-3 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("variation_C"),
						Date:      testconvert.Time(time.Now().Add(1 * time.Second)),
					},
				},
			},
			args: args{
				user:   ffcontext.NewEvaluationContext("userkey"),
				hashID: utils.Hash("flagname+userKey") % flag.MaxPercentage,
			},
			wantErr: assert.NoError,
			want:    "variation_C",
		},
		{
			name: "Progressive rollout initial before end",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(3 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_C"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
					},
				},
			},
			args:    args{},
			wantErr: assert.Error,
		},
		{
			name: "Progressive rollout no initial step",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_C"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
					},
				},
			},
			args:    args{},
			wantErr: assert.Error,
		},
		{
			name: "Progressive rollout no end step",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(3 * time.Second)),
					},
				},
			},
			args:    args{},
			wantErr: assert.Error,
		},
		{
			name: "Progressive rollout initial step with no variation",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(3 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_C"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
					},
				},
			},
			args:    args{},
			wantErr: assert.Error,
		},
		{
			name: "Progressive rollout end step with no variation",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(3 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
					},
				},
			},
			args:    args{},
			wantErr: assert.Error,
		},
		{
			name: "Progressive rollout no initial date",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(0),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_C"),
						Percentage: testconvert.Float64(100),
						Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
					},
				},
			},
			args:    args{},
			wantErr: assert.Error,
		},
		{
			name: "Progressive rollout no end date",
			rule: flag.Rule{
				Name: testconvert.String("rule1"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_B"),
						Percentage: testconvert.Float64(0),
						Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_C"),
						Percentage: testconvert.Float64(100),
					},
				},
			},
			args:    args{},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.rule.Evaluate(tt.args.user, tt.args.hashID, tt.args.isDefault)
			if !tt.wantErr(t, err, fmt.Sprintf("Evaluate(%v, %v, %v)", tt.args.user, tt.args.hashID, tt.args.isDefault)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Evaluate(%v, %v, %v)", tt.args.user, tt.args.hashID, tt.args.isDefault)
		})
	}
}

func TestRule_MergeRules(t *testing.T) {
	tests := []struct {
		name         string
		originalRule flag.Rule
		updatedRule  flag.Rule
		want         flag.Rule
	}{
		{
			name: "merge simple rule",
			originalRule: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
			},
			updatedRule: flag.Rule{
				VariationResult: testconvert.String("variation_B"),
			},
			want: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_B"),
			},
		},
		{
			name: "merge percentage",
			originalRule: flag.Rule{
				Name: testconvert.String("rule1"),
				Percentages: &map[string]float64{
					"variation_A": 10,
					"variation_B": 10,
					"variation_C": 10,
					"variation_D": 10,
					"variation_E": 60,
				},
			},
			updatedRule: flag.Rule{
				Percentages: &map[string]float64{
					"variation_D": -1,
					"variation_E": 50,
					"variation_F": 20,
				},
			},
			want: flag.Rule{
				Name: testconvert.String("rule1"),
				Percentages: &map[string]float64{
					"variation_A": 10,
					"variation_B": 10,
					"variation_C": 10,
					"variation_E": 50,
					"variation_F": 20,
				},
			},
		},
		{
			name: "merge rule with query",
			originalRule: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
				Query:           testconvert.String("key eq \"abc\""),
			},
			updatedRule: flag.Rule{
				Query: testconvert.String("key eq \"cde\""),
			},
			want: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
				Query:           testconvert.String("key eq \"cde\""),
			},
		},
		{
			name: "merge rule remove query",
			originalRule: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
				Query:           testconvert.String("key eq \"abc\""),
			},
			updatedRule: flag.Rule{
				Query: testconvert.String(""),
			},
			want: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
				Query:           testconvert.String(""),
			},
		},
		{
			name: "merge rule with progressive rollout no rollout before",
			originalRule: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
			},
			updatedRule: flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("variation_B"),
						Date:      testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("variation_C"),
						Date:      testconvert.Time(time.Date(2021, time.February, 3, 10, 10, 10, 10, time.UTC)),
					},
				},
			},
			want: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("variation_B"),
						Date:      testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("variation_C"),
						Date:      testconvert.Time(time.Date(2021, time.February, 3, 10, 10, 10, 10, time.UTC)),
					},
				},
			},
		},
		{
			name: "merge rule with progressive rollout update initial step",
			originalRule: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("variation_B"),
						Date:      testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("variation_C"),
						Date:      testconvert.Time(time.Date(2021, time.February, 3, 10, 10, 10, 10, time.UTC)),
					},
				},
			},
			updatedRule: flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_D"),
						Percentage: testconvert.Float64(40),
						Date:       testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
					},
				},
			},
			want: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_D"),
						Percentage: testconvert.Float64(40),
						Date:       testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("variation_C"),
						Date:      testconvert.Time(time.Date(2021, time.February, 3, 10, 10, 10, 10, time.UTC)),
					},
				},
			},
		},
		{
			name: "merge rule with progressive rollout update end step",
			originalRule: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("variation_B"),
						Date:      testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("variation_C"),
						Date:      testconvert.Time(time.Date(2021, time.February, 3, 10, 10, 10, 10, time.UTC)),
					},
				},
			},
			updatedRule: flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					End: &flag.ProgressiveRolloutStep{
						Date:       testconvert.Time(time.Date(2021, time.February, 12, 10, 10, 10, 10, time.UTC)),
						Percentage: testconvert.Float64(100),
					},
				},
			},
			want: flag.Rule{
				Name:            testconvert.String("rule1"),
				VariationResult: testconvert.String("variation_A"),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("variation_B"),
						Date:      testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("variation_C"),
						Date:       testconvert.Time(time.Date(2021, time.February, 12, 10, 10, 10, 10, time.UTC)),
						Percentage: testconvert.Float64(100),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.originalRule.MergeRules(tt.updatedRule)
			assert.Equal(t, tt.want, tt.originalRule)
		})
	}
}
