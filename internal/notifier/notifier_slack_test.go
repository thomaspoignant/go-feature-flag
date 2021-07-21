package notifier_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	flagv1 "github.com/thomaspoignant/go-feature-flag/internal/flagv1"
	"github.com/thomaspoignant/go-feature-flag/internal/notifier"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
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
					Updated: map[string]model.DiffUpdated{
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
									}},
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
				assert.JSONEq(t, expectedContent, mockHTTPClient.Body)
				assert.Equal(t, tt.expected.signature, mockHTTPClient.Signature)
			}
		})
	}
}
