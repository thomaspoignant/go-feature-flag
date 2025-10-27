package webhooknotifier

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/testutils"
)

func Test_webhookNotifier_Notify(t *testing.T) {
	type fields struct {
		Secret string
	}
	type expected struct {
		err       bool
		errorMsg  string
		bodyPath  string
		signature string
		headers   map[string][]string
	}
	type args struct {
		diff       notifier.DiffCache
		statusCode int
		forceError bool
		url        string
		headers    map[string][]string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		expected expected
	}{
		{
			name: "should call webhook and have valid results",
			fields: fields{
				Secret: "test-secret",
			},
			expected: expected{
				bodyPath:  "./testdata/should_call_webhook_and_have_valid_results.json",
				signature: "sha256=813bb118d9ac870a1264c2e5ce2a9a95c46246c52312ff77201d6d6b826f4ed6",
			},
			args: args{
				url:        "http://webhook.example/hook",
				statusCode: http.StatusOK,
				diff: notifier.DiffCache{
					Added: map[string]flag.Flag{
						"test-flag3": &flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface("default"),
								"False":   testconvert.Interface("false"),
								"True":    testconvert.Interface("test"),
							},
							DefaultRule: &flag.Rule{
								Name: testconvert.String("defaultRule"),
								Percentages: &map[string]float64{
									"False": 95,
									"True":  5,
								},
							},
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
								Rules: &[]flag.Rule{
									{
										Name:  testconvert.String("rule1"),
										Query: testconvert.String("key eq \"not-a-key\""),
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
							After: &flag.InternalFlag{
								Rules: &[]flag.Rule{
									{
										Name:  testconvert.String("rule1"),
										Query: testconvert.String("key eq \"not-a-key\""),
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
								Disable: testconvert.Bool(true),
							},
						},
					},
				},
			},
		},
		{
			name: "should not be signed if no secret",
			expected: expected{
				bodyPath:  "./testdata/should_not_be_signed_if_no_secret.json",
				signature: "",
			},
			args: args{
				url:        "http://webhook.example/hook",
				statusCode: http.StatusOK,
				diff: notifier.DiffCache{
					Added: map[string]flag.Flag{
						"test-flag3": &flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface("default"),
								"False":   testconvert.Interface("false"),
								"True":    testconvert.Interface("test"),
							},
							DefaultRule: &flag.Rule{
								Name: testconvert.String("defaultRule"),
								Percentages: &map[string]float64{
									"False": 95,
									"True":  5,
								},
							},
						},
					},
					Deleted: map[string]flag.Flag{},
					Updated: map[string]notifier.DiffUpdated{},
				},
			},
		},
		{
			name: "should log if http code is superior to 399",
			expected: expected{
				err:      true,
				errorMsg: "error: while calling webhook, statusCode = 400",
			},
			args: args{
				url:        "http://webhook.example/hook",
				statusCode: http.StatusBadRequest,
				diff:       notifier.DiffCache{},
			},
		},
		{
			name: "should log if error while calling webhook",
			expected: expected{
				err:      true,
				errorMsg: "error: while calling webhook: random error",
			},
			args: args{
				url:        "http://webhook.example/hook",
				statusCode: http.StatusOK,
				diff:       notifier.DiffCache{},
				forceError: true,
			},
		},
		{
			name: "no endpointURL",
			expected: expected{
				err:      true,
				errorMsg: "invalid notifier configuration, no endpointURL provided for the webhook notifier",
			},
			args: args{
				url:        "",
				statusCode: http.StatusOK,
				diff:       notifier.DiffCache{},
				forceError: true,
			},
		},
		{
			name: "should use custom Headers",
			expected: expected{
				bodyPath:  "./testdata/should_not_be_signed_if_no_secret.json",
				signature: "",
				headers: map[string][]string{
					"Authorization": {"Bearer auth_token"},
					"Content-Type":  {"application/json"},
				},
			},
			args: args{
				url:        "http://webhook.example/hook",
				statusCode: http.StatusOK,
				headers: map[string][]string{
					"Authorization": {"Bearer auth_token"},
				},
				diff: notifier.DiffCache{
					Added: map[string]flag.Flag{
						"test-flag3": &flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface("default"),
								"False":   testconvert.Interface("false"),
								"True":    testconvert.Interface("test"),
							},
							DefaultRule: &flag.Rule{
								Name: testconvert.String("defaultRule"),
								Percentages: &map[string]float64{
									"False": 95,
									"True":  5,
								},
							},
						},
					},
					Deleted: map[string]flag.Flag{},
					Updated: map[string]notifier.DiffUpdated{},
				},
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
				EndpointURL: tt.args.url,
				Secret:      tt.fields.Secret,
				Meta:        map[string]string{"hostname": "toto"},
				httpClient:  mockHTTPClient,
				init:        sync.Once{},
				Headers:     tt.args.headers,
			}

			err := c.Notify(tt.args.diff)

			if tt.expected.err {
				assert.ErrorContains(t, err, tt.expected.errorMsg)
			} else {
				assert.NoError(t, err)
				content, _ := os.ReadFile(tt.expected.bodyPath)
				assert.JSONEq(t, string(content), mockHTTPClient.Body)
				assert.Equal(t, tt.expected.signature, mockHTTPClient.Signature)
				if tt.expected.headers != nil {
					assert.Equal(t, tt.expected.headers, mockHTTPClient.Headers)
				}
			}
		})
	}
}

func Test_webhookNotifier_no_meta_data(t *testing.T) {
	mockHTTPClient := &testutils.HTTPClientMock{StatusCode: 200, ForceError: false}
	diff := notifier.DiffCache{
		Added: map[string]flag.Flag{
			"test-flag3": &flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"Default": testconvert.Interface("default"),
					"False":   testconvert.Interface("false"),
					"True":    testconvert.Interface("test"),
				},
				DefaultRule: &flag.Rule{
					Name: testconvert.String("defaultRule"),
					Percentages: &map[string]float64{
						"False": 95,
						"True":  5,
					},
				},
			},
		},
		Deleted: map[string]flag.Flag{},
		Updated: map[string]notifier.DiffUpdated{},
	}

	// no meta
	c := Notifier{
		EndpointURL: "http://webhook.example/hook",
		httpClient:  mockHTTPClient,
		init:        sync.Once{},
	}

	err := c.Notify(diff)

	assert.NoError(t, err)
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(mockHTTPClient.Body), &m)

	assert.NotEmpty(t, m["meta"])
}
