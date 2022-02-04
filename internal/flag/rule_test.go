package flag_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"testing"
	"time"
)

func TestRule_Evaluate(t *testing.T) {
	type args struct {
		user        ffuser.User
		flagName    string
		defaultRule bool
	}
	tests := []struct {
		name          string
		rule          *flag.Rule
		args          args
		wantApply     bool
		wantVariation string
		wantErr       bool
	}{
		{
			name: "empty rule",
			rule: &flag.Rule{},
			args: args{
				user:     ffuser.NewUser("toto"),
				flagName: "flag1",
			},
			wantErr:   true,
			wantApply: false,
		},
		{
			name: "Rule apply for the user with single variation",
			rule: &flag.Rule{
				Query:           testconvert.String("key == \"random-key\""),
				VariationResult: testconvert.String("valid"),
			},
			args: args{
				user:     ffuser.NewUser("random-key"),
				flagName: "flag1",
			},
			wantErr:       false,
			wantApply:     true,
			wantVariation: "valid",
		},
		{
			name: "Rule does not apply for the user with single variation",
			rule: &flag.Rule{
				Query:           testconvert.String("key == \"random-key1\""),
				VariationResult: testconvert.String("valid"),
			},
			args: args{
				user:     ffuser.NewUser("random-key"),
				flagName: "flag1",
			},
			wantErr:   false,
			wantApply: false,
		},
		{
			name: "Rule does not apply for the user with percentage variation",
			rule: &flag.Rule{
				Query: testconvert.String("key == \"random-key1\""),
				Percentages: &map[string]float64{
					"variation1": 10,
					"variation2": 90,
				},
			},
			args: args{
				user:     ffuser.NewUser("random-key"),
				flagName: "flag1",
			},
			wantErr:   false,
			wantApply: false,
		},
		{
			name: "Rule apply for the user with percentage variation (variation1)",
			rule: &flag.Rule{
				Query: testconvert.String("key == \"random-key\""),
				Percentages: &map[string]float64{
					"variation1": 50,
					"variation2": 50,
				},
			},
			args: args{
				user:     ffuser.NewUser("random-key"),
				flagName: "flag1",
			},
			wantErr:       false,
			wantApply:     true,
			wantVariation: "variation1",
		},
		{
			name: "Rule apply for the user with percentage variation (variation2)",
			rule: &flag.Rule{
				Query: testconvert.String("key == \"key1\""),
				Percentages: &map[string]float64{
					"variation1": 40,
					"variation2": 50,
					"variation3": 10,
				},
			},
			args: args{
				user:     ffuser.NewUser("key1"),
				flagName: "flag_in_variation_2",
			},
			wantErr:       false,
			wantApply:     true,
			wantVariation: "variation2",
		},
		{
			name: "Rule more that 100%",
			rule: &flag.Rule{
				Query: testconvert.String("key == \"key1\""),
				Percentages: &map[string]float64{
					"variation1": 40,
					"variation2": 100,
				},
			},
			args: args{
				user:     ffuser.NewUser("key1"),
				flagName: "flag_in_variation_2",
			},
			wantErr:   true,
			wantApply: false,
		},
		{
			name: "ProgressiveRolloutStep rollout not started",
			rule: &flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("Variation1"),
						Percentage: 0,
						Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("Variation2"),
						Percentage: 100,
						Date:       testconvert.Time(time.Now().Add(1 * time.Hour)),
					},
				},
			},
			args: args{
				user:     ffuser.NewUser("key1"),
				flagName: "flag_in_variation_2",
			},
			wantErr:       false,
			wantApply:     true,
			wantVariation: "Variation1",
		},
		{
			name: "ProgressiveRolloutStep rollout already finished",
			rule: &flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("Variation1"),
						Percentage: 0,
						Date:       testconvert.Time(time.Now().Add(-10 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("Variation2"),
						Percentage: 100,
						Date:       testconvert.Time(time.Now().Add(-1 * time.Second)),
					},
				},
			},
			args: args{
				user:     ffuser.NewUser("key1"),
				flagName: "flag_in_variation_2",
			},
			wantErr:       false,
			wantApply:     true,
			wantVariation: "Variation2",
		},
		{
			name: "ProgressiveRolloutStep rollout in the middle (Variation1)",
			rule: &flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("Variation1"),
						Percentage: 0,
						Date:       testconvert.Time(time.Now().Add(-1 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("Variation2"),
						Percentage: 100,
						Date:       testconvert.Time(time.Now().Add(10 * time.Second)),
					},
				},
			},
			args: args{
				user:     ffuser.NewUser("key1"),
				flagName: "flag_in_variation_2",
			},
			wantErr:       false,
			wantApply:     true,
			wantVariation: "Variation1",
		},
		{
			name: "ProgressiveRolloutStep rollout in the middle (Variation2)",
			rule: &flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("Variation1"),
						Percentage: 0,
						Date:       testconvert.Time(time.Now().Add(-10 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("Variation2"),
						Percentage: 100,
						Date:       testconvert.Time(time.Now().Add(10 * time.Second)),
					},
				},
			},
			args: args{
				user:     ffuser.NewUser("key1"),
				flagName: "flag_in_variation_2",
			},
			wantErr:       false,
			wantApply:     true,
			wantVariation: "Variation2",
		},
		{
			name: "ProgressiveRolloutStep rollout no percentage (will take 0 -> 100)",
			rule: &flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("Variation1"),
						Date:      testconvert.Time(time.Now().Add(-10 * time.Second)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("Variation2"),
						Date:      testconvert.Time(time.Now().Add(10 * time.Second)),
					},
				},
			},
			args: args{
				user:     ffuser.NewUser("key1"),
				flagName: "flag_in_variation_2",
			},
			wantErr:       false,
			wantApply:     true,
			wantVariation: "Variation2",
		},
		{
			name: "ProgressiveRolloutStep rollout missing date",
			rule: &flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("Variation1"),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation: testconvert.String("Variation2"),
						Date:      testconvert.Time(time.Now().Add(10 * time.Second)),
					},
				},
			},
			args: args{
				user:     ffuser.NewUser("key1"),
				flagName: "flag_in_variation_2",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashID := utils.Hash(tt.args.flagName+tt.args.user.GetKey()) % flag.MaxPercentage
			gotValue, gotVariation, err := tt.rule.Evaluate(tt.args.user, hashID, tt.args.defaultRule)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantApply, gotValue)
			assert.Equal(t, tt.wantVariation, gotVariation)
		})
	}
}

func TestRule_string(t *testing.T) {
	tests := []struct {
		name     string
		rule     *flag.Rule
		expected string
	}{
		{
			name: "Only variation",
			rule: &flag.Rule{
				VariationResult: testconvert.String("Variation1"),
			},
			expected: "variation:[Variation1]",
		},
		{
			name:     "Empty",
			rule:     &flag.Rule{},
			expected: "",
		},
		{
			name: "Percentage variation",
			rule: &flag.Rule{
				VariationResult: testconvert.String("Variation1"),
				Percentages: &map[string]float64{
					"Variation1": 10,
					"Variation2": 75,
					"Variation3": 5,
				},
			},
			expected: "variation:[Variation1], percentages:[Variation1=10.00,Variation2=75.00,Variation3=5.00]",
		},
		{
			name: "With Query",
			rule: &flag.Rule{
				VariationResult: testconvert.String("Variation1"),
				Percentages: &map[string]float64{
					"Variation1": 10,
					"Variation2": 75,
					"Variation3": 5,
				},
				Query: testconvert.String("key eq \"toto\""),
			},
			expected: "query:[key eq \"toto\"], variation:[Variation1], percentages:[Variation1=10.00,Variation2=75.00,Variation3=5.00]",
		},
		{
			name: "DtoProgressiveRollout rollout",
			rule: &flag.Rule{
				VariationResult: testconvert.String("Variation1"),
				Percentages: &map[string]float64{
					"Variation1": 10,
					"Variation2": 75,
					"Variation3": 5,
				},
				Query: testconvert.String("key eq \"toto\""),
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("Variation1"),
						Percentage: 0,
						Date:       testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
					},
					End: &flag.ProgressiveRolloutStep{
						Variation:  testconvert.String("Variation2"),
						Percentage: 100,
						Date:       testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 30, 10, time.UTC)),
					},
				},
			},
			expected: "query:[key eq \"toto\"], variation:[Variation1], percentages:[Variation1=10.00,Variation2=75.00,Variation3=5.00], progressiveRollout:[Initial:[Variation:[Variation1], Percentage:[0], Date:[2021-02-01T10:10:10Z]], End:[Variation:[Variation2], Percentage:[100], Date:[2021-02-01T10:10:30Z]]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rule.String()
			fmt.Println(got)
			assert.Equal(t, tt.expected, got)
		})
	}
}
