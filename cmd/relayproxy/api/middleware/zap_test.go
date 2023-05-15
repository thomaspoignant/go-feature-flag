package middleware_test

import (
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestZapLogger200(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	}

	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	err := middleware.ZapLogger(logger, &config.Config{})(h)(c)
	assert.Nil(t, err)
	logFields := logs.AllUntimed()[0].ContextMap()
	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, int64(200), logFields["status"])
	assert.NotNil(t, logFields["latency"])
	assert.Equal(t, "GET /something", logFields["request"])
	assert.NotNil(t, logFields["host"])
	assert.NotNil(t, logFields["size"])
}

func TestZapLogger300(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusFound, "")
	}

	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	err := middleware.ZapLogger(logger, &config.Config{})(h)(c)
	assert.Nil(t, err)
	logFields := logs.AllUntimed()[0].ContextMap()
	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, int64(302), logFields["status"])
	assert.NotNil(t, logFields["latency"])
	assert.Equal(t, "GET /something", logFields["request"])
	assert.NotNil(t, logFields["host"])
	assert.NotNil(t, logFields["size"])
}
func TestZapLogger400(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusBadRequest, "")
	}

	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	err := middleware.ZapLogger(logger, &config.Config{})(h)(c)
	assert.Nil(t, err)
	logFields := logs.AllUntimed()[0].ContextMap()
	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, int64(400), logFields["status"])
	assert.NotNil(t, logFields["latency"])
	assert.Equal(t, "GET /something", logFields["request"])
	assert.NotNil(t, logFields["host"])
	assert.NotNil(t, logFields["size"])
}

func TestZapLogger500(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusInternalServerError, "")
	}

	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	err := middleware.ZapLogger(logger, &config.Config{})(h)(c)
	assert.Nil(t, err)
	logFields := logs.AllUntimed()[0].ContextMap()
	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, int64(500), logFields["status"])
	assert.NotNil(t, logFields["latency"])
	assert.Equal(t, "GET /something", logFields["request"])
	assert.NotNil(t, logFields["host"])
	assert.NotNil(t, logFields["size"])
}

func TestZapLoggerHealth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusInternalServerError, "")
	}

	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	err := middleware.ZapLogger(logger, &config.Config{})(h)(c)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(logs.AllUntimed()))
}

func TestZapLoggerHealthDebug(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusInternalServerError, "")
	}

	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	err := middleware.ZapLogger(logger, &config.Config{Debug: true})(h)(c)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(logs.AllUntimed()))
}
