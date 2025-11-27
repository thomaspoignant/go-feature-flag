package dto_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestConvertV1DtoToInternalFlag(t *testing.T) {
	tests := []struct {
		name     string
		input    dto.DTO
		expected flag.InternalFlag
	}{
		{
			name: "all fields set",
			input: dto.DTO{
				BucketingKey: testconvert.String("bucketKey"),
				Variations: &map[string]*interface{}{
					"var1": testconvert.Interface("var1"),
					"var2": testconvert.Interface("var2"),
				},
				Rules: &[]flag.Rule{{
					Name:        testconvert.String("rule1"),
					Query:       testconvert.String("key eq \"key\""),
					Percentages: &map[string]float64{"var_true": 100, "var_false": 0}},
					{
						Name:  testconvert.String("rule2"),
						Query: testconvert.String("key eq \"key2\""),
						ProgressiveRollout: &flag.ProgressiveRollout{
							Initial: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("var1"),
								Percentage: testconvert.Float64(30),
								Date: testconvert.Time(
									time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
								),
							},
							End: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("var2"),
								Percentage: testconvert.Float64(70),
								Date: testconvert.Time(
									time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
								),
							},
						},
					},
					{
						Name:            testconvert.String("rule3"),
						Query:           testconvert.String("key eq \"key3\""),
						VariationResult: testconvert.String("var2"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("var1"),
				},
				Scheduled: &[]flag.ScheduledStep{
					{
						Date: testconvert.Time(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
						InternalFlag: flag.InternalFlag{
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("var2"),
							},
						},
					},
				},
				Experimentation: &dto.ExperimentationDto{
					Start: testconvert.Time(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
					End:   testconvert.Time(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
				},
				Metadata:    &map[string]interface{}{"key": "value"},
				TrackEvents: testconvert.Bool(true),
				Disable:     testconvert.Bool(false),
				Version:     testconvert.String("v1"),
			},
			expected: flag.InternalFlag{
				BucketingKey: testconvert.String("bucketKey"),
				Variations: &map[string]*interface{}{
					"var1": testconvert.Interface("var1"),
					"var2": testconvert.Interface("var2"),
				},
				Rules: &[]flag.Rule{{
					Name:        testconvert.String("rule1"),
					Query:       testconvert.String("key eq \"key\""),
					Percentages: &map[string]float64{"var_true": 100, "var_false": 0}},
					{
						Name:  testconvert.String("rule2"),
						Query: testconvert.String("key eq \"key2\""),
						ProgressiveRollout: &flag.ProgressiveRollout{
							Initial: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("var1"),
								Percentage: testconvert.Float64(30),
								Date: testconvert.Time(
									time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
								),
							},
							End: &flag.ProgressiveRolloutStep{
								Variation:  testconvert.String("var2"),
								Percentage: testconvert.Float64(70),
								Date: testconvert.Time(
									time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
								),
							},
						},
					},
					{
						Name:            testconvert.String("rule3"),
						Query:           testconvert.String("key eq \"key3\""),
						VariationResult: testconvert.String("var2"),
					},
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("var1"),
				},
				TrackEvents: testconvert.Bool(true),
				Disable:     testconvert.Bool(false),
				Version:     testconvert.String("v1"),
				Scheduled: &[]flag.ScheduledStep{
					{
						Date: testconvert.Time(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
						InternalFlag: flag.InternalFlag{
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("var2"),
							},
						},
					},
				},
				Experimentation: &flag.ExperimentationRollout{
					Start: testconvert.Time(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
					End:   testconvert.Time(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
				},
				Metadata: &map[string]interface{}{"key": "value"},
			},
		},
		{
			name:     "nil input",
			input:    dto.DTO{},
			expected: flag.InternalFlag{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Convert()
			assert.Equal(t, tt.expected, result)
		})
	}
}
