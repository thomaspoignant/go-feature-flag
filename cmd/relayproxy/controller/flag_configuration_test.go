package controller_test

import (
	"net/http"
	"net/http/httptest"
	"os"
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

const mockConfigFlagsLocation = "../testdata/controller/configuration/"

func TestFlagConfigurationAPICtrl_Handler_DefaultMode(t *testing.T) {
	type want struct {
		bodyLocation string
		statusCode   int
	}
	test := []struct {
		name        string
		requestBody string
		want        want
	}{
		{
			name:        "Test with empty body",
			requestBody: mockConfigFlagsLocation + "requests/empty.json",
			want: want{
				statusCode:   http.StatusOK,
				bodyLocation: mockConfigFlagsLocation + "responses/empty.json",
			},
		},
		{
			name:        "Test with empty flags ",
			requestBody: mockConfigFlagsLocation + "requests/empty-flag-array.json",
			want: want{
				statusCode:   http.StatusOK,
				bodyLocation: mockConfigFlagsLocation + "responses/empty-flag-array.json",
			},
		},
		{
			name:        "Filter flags",
			requestBody: mockConfigFlagsLocation + "requests/filter-flags.json",
			want: want{
				statusCode:   http.StatusOK,
				bodyLocation: mockConfigFlagsLocation + "responses/filter-flags.json",
			},
		},
		{
			name:        "Invalid JSON",
			requestBody: mockConfigFlagsLocation + "requests/invalid-json.json",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			// Create config for default mode (no flagsets)
			conf := config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: retrieverconf.FileRetriever,
						Path: "../testdata/controller/configuration_flags.yaml",
					},
					Exporter: &config.ExporterConf{
						Kind: config.LogExporter,
					},
				},
			}

			flagsetManager, err := service.NewFlagsetManager(&conf, zap.NewNop(), []notifier.Notifier{})
			assert.NoError(t, err, "impossible to create flagset manager")

			ctrl := controller.NewAPIFlagConfiguration(flagsetManager, metric.Metrics{})
			e := echo.New()
			rec := httptest.NewRecorder()

			// read the request body from the file
			requestBody, err := os.ReadFile(tt.requestBody)
			assert.NoError(t, err)

			req := httptest.NewRequest(echo.POST, "/v1/flag/configuration", strings.NewReader(string(requestBody)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)
			c.SetPath("/v1/flag/configuration")

			// Call the handler
			assert.NoError(t, ctrl.Handler(c))

			assert.Equal(t, tt.want.statusCode, rec.Code)

			if tt.want.bodyLocation != "" {
				wantBody, err := os.ReadFile(tt.want.bodyLocation)
				assert.NoError(t, err)
				assert.JSONEq(t, string(wantBody), rec.Body.String())
			}
		})
	}
}

func TestFlagConfigurationAPICtrl_Handler_FlagsetMode(t *testing.T) {
	const configFlagsLocation = "../testdata/controller/configuration_flags.yaml"
	const configFlagsLocation2 = "../testdata/controller/config_flags_v2.yaml"

	type want struct {
		bodyLocation string
		statusCode   int
	}
	test := []struct {
		name        string
		requestBody string
		apiKey      string
		want        want
	}{
		{
			name:        "Test with empty body for flagset1",
			requestBody: mockConfigFlagsLocation + "requests/empty.json",
			apiKey:      "flagset1-api-key",
			want: want{
				statusCode:   http.StatusOK,
				bodyLocation: mockConfigFlagsLocation + "responses/empty.json",
			},
		},
		{
			name:        "Test with empty body for flagset2",
			requestBody: mockConfigFlagsLocation + "requests/empty.json",
			apiKey:      "flagset2-api-key",
			want: want{
				statusCode:   http.StatusOK,
				bodyLocation: mockConfigFlagsLocation + "responses/empty2.json",
			},
		},
		{
			name:        "Filter flags for flagset1",
			requestBody: mockConfigFlagsLocation + "requests/filter-flags.json",
			apiKey:      "flagset1-api-key",
			want: want{
				statusCode:   http.StatusOK,
				bodyLocation: mockConfigFlagsLocation + "responses/filter-flags.json",
			},
		},
		{
			name:        "Invalid JSON",
			requestBody: mockConfigFlagsLocation + "requests/invalid-json.json",
			apiKey:      "flagset1-api-key",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:        "API key not linked to a flagset",
			requestBody: mockConfigFlagsLocation + "requests/empty.json",
			apiKey:      "invalid-api-key",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			// Create config for flagset mode
			conf := config.Config{
				FlagSets: []config.FlagSet{
					{
						Name: "flagset1",
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: retrieverconf.FileRetriever,
								Path: configFlagsLocation,
							},
							Exporter: &config.ExporterConf{
								Kind: config.LogExporter,
							},
						},
						APIKeys: []string{"flagset1-api-key"},
					},
					{
						Name: "flagset2",
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: retrieverconf.FileRetriever,
								Path: configFlagsLocation2,
							},
							Exporter: &config.ExporterConf{
								Kind: config.LogExporter,
							},
						},
						APIKeys: []string{"flagset2-api-key"},
					},
				},
			}

			flagsetManager, err := service.NewFlagsetManager(&conf, zap.NewNop(), []notifier.Notifier{})
			assert.NoError(t, err, "impossible to create flagset manager")

			ctrl := controller.NewAPIFlagConfiguration(flagsetManager, metric.Metrics{})
			e := echo.New()
			rec := httptest.NewRecorder()

			// read the request body from the file
			requestBody, err := os.ReadFile(tt.requestBody)
			assert.NoError(t, err)

			req := httptest.NewRequest(echo.POST, "/v1/flag/configuration", strings.NewReader(string(requestBody)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			// Add API key to header if provided
			if tt.apiKey != "" {
				req.Header.Set(echo.HeaderAuthorization, "Bearer "+tt.apiKey)
			}

			c := e.NewContext(req, rec)
			c.SetPath("/v1/flag/configuration")

			// Call the handler
			err = ctrl.Handler(c)

			if tt.want.statusCode == http.StatusBadRequest && tt.apiKey == "invalid-api-key" {
				assert.Error(t, err, "handler should return an error for invalid API key")
				he, ok := err.(*echo.HTTPError)
				if ok {
					assert.Equal(t, tt.want.statusCode, he.Code)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, rec.Code)

			if tt.want.bodyLocation != "" {
				wantBody, err := os.ReadFile(tt.want.bodyLocation)
				assert.NoError(t, err)
				assert.JSONEq(t, string(wantBody), rec.Body.String())
			}
		})
	}
}
