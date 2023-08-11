package cache

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"sync"
	"testing"
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
		want   notifier.DiffCache
	}{
		{
			name: "Delete flag",
			args: args{
				oldCache: map[string]flag.Flag{
					"test-flag": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(false),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("legacyRuleV0"),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					"test-flag2": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(false),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("legacyRuleV0"),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
				},
				newCache: map[string]flag.Flag{
					"test-flag": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(false),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("legacyRuleV0"),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
				},
			},
			want: notifier.DiffCache{
				Deleted: map[string]flag.Flag{
					"test-flag2": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(false),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("legacyRuleV0"),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
				},
				Added:   map[string]flag.Flag{},
				Updated: map[string]notifier.DiffUpdated{},
			},
		},
		{
			name: "Added flag",
			args: args{
				oldCache: map[string]flag.Flag{
					"test-flag": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(false),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("legacyRuleV0"),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
				},
				newCache: map[string]flag.Flag{
					"test-flag": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(false),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("legacyRuleV0"),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					"test-flag2": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(false),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("legacyRuleV0"),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
				},
			},
			want: notifier.DiffCache{
				Added: map[string]flag.Flag{
					"test-flag2": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(false),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("legacyRuleV0"),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
				},
				Deleted: map[string]flag.Flag{},
				Updated: map[string]notifier.DiffUpdated{},
			},
		},
		{
			name: "Updated flag",
			args: args{
				oldCache: map[string]flag.Flag{
					"test-flag": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(false),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("legacyRuleV0"),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
				},
				newCache: map[string]flag.Flag{
					"test-flag": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(true),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("legacyRuleV0"),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
				},
			},
			want: notifier.DiffCache{
				Added:   map[string]flag.Flag{},
				Deleted: map[string]flag.Flag{},
				Updated: map[string]notifier.DiffUpdated{
					"test-flag": {
						Before: &flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name: testconvert.String("legacyRuleV0"),
								Percentages: &map[string]float64{
									"False": 0,
									"True":  100,
								},
							},
						},
						After: &flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(true),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name: testconvert.String("legacyRuleV0"),
								Percentages: &map[string]float64{
									"False": 0,
									"True":  100,
								},
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
