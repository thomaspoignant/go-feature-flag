package ofrep_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/ofrep"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"go.uber.org/zap"
)

// FuzzOFREPEvaluate checks that any request body sent to the OFREP endpoints
// gets a JSON answer that is not a server error, and never panics the handler.
func FuzzOFREPEvaluate(f *testing.F) {
	requests, err := filepath.Glob(testdataDir + "/ofrep/*.json")
	require.NoError(f, err)
	for _, file := range requests {
		content, err := os.ReadFile(file)
		require.NoError(f, err)
		f.Add("number-flag", string(content), false)
		f.Add("", string(content), true)
	}

	flagsetManager, err := service.NewFlagsetManager(&config.Config{
		CommonFlagSet: config.CommonFlagSet{
			PollingInterval: 10000,
			FileFormat:      "yaml",
			Retrievers: &[]retrieverconf.RetrieverConf{
				{Kind: retrieverconf.FileRetriever, Path: configFlagsLocation},
			},
		},
	}, zap.NewNop(), nil, nil)
	require.NoError(f, err)
	f.Cleanup(flagsetManager.Close)

	ctrl := ofrep.NewOFREPEvaluate(flagsetManager, metric.Metrics{}, config.OfrepEventStream{})
	e := echo.New()
	e.POST("/ofrep/v1/evaluate/flags/:flagKey", ctrl.Evaluate)
	e.POST("/ofrep/v1/evaluate/flags", ctrl.BulkEvaluate)

	f.Fuzz(func(t *testing.T, flagKey, body string, bulk bool) {
		target := "/ofrep/v1/evaluate/flags/" + url.PathEscape(flagKey)
		if bulk {
			target = "/ofrep/v1/evaluate/flags"
		}
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Less(t, rec.Code, http.StatusInternalServerError, "body: %s", rec.Body.String())
		assert.True(t, json.Valid(rec.Body.Bytes()), "response is not JSON: %s", rec.Body.String())
	})
}
