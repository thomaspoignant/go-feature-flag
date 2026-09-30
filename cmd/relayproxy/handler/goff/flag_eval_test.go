package controller_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	controller "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/goff"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

const configFlagsLocation = testdataDir + "/config_flags.yaml"

func Test_flag_eval_Handler(t *testing.T) {
	type want struct {
		httpCode   int
		bodyFile   string
		handlerErr bool
		errorMsg   string
		errorCode  int
	}

	type args struct {
		flagKey  string
		bodyFile string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid flag",
			args: args{
				flagKey:  "flag-only-for-admin",
				bodyFile: testdataDir + "/flag_eval/valid_request.json",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: testdataDir + "/flag_eval/valid_response.json",
			},
		},
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:  "disable-flag",
				bodyFile: testdataDir + "/flag_eval/disable_flag_request.json",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: testdataDir + "/flag_eval/disable_flag_response.json",
			},
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:  "random-key-does-not-exist",
				bodyFile: testdataDir + "/flag_eval/flag_not_exist_request.json",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: testdataDir + "/flag_eval/flag_not_exist_response.json",
			},
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:  "test-flag-rule-not-apply",
				bodyFile: testdataDir + "/flag_eval/rule_not_apply_request.json",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: testdataDir + "/flag_eval/rule_not_apply_response.json",
			},
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:  "test-flag-rule-apply",
				bodyFile: testdataDir + "/flag_eval/rule_apply_request.json",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: testdataDir + "/flag_eval/rule_apply_response.json",
			},
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:  "test-flag-rule-apply-false",
				bodyFile: testdataDir + "/flag_eval/rule_apply_false_request.json",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: testdataDir + "/flag_eval/rule_apply_false_response.json",
			},
		},
		{
			name: "Invalid json format",
			args: args{
				flagKey:  "test-flag-rule-apply-false",
				bodyFile: testdataDir + "/flag_eval/invalid_json_request.json",
			},
			want: want{
				handlerErr: true,
				errorMsg:   "unexpected EOF",
				errorCode:  http.StatusBadRequest,
			},
		},
		{
			name: "No user key in payload",
			args: args{
				flagKey:  "test-flag-rule-apply-false",
				bodyFile: testdataDir + "/flag_eval/no_user_key_request.json",
			},
			want: want{
				handlerErr: false,
				bodyFile:   testdataDir + "/flag_eval/no_user_key_response.json",
				httpCode:   http.StatusOK,
			},
		},
		{
			name: "no flag key in URL",
			args: args{
				flagKey:  "",
				bodyFile: testdataDir + "/flag_eval/valid_request.json",
			},
			want: want{
				handlerErr: true,
				errorMsg:   "impossible to find the flag key in the URL",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config for default mode
			conf := config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: retrieverconf.FileRetriever,
						Path: configFlagsLocation,
					},
					Exporter: &config.ExporterConf{
						Kind: config.LogExporter,
					},
				},
			}

			flagsetManager, err := service.NewFlagsetManager(&conf, zap.NewNop(), []notifier.Notifier{}, nil)
			assert.NoError(t, err, "impossible to create flagset manager")

			flagEval := controller.NewFlagEval(flagsetManager, metric.Metrics{})

			e := echo.New()
			rec := httptest.NewRecorder()

			// read wantBody request file
			var bodyReq io.Reader
			if tt.args.bodyFile != "" {
				bodyReqContent, err := os.ReadFile(tt.args.bodyFile)
				assert.NoError(t, err, "request wantBody file missing %s", tt.args.bodyFile)
				bodyReq = strings.NewReader(string(bodyReqContent))
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/feature/"+tt.args.flagKey+"/eval", bodyReq)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)
			c.SetPath("/v1/feature/:flagKey/eval")
			c.SetParamNames("flagKey")
			c.SetParamValues(tt.args.flagKey)
			handlerErr := flagEval.Handler(c)

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

func Test_flag_eval_Handler_FlagEvaluationMetric(t *testing.T) {
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
			flagKeys: []string{"flag-only-for-admin", "flag-only-for-admin", "test-flag-rule-apply"},
			bodyFile: testdataDir + "/flag_eval/valid_request.json",
			want: metricHeader +
				metricName + "{flag_name=\"flag-only-for-admin\"} 2\n" +
				metricName + "{flag_name=\"test-flag-rule-apply\"} 1\n",
		},
		{
			name:     "should use the same label for all the flags that do not exist",
			flagKeys: []string{"unknown-flag-1", "unknown-flag-2", "unknown-flag-3"},
			bodyFile: testdataDir + "/flag_eval/flag_not_exist_request.json",
			want:     metricHeader + metricName + "{flag_name=\"FLAG_NOT_FOUND\"} 3\n",
		},
		{
			name:     "should not count a request that is not evaluated",
			flagKeys: []string{"flag-only-for-admin", "unknown-flag"},
			bodyFile: testdataDir + "/flag_eval/invalid_json_request.json",
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: retrieverconf.FileRetriever,
						Path: configFlagsLocation,
					},
				},
			}

			flagsetManager, err := service.NewFlagsetManager(&conf, zap.NewNop(), []notifier.Notifier{}, nil)
			require.NoError(t, err, "impossible to create flagset manager")
			defer flagsetManager.Close()

			metrics, err := metric.NewMetrics()
			require.NoError(t, err)

			flagEval := controller.NewFlagEval(flagsetManager, metrics)
			e := echo.New()

			bodyReqContent, err := os.ReadFile(tt.bodyFile)
			require.NoError(t, err, "request body file missing %s", tt.bodyFile)
			for _, flagKey := range tt.flagKeys {
				req := httptest.NewRequest(
					http.MethodPost, "/v1/feature/"+flagKey+"/eval", strings.NewReader(string(bodyReqContent)))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				c := e.NewContext(req, httptest.NewRecorder())
				c.SetPath("/v1/feature/:flagKey/eval")
				c.SetParamNames("flagKey")
				c.SetParamValues(flagKey)
				_ = flagEval.Handler(c)
			}

			assert.NoError(t, testutil.GatherAndCompare(metrics.Registry, strings.NewReader(tt.want), metricName))
		})
	}
}
