package notifier

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
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
				signature: "sha256=b60afdbdfcac21b957c68a14c0fead647f44bc71c4181433809100cb8e6690b3",
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
			name: "should not be signed if no secret",
			expected: expected{
				bodyPath:  "../../testdata/internal/notifier/webhook/should_not_be_signed_if_no_secret.json",
				signature: "",
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
					Deleted: map[string]flag.Flag{},
					Updated: map[string]model.DiffUpdated{},
				},
			},
		},
		{
			name: "should log if http code is superior to 399",
			expected: expected{
				err:    true,
				errLog: "error: while calling webhook, statusCode = 400",
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
				errLog: "error: while calling webhook: random error",
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
				fflog.Logger{Logger: log.New(logFile, "", 0)},
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
			got, err := NewWebhookNotifier(fflog.Logger{}, mockHTTPClient, tt.args.endpointURL, tt.args.secret, tt.args.meta)

			if tt.wantErr {
				assert.Error(t, err, "NewWebhookNotifier should return an error")
			} else {
				assert.NoError(t, err, "NewWebhookNotifier should not return an error. Error return: %v", err)
				assert.Equal(t, tt.want, got, "WebhookNotifier should be equals.")
			}
		})
	}
}
