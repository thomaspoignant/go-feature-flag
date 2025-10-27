package ofrep_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/ofrep"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"go.uber.org/zap"
)

const configFlagsLocation = "../testdata/controller/config_flags.yaml"

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
				bodyFile:            "../testdata/ofrep/valid_request.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/ofrep/responses/valid_response.json",
			},
		},
		{
			name: "specify flag list in context",
			args: args{
				bodyFile:            "../testdata/ofrep/valid_request_specify_flags.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/ofrep/responses/valid_response_specify_flags.json",
			},
		},
		{
			name: "Invalid context",
			args: args{
				bodyFile:            "../testdata/ofrep/invalid_context.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusBadRequest,
				bodyFile: "../testdata/ofrep/responses/invalid_context.json",
			},
		},
		{
			name: "Nil context",
			args: args{
				bodyFile:            "../testdata/ofrep/nil_context.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusBadRequest,
				bodyFile: "../testdata/ofrep/responses/nil_context.json",
			},
		},
		{
			name: "No Targeting Key in context",
			args: args{
				bodyFile:            "../testdata/ofrep/no_targeting_key_context.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusBadRequest,
				bodyFile: "../testdata/ofrep/responses/no_targeting_key_context.json",
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

			flagsetManager, err := service.NewFlagsetManager(conf, zap.NewNop(), nil)
			assert.NoError(t, err, "failed to create flagset manager")
			defer flagsetManager.Close()

			ctrl := ofrep.NewOFREPEvaluate(flagsetManager, metric.Metrics{})
			e := echo.New()
			rec := httptest.NewRecorder()

			// read wantBody request file
			var bodyReq io.Reader
			if tt.args.bodyFile != "" {
				bodyReqContent, err := os.ReadFile(tt.args.bodyFile)
				assert.NoError(t, err, "request wantBody file missing %s", tt.args.bodyFile)
				bodyReq = strings.NewReader(string(bodyReqContent))
			}

			req := httptest.NewRequest(echo.POST, "/ofrep/v1/evaluate/flags", bodyReq)
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
				bodyFile:            "../testdata/ofrep/valid_request.json",
				configFlagsLocation: configFlagsLocation,
				flagKey:             "number-flag",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/ofrep/responses/valid_evaluation.json",
			},
		},
		{
			name: "Invalid context",
			args: args{
				bodyFile:            "../testdata/ofrep/invalid_context.json",
				configFlagsLocation: configFlagsLocation,
				flagKey:             "number-flag",
			},
			want: want{
				httpCode: http.StatusBadRequest,
				bodyFile: "../testdata/ofrep/responses/invalid_context_with_key.json",
			},
		},
		{
			name: "Nil context",
			args: args{
				bodyFile:            "../testdata/ofrep/nil_context.json",
				configFlagsLocation: configFlagsLocation,
				flagKey:             "number-flag",
			},
			want: want{
				httpCode: http.StatusBadRequest,
				bodyFile: "../testdata/ofrep/responses/nil_context_with_key.json",
			},
		},
		{
			name: "No Targeting Key in context",
			args: args{
				bodyFile:            "../testdata/ofrep/no_targeting_key_context.json",
				configFlagsLocation: configFlagsLocation,
				flagKey:             "number-flag",
			},
			want: want{
				httpCode: http.StatusBadRequest,
				bodyFile: "../testdata/ofrep/responses/no_targeting_key_context_with_key.json",
			},
		},
		{
			name: "Empty flag key",
			args: args{
				bodyFile:            "../testdata/ofrep/valid_request.json",
				configFlagsLocation: configFlagsLocation,
				flagKey:             "",
			},
			want: want{
				httpCode: http.StatusNotFound,
				bodyFile: "../testdata/ofrep/responses/not_found.json",
			},
		},
		{
			name: "targeting using the field targetingKey in the rules",
			args: args{
				bodyFile:            "../testdata/ofrep/valid_targeting_key_query_request.json",
				configFlagsLocation: configFlagsLocation,
				flagKey:             "targeting-key-rule",
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/ofrep/responses/valid_targeting_key_query_response.json",
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

			flagsetManager, err := service.NewFlagsetManager(conf, zap.NewNop(), nil)
			assert.NoError(t, err, "failed to create flagset manager")
			defer flagsetManager.Close()

			ctrl := ofrep.NewOFREPEvaluate(flagsetManager, metric.Metrics{})
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
			req := httptest.NewRequest(echo.POST, "/ofrep/v1/evaluate/flags/"+flagKey, bodyReq)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			e.ServeHTTP(rec, req)
			wantBody, err := os.ReadFile(tt.want.bodyFile)
			assert.NoError(t, err, "Impossible the expected wantBody file %s", tt.want.bodyFile)
			assert.Equal(t, tt.want.httpCode, rec.Code, "Invalid HTTP Code")
			assert.JSONEq(t, string(wantBody), rec.Body.String(), "Invalid response wantBody")
		})
	}
}
