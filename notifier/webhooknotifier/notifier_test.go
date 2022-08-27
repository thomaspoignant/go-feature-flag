package webhooknotifier

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	flagv1 "github.com/thomaspoignant/go-feature-flag/internal/flagv1"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"

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
	}
	type args struct {
		diff       notifier.DiffCache
		statusCode int
		forceError bool
		url        string
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
				signature: "sha256=23effe4da9927ab72df5202a3146e6be39c12b7f6cee99f8d2e19326d8806b81",
			},
			args: args{
				url:        "http://webhook.example/hook",
				statusCode: http.StatusOK,
				diff: notifier.DiffCache{
					Added: map[string]flag.Flag{
						"test-flag3": &flagv1.FlagData{
							Percentage: testconvert.Float64(5),
							True:       testconvert.Interface("test"),
							False:      testconvert.Interface("false"),
							Default:    testconvert.Interface("default"),
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
								Rule:       testconvert.String("key eq \"not-a-key\""),
								Percentage: testconvert.Float64(100),
								True:       testconvert.Interface(true),
								False:      testconvert.Interface(false),
								Default:    testconvert.Interface(false),
							},
							After: &flagv1.FlagData{
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
				bodyPath:  "./testdata/should_not_be_signed_if_no_secret.json",
				signature: "",
			},
			args: args{
				url:        "http://webhook.example/hook",
				statusCode: http.StatusOK,
				diff: notifier.DiffCache{
					Added: map[string]flag.Flag{
						"test-flag3": &flagv1.FlagData{
							Percentage: testconvert.Float64(5),
							True:       testconvert.Interface("test"),
							False:      testconvert.Interface("false"),
							Default:    testconvert.Interface("default"),
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPClient := &testutils.HTTPClientMock{StatusCode: tt.args.statusCode, ForceError: tt.args.forceError}

			c := Notifier{
				EndpointURL: tt.args.url,
				Secret:      tt.fields.Secret,
				Meta:        map[string]string{"hostname": "toto"},
				httpClient:  mockHTTPClient,
				init:        sync.Once{},
			}

			w := sync.WaitGroup{}
			w.Add(1)
			err := c.Notify(tt.args.diff, &w)

			if tt.expected.err {
				assert.ErrorContains(t, err, tt.expected.errorMsg)
			} else {
				assert.NoError(t, err)
				content, _ := os.ReadFile(tt.expected.bodyPath)
				assert.JSONEq(t, string(content), mockHTTPClient.Body)
				assert.Equal(t, tt.expected.signature, mockHTTPClient.Signature)
			}
		})
	}
}

func Test_webhookNotifier_no_meta_data(t *testing.T) {
	mockHTTPClient := &testutils.HTTPClientMock{StatusCode: 200, ForceError: false}
	diff := notifier.DiffCache{
		Added: map[string]flag.Flag{
			"test-flag3": &flagv1.FlagData{
				Percentage: testconvert.Float64(5),
				True:       testconvert.Interface("test"),
				False:      testconvert.Interface("false"),
				Default:    testconvert.Interface("default"),
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

	w := sync.WaitGroup{}
	w.Add(1)
	err := c.Notify(diff, &w)

	assert.NoError(t, err)
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(mockHTTPClient.Body), &m)

	assert.NotEmpty(t, m["meta"])
}
