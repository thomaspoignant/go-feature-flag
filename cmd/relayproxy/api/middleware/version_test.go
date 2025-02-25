package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	middleware2 "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

func TestVersion(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/whatever", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	conf := &config.Config{
		Version: "1.0.0",
	}
	middleware := middleware2.VersionHeader(middleware2.VersionHeaderConfig{
		RelayProxyConfig: conf,
	})
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "Authorized")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", rec.Header().Get("X-GOFEATUREFLAG-VERSION"))
}

func TestNoVersion(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/whatever", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	conf := &config.Config{
		Version:              "1.0.0",
		DisableVersionHeader: true,
	}
	middleware := middleware2.VersionHeader(middleware2.VersionHeaderConfig{
		Skipper: func(c echo.Context) bool {
			return conf.DisableVersionHeader
		},
		RelayProxyConfig: conf,
	})
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "Authorized")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Empty(t, rec.Header().Get("X-GOFEATUREFLAG-VERSION"))
}
