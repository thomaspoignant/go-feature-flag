package manifest_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/manifest"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/helper"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

const (
	manifestTestdataDir        = "../../testdata/manifest"
	manifestConfigFlags        = manifestTestdataDir + "/config_flags.yaml"
	manifestConfigFlagsV2      = manifestTestdataDir + "/config_flags_v2.yaml"
	manifestSuccessResponse    = manifestTestdataDir + "/responses/valid_manifest.json"
	manifestSuccessResponseV2  = manifestTestdataDir + "/responses/valid_manifest_v2.json"
	manifestCapabilitiesHeader = "X-Manifest-Capabilities"
	manifestCapabilitiesValue  = "read"
	manifestEndpoint           = "/openfeature/v0/manifest"
)

func TestManifestCtrl_GetManifest_DefaultMode(t *testing.T) {
	conf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			PollingInterval: 10000,
			FileFormat:      "yaml",
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: retrieverconf.FileRetriever,
					Path: manifestConfigFlags,
				},
			},
		},
	}

	flagsetManager, err := service.NewFlagsetManager(conf, zap.NewNop(), []notifier.Notifier{})
	assert.NoError(t, err, "failed to create flagset manager")
	defer flagsetManager.Close()

	ctrl := manifest.NewManifest(flagsetManager, metric.Metrics{}, zap.NewNop())
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(echo.GET, manifestEndpoint, nil)
	c := e.NewContext(req, rec)
	c.SetPath(manifestEndpoint)

	assert.NoError(t, ctrl.GetManifest(c))
	assert.Equal(t, http.StatusOK, rec.Code, "Invalid HTTP Code")
	assert.Equal(
		t,
		manifestCapabilitiesValue,
		rec.Header().Get(manifestCapabilitiesHeader),
		"capabilities header missing or wrong",
	)
	assertManifestEqual(t, manifestSuccessResponse, rec.Body.Bytes())
}

func TestManifestCtrl_GetManifest_FlagsetMode(t *testing.T) {
	type want struct {
		httpCode         int
		bodyFile         string
		handlerErr       bool
		errorCode        int
		errorMsgContains string
	}

	tests := []struct {
		name   string
		apiKey string
		want   want
	}{
		{
			name:   "valid manifest for flagset1",
			apiKey: "flagset1-api-key",
			want: want{
				httpCode: http.StatusOK,
				bodyFile: manifestSuccessResponse,
			},
		},
		{
			name:   "valid manifest for flagset2",
			apiKey: "flagset2-api-key",
			want: want{
				httpCode: http.StatusOK,
				bodyFile: manifestSuccessResponseV2,
			},
		},
		{
			name:   "API key not linked to a flagset",
			apiKey: "invalid-api-key",
			want: want{
				handlerErr:       true,
				errorCode:        http.StatusBadRequest,
				errorMsgContains: "error while getting flagset",
			},
		},
		{
			name:   "missing API key",
			apiKey: "",
			want: want{
				handlerErr:       true,
				errorCode:        http.StatusBadRequest,
				errorMsgContains: "error while getting flagset",
			},
		},
	}

	conf := &config.Config{
		FlagSets: []config.FlagSet{
			{
				Name: "flagset1",
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 10000,
					FileFormat:      "yaml",
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: retrieverconf.FileRetriever,
							Path: manifestConfigFlags,
						},
					},
				},
				APIKeys: []string{"flagset1-api-key"},
			},
			{
				Name: "flagset2",
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 10000,
					FileFormat:      "yaml",
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: retrieverconf.FileRetriever,
							Path: manifestConfigFlagsV2,
						},
					},
				},
				APIKeys: []string{"flagset2-api-key"},
			},
		},
	}

	flagsetManager, err := service.NewFlagsetManager(conf, zap.NewNop(), []notifier.Notifier{})
	assert.NoError(t, err, "failed to create flagset manager")
	defer flagsetManager.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := manifest.NewManifest(flagsetManager, metric.Metrics{}, zap.NewNop())
			e := echo.New()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(echo.GET, manifestEndpoint, nil)
			if tt.apiKey != "" {
				req.Header.Set(helper.AuthorizationHeader, helper.BearerPrefix+tt.apiKey)
			}
			c := e.NewContext(req, rec)
			c.SetPath(manifestEndpoint)

			handlerErr := ctrl.GetManifest(c)

			if tt.want.handlerErr {
				assert.Error(t, handlerErr, "handler should return an error")
				he, ok := handlerErr.(*echo.HTTPError)
				assert.True(t, ok, "handler error should be an echo.HTTPError")
				assert.Equal(t, tt.want.errorCode, he.Code)
				assert.Contains(t, he.Message, tt.want.errorMsgContains)
				return
			}

			assert.NoError(t, handlerErr)
			assert.Equal(t, tt.want.httpCode, rec.Code, "Invalid HTTP Code")
			assert.Equal(
				t,
				manifestCapabilitiesValue,
				rec.Header().Get(manifestCapabilitiesHeader),
				"capabilities header missing or wrong",
			)

			if tt.want.bodyFile != "" {
				assertManifestEqual(t, tt.want.bodyFile, rec.Body.Bytes())
			}
		})
	}
}

func TestManifestCtrl_GetManifest_NilFlagsetManager(t *testing.T) {
	ctrl := manifest.NewManifest(nil, metric.Metrics{}, zap.NewNop())
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(echo.GET, manifestEndpoint, nil)
	c := e.NewContext(req, rec)
	c.SetPath(manifestEndpoint)

	handlerErr := ctrl.GetManifest(c)

	assert.Error(t, handlerErr, "handler should return an error when flagset manager is nil")
	he, ok := handlerErr.(*echo.HTTPError)
	assert.True(t, ok, "handler error should be an echo.HTTPError")
	assert.Equal(t, http.StatusInternalServerError, he.Code)
	assert.Equal(t, "flagset manager is not initialized", he.Message)
}

// assertManifestEqual deserializes the actual body and the expected file into
// ManifestSuccessResponse structs and compares them element-wise. This is needed
// because the manifest is built from a Go map, so the order of flags in the
// response is not deterministic.
func assertManifestEqual(t *testing.T, expectedFile string, actualBody []byte) {
	t.Helper()

	expectedRaw, err := os.ReadFile(expectedFile)
	assert.NoError(t, err, "Impossible to read expected body file %s", expectedFile)

	var expected model.ManifestSuccessResponse
	assert.NoError(t, json.Unmarshal(expectedRaw, &expected), "invalid expected manifest JSON")

	var actual model.ManifestSuccessResponse
	assert.NoError(t, json.Unmarshal(actualBody, &actual), "invalid actual manifest JSON")

	assert.ElementsMatch(t, expected.Flags, actual.Flags, "manifest flags mismatch")
}
