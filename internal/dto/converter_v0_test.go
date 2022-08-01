package dto_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/dto"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
)

func TestConvertV0DtoToInternalFlag(t *testing.T) {
	tests := []struct {
		name string
		d    dto.DTOv0
		want flag.InternalFlag
	}{
		{
			name: "Simplest flag, no converter provided",
			d: dto.DTOv0{
				True:    testconvert.Interface("true"),
				False:   testconvert.Interface("false"),
				Default: testconvert.Interface("default"),
			},
			want: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("true"),
				},
				Rules: nil,
				DefaultRule: &flag.Rule{
					Name: testconvert.String("legacyDefaultRule"),
					Percentages: &map[string]float64{
						"True":  0,
						"False": 100,
					},
				},
			},
		},
		{
			name: "Flag with percentage",
			d: dto.DTOv0{
				Percentage: testconvert.Float64(10),
				True:       testconvert.Interface("true"),
				False:      testconvert.Interface("false"),
				Default:    testconvert.Interface("default"),
			},
			want: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("true"),
				},
				DefaultRule: &flag.Rule{
					Name: testconvert.String("legacyDefaultRule"),
					Percentages: &map[string]float64{
						"True":  10,
						"False": 90,
					},
				},
			},
		},
		{
			name: "Flag with 100 percentage",
			d: dto.DTOv0{
				Percentage: testconvert.Float64(100),
				True:       testconvert.Interface("true"),
				False:      testconvert.Interface("false"),
				Default:    testconvert.Interface("default"),
			},
			want: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("true"),
				},
				Rules: nil,
				DefaultRule: &flag.Rule{
					Name: testconvert.String("legacyDefaultRule"),
					Percentages: &map[string]float64{
						"True":  100,
						"False": 0,
					},
				},
			},
		},
		{
			name: "Flag with rule not match",
			d: dto.DTOv0{
				Rule:    testconvert.String("key eq \"random\""),
				True:    testconvert.Interface("true"),
				False:   testconvert.Interface("false"),
				Default: testconvert.Interface("default"),
			},
			want: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("true"),
				},
				Rules: &[]flag.Rule{
					{
						Name:  testconvert.String("legacyRuleV0"),
						Query: testconvert.String("key eq \"random\""),
						Percentages: &map[string]float64{
							"True":  0,
							"False": 100,
						},
					},
				},
				DefaultRule: &flag.Rule{
					Name:            testconvert.String("legacyDefaultRule"),
					VariationResult: testconvert.String("Default"),
				},
			},
		},
		{
			name: "Flag with rule match",
			d: dto.DTOv0{
				Rule:    testconvert.String("key eq \"test-user\""),
				True:    testconvert.Interface("true"),
				False:   testconvert.Interface("false"),
				Default: testconvert.Interface("default"),
			},
			want: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("true"),
				},
				Rules: &[]flag.Rule{
					{
						Name:  testconvert.String("legacyRuleV0"),
						Query: testconvert.String("key eq \"test-user\""),
						Percentages: &map[string]float64{
							"True":  0,
							"False": 100,
						},
					},
				},
				DefaultRule: &flag.Rule{
					Name:            testconvert.String("legacyDefaultRule"),
					VariationResult: testconvert.String("Default"),
				},
			},
		},
		{
			name: "Flag with rule match + 10% percentage",
			d: dto.DTOv0{
				Rule:       testconvert.String("key eq \"test-user\""),
				Percentage: testconvert.Float64(10),
				True:       testconvert.Interface("true"),
				False:      testconvert.Interface("false"),
				Default:    testconvert.Interface("default"),
			},
			want: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("true"),
				},
				Rules: &[]flag.Rule{
					{
						Name:  testconvert.String("legacyRuleV0"),
						Query: testconvert.String("key eq \"test-user\""),
						Percentages: &map[string]float64{
							"True":  10,
							"False": 90,
						},
					},
				},
				DefaultRule: &flag.Rule{
					Name:            testconvert.String("legacyDefaultRule"),
					VariationResult: testconvert.String("Default"),
				},
			},
		},
		{
			name: "Flag with query + experimentation rollout",
			d: dto.DTOv0{
				Rule:       testconvert.String("key eq \"test-user\""),
				Percentage: testconvert.Float64(100),
				True:       testconvert.Interface("true"),
				False:      testconvert.Interface("false"),
				Default:    testconvert.Interface("default"),
				Rollout: &dto.RolloutV0{
					Experimentation: &dto.ExperimentationV0{
						Start: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
						End:   testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
					},
				},
			},
			want: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("true"),
				},
				Rules: &[]flag.Rule{
					{
						Name:  testconvert.String("legacyRuleV0"),
						Query: testconvert.String("key eq \"test-user\""),
						Percentages: &map[string]float64{
							"True":  100,
							"False": 0,
						},
					},
				},
				DefaultRule: &flag.Rule{
					Name:            testconvert.String("legacyDefaultRule"),
					VariationResult: testconvert.String("Default"),
				},
				Rollout: &flag.Rollout{
					Experimentation: &flag.ExperimentationRollout{
						Start: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
						End:   testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
					},
				},
			},
		},
		{
			name: "Flag with query + progressive rollout",
			d: dto.DTOv0{
				Rule:       testconvert.String("key eq \"test-user\""),
				Percentage: testconvert.Float64(100),
				True:       testconvert.Interface("true"),
				False:      testconvert.Interface("false"),
				Default:    testconvert.Interface("default"),
				Rollout: &dto.RolloutV0{
					Progressive: &dto.ProgressiveV0{
						Percentage: dto.ProgressivePercentageV0{
							Initial: 0,
							End:     100,
						},
						ReleaseRamp: dto.ProgressiveReleaseRampV0{
							Start: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
							End:   testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
						},
					},
				},
			},
			want: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("true"),
				},
				Rules: &[]flag.Rule{
					{
						Name:  testconvert.String("legacyRuleV0"),
						Query: testconvert.String("key eq \"test-user\""),
						ProgressiveRollout: &flag.ProgressiveRollout{
							Initial: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("False"),
								Percentage: testconvert.Float64(0),
								Date:       testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
							},
							End: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("True"),
								Percentage: testconvert.Float64(100),
								Date:       testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
							},
						},
					},
				},
				DefaultRule: &flag.Rule{
					Name:            testconvert.String("legacyDefaultRule"),
					VariationResult: testconvert.String("Default"),
				},
			},
		},
		{
			name: "Flag without query + progressive rollout",
			d: dto.DTOv0{
				True:    testconvert.Interface("true"),
				False:   testconvert.Interface("false"),
				Default: testconvert.Interface("default"),
				Rollout: &dto.RolloutV0{
					Progressive: &dto.ProgressiveV0{
						Percentage: dto.ProgressivePercentageV0{
							Initial: 0,
							End:     100,
						},
						ReleaseRamp: dto.ProgressiveReleaseRampV0{
							Start: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
							End:   testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
						},
					},
				},
			},
			want: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("true"),
				},
				DefaultRule: &flag.Rule{
					Name: testconvert.String("legacyDefaultRule"),
					ProgressiveRollout: &flag.ProgressiveRollout{
						Initial: &flag.ProgressiveRolloutStep{
							Variation:  testconvert.String("False"),
							Percentage: testconvert.Float64(0),
							Date:       testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
						},
						End: &flag.ProgressiveRolloutStep{
							Variation:  testconvert.String("True"),
							Percentage: testconvert.Float64(100),
							Date:       testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
						},
					},
				},
			},
		},
		{
			name: "Flag with percentage + scheduled step",
			d: dto.DTOv0{
				Percentage: testconvert.Float64(10),
				True:       testconvert.Interface("true"),
				False:      testconvert.Interface("false"),
				Default:    testconvert.Interface("default"),
				Rollout: &dto.RolloutV0{
					Scheduled: &dto.ScheduledRolloutV0{Steps: []dto.ScheduledStepV0{
						{
							DTOv0: dto.DTOv0{
								Percentage: testconvert.Float64(20),
							},
							Date: testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
						},
						{
							DTOv0: dto.DTOv0{
								True: testconvert.Interface("true2"),
							},
							Date: testconvert.Time(time.Date(2021, time.February, 3, 10, 10, 10, 10, time.UTC)),
						},
						{
							DTOv0: dto.DTOv0{
								Rule: testconvert.String("key eq \"test-user\""),
							},
							Date: testconvert.Time(time.Date(2021, time.February, 4, 10, 10, 10, 10, time.UTC)),
						},
					}},
				},
			},
			want: flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("true"),
				},
				DefaultRule: &flag.Rule{
					Name: testconvert.String("legacyDefaultRule"),
					Percentages: &map[string]float64{
						"True":  10,
						"False": 90,
					},
				},
				Rollout: &flag.Rollout{
					Scheduled: &[]flag.ScheduledStep{
						{
							InternalFlag: flag.InternalFlag{
								Variations: &map[string]*interface{}{
									"Default": testconvert.Interface("default"),
									"False":   testconvert.Interface("false"),
									"True":    testconvert.Interface("true"),
								},
								DefaultRule: &flag.Rule{
									Name: testconvert.String("legacyDefaultRule"),
									Percentages: &map[string]float64{
										"False": 80,
										"True":  20,
									},
								},
							},
							Date: testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
						},
						{
							InternalFlag: flag.InternalFlag{
								Variations: &map[string]*interface{}{
									"Default": testconvert.Interface("default"),
									"False":   testconvert.Interface("false"),
									"True":    testconvert.Interface("true2"),
								},
								DefaultRule: &flag.Rule{
									Name: testconvert.String("legacyDefaultRule"),
									Percentages: &map[string]float64{
										"False": 80,
										"True":  20,
									},
								},
							},
							Date: testconvert.Time(time.Date(2021, time.February, 3, 10, 10, 10, 10, time.UTC)),
						},
						{
							InternalFlag: flag.InternalFlag{
								Variations: &map[string]*interface{}{
									"Default": testconvert.Interface("default"),
									"False":   testconvert.Interface("false"),
									"True":    testconvert.Interface("true2"),
								},
								Rules: &[]flag.Rule{
									{
										Name:  testconvert.String("legacyRuleV0"),
										Query: testconvert.String("key eq \"test-user\""),
										Percentages: &map[string]float64{
											"False": 80,
											"True":  20,
										},
									},
								},
								DefaultRule: &flag.Rule{
									Name:            testconvert.String("legacyDefaultRule"),
									VariationResult: testconvert.String("Default"),
								},
							},
							Date: testconvert.Time(time.Date(2021, time.February, 4, 10, 10, 10, 10, time.UTC)),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, dto.ConvertV0DtoToInternalFlag(tt.d, false))
		})
	}
}
