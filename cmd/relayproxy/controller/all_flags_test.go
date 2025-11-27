package controller_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

func Test_all_flag_Handler_DefaultMode(t *testing.T) {
	const configFlagsLocation = "../testdata/controller/config_flags.yaml"

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
			name: "No user key in payload - should try to evaluate flags individually",
			// the API should return targeting key missing for all flags that require bucketing
			// and it should also perform all the static evaluations for non-bucketing flags
			args: args{
				bodyFile:            "../testdata/controller/all_flags/no_user_key_request.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				handlerErr: false,
				httpCode:   http.StatusOK,
				bodyFile:   "../testdata/controller/all_flags/no_user_key_response_updated.json",
			},
		},
		{
			name: "specify flags in evaluation context",
			args: args{
				bodyFile:            "../testdata/controller/all_flags/valid_request_specify_flags.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/controller/all_flags/valid_response_specify_flags.json",
			},
		},
		{
			name: "user context without custom field should not crash",
			args: args{
				bodyFile:            "../testdata/controller/all_flags/invalid_user_context_request.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/controller/all_flags/invalid_user_context_response.json",
			},
		},
		{
			name: "evaluation context without custom field should not crash",
			args: args{
				bodyFile:            "../testdata/controller/all_flags/invalid_evaluation_context_request.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/controller/all_flags/invalid_evaluation_context_response.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: retrieverconf.FileRetriever,
						Path: tt.args.configFlagsLocation,
					},
					Exporter: &config.ExporterConf{
						Kind: config.LogExporter,
					},
				},
			}
			flagsetManager, err := service.NewFlagsetManager(&conf, zap.NewNop(), []notifier.Notifier{})
			assert.NoError(t, err, "impossible to create flagset manager")

			ctrl := controller.NewAllFlags(flagsetManager, metric.Metrics{})

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

func Test_all_flag_Handler_FlagsetMode(t *testing.T) {
	const configFlagsLocation = "../testdata/controller/config_flags.yaml"

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
			name: "No user key in payload - should now pass and evaluate flags individually",
			// the API should return targeting key missing for all flags that require bucketing
			// and it should also perform all the static evaluations for non-bucketing flags
			args: args{
				bodyFile:            "../testdata/controller/all_flags/no_user_key_request.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				handlerErr: false, // No longer fails at API validation level
				httpCode:   http.StatusOK,
				bodyFile:   "../testdata/controller/all_flags/no_user_key_response_updated.json",
			},
		},
		{
			name: "specify flags in evaluation context",
			args: args{
				bodyFile:            "../testdata/controller/all_flags/valid_request_specify_flags.json",
				configFlagsLocation: configFlagsLocation,
			},
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/controller/all_flags/valid_response_specify_flags.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := config.Config{
				FlagSets: []config.FlagSet{
					{
						APIKeys: []string{"test-api-key"},
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: retrieverconf.FileRetriever,
								Path: tt.args.configFlagsLocation,
							},
							Exporter: &config.ExporterConf{
								Kind: config.LogExporter,
							},
						}},
				},
			}
			flagsetManager, err := service.NewFlagsetManager(&conf, zap.NewNop(), []notifier.Notifier{})
			assert.NoError(t, err, "impossible to create flagset manager")

			ctrl := controller.NewAllFlags(flagsetManager, metric.Metrics{})

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
			req.Header.Set(echo.HeaderAuthorization, "Bearer test-api-key")
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
