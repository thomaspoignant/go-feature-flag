package cache_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
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
					Variations: &map[string]*interface{}{
						"test":  testconvert.Interface(true),
						"test2": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{"test": 10, "test2": 90},
					},
				},
			},
			want: map[string]flag.Flag{
				"test": &flag.FlagData{
					Variations: &map[string]*interface{}{
						"test":  testconvert.Interface(true),
						"test2": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{"test": 10, "test2": 90},
					},
				},
			},
		},
		{
			name: "all with multiple flags",
			param: map[string]flag.DtoFlag{
				"test": {
					Variations: &map[string]*interface{}{
						"test":  testconvert.Interface(true),
						"test2": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{"test": 10, "test2": 90},
					},
				},
				"test1": {
					Variations: &map[string]*interface{}{
						"test":  testconvert.Interface(true),
						"test2": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{"test": 0, "test2": 100},
					},
				},
			},
			want: map[string]flag.Flag{
				"test": &flag.FlagData{
					Variations: &map[string]*interface{}{
						"test":  testconvert.Interface(true),
						"test2": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{"test": 10, "test2": 90},
					},
				},
				"test1": &flag.FlagData{
					Variations: &map[string]*interface{}{
						"test":  testconvert.Interface(true),
						"test2": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{"test": 0, "test2": 100},
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
			c := cache.NewInMemoryCache()
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
					Variations: &map[string]*interface{}{
						"True":    testconvert.Interface(true),
						"False":   testconvert.Interface(false),
						"Default": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						Query: testconvert.String("key eq \"random-key\""),
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
			c := cache.NewInMemoryCache()
			c.Init(tt.param)
			got := c.Copy()
			assert.Equal(t, c, got)
		})
	}
}
