package slacknotifier

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/testutils"
)

func TestSlackNotifier_Notify(t *testing.T) {
	type args struct {
		diff       notifier.DiffCache
		statusCode int
		forceError bool
		url        string
	}
	type expected struct {
		err       bool
		errMsg    string
		bodyPath  string
		signature string
	}
	tests := []struct {
		name     string
		args     args
		expected expected
	}{
		{
			name: "should call webhook and have valid results",
			expected: expected{
				bodyPath: "./testdata/should_call_webhook_and_have_valid_results.json",
			},
			args: args{
				url:        "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				statusCode: http.StatusOK,
				diff: notifier.DiffCache{
					Added: map[string]flag.Flag{
						"test-flag3": &flag.InternalFlag{
							Rules: &[]flag.Rule{
								{
									Name:  testconvert.String("rule1"),
									Query: testconvert.String("key eq \"random-key\""),
									Percentages: &map[string]float64{
										"False": 95,
										"True":  5,
									},
								},
							},
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface("default"),
								"False":   testconvert.Interface("false"),
								"True":    testconvert.Interface("test"),
							},
							DefaultRule: &flag.Rule{
								Name:            testconvert.String("defaultRule"),
								VariationResult: testconvert.String("Default"),
							},
							TrackEvents: testconvert.Bool(true),
							Disable:     testconvert.Bool(false),
							Version:     testconvert.String("1.1"),
						},
					},
					Deleted: map[string]flag.Flag{
						"test-flag": &flag.InternalFlag{
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
							Variations: &map[string]*interface{}{
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
					Updated: map[string]notifier.DiffUpdated{
						"test-flag2": {
							Before: &flag.InternalFlag{
								Variations: &map[string]*interface{}{
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
								Experimentation: &flag.ExperimentationRollout{
									Start: testconvert.Time(time.Unix(1095379400, 0)),
									End:   testconvert.Time(time.Unix(1095371000, 0)),
								},
							},
							After: &flag.InternalFlag{
								Variations: &map[string]*interface{}{
									"Default": testconvert.Interface("strDefault"),
									"False":   testconvert.Interface("strFalse"),
									"True":    testconvert.Interface("strTrue"),
								},
								Rules: &[]flag.Rule{
									{
										Name:  testconvert.String("rule1"),
										Query: testconvert.String("key eq \"not-a-ke\""),
										Percentages: &map[string]float64{
											"False": 20,
											"True":  80,
										},
									},
								},
								DefaultRule: &flag.Rule{
									Name:            testconvert.String("defaultRule"),
									VariationResult: testconvert.String("Default"),
								},
								Disable:     testconvert.Bool(true),
								TrackEvents: testconvert.Bool(false),
								Version:     testconvert.String("1.1"),
							},
						},
					},
				},
			},
		},
		{
			name: "should err if http code is superior to 399",
			expected: expected{
				err:    true,
				errMsg: "error: (Slack Notifier) while calling slack webhook, statusCode = 400",
			},
			args: args{
				url:        "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				statusCode: http.StatusBadRequest,
				diff:       notifier.DiffCache{},
			},
		},
		{
			name: "should err if error while calling webhook",
			expected: expected{
				err:    true,
				errMsg: "error: (Slack Notifier) error: while calling webhook: random error",
			},
			args: args{
				url:        "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				statusCode: http.StatusOK,
				diff:       notifier.DiffCache{},
				forceError: true,
			},
		},
		{
			name: "missing slack url",
			expected: expected{
				err:    true,
				errMsg: "error: (Slack Notifier) invalid notifier configuration, no SlackWebhookURL provided for the slack notifier",
			},
			args: args{
				statusCode: http.StatusOK,
				diff:       notifier.DiffCache{},
			},
		},
		{
			name: "invalid slack url",
			expected: expected{
				err:    true,
				errMsg: "error: (Slack Notifier) invalid SlackWebhookURL: https://{}hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
			},
			args: args{
				url:        "https://{}hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				statusCode: http.StatusOK,
				diff:       notifier.DiffCache{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPClient := &testutils.HTTPClientMock{
				StatusCode: tt.args.statusCode,
				ForceError: tt.args.forceError,
			}

			c := Notifier{
				SlackWebhookURL: tt.args.url,
				httpClient:      mockHTTPClient,
			}

			err := c.Notify(tt.args.diff)

			if tt.expected.err {
				assert.ErrorContains(t, err, tt.expected.errMsg)
			} else {
				assert.NoError(t, err)
				hostname, _ := os.Hostname()
				content, _ := os.ReadFile(tt.expected.bodyPath)
				expectedContent := strings.ReplaceAll(string(content), "{{hostname}}", hostname)
				assert.JSONEq(t, expectedContent, mockHTTPClient.Body)
				assert.Equal(t, tt.expected.signature, mockHTTPClient.Signature)
			}
		})
	}
}
