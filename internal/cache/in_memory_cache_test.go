package cache_test

import (
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/dto"
	"github.com/thomaspoignant/go-feature-flag/internal/flagv1"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
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
				"test": &flagv1.FlagData{
					Percentage: testconvert.Float64(40),
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
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
				"test": &flagv1.FlagData{
					Percentage: testconvert.Float64(40),
					True:       testconvert.Interface("true"),
					False:      testconvert.Interface("false"),
					Default:    testconvert.Interface("default"),
				},
				"test1": &flagv1.FlagData{
					Percentage: testconvert.Float64(30),
					True:       testconvert.Interface(true),
					False:      testconvert.Interface(false),
					Default:    testconvert.Interface(false),
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
