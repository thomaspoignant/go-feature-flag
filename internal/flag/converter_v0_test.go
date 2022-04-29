package flag_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"testing"
	"time"
)

func TestConvertV0DtoToFlag(t *testing.T) {
	type args struct {
		dtoFlag        flag.DtoFlag
		isScheduleStep bool
	}
	tests := []struct {
		name string
		args args
		want flag.FlagData
	}{
		{
			name: "convert simple v0 flag",
			args: args{
				dtoFlag: flag.DtoFlag{
					Rule:       testconvert.String("key eq \"batman\""),
					Percentage: testconvert.Float64(40),
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
				},
			},
			want: flag.FlagData{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("true"),
				},
				Rules: &map[string]flag.Rule{
					flag.LegacyRuleName: {
						Query: testconvert.String("key eq \"batman\""),
						Percentages: &map[string]float64{
							"True":  40,
							"False": 60,
						},
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
			},
		},
		{
			name: "convert v0 flag with progressive rollout",
			args: args{
				dtoFlag: flag.DtoFlag{
					Rule:       testconvert.String("key eq \"batman\""),
					Percentage: testconvert.Float64(40),
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Rollout: &flag.DtoRollout{
						Progressive: &flag.DtoProgressiveRollout{
							Percentage: flag.DtoProgressivePercentage{
								Initial: 5,
								End:     95,
							},
							ReleaseRamp: flag.DtoProgressiveReleaseRamp{
								Start: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
								End:   testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
							},
						},
					},
				},
			},
			want: flag.FlagData{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("true"),
				},
				Rules: &map[string]flag.Rule{
					flag.LegacyRuleName: {
						Query: testconvert.String("key eq \"batman\""),
						ProgressiveRollout: &flag.ProgressiveRollout{
							Initial: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("False"),
								Percentage: 5,
								Date:       testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
							},
							End: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("True"),
								Percentage: 95,
								Date:       testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
							},
						},
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
			},
		},
		{
			name: "convert v0 flag with scheduled rollout",
			args: args{
				dtoFlag: flag.DtoFlag{
					Rule:       testconvert.String("key eq \"batman\""),
					Percentage: testconvert.Float64(40),
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Rollout: &flag.DtoRollout{
						Scheduled: &flag.DtoScheduledRollout{Steps: []flag.DtoScheduledStep{
							{
								DtoFlag: flag.DtoFlag{
									Rule: testconvert.String("key eq \"superman\""),
								},
								Date: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
							},
							{
								DtoFlag: flag.DtoFlag{
									True: testconvert.Interface("trueXXX"),
								},
								Date: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
							},
						}},
					},
				},
			},
			want: flag.FlagData{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("true"),
				},
				Rules: &map[string]flag.Rule{
					flag.LegacyRuleName: {
						Query: testconvert.String("key eq \"batman\""),
						Percentages: &map[string]float64{
							"True":  40,
							"False": 60,
						},
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
				Rollout: &flag.Rollout{
					Scheduled: &flag.ScheduledRollout{Steps: []flag.ScheduledStep{
						{
							FlagData: flag.FlagData{
								Rules: &map[string]flag.Rule{
									flag.LegacyRuleName: {
										Query: testconvert.String("key eq \"superman\""),
									},
								},
							},
							Date: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
						},
						{
							FlagData: flag.FlagData{
								Variations: &map[string]*interface{}{
									"True": testconvert.Interface("trueXXX"),
								},
							},
							Date: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
						},
					}},
				},
			},
		},
		{
			name: "convert v0 flag with experimentation",
			args: args{
				dtoFlag: flag.DtoFlag{
					Rule:       testconvert.String("key eq \"batman\""),
					Percentage: testconvert.Float64(40),
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Rollout: &flag.DtoRollout{
						Experimentation: &flag.Experimentation{
							Start: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
							End:   testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
						},
					},
				},
			},
			want: flag.FlagData{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("true"),
				},
				Rules: &map[string]flag.Rule{
					flag.LegacyRuleName: {
						Query: testconvert.String("key eq \"batman\""),
						Percentages: &map[string]float64{
							"True":  40,
							"False": 60,
						},
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
				Rollout: &flag.Rollout{
					Experimentation: &flag.Experimentation{
						Start: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
						End:   testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.args.dtoFlag.ConvertToFlagData(tt.args.isScheduleStep), "ConvertV0DtoToFlag(%v, %v)", tt.args.dtoFlag, tt.args.isScheduleStep)
		})
	}
}
