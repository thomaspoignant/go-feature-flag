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
			want: map[string]flag.Flag{
				"test": &flag.FlagData{
					Variations: &map[string]*interface{}{
						"A": testconvert.Interface("true"),
						"B": testconvert.Interface("false"),
						"C": testconvert.Interface("C"),
						"D": testconvert.Interface("D"),
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{
							"A": 10,
							"B": 10,
							"C": 10,
							"D": 70,
						},
					},
					Rules: &map[string]flag.Rule{
						"testRule": {
							Query:           testconvert.String("key eq \"marty\""),
							VariationResult: testconvert.String("A"),
						},
					},
				},
			},
		},
		{
			name: "all with multiple flags",
			param: map[string]flag.DtoFlag{
				"test": {
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
						VariationResult: testconvert.String("D"),
					},
				},
				"test1": {
					Variations: &map[string]*interface{}{
						"E": testconvert.Interface("E"),
						"F": testconvert.Interface("F"),
						"H": testconvert.Interface("HH"),
					},
					Rules: &map[string]flag.Rule{
						"testRule": {
							Query:           testconvert.String("key eq \"mcfly\""),
							VariationResult: testconvert.String("G"),
						},
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{
							"E": 10,
							"F": 10,
							"H": 80,
						},
					},
				},
			},
			want: map[string]flag.Flag{
				"test": &flag.FlagData{
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
						VariationResult: testconvert.String("D"),
					},
				},
				"test1": &flag.FlagData{
					Variations: &map[string]*interface{}{
						"E": testconvert.Interface("E"),
						"F": testconvert.Interface("F"),
						"H": testconvert.Interface("HH"),
					},
					Rules: &map[string]flag.Rule{
						"testRule": {
							Query:           testconvert.String("key eq \"mcfly\""),
							VariationResult: testconvert.String("G"),
						},
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{
							"E": 10,
							"F": 10,
							"H": 80,
						},
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
					Variations: &map[string]*interface{}{
						"E": testconvert.Interface("E"),
						"F": testconvert.Interface("F"),
						"H": testconvert.Interface("HH"),
					},
					Rules: &map[string]flag.Rule{
						"testRule": {
							Query:           testconvert.String("key eq \"mcfly\""),
							VariationResult: testconvert.String("G"),
						},
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{
							"E": 10,
							"F": 10,
							"H": 80,
						},
					},
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
