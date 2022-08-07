package dto_test

import (
	"encoding/json"
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/testutils/flagv1"
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
						//{ScheduledStepV0
						//	DTOv0: dto.DTOv0{
						//		Percentage: testconvert.Float64(20),
						//	},
						//	Date: testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
						// },
						//{
						//	DTOv0: dto.DTOv0{
						//		True: testconvert.Interface("true2"),
						//	},
						//	Date: testconvert.Time(time.Date(2021, time.February, 3, 10, 10, 10, 10, time.UTC)),
						// },
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
			res := dto.ConvertV0DtoToInternalFlag(tt.d, false)
			m, _ := json.Marshal(res)
			fmt.Println(string(m))
			assert.Equal(t, tt.want, dto.ConvertV0DtoToInternalFlag(tt.d, false))
		})
	}
}

func TestConvertV0ScheduleStep(t *testing.T) {
	tests := []struct {
		name string
		dto  dto.DTO
	}{
		{
			name: "Update a rule that exists already",
			dto: dto.DTO{
				DTOv0: dto.DTOv0{
					Rule:       testconvert.String("key eq \"yo\""),
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Percentage: testconvert.Float64(95),
					Rollout: &dto.RolloutV0{
						Scheduled: &dto.ScheduledRolloutV0{Steps: []dto.ScheduledStepV0{
							{
								DTOv0: dto.DTOv0{
									Rule: testconvert.String("anonymous eq false"),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
						}},
					},
				},
			},
		},
		{
			name: "Update a rule that exists already + percentages",
			dto: dto.DTO{
				DTOv0: dto.DTOv0{
					Rule:       testconvert.String("key eq \"yo\""),
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Percentage: testconvert.Float64(95),
					Rollout: &dto.RolloutV0{
						Scheduled: &dto.ScheduledRolloutV0{Steps: []dto.ScheduledStepV0{
							{
								DTOv0: dto.DTOv0{
									Rule:       testconvert.String("anonymous eq false"),
									Percentage: testconvert.Float64(5),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
						}},
					},
				},
			},
		},
		{
			name: "No rule, update only percentages",
			dto: dto.DTO{
				DTOv0: dto.DTOv0{
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Percentage: testconvert.Float64(100),
					Rollout: &dto.RolloutV0{
						Scheduled: &dto.ScheduledRolloutV0{Steps: []dto.ScheduledStepV0{
							{
								DTOv0: dto.DTOv0{
									Percentage: testconvert.Float64(10),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
						}},
					},
				},
			},
		},
		{
			name: "No rule, add rule which not match",
			dto: dto.DTO{
				DTOv0: dto.DTOv0{
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Percentage: testconvert.Float64(100),
					Rollout: &dto.RolloutV0{
						Scheduled: &dto.ScheduledRolloutV0{Steps: []dto.ScheduledStepV0{
							{
								DTOv0: dto.DTOv0{
									Rule: testconvert.String("key eq \"ko\""),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
						}},
					},
				},
			},
		},
		{
			name: "No rule, add rule which match",
			dto: dto.DTO{
				DTOv0: dto.DTOv0{
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Percentage: testconvert.Float64(100),
					Rollout: &dto.RolloutV0{
						Scheduled: &dto.ScheduledRolloutV0{Steps: []dto.ScheduledStepV0{
							{
								DTOv0: dto.DTOv0{
									Rule: testconvert.String("key eq \"yo\""),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
						}},
					},
				},
			},
		},
		{
			name: "Change value of a variation",
			dto: dto.DTO{
				DTOv0: dto.DTOv0{
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Percentage: testconvert.Float64(100),
					Rollout: &dto.RolloutV0{
						Scheduled: &dto.ScheduledRolloutV0{Steps: []dto.ScheduledStepV0{
							{
								DTOv0: dto.DTOv0{
									True: testconvert.Interface("newValue"),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
						}},
					},
				},
			},
		},
		{
			name: "Change percentage with no rule",
			dto: dto.DTO{
				DTOv0: dto.DTOv0{
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Percentage: testconvert.Float64(100),
					Rollout: &dto.RolloutV0{
						Scheduled: &dto.ScheduledRolloutV0{Steps: []dto.ScheduledStepV0{
							{
								DTOv0: dto.DTOv0{
									Percentage: testconvert.Float64(10),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
						}},
					},
				},
			},
		},
		{
			name: "Change percentage and add rule (not in percentage)",
			dto: dto.DTO{
				DTOv0: dto.DTOv0{
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Percentage: testconvert.Float64(100),
					Rollout: &dto.RolloutV0{
						Scheduled: &dto.ScheduledRolloutV0{Steps: []dto.ScheduledStepV0{
							{
								DTOv0: dto.DTOv0{
									Rule:       testconvert.String("key eq \"yo\""),
									Percentage: testconvert.Float64(10),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
						}},
					},
				},
			},
		},
		{
			name: "Change percentage and add rule (in percentage)",
			dto: dto.DTO{
				DTOv0: dto.DTOv0{
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Percentage: testconvert.Float64(100),
					Rollout: &dto.RolloutV0{
						Scheduled: &dto.ScheduledRolloutV0{Steps: []dto.ScheduledStepV0{
							{
								DTOv0: dto.DTOv0{
									Rule:       testconvert.String("key eq \"yo\""),
									Percentage: testconvert.Float64(50),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
						}},
					},
				},
			},
		},
		{
			name: "Add rule and remove rule + change percentages",
			dto: dto.DTO{
				DTOv0: dto.DTOv0{
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Rule:       testconvert.String("key eq \"yo\""),
					Percentage: testconvert.Float64(100),
					Rollout: &dto.RolloutV0{
						Scheduled: &dto.ScheduledRolloutV0{Steps: []dto.ScheduledStepV0{
							{
								DTOv0: dto.DTOv0{
									Rule: testconvert.String("key eq \"yo\""),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
							{
								DTOv0: dto.DTOv0{
									Rule: testconvert.String(""),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
							{
								DTOv0: dto.DTOv0{
									Percentage: testconvert.Float64(10),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
						}},
					},
				},
			},
		},
		{
			name: "Rule with percentages, remove rule",
			dto: dto.DTO{
				DTOv0: dto.DTOv0{
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Rule:       testconvert.String("key eq \"yo\""),
					Percentage: testconvert.Float64(95),
					Rollout: &dto.RolloutV0{
						Scheduled: &dto.ScheduledRolloutV0{Steps: []dto.ScheduledStepV0{
							{
								DTOv0: dto.DTOv0{
									Rule: testconvert.String(""),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
						}},
					},
				},
			},
		},
		{
			name: "XXX",
			dto: dto.DTO{
				DTOv0: dto.DTOv0{
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
					Rule:       testconvert.String("key eq \"yo\""),
					Percentage: testconvert.Float64(100),
					Rollout: &dto.RolloutV0{
						Scheduled: &dto.ScheduledRolloutV0{Steps: []dto.ScheduledStepV0{
							{
								DTOv0: dto.DTOv0{
									Rule: testconvert.String("key eq \"oy\""),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
							{
								DTOv0: dto.DTOv0{
									Percentage: testconvert.Float64(95),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
							{
								DTOv0: dto.DTOv0{
									Rule: testconvert.String("key eq \"yo\""),
								},
								Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
							},
						}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := ffuser.NewUser("yo")
			flagName := "yo"

			convertInternalFlag := tt.dto.Convert()
			gotInternalFlag, _ := convertInternalFlag.Value(flagName, u, flag.EvaluationContext{})

			convertFlagv1 := flagv1.ConvertDtoToV1(tt.dto.DTOv0)
			gotFlagV1, _ := convertFlagv1.Value(flagName, u, flag.EvaluationContext{})

			fmt.Println(gotFlagV1, gotInternalFlag)
			assert.Equal(t, gotFlagV1, gotInternalFlag)
		})
	}

}
