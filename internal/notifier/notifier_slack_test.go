package notifier

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/testutils"
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
			name: "should call webhook and have valid results",
			expected: expected{
				bodyPath: "../../testdata/internal/notifier/slack/should_call_webhook_and_have_valid_results.json",
			},
			args: args{
				statusCode: http.StatusOK,
				diff: model.DiffCache{
					Added: map[string]model.Flag{
						"test-flag3": {
							Percentage:  5,
							True:        "test",
							False:       "false",
							Default:     "default",
							Rule:        "key eq \"random-key\"",
							TrackEvents: testutils.Bool(true),
						},
					},
					Deleted: map[string]model.Flag{
						"test-flag": {
							Rule:       "key eq \"random-key\"",
							Percentage: 100,
							True:       true,
							False:      false,
							Default:    false,
						},
					},
					Updated: map[string]model.DiffUpdated{
						"test-flag2": {
							Before: model.Flag{
								Rule:        "key eq \"not-a-key\"",
								Percentage:  100,
								True:        true,
								False:       false,
								Default:     false,
								TrackEvents: testutils.Bool(true),
							},
							After: model.Flag{
								Rule:        "key eq \"not-a-ke\"",
								Percentage:  80,
								True:        "strTrue",
								False:       "strFalse",
								Default:     "strDefault",
								Disable:     true,
								TrackEvents: testutils.Bool(false),
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

			mockHTTPClient := &httpClientMock{statusCode: tt.args.statusCode, forceError: tt.args.forceError}

			c := NewSlackNotifier(
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
				assert.JSONEq(t, expectedContent, mockHTTPClient.body)
				assert.Equal(t, tt.expected.signature, mockHTTPClient.signature)
			}
		})
	}
}
