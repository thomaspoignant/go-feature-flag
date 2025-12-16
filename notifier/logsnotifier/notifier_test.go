package logsnotifier

import (
	"log/slog"
	"testing"

	"github.com/thejerf/slogassert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

func TestLogNotifier_Notify(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name        string
		args        args
		diff        notifier.DiffCache
		expectedLog *slogassert.LogMessageMatch
	}{
		{
			name: "Flag deleted",
			diff: notifier.DiffCache{
				Deleted: map[string]flag.Flag{
					"test-flag": &flag.InternalFlag{
						Variations: &map[string]*any{
							"Default": testconvert.Interface(false),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
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
				Updated: map[string]notifier.DiffUpdated{},
				Added:   map[string]flag.Flag{},
			},
			expectedLog: &slogassert.LogMessageMatch{
				Message: "flag removed",
				Level:   slog.LevelInfo,
				Attrs: map[string]any{
					"key": "test-flag",
				},
				AllAttrsMatch: true,
			},
		},
		{
			name: "Update flag",
			diff: notifier.DiffCache{
				Deleted: map[string]flag.Flag{},
				Updated: map[string]notifier.DiffUpdated{
					"test-flag": {
						Before: &flag.InternalFlag{
							Rules: &[]flag.Rule{
								{
									Name:  testconvert.String("rule1"),
									Query: testconvert.String("key eq \"random-key\""),
									Percentages: &map[string]float64{
										"False": 0,
										"True":  100,
									},
								},
							},
							Variations: &map[string]*any{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name:            testconvert.String("defaultRule"),
								VariationResult: testconvert.String("Default"),
							},
						},
						After: &flag.InternalFlag{
							Variations: &map[string]*any{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
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
				Added: map[string]flag.Flag{},
			},
			expectedLog: &slogassert.LogMessageMatch{
				Message: "flag updated",
				Level:   slog.LevelInfo,
				Attrs: map[string]any{
					"key": "test-flag",
				},
				AllAttrsMatch: true,
			},
		},
		{
			name: "Disable flag",
			diff: notifier.DiffCache{
				Deleted: map[string]flag.Flag{},
				Updated: map[string]notifier.DiffUpdated{
					"test-flag": {
						Before: &flag.InternalFlag{
							Rules: &[]flag.Rule{
								{
									Name:  testconvert.String("rule1"),
									Query: testconvert.String("key eq \"random-key\""),
									Percentages: &map[string]float64{
										"False": 0,
										"True":  100,
									},
								},
							},
							Variations: &map[string]*any{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name:            testconvert.String("defaultRule"),
								VariationResult: testconvert.String("Default"),
							},
						},
						After: &flag.InternalFlag{
							Rules: &[]flag.Rule{
								{
									Name:  testconvert.String("rule1"),
									Query: testconvert.String("key eq \"random-key\""),
									Percentages: &map[string]float64{
										"False": 0,
										"True":  100,
									},
								},
							},
							Variations: &map[string]*any{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name:            testconvert.String("defaultRule"),
								VariationResult: testconvert.String("Default"),
							},
							Disable: testconvert.Bool(true),
						},
					},
				},
				Added: map[string]flag.Flag{},
			},
			expectedLog: &slogassert.LogMessageMatch{
				Message: "flag is turned OFF",
				Level:   slog.LevelInfo,
				Attrs: map[string]any{
					"key": "test-flag",
				},
				AllAttrsMatch: true,
			},
		},
		{
			name: "Add flag",
			diff: notifier.DiffCache{
				Deleted: map[string]flag.Flag{},
				Updated: map[string]notifier.DiffUpdated{},
				Added: map[string]flag.Flag{
					"add-test-flag": &flag.InternalFlag{
						Rules: &[]flag.Rule{
							{
								Name:  testconvert.String("rule1"),
								Query: testconvert.String("key eq \"random-key\""),
								Percentages: &map[string]float64{
									"False": 0,
									"True":  100,
								},
							},
						},
						Variations: &map[string]*any{
							"Default": testconvert.Interface(false),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name:            testconvert.String("defaultRule"),
							VariationResult: testconvert.String("Default"),
						},
					},
				},
			},
			expectedLog: &slogassert.LogMessageMatch{
				Message: "flag added",
				Level:   slog.LevelInfo,
				Attrs: map[string]any{
					"key": "add-test-flag",
				},
				AllAttrsMatch: true,
			},
		},
		{
			name: "Enable flag",
			diff: notifier.DiffCache{
				Deleted: map[string]flag.Flag{},
				Updated: map[string]notifier.DiffUpdated{
					"test-flag": {
						After: &flag.InternalFlag{
							Rules: &[]flag.Rule{
								{
									Name:  testconvert.String("rule1"),
									Query: testconvert.String("key eq \"random-key\""),
									Percentages: &map[string]float64{
										"False": 0,
										"True":  100,
									},
								},
							},
							Variations: &map[string]*any{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name:            testconvert.String("defaultRule"),
								VariationResult: testconvert.String("Default"),
							},
						},
						Before: &flag.InternalFlag{
							Rules: &[]flag.Rule{
								{
									Name:  testconvert.String("rule1"),
									Query: testconvert.String("key eq \"random-key\""),
									Percentages: &map[string]float64{
										"False": 0,
										"True":  100,
									},
								},
							},
							Variations: &map[string]*any{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name:            testconvert.String("defaultRule"),
								VariationResult: testconvert.String("Default"),
							},
							Disable: testconvert.Bool(true),
						},
					},
				},
				Added: map[string]flag.Flag{},
			},
			expectedLog: &slogassert.LogMessageMatch{
				Message: "flag is turned ON",
				Level:   slog.LevelInfo,
				Attrs: map[string]any{
					"key": "test-flag",
				},
				AllAttrsMatch: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := slogassert.New(t, slog.LevelDebug, nil)
			logger := slog.New(handler)
			c := &Notifier{
				Logger: &fflog.FFLogger{LeveledLogger: logger},
			}
			_ = c.Notify(tt.diff)
			handler.AssertPrecise(*tt.expectedLog)
		})
	}
}
