package notifier_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/internal/notifier"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestSlackNotifier_Notify(t *testing.T) {
	type args struct {
		diff       model.DiffCache
		statusCode int
		forceError bool
	}
	type expected struct {
		err       bool
		errLog    string
		bodyPath  string
		signature string
	}
	tests := []struct {
		name     string
		args     args
		expected expected
	}{
		{
			name: "should call slack webhook and have valid results",
			expected: expected{
				bodyPath: "../../testdata/internal/notifier/slack/should_call_webhook_and_have_valid_results.json",
			},
			args: args{
				statusCode: http.StatusOK,
				diff: model.DiffCache{
					Added: map[string]flag.Flag{
						"test-flag3": &flag.FlagData{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface("default"),
								"False":   testconvert.Interface("false"),
								"True":    testconvert.Interface("test"),
							},
							Rules: &map[string]flag.Rule{
								"defaultRule": {
									Percentages: &map[string]float64{
										"True":  5,
										"False": 95,
									},
								},
							},
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("Default"),
							},
						},
					},
					Deleted: map[string]flag.Flag{
						"test-flag": &flag.FlagData{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							Rules: &map[string]flag.Rule{
								"defaultRule": {
									Query: testconvert.String("key eq \"random-key\""),
									Percentages: &map[string]float64{
										"True":  100,
										"False": 0,
									},
								},
							},
							DefaultRule: &flag.Rule{
								VariationResult: testconvert.String("Default"),
							},
						},
					},
					Updated: map[string]model.DiffUpdated{
						"test-flag2": {
							Before: &flag.FlagData{
								Variations: &map[string]*interface{}{
									"Default": testconvert.Interface(false),
									"False":   testconvert.Interface(false),
									"True":    testconvert.Interface(true),
								},
								Rules: &map[string]flag.Rule{
									"defaultRule": {
										Query: testconvert.String("key eq \"not-a-key\""),
										Percentages: &map[string]float64{
											"True":  100,
											"False": 0,
										},
									},
								},
								DefaultRule: &flag.Rule{
									VariationResult: testconvert.String("Default"),
								},
							},
							After: &flag.FlagData{
								Variations: &map[string]*interface{}{
									"Default": testconvert.Interface(false),
									"False":   testconvert.Interface(false),
									"True":    testconvert.Interface(true),
								},
								Rules: &map[string]flag.Rule{
									"defaultRule": {
										Query: testconvert.String("key eq \"not-a-key\""),
										Percentages: &map[string]float64{
											"True":  100,
											"False": 0,
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
				},
			},
		},
		{
			name: "should log if http code is superior to 399",
			expected: expected{
				err:    true,
				errLog: "^\\[" + testutils.RFC3339Regex + "\\] error: \\(SlackNotifier\\) while calling slack webhook, statusCode = 400",
			},
			args: args{
				statusCode: http.StatusBadRequest,
				diff:       model.DiffCache{},
			},
		},
		{
			name: "should log if error while calling webhook",
			expected: expected{
				err:    true,
				errLog: "^\\[" + testutils.RFC3339Regex + "\\] error: \\(SlackNotifier\\) error: while calling webhook: random error",
			},
			args: args{
				statusCode: http.StatusOK,
				diff:       model.DiffCache{},
				forceError: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logFile, _ := ioutil.TempFile("", "")
			defer logFile.Close()
			defer os.Remove(logFile.Name())

			mockHTTPClient := &testutils.HTTPClientMock{StatusCode: tt.args.statusCode, ForceError: tt.args.forceError}

			c := notifier.NewSlackNotifier(
				log.New(logFile, "", 0),
				mockHTTPClient,
				"https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
			)

			w := sync.WaitGroup{}
			w.Add(1)
			c.Notify(tt.args.diff, &w)

			if tt.expected.err {
				log, _ := ioutil.ReadFile(logFile.Name())
				assert.Regexp(t, tt.expected.errLog, string(log))
			} else {
				hostname, _ := os.Hostname()
				content, _ := ioutil.ReadFile(tt.expected.bodyPath)
				expectedContent := strings.ReplaceAll(string(content), "{{hostname}}", hostname)
				fmt.Println(mockHTTPClient.Body)
				assert.JSONEq(t, expectedContent, mockHTTPClient.Body)
				assert.Equal(t, tt.expected.signature, mockHTTPClient.Signature)
			}
		})
	}
}
