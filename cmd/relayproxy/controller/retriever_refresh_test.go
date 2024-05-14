package controller_test

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_retriever_refresh_Handler_no_goff(t *testing.T) {
	ctrl := controller.NewForceFlagsRefresh(nil, metric.Metrics{})
	e := echo.New()
	rec := httptest.NewRecorder()

	req := httptest.NewRequest(echo.POST, "/admin/v1/retriver/refresh", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)
	handlerErr := ctrl.Handler(c)
	assert.Error(t, handlerErr)
	assert.Equal(t, "code=500, message=forceFlagsRefresh: goFF is not initialized", handlerErr.Error())
}

func Test_retriever_refresh_Handler_valid(t *testing.T) {
	gffClient, err := ffclient.New(ffclient.Config{
		PollingInterval: 15 * time.Minute,
		Retriever:       &fileretriever.Retriever{Path: "../../../testdata/flag-config.yaml"},
		Offline:         false,
	})
	assert.NoError(t, err)
	previousRefresh := gffClient.GetCacheRefreshDate()
	ctrl := controller.NewForceFlagsRefresh(gffClient, metric.Metrics{})
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(echo.POST, "/admin/v1/retriver/refresh", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)
	handlerErr := ctrl.Handler(c)
	assert.NoError(t, handlerErr)

	assert.NotEqual(t, previousRefresh, gffClient.GetCacheRefreshDate())
}
