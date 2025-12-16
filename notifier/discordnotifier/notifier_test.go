package discordnotifier

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/testutils"
)

func TestDiscordNotifier_Notify(t *testing.T) {
	genericFlag := &flag.InternalFlag{
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
		Variations: &map[string]*any{
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
	}

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
				url:        "https://discord.com/api/webhooks/000000000000000000/XXXXXXXXXXXXXXXXXXXXXXXX",
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
							Variations: &map[string]*any{
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
					Updated: map[string]notifier.DiffUpdated{
						"test-flag2": {
							Before: &flag.InternalFlag{
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
								Experimentation: &flag.ExperimentationRollout{
									Start: testconvert.Time(time.Unix(1095379400, 0)),
									End:   testconvert.Time(time.Unix(1095371000, 0)),
								},
							},
							After: &flag.InternalFlag{
								Variations: &map[string]*any{
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
			name: "should not display all the flags if reached the limit of embeds",
			expected: expected{
				bodyPath: "./testdata/should_call_webhook_and_have_valid_results_limit_max.json",
			},
			args: args{
				url:        "https://discord.com/api/webhooks/000000000000000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				statusCode: http.StatusOK,
				diff: notifier.DiffCache{
					Added: map[string]flag.Flag{
						"test-flag3": genericFlag,
						"test-flag4": genericFlag,
						"test-flag5": genericFlag,
						"test-flag6": genericFlag,
					},
					Deleted: map[string]flag.Flag{
						"test-flag":   genericFlag,
						"test-flag1":  genericFlag,
						"test-flag2":  genericFlag,
						"test-flag7":  genericFlag,
						"test-flag8":  genericFlag,
						"test-flag9":  genericFlag,
						"test-flag10": genericFlag,
					},
				},
			},
		},
		{
			name: "should err if http code is superior to 399",
			expected: expected{
				err:    true,
				errMsg: "error: (Discord Notifier) webhook call failed with statusCode = 400",
			},
			args: args{
				url:        "https://discord.com/api/webhooks/000000000000000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				statusCode: http.StatusBadRequest,
				diff:       notifier.DiffCache{},
			},
		},
		{
			name: "should err if error while calling webhook",
			expected: expected{
				err:    true,
				errMsg: "error: (Discord Notifier) error calling webhook: random error",
			},
			args: args{
				url:        "https://discord.com/api/webhooks/000000000000000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				statusCode: http.StatusOK,
				diff:       notifier.DiffCache{},
				forceError: true,
			},
		},
		{
			name: "missing discord url",
			expected: expected{
				err:    true,
				errMsg: "error: (Discord Notifier) invalid notifier configuration, no DiscordWebhookURL provided",
			},
			args: args{
				statusCode: http.StatusOK,
				diff:       notifier.DiffCache{},
			},
		},
		{
			name: "invalid discord url",
			expected: expected{
				err:    true,
				errMsg: "error: (Discord Notifier) invalid DiscordWebhookURL: https://{}discord.com/api/webhooks/000000000000000000/XXXXXXXXXXXXXXXXXXXXXXXX",
			},
			args: args{
				url:        "https://{}discord.com/api/webhooks/000000000000000000/XXXXXXXXXXXXXXXXXXXXXXXX",
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
				DiscordWebhookURL: tt.args.url,
				httpClient:        mockHTTPClient,
			}

			err := c.Notify(tt.args.diff)

			if tt.expected.err {
				assert.ErrorContains(t, err, tt.expected.errMsg)
			} else {
				assert.NoError(t, err)
				hostname, _ := os.Hostname()
				content, err := os.ReadFile(tt.expected.bodyPath)
				require.NoError(t, err)
				expectedContent := strings.ReplaceAll(string(content), "{{hostname}}", hostname)
				assert.JSONEq(t, expectedContent, mockHTTPClient.Body)
				assert.Equal(t, tt.expected.signature, mockHTTPClient.Signature)
			}
		})
	}
}
