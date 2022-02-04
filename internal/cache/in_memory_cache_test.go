package cache_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"testing"
)

func TestAll(t *testing.T) {
	tests := []struct {
		name  string
		param map[string]flag.DtoFlag
		want  map[string]flag.Flag
	}{
		{
			name: "all with 1 flag",
			param: map[string]flag.DtoFlag{
				"test": {
					Percentage: testconvert.Float64(40),
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
				},
			},
			want: map[string]flag.Flag{
				"test": &flag.FlagData{
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface("default"),
						"False":   testconvert.Interface("false"),
						"True":    testconvert.Interface("true"),
					},
					Rules: &map[string]flag.Rule{
						flag.LegacyRuleName: {
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
		},
		{
			name: "all with multiple flags",
			param: map[string]flag.DtoFlag{
				"test": {
					Percentage: testconvert.Float64(40),
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
				},
				"test1": {
					Percentage: testconvert.Float64(30),
					True:       testconvert.Interface(true),
					False:      testconvert.Interface(false),
					Default:    testconvert.Interface(false),
				},
			},
			want: map[string]flag.Flag{
				"test": &flag.FlagData{
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface("default"),
						"False":   testconvert.Interface("false"),
						"True":    testconvert.Interface("true"),
					},
					Rules: &map[string]flag.Rule{
						flag.LegacyRuleName: {
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
				"test1": &flag.FlagData{
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(false),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					Rules: &map[string]flag.Rule{
						flag.LegacyRuleName: {
							Percentages: &map[string]float64{
								"True":  30,
								"False": 70,
							},
						},
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("Default"),
					},
				},
			},
		},
		{
			name:  "empty",
			param: map[string]flag.DtoFlag{},
			want:  map[string]flag.Flag{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cache.NewInMemoryCache(fflog.Logger{})
			c.Init(tt.param)
			assert.Equal(t, tt.want, c.All())
		})
	}
}

func TestCopy(t *testing.T) {
	tests := []struct {
		name  string
		param map[string]flag.DtoFlag
	}{
		{
			name: "copy with 1 flag",
			param: map[string]flag.DtoFlag{
				"test": {
					Percentage: testconvert.Float64(40),
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cache.NewInMemoryCache(fflog.Logger{})
			c.Init(tt.param)
			got := c.Copy()
			assert.Equal(t, c, got)
		})
	}
}
