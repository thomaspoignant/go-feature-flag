package notifier

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/testutils"
)

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
				bodyPath:  "../../testdata/internal/notifier/webhook/should_call_webhook_and_have_valid_results.json",
				signature: "sha256=23effe4da9927ab72df5202a3146e6be39c12b7f6cee99f8d2e19326d8806b81",
			},
			args: args{
				statusCode: http.StatusOK,
				diff: model.DiffCache{
					Added: map[string]model.Flag{
						"test-flag3": &model.FlagData{
							Percentage: testconvert.Float64(5),
							True:       testconvert.Interface("test"),
							False:      testconvert.Interface("false"),
							Default:    testconvert.Interface("default"),
						},
					},
					Deleted: map[string]model.Flag{
						"test-flag": &model.FlagData{
							Rule:       testconvert.String("key eq \"random-key\""),
							Percentage: testconvert.Float64(100),
							True:       testconvert.Interface(true),
							False:      testconvert.Interface(false),
							Default:    testconvert.Interface(false),
						},
					},
					Updated: map[string]model.DiffUpdated{
						"test-flag2": {
							Before: &model.FlagData{
								Rule:       testconvert.String("key eq \"not-a-key\""),
								Percentage: testconvert.Float64(100),
								True:       testconvert.Interface(true),
								False:      testconvert.Interface(false),
								Default:    testconvert.Interface(false),
							},
							After: &model.FlagData{
								Rule:       testconvert.String("key eq \"not-a-key\""),
								Percentage: testconvert.Float64(100),
								True:       testconvert.Interface(true),
								False:      testconvert.Interface(false),
								Default:    testconvert.Interface(false),
								Disable:    testconvert.Bool(true),
							},
						},
					},
				},
			},
		},
		{
			name: "should not be signed if no secret",
			expected: expected{
				bodyPath:  "../../testdata/internal/notifier/webhook/should_not_be_signed_if_no_secret.json",
				signature: "",
			},
			args: args{
				statusCode: http.StatusOK,
				diff: model.DiffCache{
					Added: map[string]model.Flag{
						"test-flag3": &model.FlagData{
							Percentage: testconvert.Float64(5),
							True:       testconvert.Interface("test"),
							False:      testconvert.Interface("false"),
							Default:    testconvert.Interface("default"),
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
				errLog: "^\\[" + testutils.RFC3339Regex + "\\] error: while calling webhook, statusCode = 400",
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
				errLog: "^\\[" + testutils.RFC3339Regex + "\\] error: while calling webhook: random error",
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

			c, _ := NewWebhookNotifier(
				log.New(logFile, "", 0),
				mockHTTPClient,
				"http://webhook.example/hook",
				tt.fields.Secret,
				map[string]string{"hostname": "toto"},
			)

			w := sync.WaitGroup{}
			w.Add(1)
			c.Notify(tt.args.diff, &w)

			if tt.expected.err {
				log, _ := ioutil.ReadFile(logFile.Name())
				assert.Regexp(t, tt.expected.errLog, string(log))
			} else {
				content, _ := ioutil.ReadFile(tt.expected.bodyPath)
				assert.JSONEq(t, string(content), mockHTTPClient.Body)
				assert.Equal(t, tt.expected.signature, mockHTTPClient.Signature)
			}
		})
	}
}

func TestNewWebhookNotifier(t *testing.T) {
	mockHTTPClient := &testutils.HTTPClientMock{StatusCode: 200, ForceError: false}
	hostname, _ := os.Hostname()

	type args struct {
		endpointURL string
		secret      string
		meta        map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    WebhookNotifier
		wantErr bool
	}{
		{
			name: "Invalid URL",
			args: args{
				endpointURL: " http://example.com",
			},
			wantErr: true,
		},
		{
			name: "No meta",
			args: args{
				endpointURL: "http://example.com",
			},
			wantErr: false,
			want: WebhookNotifier{
				HTTPClient:  mockHTTPClient,
				EndpointURL: url.URL{Host: "example.com", Scheme: "http"},
				Secret:      "",
				Meta:        map[string]string{"hostname": hostname},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWebhookNotifier(nil, mockHTTPClient, tt.args.endpointURL, tt.args.secret, tt.args.meta)

			if tt.wantErr {
				assert.Error(t, err, "NewWebhookNotifier should return an error")
			} else {
				assert.NoError(t, err, "NewWebhookNotifier should not return an error. Error return: %v", err)
				assert.Equal(t, tt.want, got, "WebhookNotifier should be equals.")
			}
		})
	}
}
