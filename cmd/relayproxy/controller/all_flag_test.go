package controller_test

import (
	"context"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
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

func Test_all_flag_Handler(t *testing.T) {
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
				bodyFile:            "../testdata/controller/all_flags/valid_request.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/controller/all_flags/valid_response.json",
			},
		},
		{
			name: "Invalid json format",
			args: args{
				bodyFile:            "../testdata/controller/all_flags/invalid_json_request.json",
				configFlagsLocation: configFlagsLocation,
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
				bodyFile:            "../testdata/controller/all_flags/no_user_key_request.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				handlerErr: true,
				errorMsg:   "empty key for evaluation context, impossible to retrieve flags",
				errorCode:  http.StatusBadRequest,
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
					Path: tt.args.configFlagsLocation,
				},
				DataExporter: ffclient.DataExporter{
					FlushInterval:    10 * time.Second,
					MaxEventInMemory: 10000,
					Exporter:         &logsexporter.Exporter{},
				},
			})
			defer goFF.Close()
			ctrl := controller.NewAllFlags(goFF, metric.Metrics{})

			e := echo.New()
			rec := httptest.NewRecorder()

			// read wantBody request file
			var bodyReq io.Reader
			if tt.args.bodyFile != "" {
				bodyReqContent, err := os.ReadFile(tt.args.bodyFile)
				assert.NoError(t, err, "request wantBody file missing %s", tt.args.bodyFile)
				bodyReq = strings.NewReader(string(bodyReqContent))
			}

			req := httptest.NewRequest(echo.POST, "/v1/allflags", bodyReq)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)
			c.SetPath("/v1/allflags")
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

			wantBody, err := os.ReadFile(tt.want.bodyFile)

			// replace the timestamps in the response
			regex := regexp.MustCompile(`\d{10}`)
			replacedStr := regex.ReplaceAllString(rec.Body.String(), "1652273630")

			// validate the result
			assert.NoError(t, err, "Impossible the expected wantBody file %s", tt.want.bodyFile)
			assert.Equal(t, tt.want.httpCode, rec.Code, "Invalid HTTP Code")
			assert.JSONEq(t, string(wantBody), replacedStr, "Invalid response wantBody")
		})
	}
}
