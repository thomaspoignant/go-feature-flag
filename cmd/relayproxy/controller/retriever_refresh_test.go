package controller_test

import (
	"net/http/httptest"
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

func Test_retriever_refresh_Handler_no_goff(t *testing.T) {
	ctrl := controller.NewForceFlagsRefresh(nil, metric.Metrics{})
	e := echo.New()
	rec := httptest.NewRecorder()

	req := httptest.NewRequest(echo.POST, "/admin/v1/retriever/refresh", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)
	handlerErr := ctrl.Handler(c)
	assert.Error(t, handlerErr)
	assert.Equal(
		t,
		"code=500, message=flagset manager is not initialized",
		handlerErr.Error(),
	)
}

func Test_retriever_refresh_Handler_valid(t *testing.T) {
	// Create config for default mode
	conf := config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retriever: &retrieverconf.RetrieverConf{
				Kind: retrieverconf.FileRetriever,
				Path: "../../../testdata/flag-config.yaml",
			},
		},
	}

	flagsetManager, err := service.NewFlagsetManager(&conf, zap.NewNop(), []notifier.Notifier{})
	assert.NoError(t, err, "impossible to create flagset manager")

	// Get the default flagset to check refresh date
	defaultFlagset := flagsetManager.Default()
	previousRefresh := defaultFlagset.GetCacheRefreshDate()

	ctrl := controller.NewForceFlagsRefresh(flagsetManager, metric.Metrics{})
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(echo.POST, "/admin/v1/retriever/refresh", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)
	handlerErr := ctrl.Handler(c)
	assert.NoError(t, handlerErr)

	assert.NotEqual(t, previousRefresh, defaultFlagset.GetCacheRefreshDate())
}
