package notifier

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/testutils"
)

func TestLogNotifier_Notify(t *testing.T) {
	type args struct {
		diff model.DiffCache
		wg   *sync.WaitGroup
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "Flag deleted",
			args: args{
				diff: model.DiffCache{
					Deleted: map[string]flag.Flag{
						"test-flag": &flag.FlagData{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface("default"),
								"False":   testconvert.Interface("false"),
								"True":    testconvert.Interface("true"),
							},
							Rules: &map[string]flag.Rule{
								"defaultRule": {
									Percentages: &map[string]float64{
										"True":  40,
										"False": 60,
									},
								},
							},
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("Default"),
							},
						},
					},
					Updated: map[string]model.DiffUpdated{},
					Added:   map[string]flag.Flag{},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag test-flag removed",
		},
		{
			name: "Update flag",
			args: args{
				diff: model.DiffCache{
					Deleted: map[string]flag.Flag{},
					Updated: map[string]model.DiffUpdated{
						"test-flag": {
							Before: &flag.FlagData{
								Variations: &map[string]*interface{}{
									"Default": testconvert.Interface("default"),
									"False":   testconvert.Interface("false"),
									"True":    testconvert.Interface("true"),
								},
								Rules: &map[string]flag.Rule{
									"defaultRule": {
										Percentages: &map[string]float64{
											"True":  40,
											"False": 60,
										},
									},
								},
								DefaultRule: &flag.Rule{
									VariationResult: testconvert.String("Default"),
								},
							},
							After: &flag.FlagData{
								Variations: &map[string]*interface{}{
									"Default": testconvert.Interface("default"),
									"False":   testconvert.Interface("false"),
									"True":    testconvert.Interface("true"),
								},
								Rules: &map[string]flag.Rule{
									"defaultRule": {
										Percentages: &map[string]float64{
											"True":  10,
											"False": 90,
										},
									},
								},
								DefaultRule: &flag.Rule{
									VariationResult: testconvert.String("Default"),
								},
							},
						},
					},
					Added: map[string]flag.Flag{},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag test-flag updated, old=\\[Variations:\\[Default=default,False=false,True=true\\], Rules:\\[\\[percentages:\\[False=60.00,True=40.00\\]\\]\\], DefaultRule:\\[variation:\\[Default\\]\\]\\], new=\\[Variations:\\[Default=default,False=false,True=true\\], Rules:\\[\\[percentages:\\[False=90.00,True=10.00\\]\\]\\], DefaultRule:\\[variation:\\[Default\\]\\]\\]",
		},
		{
			name: "Disable flag",
			args: args{
				diff: model.DiffCache{
					Deleted: map[string]flag.Flag{},
					Updated: map[string]model.DiffUpdated{
						"test-flag": {
							Before: &flag.FlagData{
								Variations: &map[string]*interface{}{
									"Default": testconvert.Interface("default"),
									"False":   testconvert.Interface("false"),
									"True":    testconvert.Interface("true"),
								},
								Rules: &map[string]flag.Rule{
									"defaultRule": {
										Percentages: &map[string]float64{
											"True":  10,
											"False": 90,
										},
									},
								},
								DefaultRule: &flag.Rule{
									VariationResult: testconvert.String("Default"),
								},
							},
							After: &flag.FlagData{
								Variations: &map[string]*interface{}{
									"Default": testconvert.Interface("default"),
									"False":   testconvert.Interface("false"),
									"True":    testconvert.Interface("true"),
								},
								Rules: &map[string]flag.Rule{
									"defaultRule": {
										Percentages: &map[string]float64{
											"True":  10,
											"False": 90,
										},
									},
								},
								DefaultRule: &flag.Rule{
									VariationResult: testconvert.String("Default"),
								},
								Disable: testconvert.Bool(true),
							},
						},
					},
					Added: map[string]flag.Flag{},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag test-flag is turned OFF",
		},
		{
			name: "Add flag",
			args: args{
				diff: model.DiffCache{
					Deleted: map[string]flag.Flag{},
					Updated: map[string]model.DiffUpdated{},
					Added: map[string]flag.Flag{
						"add-test-flag": &flag.FlagData{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface("default"),
								"False":   testconvert.Interface("false"),
								"True":    testconvert.Interface("true"),
							},
							Rules: &map[string]flag.Rule{
								"defaultRule": {
									Percentages: &map[string]float64{
										"True":  10,
										"False": 90,
									},
								},
							},
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("Default"),
							},
						},
					},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag add-test-flag added",
		},
		{
			name: "Enable flag",
			args: args{
				diff: model.DiffCache{
					Deleted: map[string]flag.Flag{},
					Updated: map[string]model.DiffUpdated{
						"test-flag": {
							Before: &flag.FlagData{
								Variations: &map[string]*interface{}{
									"Default": testconvert.Interface("default"),
									"False":   testconvert.Interface("false"),
									"True":    testconvert.Interface("true"),
								},
								Rules: &map[string]flag.Rule{
									"defaultRule": {
										Percentages: &map[string]float64{
											"True":  10,
											"False": 90,
										},
									},
								},
								DefaultRule: &flag.Rule{
									VariationResult: testconvert.String("Default"),
								},
								Disable: testconvert.Bool(true),
							},
							After: &flag.FlagData{
								Variations: &map[string]*interface{}{
									"Default": testconvert.Interface("default"),
									"False":   testconvert.Interface("false"),
									"True":    testconvert.Interface("true"),
								},
								Rules: &map[string]flag.Rule{
									"defaultRule": {
										Percentages: &map[string]float64{
											"True":  10,
											"False": 90,
										},
									},
								},
								DefaultRule: &flag.Rule{
									VariationResult: testconvert.String("Default"),
								},
							},
						},
					},
					Added: map[string]flag.Flag{},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag test-flag is turned ON \\(flag=\\[Variations:\\[Default=default,False=false,True=true\\], Rules:\\[\\[percentages:\\[False=90.00,True=10.00\\]\\]\\], DefaultRule:\\[variation:\\[Default\\]\\]\\]\\)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logOutput, _ := ioutil.TempFile("", "")
			defer os.Remove(logOutput.Name())

			c := &LogNotifier{
				Logger: fflog.Logger{Logger: log.New(logOutput, "", 0)},
			}
			tt.args.wg.Add(1)
			c.Notify(tt.args.diff, tt.args.wg)
			log, _ := ioutil.ReadFile(logOutput.Name())
			assert.Regexp(t, tt.expected, string(log))
		})
	}
}
