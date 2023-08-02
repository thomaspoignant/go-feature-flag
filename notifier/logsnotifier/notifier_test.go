package logsnotifier

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"

	"github.com/thomaspoignant/go-feature-flag/testutils"
)

func TestLogNotifier_Notify(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name     string
		args     args
		diff     notifier.DiffCache
		expected string
	}{
		{
			name: "Flag deleted",
			diff: notifier.DiffCache{
				Deleted: map[string]flag.Flag{
					"test-flag": &flag.InternalFlag{
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(false),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name: testconvert.String("legacyDefaultRule"),
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
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag test-flag removed",
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
									Name:  testconvert.String("legacyRuleV0"),
									Query: testconvert.String("key eq \"random-key\""),
									Percentages: &map[string]float64{
										"False": 0,
										"True":  100,
									},
								},
							},
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name:            testconvert.String("legacyDefaultRule"),
								VariationResult: testconvert.String("Default"),
							},
						},
						After: &flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name: testconvert.String("legacyDefaultRule"),
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
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag test-flag updated",
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
									Name:  testconvert.String("legacyRuleV0"),
									Query: testconvert.String("key eq \"random-key\""),
									Percentages: &map[string]float64{
										"False": 0,
										"True":  100,
									},
								},
							},
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name:            testconvert.String("legacyDefaultRule"),
								VariationResult: testconvert.String("Default"),
							},
						},
						After: &flag.InternalFlag{
							Rules: &[]flag.Rule{
								{
									Name:  testconvert.String("legacyRuleV0"),
									Query: testconvert.String("key eq \"random-key\""),
									Percentages: &map[string]float64{
										"False": 0,
										"True":  100,
									},
								},
							},
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name:            testconvert.String("legacyDefaultRule"),
								VariationResult: testconvert.String("Default"),
							},
							Disable: testconvert.Bool(true),
						},
					},
				},
				Added: map[string]flag.Flag{},
			},
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag test-flag is turned OFF",
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
								Name:  testconvert.String("legacyRuleV0"),
								Query: testconvert.String("key eq \"random-key\""),
								Percentages: &map[string]float64{
									"False": 0,
									"True":  100,
								},
							},
						},
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(false),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name:            testconvert.String("legacyDefaultRule"),
							VariationResult: testconvert.String("Default"),
						},
					},
				},
			},
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag add-test-flag added",
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
									Name:  testconvert.String("legacyRuleV0"),
									Query: testconvert.String("key eq \"random-key\""),
									Percentages: &map[string]float64{
										"False": 0,
										"True":  100,
									},
								},
							},
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name:            testconvert.String("legacyDefaultRule"),
								VariationResult: testconvert.String("Default"),
							},
						},
						Before: &flag.InternalFlag{
							Rules: &[]flag.Rule{
								{
									Name:  testconvert.String("legacyRuleV0"),
									Query: testconvert.String("key eq \"random-key\""),
									Percentages: &map[string]float64{
										"False": 0,
										"True":  100,
									},
								},
							},
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name:            testconvert.String("legacyDefaultRule"),
								VariationResult: testconvert.String("Default"),
							},
							Disable: testconvert.Bool(true),
						},
					},
				},
				Added: map[string]flag.Flag{},
			},
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag test-flag is turned ON",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logOutput, _ := os.CreateTemp("", "")
			defer os.Remove(logOutput.Name())

			c := &Notifier{
				Logger: log.New(logOutput, "", 0),
			}
			_ = c.Notify(tt.diff)
			log, _ := os.ReadFile(logOutput.Name())
			assert.Regexp(t, tt.expected, string(log))
		})
	}
}
