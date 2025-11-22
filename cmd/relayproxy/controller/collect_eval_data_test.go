package controller_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
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
		name   string
		args   args
		want   want
		config config.Config
		apiKey string
	}{
		{
			name: "valid usecase",
			config: config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 10,
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: configFlagsLocation,
						},
					},
				},
			},
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
			name:   "valid usecase flagset",
			apiKey: "test",
			config: config.Config{
				FlagSets: []config.FlagSet{
					{
						APIKeys: []string{"test"},
						CommonFlagSet: config.CommonFlagSet{
							PollingInterval: 10,
							Retrievers: &[]retrieverconf.RetrieverConf{
								{
									Kind: "file",
									Path: configFlagsLocation,
								},
							},
						},
					},
				},
			},
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
			config: config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 10,
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: configFlagsLocation,
						},
					},
				},
			},
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
			name:   "valid with source field flagset",
			apiKey: "test",
			config: config.Config{
				FlagSets: []config.FlagSet{
					{
						APIKeys: []string{"test"},
						CommonFlagSet: config.CommonFlagSet{
							PollingInterval: 10,
							Retrievers: &[]retrieverconf.RetrieverConf{
								{
									Kind: "file",
									Path: configFlagsLocation,
								},
							},
						},
					},
				},
			},
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
			config: config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 10,
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: configFlagsLocation,
						},
					},
				},
			},
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
			config: config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 10,
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: configFlagsLocation,
						},
					},
				},
			},
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
		{
			name: "be sure that the creation date is a unix timestamp",
			config: config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 10,
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: configFlagsLocation,
						},
					},
				},
			},
			args: args{
				"../testdata/controller/collect_eval_data/valid_request_with_timestamp_ms.json",
			},
			want: want{
				httpCode:          http.StatusOK,
				bodyFile:          "../testdata/controller/collect_eval_data/valid_response.json",
				collectedDataFile: "../testdata/controller/collect_eval_data/valid_collected_data_with_timestamp_ms.json",
			},
		},
		{
			name: "should have the metadata in the exporter",
			config: config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 10,
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: configFlagsLocation,
						},
					},
				},
			},
			args: args{
				"../testdata/controller/collect_eval_data/valid_request_metadata.json",
			},
			want: want{
				httpCode:          http.StatusOK,
				bodyFile:          "../testdata/controller/collect_eval_data/valid_response_metadata.json",
				collectedDataFile: "../testdata/controller/collect_eval_data/valid_collected_data_metadata.json",
			},
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			exporterFile, err := os.CreateTemp("", "exporter.json")
			assert.NoError(t, err)
			defer os.Remove(exporterFile.Name())

			if len(tt.config.FlagSets) > 0 {
				out := make([]config.FlagSet, len(tt.config.FlagSets))
				for index, flagSet := range tt.config.FlagSets {
					flagSet.Exporters = &[]config.ExporterConf{
						{
							Kind:     "file",
							Filename: exporterFile.Name(),
						},
					}
					out[index] = flagSet
				}
				tt.config.FlagSets = out
			} else {
				tt.config.Exporters = &[]config.ExporterConf{
					{
						Kind:     "file",
						Filename: exporterFile.Name(),
					},
				}
			}

			flagsetManager, err := service.NewFlagsetManager(&tt.config, zap.NewNop(), []notifier.Notifier{})
			assert.NoError(t, err)

			logger, err := zap.NewDevelopment()
			require.NoError(t, err)
			ctrl := controller.NewCollectEvalData(flagsetManager, metric.Metrics{}, logger)

			e := echo.New()
			rec := httptest.NewRecorder()

			// read wantBody request file
			var bodyReq io.Reader
			if tt.args.bodyFile != "" {
				bodyReqContent, err := os.ReadFile(tt.args.bodyFile)
				assert.NoError(t, err, "request wantBody file missing %s", tt.args.bodyFile)
				bodyReq = strings.NewReader(string(bodyReqContent))
			}

			req := httptest.NewRequest(echo.POST, "/v1/data/collector", bodyReq)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			if tt.apiKey != "" {
				req.Header.Set("Authorization", "Bearer "+tt.apiKey)
			}
			c := e.NewContext(req, rec)
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
			flagsetManager.Close()
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
			assert.JSONEq(t, string(wantCollectData), string(exportedData), "Invalid exported data")
		})
	}
}

func TestCollectEvalData_Handler_cancellation(t *testing.T) {
	flagsetManager, err := service.NewFlagsetManager(&config.Config{
		CommonFlagSet: config.CommonFlagSet{
			PollingInterval: 10,
			Retrievers:      &[]retrieverconf.RetrieverConf{{Kind: "file", Path: configFlagsLocation}},
		},
	}, zap.NewNop(), []notifier.Notifier{})
	require.NoError(t, err)

	t.Cleanup(func() { flagsetManager.Close() })

	ctrl := controller.NewCollectEvalData(flagsetManager, metric.Metrics{}, zap.NewNop())

	// large payload with 20,000 events to ensure processing takes time
	event := `{"kind":"feature","contextKind":"user","userKey":"u","creationDate":1680246000,"key":"f","variation":"v","value":"true","default":false,"version":"1","source":"PROVIDER_CACHE"}`
	body := `{"events":[` + strings.Repeat(event+",", 19999) + event + `]}`

	ctx, cancel := context.WithCancel(t.Context())
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/v1/data/collector", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	e := echo.New()
	c := e.NewContext(req, httptest.NewRecorder())

	// run handler in background, cancel after 10ms
	done := make(chan error, 1)
	go func() {
		done <- ctrl.Handler(c)
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	err = <-done

	require.ErrorIs(t, err, context.Canceled)
	assert.Contains(t, err.Error(), "context cancelled after processing")
	// count will be less than the total if cancellation worked
	assert.NotContains(t, err.Error(), "20000/20000")
}

func Test_collect_tracking_and_evaluation_events(t *testing.T) {
	tests := []struct {
		name   string
		config config.Config
		apiKey string
	}{
		{
			name: "default mode",
			config: config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 10,
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: configFlagsLocation,
						},
					},
				},
			},
		},
		{
			name:   "flagset mode",
			apiKey: "test",
			config: config.Config{
				FlagSets: []config.FlagSet{
					{
						APIKeys: []string{"test"},
						CommonFlagSet: config.CommonFlagSet{
							PollingInterval: 10,
							Retrievers: &[]retrieverconf.RetrieverConf{
								{
									Kind: "file",
									Path: configFlagsLocation,
								},
							},
						},
					},
				},
			},
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			evalExporter, err := os.CreateTemp("", "evalExport.json")
			assert.NoError(t, err)
			trackingExporter, err := os.CreateTemp("", "trackExport.json")
			assert.NoError(t, err)
			defer func() {
				_ = os.Remove(evalExporter.Name())
				_ = os.Remove(trackingExporter.Name())
			}()

			if len(tt.config.FlagSets) > 0 {
				out := make([]config.FlagSet, len(tt.config.FlagSets))

				for index, flagSet := range tt.config.FlagSets {
					flagSet.Exporters = &[]config.ExporterConf{
						{
							Kind:             config.FileExporter,
							Filename:         evalExporter.Name(),
							MaxEventInMemory: 10000,
							FlushInterval:    int64(10 * time.Second),
						},
						{
							Kind:              config.FileExporter,
							Filename:          trackingExporter.Name(),
							ExporterEventType: ffclient.TrackingEventExporter,
							MaxEventInMemory:  10000,
							FlushInterval:     int64(10 * time.Second),
						},
					}
					out[index] = flagSet
				}
				tt.config.FlagSets = out
			} else {
				tt.config.Exporters = &[]config.ExporterConf{
					{
						Kind:             "file",
						Filename:         evalExporter.Name(),
						MaxEventInMemory: 10000,
						FlushInterval:    int64(10 * time.Second),
					},
					{
						Kind:              "file",
						Filename:          trackingExporter.Name(),
						ExporterEventType: ffclient.TrackingEventExporter,
						MaxEventInMemory:  10000,
						FlushInterval:     int64(10 * time.Second),
					},
				}
			}

			flagsetManager, err := service.NewFlagsetManager(&tt.config, zap.NewNop(), []notifier.Notifier{})
			assert.NoError(t, err)
			ctrl := controller.NewCollectEvalData(flagsetManager, metric.Metrics{}, zap.NewNop())
			bodyReq, err := os.ReadFile(
				"../testdata/controller/collect_eval_data/valid_request_mix_tracking_evaluation.json")
			assert.NoError(t, err)
			e := echo.New()
			rec := httptest.NewRecorder()

			req := httptest.NewRequest(echo.POST, "/v1/data/collector", strings.NewReader(string(bodyReq)))
			if tt.apiKey != "" {
				req.Header.Set("Authorization", "Bearer "+tt.apiKey)
			}
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)
			c.SetPath("/v1/data/collector")
			handlerErr := ctrl.Handler(c)
			assert.NoError(t, handlerErr)
			flagsetManager.Close()
			evalEvents, err := os.ReadFile(evalExporter.Name())
			assert.NoError(t, err)
			want := "{\"kind\":\"feature\",\"contextKind\":\"user\",\"userKey\":\"94a25909-20d8-40cc-8500-fee99b569345\",\"creationDate\":1680246000,\"key\":\"my-feature-flag\",\"variation\":\"admin-variation\",\"value\":\"string\",\"default\":false,\"version\":\"v1.0.0\",\"source\":\"PROVIDER_CACHE\",\"metadata\":{\"environment\":\"production\",\"sdkVersion\":\"v1.0.0\",\"source\":\"my-source\",\"timestamp\":1680246000}}\n"
			assert.JSONEq(t, want, string(evalEvents), "Invalid exported data")
			wantTracking := "{\"kind\":\"tracking\",\"contextKind\":\"user\",\"userKey\":\"94a25909-20d8-40cc-8500-fee99b569345\",\"creationDate\":1680246020,\"key\":\"my-feature-flag\",\"evaluationContext\":{\"admin\":true,\"name\":\"john doe\",\"targetingKey\":\"94a25909-20d8-40cc-8500-fee99b569345\"},\"trackingEventDetails\":{\"value\":\"string\",\"version\":\"v1.0.0\"}}\n"
			trackingEvents, err := os.ReadFile(trackingExporter.Name())
			assert.NoError(t, err)
			assert.JSONEq(t, wantTracking, string(trackingEvents), "Invalid exported data")
		})
	}
}
