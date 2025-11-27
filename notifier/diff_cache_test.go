package notifier_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

func TestDiffCache_HasDiff(t *testing.T) {
	type fields struct {
		Deleted map[string]flag.Flag
		Added   map[string]flag.Flag
		Updated map[string]notifier.DiffUpdated
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "null fields",
			fields: fields{},
			want:   false,
		},
		{
			name: "empty fields",
			fields: fields{
				Deleted: map[string]flag.Flag{},
				Added:   map[string]flag.Flag{},
				Updated: map[string]notifier.DiffUpdated{},
			},
			want: false,
		},
		{
			name: "only Deleted",
			fields: fields{
				Deleted: map[string]flag.Flag{
					"flag": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(true),
							"True":    testconvert.Interface(true),
							"False":   testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("defaultRule"),
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
			want: true,
		},
		{
			name: "only Added",
			fields: fields{
				Added: map[string]flag.Flag{
					"flag": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(true),
							"True":    testconvert.Interface(true),
							"False":   testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("defaultRule"),
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
			want: true,
		},
		{
			name: "only Updated",
			fields: fields{
				Added:   map[string]flag.Flag{},
				Deleted: map[string]flag.Flag{},
				Updated: map[string]notifier.DiffUpdated{
					"flag": {
						Before: &flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(true),
								"True":    testconvert.Interface(true),
								"False":   testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name: testconvert.String("defaultRule"),
								Percentages: &map[string]float64{
									"False": 0,
									"True":  100,
								},
							},
						},
						After: &flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(false),
								"True":    testconvert.Interface(true),
								"False":   testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name: testconvert.String("defaultRule"),
								Percentages: &map[string]float64{
									"False": 0,
									"True":  100,
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "all fields",
			fields: fields{
				Added: map[string]flag.Flag{
					"flag": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(true),
							"True":    testconvert.Interface(true),
							"False":   testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("defaultRule"),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
				},
				Deleted: map[string]flag.Flag{
					"flag": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(true),
							"True":    testconvert.Interface(true),
							"False":   testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("defaultRule"),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
				},
				Updated: map[string]notifier.DiffUpdated{
					"flag": {
						Before: &flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(true),
								"True":    testconvert.Interface(true),
								"False":   testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name: testconvert.String("defaultRule"),
								Percentages: &map[string]float64{
									"False": 0,
									"True":  100,
								},
							},
						},
						After: &flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(false),
								"True":    testconvert.Interface(true),
								"False":   testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name: testconvert.String("defaultRule"),
								Percentages: &map[string]float64{
									"False": 0,
									"True":  100,
								},
							},
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := notifier.DiffCache{
				Deleted: tt.fields.Deleted,
				Added:   tt.fields.Added,
				Updated: tt.fields.Updated,
			}
			assert.Equal(t, tt.want, d.HasDiff())
		})
	}
}
