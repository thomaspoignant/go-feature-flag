package cache_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/model/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestAll(t *testing.T) {
	tests := []struct {
		name  string
		param map[string]dto.DTO
		want  map[string]flag.Flag
	}{
		{
			name: "all with 1 flag",
			param: map[string]dto.DTO{
				"test": {
					Variations: &map[string]*interface{}{
						"True":    testconvert.Interface("true"),
						"False":   testconvert.Interface("false"),
						"Default": testconvert.Interface("default"),
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{
							"True":  40,
							"False": 60,
						},
					},
				},
			},
			want: map[string]flag.Flag{
				"test": &flag.InternalFlag{
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface("default"),
						"False":   testconvert.Interface("false"),
						"True":    testconvert.Interface("true"),
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{
							"False": 60,
							"True":  40,
						},
					},
				},
			},
		},
		{
			name: "all with multiple flags",
			param: map[string]dto.DTO{
				"test": {
					Variations: &map[string]*interface{}{
						"True":    testconvert.Interface("true"),
						"False":   testconvert.Interface("false"),
						"Default": testconvert.Interface("default"),
					},
					DefaultRule: &flag.Rule{
						Name: testconvert.String("defaultRule"),
						Percentages: &map[string]float64{
							"True":  40,
							"False": 60,
						},
					},
				},
				"test1": {
					Variations: &map[string]*interface{}{
						"True":    testconvert.Interface(true),
						"False":   testconvert.Interface(false),
						"Default": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						Name: testconvert.String("defaultRule"),
						Percentages: &map[string]float64{
							"True":  30,
							"False": 70,
						},
					},
				},
			},
			want: map[string]flag.Flag{
				"test": &flag.InternalFlag{
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface("default"),
						"False":   testconvert.Interface("false"),
						"True":    testconvert.Interface("true"),
					},
					DefaultRule: &flag.Rule{
						Name: testconvert.String("defaultRule"),
						Percentages: &map[string]float64{
							"False": 60,
							"True":  40,
						},
					},
				},
				"test1": &flag.InternalFlag{
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(false),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						Name: testconvert.String("defaultRule"),
						Percentages: &map[string]float64{
							"False": 70,
							"True":  30,
						},
					},
				},
			},
		},
		{
			name:  "empty",
			param: map[string]dto.DTO{},
			want:  map[string]flag.Flag{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cache.NewInMemoryCache(nil)
			c.Init(tt.param)
			assert.Equal(t, tt.want, c.All())
		})
	}
}

func TestCopy(t *testing.T) {
	tests := []struct {
		name  string
		param map[string]dto.DTO
	}{
		{
			name: "copy with 1 flag",
			param: map[string]dto.DTO{
				"test": {
					Variations: &map[string]*interface{}{
						"True":    testconvert.Interface("true"),
						"False":   testconvert.Interface("false"),
						"Default": testconvert.Interface("default"),
					},
					DefaultRule: &flag.Rule{
						Name: testconvert.String("defaultRule"),
						Percentages: &map[string]float64{
							"True":  40,
							"False": 60,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cache.NewInMemoryCache(nil)
			c.Init(tt.param)
			got := c.Copy()
			assert.Equal(t, c, got)
		})
	}
}
