package cache

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"sync"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/internal/notifier"
)

func Test_notificationService_getDifferences(t *testing.T) {
	type fields struct {
		Notifiers []notifier.Notifier
	}
	type args struct {
		oldCache map[string]flag.Flag
		newCache map[string]flag.Flag
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   model.DiffCache
	}{
		{
			name: "Delete flag",
			args: args{
				oldCache: map[string]flag.Flag{
					"test-flag": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"True":    testconvert.Interface(true),
							"False":   testconvert.Interface(false),
							"Default": testconvert.Interface(false),
						},
						Rules: nil,
						DefaultRule: &flag.Rule{
							Percentages: &map[string]float64{"True": 100, "False": 0},
						},
					},
					"test-flag2": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"True":    testconvert.Interface(true),
							"False":   testconvert.Interface(false),
							"Default": testconvert.Interface(false),
						},
						Rules: nil,
						DefaultRule: &flag.Rule{
							Percentages: &map[string]float64{"True": 100, "False": 0},
						},
					},
				},
				newCache: map[string]flag.Flag{
					"test-flag": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"True":    testconvert.Interface(true),
							"False":   testconvert.Interface(false),
							"Default": testconvert.Interface(false),
						},
						Rules: nil,
						DefaultRule: &flag.Rule{
							Percentages: &map[string]float64{"True": 100, "False": 0},
						},
					},
				},
			},
			want: model.DiffCache{
				Deleted: map[string]flag.Flag{
					"test-flag2": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"True":    testconvert.Interface(true),
							"False":   testconvert.Interface(false),
							"Default": testconvert.Interface(false),
						},
						Rules: nil,
						DefaultRule: &flag.Rule{
							Percentages: &map[string]float64{"True": 100, "False": 0},
						},
					},
				},
				Added:   map[string]flag.Flag{},
				Updated: map[string]model.DiffUpdated{},
			},
		},
		{
			name: "Added flag",
			args: args{
				oldCache: map[string]flag.Flag{
					"test-flag": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"True":    testconvert.Interface(true),
							"False":   testconvert.Interface(false),
							"Default": testconvert.Interface(false),
						},
						Rules: nil,
						DefaultRule: &flag.Rule{
							Percentages: &map[string]float64{"True": 100, "False": 0},
						},
					},
				},
				newCache: map[string]flag.Flag{
					"test-flag": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"True":    testconvert.Interface(true),
							"False":   testconvert.Interface(false),
							"Default": testconvert.Interface(false),
						},
						Rules: nil,
						DefaultRule: &flag.Rule{
							Percentages: &map[string]float64{"True": 100, "False": 0},
						},
					},
					"test-flag2": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"True":    testconvert.Interface(true),
							"False":   testconvert.Interface(false),
							"Default": testconvert.Interface(false),
						},
						Rules: nil,
						DefaultRule: &flag.Rule{
							Percentages: &map[string]float64{"True": 100, "False": 0},
						},
					},
				},
			},
			want: model.DiffCache{
				Added: map[string]flag.Flag{
					"test-flag2": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"True":    testconvert.Interface(true),
							"False":   testconvert.Interface(false),
							"Default": testconvert.Interface(false),
						},
						Rules: nil,
						DefaultRule: &flag.Rule{
							Percentages: &map[string]float64{"True": 100, "False": 0},
						},
					},
				},
				Deleted: map[string]flag.Flag{},
				Updated: map[string]model.DiffUpdated{},
			},
		},
		{
			name: "Updated flag",
			args: args{
				oldCache: map[string]flag.Flag{
					"test-flag": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"True":    testconvert.Interface(true),
							"False":   testconvert.Interface(false),
							"Default": testconvert.Interface(false),
						},
						Rules: nil,
						DefaultRule: &flag.Rule{
							Percentages: &map[string]float64{"True": 100, "False": 0},
						},
					},
				},
				newCache: map[string]flag.Flag{
					"test-flag": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"True":    testconvert.Interface(true),
							"False":   testconvert.Interface(false),
							"Default": testconvert.Interface(true),
						},
						Rules: nil,
						DefaultRule: &flag.Rule{
							Percentages: &map[string]float64{"True": 100, "False": 0},
						},
					},
				},
			},
			want: model.DiffCache{
				Added:   map[string]flag.Flag{},
				Deleted: map[string]flag.Flag{},
				Updated: map[string]model.DiffUpdated{
					"test-flag": {
						Before: &flag.FlagData{
							Variations: &map[string]*interface{}{
								"True":    testconvert.Interface(true),
								"False":   testconvert.Interface(false),
								"Default": testconvert.Interface(false),
							},
							Rules: nil,
							DefaultRule: &flag.Rule{
								Percentages: &map[string]float64{"True": 100, "False": 0},
							},
						},
						After: &flag.FlagData{
							Variations: &map[string]*interface{}{
								"True":    testconvert.Interface(true),
								"False":   testconvert.Interface(false),
								"Default": testconvert.Interface(true),
							},
							Rules: nil,
							DefaultRule: &flag.Rule{
								Percentages: &map[string]float64{"True": 100, "False": 0},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &notificationService{
				Notifiers: tt.fields.Notifiers,
				waitGroup: &sync.WaitGroup{},
			}
			got := c.getDifferences(tt.args.oldCache, tt.args.newCache)
			assert.Equal(t, tt.want, got)
		})
	}
}
