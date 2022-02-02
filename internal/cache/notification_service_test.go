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
							"A": testconvert.Interface(true),
							"B": testconvert.Interface(false),
						},
						DefaultRule: &flag.Rule{
							VariationResult: testconvert.String("A"),
						},
						Version: testconvert.String("1.1.0"),
					},
					"test-flag2": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"A": testconvert.Interface("true"),
							"B": testconvert.Interface("false"),
						},
						DefaultRule: &flag.Rule{
							VariationResult: testconvert.String("B"),
						},
						Version: testconvert.String("0.0.0-beta"),
					},
				},
				newCache: map[string]flag.Flag{
					"test-flag": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"A": testconvert.Interface(true),
							"B": testconvert.Interface(false),
						},
						DefaultRule: &flag.Rule{
							VariationResult: testconvert.String("A"),
						},
						Version: testconvert.String("1.1.0"),
					},
				},
			},
			want: model.DiffCache{
				Deleted: map[string]flag.Flag{
					"test-flag2": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"A": testconvert.Interface("true"),
							"B": testconvert.Interface("false"),
						},
						DefaultRule: &flag.Rule{
							VariationResult: testconvert.String("B"),
						},
						Version: testconvert.String("0.0.0-beta"),
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
							"A": testconvert.Interface("true"),
							"B": testconvert.Interface("false"),
						},
						DefaultRule: &flag.Rule{
							VariationResult: testconvert.String("B"),
						},
						Version: testconvert.String("0.0.0-beta"),
					},
				},
				newCache: map[string]flag.Flag{
					"test-flag": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"A": testconvert.Interface("true"),
							"B": testconvert.Interface("false"),
						},
						DefaultRule: &flag.Rule{
							VariationResult: testconvert.String("B"),
						},
						Version: testconvert.String("0.0.0-beta"),
					},
					"test-flag2": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"A": testconvert.Interface("true"),
							"B": testconvert.Interface("false"),
						},
						DefaultRule: &flag.Rule{
							VariationResult: testconvert.String("A"),
						},
						Version: testconvert.String("0.0.0-beta"),
					},
				},
			},
			want: model.DiffCache{
				Added: map[string]flag.Flag{
					"test-flag2": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"A": testconvert.Interface("true"),
							"B": testconvert.Interface("false"),
						},
						DefaultRule: &flag.Rule{
							VariationResult: testconvert.String("A"),
						},
						Version: testconvert.String("0.0.0-beta"),
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
							"A": testconvert.Interface("true"),
							"B": testconvert.Interface("false"),
						},
						DefaultRule: &flag.Rule{
							VariationResult: testconvert.String("B"),
						},
						Version: testconvert.String("0.0.0-beta"),
					},
				},
				newCache: map[string]flag.Flag{
					"test-flag": &flag.FlagData{
						Variations: &map[string]*interface{}{
							"A": testconvert.Interface("true"),
							"B": testconvert.Interface("false-updated"),
						},
						DefaultRule: &flag.Rule{
							VariationResult: testconvert.String("B"),
						},
						Version: testconvert.String("0.0.0-beta"),
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
								"A": testconvert.Interface("true"),
								"B": testconvert.Interface("false"),
							},
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("B"),
							},
							Version: testconvert.String("0.0.0-beta"),
						},
						After: &flag.FlagData{
							Variations: &map[string]*interface{}{
								"A": testconvert.Interface("true"),
								"B": testconvert.Interface("false-updated"),
							},
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("B"),
							},
							Version: testconvert.String("0.0.0-beta"),
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
