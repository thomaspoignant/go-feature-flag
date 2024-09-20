package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestBodyLoggerInDebug(t *testing.T) {
	e := echo.New()
	body := `{"context": {"custom":"label"}}`
	req := httptest.NewRequest(http.MethodPost, "/something", bytes.NewBuffer([]byte(body)))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	}

	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	err := middleware.BodyLogger(logger, &config.Config{LogLevel: "DEBUG"})(h)(c)
	assert.Nil(t, err)
	logFields := logs.AllUntimed()[0].ContextMap()
	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, body, logFields["body"])
}

func TestBodyLoggerNotInDebug(t *testing.T) {
	e := echo.New()
	body := `{"context": {"custom":"label"}}`
	req := httptest.NewRequest(http.MethodPost, "/something", bytes.NewBuffer([]byte(body)))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusFound, "")
	}

	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	err := middleware.BodyLogger(logger, &config.Config{})(h)(c)
	assert.Nil(t, err)
	assert.Equal(t, 0, logs.Len())
}
