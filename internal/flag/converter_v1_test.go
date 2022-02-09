package flag_test

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"testing"
	"time"
)

func TestConvertV1DtoToFlag(t *testing.T) {
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
			name: "convert simple v1 flag",
			args: args{
				dtoFlag: flag.DtoFlag{
					Variations: &map[string]*interface{}{
						"A": testconvert.Interface("true"),
						"B": testconvert.Interface("false"),
						"C": testconvert.Interface("C"),
						"D": testconvert.Interface("D"),
					},
					Rules: &map[string]flag.Rule{
						"testRule": {
							Query:           testconvert.String("key eq \"marty\""),
							VariationResult: testconvert.String("A"),
						},
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{
							"A": 10,
							"B": 10,
							"C": 10,
							"D": 70,
						},
					},
				},
			},
			want: flag.FlagData{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("true"),
					"B": testconvert.Interface("false"),
					"C": testconvert.Interface("C"),
					"D": testconvert.Interface("D"),
				},
				Rules: &map[string]flag.Rule{
					"testRule": {
						Query:           testconvert.String("key eq \"marty\""),
						VariationResult: testconvert.String("A"),
					},
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"A": 10,
						"B": 10,
						"C": 10,
						"D": 70,
					},
				},
			},
		},
		{
			name: "convert v1 flag with progressive rollout",
			args: args{
				dtoFlag: flag.DtoFlag{
					Variations: &map[string]*interface{}{
						"A": testconvert.Interface("true"),
						"B": testconvert.Interface("false"),
						"C": testconvert.Interface("C"),
						"D": testconvert.Interface("D"),
					},
					Rules: &map[string]flag.Rule{
						"testRule": {
							Query: testconvert.String("key eq \"marty\""),
							ProgressiveRollout: &flag.ProgressiveRollout{
								Initial: &flag.ProgressiveRolloutStep{
									Variation:  testconvert.String("A"),
									Percentage: 0,
									Date:       testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
								},
								End: &flag.ProgressiveRolloutStep{
									Variation:  testconvert.String("b"),
									Percentage: 100,
									Date:       testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
								},
							},
						},
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{
							"A": 10,
							"B": 10,
							"C": 10,
							"D": 70,
						},
					},
				},
			},
			want: flag.FlagData{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("true"),
					"B": testconvert.Interface("false"),
					"C": testconvert.Interface("C"),
					"D": testconvert.Interface("D"),
				},
				Rules: &map[string]flag.Rule{
					"testRule": {
						Query: testconvert.String("key eq \"marty\""),
						ProgressiveRollout: &flag.ProgressiveRollout{
							Initial: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("A"),
								Percentage: 0,
								Date:       testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
							},
							End: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("b"),
								Percentage: 100,
								Date:       testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
							},
						},
					},
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"A": 10,
						"B": 10,
						"C": 10,
						"D": 70,
					},
				},
			},
		},
		{
			name: "convert v1 flag with scheduled rollout",
			args: args{
				dtoFlag: flag.DtoFlag{
					Variations: &map[string]*interface{}{
						"A": testconvert.Interface("true"),
						"B": testconvert.Interface("false"),
						"C": testconvert.Interface("C"),
						"D": testconvert.Interface("D"),
					},
					Rules: &map[string]flag.Rule{
						"testRule": {
							Query:           testconvert.String("key eq \"marty\""),
							VariationResult: testconvert.String("A"),
						},
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{
							"A": 10,
							"B": 10,
							"C": 10,
							"D": 70,
						},
					},
					Rollout: &flag.DtoRollout{
						Scheduled: &flag.DtoScheduledRollout{Steps: []flag.DtoScheduledStep{
							{
								DtoFlag: flag.DtoFlag{
									Rules: &map[string]flag.Rule{
										"testRule": {
											Query: testconvert.String("key eq \"superman\""),
										},
									},
								},
								Date: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
							},
							{
								DtoFlag: flag.DtoFlag{
									Variations: &map[string]*interface{}{
										"A": testconvert.Interface("trueXXX"),
									},
								},
								Date: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
							},
						}},
					},
				},
			},
			want: flag.FlagData{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface("true"),
					"B": testconvert.Interface("false"),
					"C": testconvert.Interface("C"),
					"D": testconvert.Interface("D"),
				},
				Rules: &map[string]flag.Rule{
					"testRule": {
						Query:           testconvert.String("key eq \"marty\""),
						VariationResult: testconvert.String("A"),
					},
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"A": 10,
						"B": 10,
						"C": 10,
						"D": 70,
					},
				},
				Rollout: &flag.Rollout{
					Scheduled: &flag.ScheduledRollout{Steps: []flag.ScheduledStep{
						{
							FlagData: flag.FlagData{
								Rules: &map[string]flag.Rule{
									"testRule": {
										Query: testconvert.String("key eq \"superman\""),
									},
								},
							},
							Date: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
						},
						{
							FlagData: flag.FlagData{
								Variations: &map[string]*interface{}{
									"A": testconvert.Interface("trueXXX"),
								},
							},
							Date: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
						},
					}},
				},
			},
		},
		{
			name: "convert v1 flag with experimentation",
			args: args{
				dtoFlag: flag.DtoFlag{
					Variations: &map[string]*interface{}{
						"A": testconvert.Interface("true"),
						"B": testconvert.Interface("false"),
						"C": testconvert.Interface("C"),
						"D": testconvert.Interface("D"),
					},
					Rules: &map[string]flag.Rule{
						"testRule": {
							Query:           testconvert.String("key eq \"doc\""),
							VariationResult: testconvert.String("A"),
						},
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{
							"A": 10,
							"B": 10,
							"C": 10,
							"D": 70,
						},
					},
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
					"A": testconvert.Interface("true"),
					"B": testconvert.Interface("false"),
					"C": testconvert.Interface("C"),
					"D": testconvert.Interface("D"),
				},
				Rules: &map[string]flag.Rule{
					"testRule": {
						Query:           testconvert.String("key eq \"doc\""),
						VariationResult: testconvert.String("A"),
					},
				},
				DefaultRule: &flag.Rule{
					Percentages: &map[string]float64{
						"A": 10,
						"B": 10,
						"C": 10,
						"D": 70,
					},
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
			fmt.Println(cmp.Diff(tt.want, flag.ConvertV1DtoToFlag(tt.args.dtoFlag, tt.args.isScheduleStep)))
			assert.Equalf(t, tt.want, flag.ConvertV1DtoToFlag(tt.args.dtoFlag, tt.args.isScheduleStep), "ConvertV1DtoToFlag(%v, %v)", tt.args.dtoFlag, tt.args.isScheduleStep)
		})
	}
}
