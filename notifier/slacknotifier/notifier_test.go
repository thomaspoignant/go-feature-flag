package slacknotifier

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	flagv1 "github.com/thomaspoignant/go-feature-flag/internal/flagv1"
	"github.com/thomaspoignant/go-feature-flag/notifier"

	"github.com/thomaspoignant/go-feature-flag/testutils"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
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
				bodyPath: "../../testdata/internal/notifier/slack/should_call_webhook_and_have_valid_results.json",
			},
			args: args{
				url:        "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				statusCode: http.StatusOK,
				diff: notifier.DiffCache{
					Added: map[string]flag.Flag{
						"test-flag3": &flagv1.FlagData{
							Percentage:  testconvert.Float64(5),
							True:        testconvert.Interface("test"),
							False:       testconvert.Interface("false"),
							Default:     testconvert.Interface("default"),
							Rule:        testconvert.String("key eq \"random-key\""),
							TrackEvents: testconvert.Bool(true),
							Disable:     testconvert.Bool(false),
							Version:     testconvert.Float64(1.1),
						},
					},
					Deleted: map[string]flag.Flag{
						"test-flag": &flagv1.FlagData{
							Rule:       testconvert.String("key eq \"random-key\""),
							Percentage: testconvert.Float64(100),
							True:       testconvert.Interface(true),
							False:      testconvert.Interface(false),
							Default:    testconvert.Interface(false),
						},
					},
					Updated: map[string]notifier.DiffUpdated{
						"test-flag2": {
							Before: &flagv1.FlagData{
								Percentage:  testconvert.Float64(100),
								True:        testconvert.Interface(true),
								False:       testconvert.Interface(false),
								Default:     testconvert.Interface(false),
								Disable:     testconvert.Bool(false),
								TrackEvents: testconvert.Bool(true),
								Rollout: &flagv1.Rollout{
									Experimentation: &flagv1.Experimentation{
										Start: testconvert.Time(time.Unix(1095379400, 0)),
										End:   testconvert.Time(time.Unix(1095371000, 0)),
									},
								},
							},
							After: &flagv1.FlagData{
								Rule:        testconvert.String("key eq \"not-a-ke\""),
								Percentage:  testconvert.Float64(80),
								True:        testconvert.Interface("strTrue"),
								False:       testconvert.Interface("strFalse"),
								Default:     testconvert.Interface("strDefault"),
								Disable:     testconvert.Bool(true),
								TrackEvents: testconvert.Bool(false),
								Version:     testconvert.Float64(1.1),
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
				forceError: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPClient := &testutils.HTTPClientMock{StatusCode: tt.args.statusCode, ForceError: tt.args.forceError}

			slackURL, _ := url.Parse(tt.args.url)
			c := Notifier{
				SlackWebhookURL: *slackURL,
				httpClient:      mockHTTPClient,
			}

			w := sync.WaitGroup{}
			w.Add(1)
			err := c.Notify(tt.args.diff, &w)

			if tt.expected.err {
				assert.ErrorContains(t, err, tt.expected.errMsg)
			} else {
				assert.NoError(t, err)
				hostname, _ := os.Hostname()
				content, _ := ioutil.ReadFile(tt.expected.bodyPath)
				expectedContent := strings.ReplaceAll(string(content), "{{hostname}}", hostname)
				assert.JSONEq(t, expectedContent, mockHTTPClient.Body)
				assert.Equal(t, tt.expected.signature, mockHTTPClient.Signature)
			}
		})
	}
}
