package controller_test

import (
	"context"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/exporter/logsexporter"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

const configFlagsLocation = "../testdata/controller/config_flags.yaml"

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
				bodyFile: "../testdata/controller/flag_eval/valid_request.json",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/controller/flag_eval/valid_response.json",
			},
		},
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:  "disable-flag",
				bodyFile: "../testdata/controller/flag_eval/disable_flag_request.json",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/controller/flag_eval/disable_flag_response.json",
			},
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:  "random-key-does-not-exist",
				bodyFile: "../testdata/controller/flag_eval/flag_not_exist_request.json",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/controller/flag_eval/flag_not_exist_response.json",
			},
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:  "test-flag-rule-not-apply",
				bodyFile: "../testdata/controller/flag_eval/rule_not_apply_request.json",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/controller/flag_eval/rule_not_apply_response.json",
			},
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:  "test-flag-rule-apply",
				bodyFile: "../testdata/controller/flag_eval/rule_apply_request.json",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/controller/flag_eval/rule_apply_response.json",
			},
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:  "test-flag-rule-apply-false",
				bodyFile: "../testdata/controller/flag_eval/rule_apply_false_request.json",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/controller/flag_eval/rule_apply_false_response.json",
			},
		},
		{
			name: "Invalid json format",
			args: args{
				flagKey:  "test-flag-rule-apply-false",
				bodyFile: "../testdata/controller/flag_eval/invalid_json_request.json",
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
				bodyFile: "../testdata/controller/flag_eval/no_user_key_request.json",
			},
			want: want{
				handlerErr: true,
				errorMsg:   "empty key for user, impossible to retrieve flags",
				errorCode:  http.StatusBadRequest,
			},
		},
		{
			name: "no flag key in URL",
			args: args{
				flagKey:  "",
				bodyFile: "../testdata/controller/flag_eval/valid_request.json",
			},
			want: want{
				handlerErr: true,
				errorMsg:   "impossible to find the flag key in the URL",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init go-feature-flag
			goFF, _ := ffclient.New(ffclient.Config{
				PollingInterval: 10 * time.Second,
				Logger:          log.New(os.Stdout, "", 0),
				Context:         context.Background(),
				Retriever: &fileretriever.Retriever{
					Path: configFlagsLocation,
				},
				DataExporter: ffclient.DataExporter{
					FlushInterval:    10 * time.Second,
					MaxEventInMemory: 10000,
					Exporter:         &logsexporter.Exporter{},
				},
			})
			defer goFF.Close()

			flagEval := controller.NewFlagEval(goFF)

			e := echo.New()
			rec := httptest.NewRecorder()

			// read wantBody request file
			var bodyReq io.Reader
			if tt.args.bodyFile != "" {
				bodyReqContent, err := ioutil.ReadFile(tt.args.bodyFile)
				assert.NoError(t, err, "request wantBody file missing %s", tt.args.bodyFile)
				bodyReq = strings.NewReader(string(bodyReqContent))
			}

			metrics := metric.NewMetrics()
			prometheus := prometheus.NewPrometheus("gofeatureflag", nil, metrics.MetricList())
			prometheus.Use(e)

			req := httptest.NewRequest(echo.POST, "/v1/feature/"+tt.args.flagKey+"/eval", bodyReq)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)
			c.Set(metric.CustomMetrics, metrics)
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

			wantBody, err := ioutil.ReadFile(tt.want.bodyFile)
			assert.NoError(t, err, "Impossible the expected wantBody file %s", tt.want.bodyFile)
			assert.Equal(t, tt.want.httpCode, rec.Code, "Invalid HTTP Code")
			assert.JSONEq(t, string(wantBody), rec.Body.String(), "Invalid response wantBody")
		})
	}
}
