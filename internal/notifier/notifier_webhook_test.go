package notifier

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/testutil"
)

type httpClientMock struct {
	forceError bool
	statusCode int
	body       string
	signature  string
}

func (h *httpClientMock) Do(req *http.Request) (*http.Response, error) {
	if h.forceError {
		return nil, errors.New("random error")
	}

	b, _ := ioutil.ReadAll(req.Body)
	h.body = string(b)
	h.signature = req.Header.Get("X-Hub-Signature-256")
	resp := &http.Response{
		Body: ioutil.NopCloser(bytes.NewReader([]byte(""))),
	}
	resp.StatusCode = h.statusCode
	return resp, nil
}

func Test_webhookNotifier_Notify(t *testing.T) {
	type fields struct {
		Secret string
	}
	type expected struct {
		err       bool
		errLog    string
		bodyPath  string
		signature string
	}
	type args struct {
		diff       model.DiffCache
		statusCode int
		forceError bool
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
				bodyPath:  "../../testdata/internal/cache/notifier/webhook/should_call_webhook_and_have_valid_results.json",
				signature: "sha256=9859701f2e692d33b0cf7ed4546c56dc0d0df8d587e95472f36592c482cd835d",
			},
			args: args{
				statusCode: http.StatusOK,
				diff: model.DiffCache{
					Added: map[string]model.Flag{
						"test-flag3": {
							Percentage: 5,
							True:       "test",
							False:      "false",
							Default:    "default",
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
								Rule:       "key eq \"not-a-key\"",
								Percentage: 100,
								True:       true,
								False:      false,
								Default:    false,
							},
							After: model.Flag{
								Rule:       "key eq \"not-a-key\"",
								Percentage: 100,
								True:       true,
								False:      false,
								Default:    false,
								Disable:    true,
							},
						},
					},
				},
			},
		},
		{
			name: "should not be signed if no secret",
			expected: expected{
				bodyPath:  "../../testdata/internal/cache/notifier/webhook/should_not_be_signed_if_no_secret.json",
				signature: "",
			},
			args: args{
				statusCode: http.StatusOK,
				diff: model.DiffCache{
					Added: map[string]model.Flag{
						"test-flag3": {
							Percentage: 5,
							True:       "test",
							False:      "false",
							Default:    "default",
						},
					},
					Deleted: map[string]model.Flag{},
					Updated: map[string]model.DiffUpdated{},
				},
			},
		},
		{
			name: "should log if http code is superior to 399",
			expected: expected{
				err:    true,
				errLog: "\\[" + testutil.RFC3339Regex + "\\] error: while calling webhook, statusCode = 400",
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
				errLog: "\\[" + testutil.RFC3339Regex + "\\] error: while calling webhook: random error",
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

			webhookURL, _ := url.Parse("http://webhook.example/hook")
			mockHTTPClient := &httpClientMock{statusCode: tt.args.statusCode, forceError: tt.args.forceError}

			c := WebhookNotifier{
				Logger:     log.New(logFile, "", 0),
				HTTPClient: mockHTTPClient,
				PayloadURL: *webhookURL,
				Secret:     tt.fields.Secret,
				Meta:       map[string]string{"hostname": "toto"},
			}

			w := sync.WaitGroup{}
			w.Add(1)
			c.Notify(tt.args.diff, &w)

			if tt.expected.err {
				log, _ := ioutil.ReadFile(logFile.Name())
				fmt.Println(string(log))
				assert.Regexp(t, tt.expected.errLog, string(log))
			} else {
				content, _ := ioutil.ReadFile(tt.expected.bodyPath)
				assert.JSONEq(t, string(content), mockHTTPClient.body)
				assert.Equal(t, tt.expected.signature, mockHTTPClient.signature)
			}
		})
	}
}
