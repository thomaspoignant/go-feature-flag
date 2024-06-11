package middleware_test

import (
	"github.com/labstack/echo/v4"
	middleware2 "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
	middleware := middleware2.VersionHeader(conf)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "Authorized")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", rec.Header().Get("X-GOFEATUREFLAG-VERSION"))
}
