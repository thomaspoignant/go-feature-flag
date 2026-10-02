package ofrep_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/ofrep"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"go.uber.org/zap"
)

const testdataDir = "../../testdata"
const configFlagsLocation = testdataDir + "/goff/config_flags.yaml"

func Test_Bulk_Evaluation(t *testing.T) {
	type want struct {
		httpCode   int
		bodyFile   string
		handlerErr bool
		errorMsg   string
		errorCode  int
	}

	type args struct {
		bodyFile            string
		configFlagsLocation string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid flag",
			args: args{
				bodyFile:            testdataDir + "/ofrep/valid_request.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: testdataDir + "/ofrep/responses/valid_response.json",
			},
		},
		{
			name: "specify flag list in context",
			args: args{
				bodyFile:            testdataDir + "/ofrep/valid_request_specify_flags.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: testdataDir + "/ofrep/responses/valid_response_specify_flags.json",
			},
		},
		{
			name: "Invalid context",
			args: args{
				bodyFile:            testdataDir + "/ofrep/invalid_context.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusBadRequest,
				bodyFile: testdataDir + "/ofrep/responses/invalid_context.json",
			},
		},
		{
			name: "Empty body",
			args: args{
				bodyFile:            testdataDir + "/ofrep/empty_body.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusBadRequest,
				bodyFile: testdataDir + "/ofrep/responses/empty_body.json",
			},
		},
		{
			name: "Nil context",
			args: args{
				bodyFile:            testdataDir + "/ofrep/nil_context.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusBadRequest,
				bodyFile: testdataDir + "/ofrep/responses/nil_context.json",
			},
		},
		{
			name: "No Targeting Key in context",
			// in this case we don't have a targetingKey, so we will evaluate the flags individually
			// if the flag requires bucketing, we will return a targeting key missing error
			args: args{
				bodyFile:            testdataDir + "/ofrep/no_targeting_key_context.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: testdataDir + "/ofrep/responses/no_targeting_key_context.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create flagset manager with configuration
			conf := &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 10000, // 10 seconds in milliseconds
					FileFormat:      "yaml",
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: retrieverconf.FileRetriever,
							Path: tt.args.configFlagsLocation,
						},
					},
				},
			}

			flagsetManager, err := service.NewFlagsetManager(conf, zap.NewNop(), nil, nil)
			assert.NoError(t, err, "failed to create flagset manager")
			defer flagsetManager.Close()

			ctrl := ofrep.NewOFREPEvaluate(flagsetManager, metric.Metrics{}, config.OfrepEventStream{})
			e := echo.New()
			rec := httptest.NewRecorder()

			// read wantBody request file
			var bodyReq io.Reader
			if tt.args.bodyFile != "" {
				bodyReqContent, err := os.ReadFile(tt.args.bodyFile)
				assert.NoError(t, err, "request wantBody file missing %s", tt.args.bodyFile)
				bodyReq = strings.NewReader(string(bodyReqContent))
			}

			req := httptest.NewRequest(http.MethodPost, "/ofrep/v1/evaluate/flags", bodyReq)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)

			c.SetPath("/ofrep/v1/evaluate/flags")
			handlerErr := ctrl.BulkEvaluate(c)

			if tt.want.handlerErr {
				assert.Error(t, handlerErr, "handler should return an error")
				he, ok := handlerErr.(*echo.HTTPError)
				if ok {
					assert.Equal(t, tt.want.errorCode, he.Code)
					assert.Equal(t, tt.want.errorMsg, he.Message)
				} else {
					assert.Equal(t, tt.want.errorMsg, handlerErr.Error())
				}
				return
			}

			wantBody, err := os.ReadFile(tt.want.bodyFile)

			assert.NoError(t, err, "Impossible the expected wantBody file %s", tt.want.bodyFile)
			assert.Equal(t, tt.want.httpCode, rec.Code, "Invalid HTTP Code")
			assert.JSONEq(t, string(wantBody), rec.Body.String(), "Invalid response wantBody")
		})
	}
}

func Test_Bulk_Evaluation_EventStreams(t *testing.T) {
	type args struct {
		eventStream config.OfrepEventStream
		headers     map[string]string
	}

	tests := []struct {
		name string
		args args
		want []model.OFREPEventStream
	}{
		{
			name: "disabled when no base URL is configured",
			args: args{},
			want: nil,
		},
		{
			name: "base URL and inactivity delay from the configuration",
			args: args{
				eventStream: config.OfrepEventStream{
					BaseURL:            "https://gofeatureflag.example.com/",
					InactivityDelaySec: 60,
				},
			},
			want: []model.OFREPEventStream{
				{
					Type:               "sse",
					URL:                "https://gofeatureflag.example.com/stream/v1/sse/flag/change",
					InactivityDelaySec: 60,
				},
			},
		},
		{
			name: "base URL with a path prefix",
			args: args{
				eventStream: config.OfrepEventStream{BaseURL: "https://example.com/goff"},
			},
			want: []model.OFREPEventStream{
				{Type: "sse", URL: "https://example.com/goff/stream/v1/sse/flag/change"},
			},
		},
		{
			name: "API key from the X-API-Key header",
			args: args{
				eventStream: config.OfrepEventStream{BaseURL: "https://example.com"},
				headers:     map[string]string{"X-API-Key": "my-key"},
			},
			want: []model.OFREPEventStream{
				{Type: "sse", URL: "https://example.com/stream/v1/sse/flag/change?apiKey=my-key"},
			},
		},
		{
			name: "API key from the Authorization header is escaped",
			args: args{
				eventStream: config.OfrepEventStream{BaseURL: "https://example.com"},
				headers:     map[string]string{"Authorization": "Bearer my key&x=1"},
			},
			want: []model.OFREPEventStream{
				{Type: "sse", URL: "https://example.com/stream/v1/sse/flag/change?apiKey=my+key%26x%3D1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 10000,
					FileFormat:      "yaml",
					Retrievers: &[]retrieverconf.RetrieverConf{
						{Kind: retrieverconf.FileRetriever, Path: configFlagsLocation},
					},
				},
			}
			flagsetManager, err := service.NewFlagsetManager(conf, zap.NewNop(), nil, nil)
			require.NoError(t, err)
			defer flagsetManager.Close()

			ctrl := ofrep.NewOFREPEvaluate(flagsetManager, metric.Metrics{}, tt.args.eventStream)
			body, err := os.ReadFile(testdataDir + "/ofrep/valid_request.json")
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/ofrep/v1/evaluate/flags", strings.NewReader(string(body)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			for k, v := range tt.args.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)
			c.SetPath("/ofrep/v1/evaluate/flags")

			require.NoError(t, ctrl.BulkEvaluate(c))
			assert.Equal(t, http.StatusOK, rec.Code)
			var got model.OFREPBulkEvaluateSuccessResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
			assert.Equal(t, tt.want, got.EventStreams)
		})
	}
}

func Test_Evaluate(t *testing.T) {
	type want struct {
		httpCode int
		bodyFile string
	}

	type args struct {
		bodyFile            string
		configFlagsLocation string
		flagKey             string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid evaluation",
			args: args{
				bodyFile:            testdataDir + "/ofrep/valid_request.json",
				configFlagsLocation: configFlagsLocation,
				flagKey:             "number-flag",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: testdataDir + "/ofrep/responses/valid_evaluation.json",
			},
		},
		{
			name: "Invalid context",
			args: args{
				bodyFile:            testdataDir + "/ofrep/invalid_context.json",
				configFlagsLocation: configFlagsLocation,
				flagKey:             "number-flag",
			},
			want: want{
				httpCode: http.StatusBadRequest,
				bodyFile: testdataDir + "/ofrep/responses/invalid_context_with_key.json",
			},
		},
		{
			name: "Nil context",
			args: args{
				bodyFile:            testdataDir + "/ofrep/nil_context.json",
				configFlagsLocation: configFlagsLocation,
				flagKey:             "number-flag",
			},
			want: want{
				httpCode: http.StatusBadRequest,
				bodyFile: testdataDir + "/ofrep/responses/nil_context_with_key.json",
			},
		},
		{
			name: "No Targeting Key for bucketing-required flag - should return 400 from core evaluation",
			args: args{
				bodyFile:            testdataDir + "/ofrep/no_targeting_key_context.json",
				configFlagsLocation: configFlagsLocation,
				flagKey:             "number-flag", // This flag has percentage rules, requires bucketing
			},
			want: want{
				httpCode: http.StatusBadRequest,
				bodyFile: testdataDir + "/ofrep/responses/no_targeting_key_bucketing_flag.json",
			},
		},
		{
			name: "No Targeting Key for non-bucketing flag - should succeed",
			args: args{
				bodyFile:            testdataDir + "/ofrep/no_targeting_key_context.json",
				configFlagsLocation: configFlagsLocation,
				flagKey:             "targeting-key-rule", // This flag has no percentages, doesn't require bucketing
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: testdataDir + "/ofrep/responses/no_targeting_key_static_flag.json",
			},
		},
		{
			name: "Percentage-based rule in flag without targeting key should return 400 error",
			args: args{
				bodyFile:            testdataDir + "/ofrep/no_targeting_key_context.json",
				configFlagsLocation: configFlagsLocation,
				flagKey:             "flag-only-for-admin", // This flag has percentage rules, requires bucketing
			},
			want: want{
				httpCode: http.StatusBadRequest, // Core evaluation returns 400 for missing targeting key
				bodyFile: testdataDir + "/ofrep/responses/percentage_flag_no_key_error.json",
			},
		},
		{
			name: "Empty flag key",
			args: args{
				bodyFile:            testdataDir + "/ofrep/valid_request.json",
				configFlagsLocation: configFlagsLocation,
				flagKey:             "",
			},
			want: want{
				httpCode: http.StatusNotFound,
				bodyFile: testdataDir + "/ofrep/responses/not_found.json",
			},
		},
		{
			name: "targeting using the field targetingKey in the rules",
			args: args{
				bodyFile:            testdataDir + "/ofrep/valid_targeting_key_query_request.json",
				configFlagsLocation: configFlagsLocation,
				flagKey:             "targeting-key-rule",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: testdataDir + "/ofrep/responses/valid_targeting_key_query_response.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create flagset manager with configuration
			conf := &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 10000, // 10 seconds in milliseconds
					FileFormat:      "yaml",
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: retrieverconf.FileRetriever,
							Path: tt.args.configFlagsLocation,
						},
					},
				},
			}

			flagsetManager, err := service.NewFlagsetManager(conf, zap.NewNop(), nil, nil)
			assert.NoError(t, err, "failed to create flagset manager")
			defer flagsetManager.Close()

			ctrl := ofrep.NewOFREPEvaluate(flagsetManager, metric.Metrics{}, config.OfrepEventStream{})
			e := echo.New()
			e.POST("/ofrep/v1/evaluate/flags/:flagKey", ctrl.Evaluate)
			rec := httptest.NewRecorder()

			flagKey := tt.args.flagKey

			// read wantBody request file
			var bodyReq io.Reader
			if tt.args.bodyFile != "" {
				bodyReqContent, err := os.ReadFile(tt.args.bodyFile)
				assert.NoError(t, err, "request wantBody file missing %s", tt.args.bodyFile)
				bodyReq = strings.NewReader(string(bodyReqContent))
			}
			req := httptest.NewRequest(http.MethodPost, "/ofrep/v1/evaluate/flags/"+flagKey, bodyReq)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			e.ServeHTTP(rec, req)
			wantBody, err := os.ReadFile(tt.want.bodyFile)
			assert.NoError(t, err, "Impossible the expected wantBody file %s", tt.want.bodyFile)
			assert.Equal(t, tt.want.httpCode, rec.Code, "Invalid HTTP Code")
			assert.JSONEq(t, string(wantBody), rec.Body.String(), "Invalid response wantBody")
		})
	}
}

func Test_Evaluate_FlagEvaluationMetric(t *testing.T) {
	const metricName = "gofeatureflag_flag_evaluations_total"
	const metricHeader = "# HELP " + metricName + " Counter events for number of flag evaluation.\n" +
		"# TYPE " + metricName + " counter\n"

	tests := []struct {
		name     string
		flagKeys []string
		bodyFile string
		want     string
	}{
		{
			name:     "should use the flag key as label if the flag exists",
			flagKeys: []string{"number-flag", "number-flag", "targeting-key-rule"},
			bodyFile: testdataDir + "/ofrep/valid_request.json",
			want: metricHeader +
				metricName + "{flag_name=\"number-flag\"} 2\n" +
				metricName + "{flag_name=\"targeting-key-rule\"} 1\n",
		},
		{
			name:     "should use the same label for all the flags that do not exist",
			flagKeys: []string{"unknown-flag-1", "unknown-flag-2", "unknown-flag-3"},
			bodyFile: testdataDir + "/ofrep/valid_request.json",
			want:     metricHeader + metricName + "{flag_name=\"FLAG_NOT_FOUND\"} 3\n",
		},
		{
			name:     "should not count a request that is not evaluated",
			flagKeys: []string{"number-flag", "unknown-flag"},
			bodyFile: testdataDir + "/ofrep/invalid_context.json",
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 10000, // 10 seconds in milliseconds
					FileFormat:      "yaml",
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: retrieverconf.FileRetriever,
							Path: configFlagsLocation,
						},
					},
				},
			}

			flagsetManager, err := service.NewFlagsetManager(conf, zap.NewNop(), nil, nil)
			require.NoError(t, err, "failed to create flagset manager")
			defer flagsetManager.Close()

			metrics, err := metric.NewMetrics()
			require.NoError(t, err)

			ctrl := ofrep.NewOFREPEvaluate(flagsetManager, metrics, config.OfrepEventStream{})
			e := echo.New()
			e.POST("/ofrep/v1/evaluate/flags/:flagKey", ctrl.Evaluate)

			bodyReqContent, err := os.ReadFile(tt.bodyFile)
			require.NoError(t, err, "request body file missing %s", tt.bodyFile)
			for _, flagKey := range tt.flagKeys {
				req := httptest.NewRequest(
					http.MethodPost, "/ofrep/v1/evaluate/flags/"+flagKey, strings.NewReader(string(bodyReqContent)))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				e.ServeHTTP(httptest.NewRecorder(), req)
			}

			assert.NoError(t, testutil.GatherAndCompare(metrics.Registry, strings.NewReader(tt.want), metricName))
		})
	}
}
