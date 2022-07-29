package cache_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/dto"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
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
					DTOv0: dto.DTOv0{
						Percentage: testconvert.Float64(40),
						True:       testconvert.Interface("true"),
						False:      testconvert.Interface("false"),
						Default:    testconvert.Interface("default"),
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
						Name: testconvert.String("legacyDefaultRule"),
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
					DTOv0: dto.DTOv0{
						Percentage: testconvert.Float64(40),
						True:       testconvert.Interface("true"),
						False:      testconvert.Interface("false"),
						Default:    testconvert.Interface("default"),
					},
				},
				"test1": {
					DTOv0: dto.DTOv0{
						Percentage: testconvert.Float64(30),
						True:       testconvert.Interface(true),
						False:      testconvert.Interface(false),
						Default:    testconvert.Interface(false),
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
						Name: testconvert.String("legacyDefaultRule"),
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
						Name: testconvert.String("legacyDefaultRule"),
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
			c := cache.NewInMemoryCache()
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
					DTOv0: dto.DTOv0{
						Percentage: testconvert.Float64(40),
						True:       testconvert.Interface("true"),
						False:      testconvert.Interface("false"),
						Default:    testconvert.Interface("default"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cache.NewInMemoryCache()
			c.Init(tt.param)
			got := c.Copy()
			assert.Equal(t, c, got)
		})
	}
}
