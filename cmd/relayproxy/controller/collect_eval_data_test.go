package controller_test

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func Test_collect_eval_data_Handler(t *testing.T) {
	type want struct {
		httpCode          int
		bodyFile          string
		handlerErr        bool
		errorMsg          string
		errorCode         int
		collectedDataFile string
	}
	type args struct {
		bodyFile string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid usecase",
			args: args{
				bodyFile: "../testdata/controller/collect_eval_data/valid_request.json",
			},
			want: want{
				httpCode:          http.StatusOK,
				bodyFile:          "../testdata/controller/collect_eval_data/valid_response.json",
				collectedDataFile: "../testdata/controller/collect_eval_data/valid_collected_data.json",
			},
		},
		{
			name: "valid with source field",
			args: args{
				bodyFile: "../testdata/controller/collect_eval_data/request_with_source_field.json",
			},
			want: want{
				httpCode:          http.StatusOK,
				bodyFile:          "../testdata/controller/collect_eval_data/valid_response.json",
				collectedDataFile: "../testdata/controller/collect_eval_data/collected_data_with_source_field.json",
			},
		},
		{
			name: "invalid json",
			args: args{
				bodyFile: "../testdata/controller/collect_eval_data/invalid_request.json",
			},
			want: want{
				handlerErr: true,
				httpCode:   http.StatusBadRequest,
				errorMsg: "collectEvalData: invalid input data code=400, message=Syntax error: offset=322, " +
					"error=invalid character '}' after array element, internal=invalid character '}' after array " +
					"element",
				errorCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid data field",
			args: args{
				bodyFile: "../testdata/controller/collect_eval_data/invalid_request_data_null.json",
			},
			want: want{
				handlerErr: true,
				httpCode:   http.StatusBadRequest,
				errorMsg:   "collectEvalData: invalid input data",
				errorCode:  http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporterFile, err := os.CreateTemp("", "exporter.json")
			assert.NoError(t, err)
			defer os.Remove(exporterFile.Name())

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
					Exporter:         &fileexporter.Exporter{Filename: exporterFile.Name()},
				},
			})
			ctrl := controller.NewCollectEvalData(goFF)

			e := echo.New()
			rec := httptest.NewRecorder()

			// read wantBody request file
			var bodyReq io.Reader
			if tt.args.bodyFile != "" {
				bodyReqContent, err := os.ReadFile(tt.args.bodyFile)
				assert.NoError(t, err, "request wantBody file missing %s", tt.args.bodyFile)
				bodyReq = strings.NewReader(string(bodyReqContent))
			}

			metrics := metric.NewMetrics()
			prometheus := prometheus.NewPrometheus("gofeatureflag", nil, metrics.MetricList())
			prometheus.Use(e)

			req := httptest.NewRequest(echo.POST, "/v1/data/collector", bodyReq)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)
			c.Set(metric.CustomMetrics, metrics)
			c.SetPath("/v1/data/collector")
			handlerErr := ctrl.Handler(c)

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

			goFF.Close()
			wantBody, err := os.ReadFile(tt.want.bodyFile)
			assert.NoError(t, err)
			wantCollectData, err := os.ReadFile(tt.want.collectedDataFile)
			assert.NoError(t, err)
			exportedData, err := os.ReadFile(exporterFile.Name())
			assert.NoError(t, err)

			// replace the timestamps in the response
			regex := regexp.MustCompile(`\d{10}`)
			replacedStr := regex.ReplaceAllString(rec.Body.String(), "1652273630")

			// validate the result
			assert.NoError(t, err, "Impossible the expected wantBody file %s", tt.want.bodyFile)
			assert.Equal(t, tt.want.httpCode, rec.Code, "Invalid HTTP Code")
			assert.JSONEq(t, string(wantBody), replacedStr, "Invalid response wantBody")
			assert.Equal(t, string(wantCollectData), string(exportedData), "Invalid exported data")
		})
	}
}
