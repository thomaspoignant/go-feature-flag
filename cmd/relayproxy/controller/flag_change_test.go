package controller_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

func TestPIFlagChange_WithConfigChange(t *testing.T) {
	tests := []struct {
		name   string
		config config.Config
		apiKey string
	}{
		{
			name: "default mode",
			config: config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: "../testdata/controller/config_flags.yaml",
						},
					},
				},
			},
		},
		{
			name: "flagset mode",
			config: config.Config{
				FlagSets: []config.FlagSet{
					{
						APIKeys: []string{"test"},
						CommonFlagSet: config.CommonFlagSet{
							PollingInterval: 1000,
							Retrievers: &[]retrieverconf.RetrieverConf{
								{
									Kind: "file",
									Path: "../testdata/controller/config_flags.yaml",
								},
							},
						},
					},
				},
			},
			apiKey: "test",
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "")
			assert.NoError(t, err)
			defer func() {
				_ = file.Close()
				_ = os.Remove(file.Name())
			}()

			content, err := os.ReadFile("../testdata/controller/config_flags.yaml")
			assert.NoError(t, err)

			errWF := os.WriteFile(file.Name(), content, 0644)
			assert.NoError(t, errWF)
			file.Close()

			// Update the config to use the temp file
			if tt.config.Retrievers != nil {
				(*tt.config.Retrievers)[0].Path = file.Name()
			}
			if len(tt.config.FlagSets) > 0 {
				tt.config.FlagSets[0].Retrievers = &[]retrieverconf.RetrieverConf{
					{
						Kind: "file",
						Path: file.Name(),
					},
				}
			}

			flagsetManager, err := service.NewFlagsetManager(&tt.config, zap.NewNop(), []notifier.Notifier{})
			assert.NoError(t, err)
			defer flagsetManager.Close()

			ctrl := controller.NewAPIFlagChange(flagsetManager, metric.Metrics{})

			e := echo.New()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(echo.GET, "/v1/flag/change", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			if tt.apiKey != "" {
				req.Header.Set("Authorization", "Bearer "+tt.apiKey)
			}
			c := e.NewContext(req, rec)
			c.SetPath("/v1/flag/change")
			handlerErr := ctrl.Handler(c)
			assert.NoError(t, handlerErr)

			want, _ := os.ReadFile("../testdata/controller/flag_change/flag_change_with_config_change.json")
			assert.JSONEq(t, string(want), rec.Body.String())
			assert.Equal(t, http.StatusOK, rec.Code)

			content, err = os.ReadFile("../testdata/controller/config_flags_v2.yaml")
			assert.NoError(t, err)

			errWF = os.WriteFile(file.Name(), content, 0644)
			assert.NoError(t, errWF)

			time.Sleep(1500 * time.Millisecond)

			rec2 := httptest.NewRecorder()
			c2 := e.NewContext(req, rec2)
			c2.SetPath("/v1/flag/change")
			handlerErr2 := ctrl.Handler(c2)
			assert.NoError(t, handlerErr2)
			assert.NotEqual(t, want, rec2.Body.String())
		})
	}
}

func TestPIFlagChange_WithoutConfigChange(t *testing.T) {
	tests := []struct {
		name   string
		config config.Config
		apiKey string
	}{
		{
			name: "default mode",
			config: config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: "../testdata/controller/config_flags.yaml",
						},
					},
				},
			},
		},
		{
			name: "flagset mode",
			config: config.Config{
				FlagSets: []config.FlagSet{
					{
						APIKeys: []string{"test"},
						CommonFlagSet: config.CommonFlagSet{
							PollingInterval: 1000,
							Retrievers: &[]retrieverconf.RetrieverConf{
								{
									Kind: "file",
									Path: "../testdata/controller/config_flags.yaml",
								},
							},
						},
					},
				},
			},
			apiKey: "test",
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "")
			assert.NoError(t, err)
			defer func() {
				_ = file.Close()
				_ = os.Remove(file.Name())
			}()

			content, err := os.ReadFile("../testdata/controller/config_flags.yaml")
			assert.NoError(t, err)

			errWF := os.WriteFile(file.Name(), content, 0644)
			assert.NoError(t, errWF)
			file.Close()

			// Update the config to use the temp file
			if tt.config.Retrievers != nil {
				(*tt.config.Retrievers)[0].Path = file.Name()
			}
			if len(tt.config.FlagSets) > 0 {
				tt.config.FlagSets[0].Retrievers = &[]retrieverconf.RetrieverConf{
					{
						Kind: "file",
						Path: file.Name(),
					},
				}
			}

			flagsetManager, err := service.NewFlagsetManager(&tt.config, zap.NewNop(), []notifier.Notifier{})
			assert.NoError(t, err)
			defer flagsetManager.Close()

			ctrl := controller.NewAPIFlagChange(flagsetManager, metric.Metrics{})

			e := echo.New()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(echo.GET, "/v1/flag/change", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			if tt.apiKey != "" {
				req.Header.Set("Authorization", "Bearer "+tt.apiKey)
			}
			c := e.NewContext(req, rec)
			c.SetPath("/v1/flag/change")
			handlerErr := ctrl.Handler(c)
			assert.NoError(t, handlerErr)

			want, _ := os.ReadFile(
				"../testdata/controller/flag_change/flag_change_without_config_change.json",
			)
			assert.JSONEq(t, string(want), rec.Body.String())
			assert.Equal(t, http.StatusOK, rec.Code)

			time.Sleep(1500 * time.Millisecond)

			rec2 := httptest.NewRecorder()
			c2 := e.NewContext(req, rec2)
			c2.SetPath("/v1/flag/change")
			handlerErr2 := ctrl.Handler(c2)
			assert.NoError(t, handlerErr2)
			assert.JSONEq(t, string(want), rec2.Body.String())
		})
	}
}
