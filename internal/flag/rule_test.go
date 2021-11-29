package flag_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"testing"
	"time"
)

func TestRule_Evaluate(t *testing.T) {
	type fields struct {
		Query              *string
		VariationResult    *string
		Percentages        *map[string]float64
		ProgressiveRollout *flag.ProgressiveRollout
	}
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
			name: "Progressive rollout not started",
			rule: &flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.Progressive{
						Variation:  testconvert.String("Variation1"),
						Percentage: 0,
						Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
					},
					End: &flag.Progressive{
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
			name: "Progressive rollout already finished",
			rule: &flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.Progressive{
						Variation:  testconvert.String("Variation1"),
						Percentage: 0,
						Date:       testconvert.Time(time.Now().Add(-10 * time.Second)),
					},
					End: &flag.Progressive{
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
			name: "Progressive rollout in the middle (Variation1)",
			rule: &flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.Progressive{
						Variation:  testconvert.String("Variation1"),
						Percentage: 0,
						Date:       testconvert.Time(time.Now().Add(-1 * time.Second)),
					},
					End: &flag.Progressive{
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
			name: "Progressive rollout in the middle (Variation2)",
			rule: &flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.Progressive{
						Variation:  testconvert.String("Variation1"),
						Percentage: 0,
						Date:       testconvert.Time(time.Now().Add(-10 * time.Second)),
					},
					End: &flag.Progressive{
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
			name: "Progressive rollout no percentage (will take 0 -> 100)",
			rule: &flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.Progressive{
						Variation: testconvert.String("Variation1"),
						Date:      testconvert.Time(time.Now().Add(-10 * time.Second)),
					},
					End: &flag.Progressive{
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
			name: "Progressive rollout missing date",
			rule: &flag.Rule{
				ProgressiveRollout: &flag.ProgressiveRollout{
					Initial: &flag.Progressive{
						Variation: testconvert.String("Variation1"),
					},
					End: &flag.Progressive{
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
