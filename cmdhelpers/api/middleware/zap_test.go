package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/api/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestZapLogger200(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, "")
	}

	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	err := middleware.ZapLogger(logger, false)(h)(c)
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

	h := func(c *echo.Context) error {
		return c.String(http.StatusFound, "")
	}

	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	err := middleware.ZapLogger(logger, false)(h)(c)
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

	h := func(c *echo.Context) error {
		return c.String(http.StatusBadRequest, "")
	}

	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	err := middleware.ZapLogger(logger, false)(h)(c)
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

	h := func(c *echo.Context) error {
		return c.String(http.StatusInternalServerError, "")
	}

	obs, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(obs)
	err := middleware.ZapLogger(logger, false)(h)(c)
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

	h := func(c *echo.Context) error {
		return c.String(http.StatusInternalServerError, "")
	}

	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	err := middleware.ZapLogger(logger, false)(h)(c)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(logs.AllUntimed()))
}

func TestZapLoggerHealthDebug(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c *echo.Context) error {
		return c.String(http.StatusInternalServerError, "")
	}

	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	err := middleware.ZapLogger(logger, true)(h)(c)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(logs.AllUntimed()))
}

func TestZapLoggerRedactsAPIKey(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want string
	}{
		{
			name: "no query",
			uri:  "/ws/v1/flag/change",
			want: "GET /ws/v1/flag/change",
		},
		{
			name: "apiKey only",
			uri:  "/stream/v1/sse/flag/change?apiKey=secret",
			want: "GET /stream/v1/sse/flag/change?apiKey=REDACTED",
		},
		{
			name: "apiKey among other params",
			uri:  "/stream/v1/ws/flag/change?a=1&apiKey=secret&b=2",
			want: "GET /stream/v1/ws/flag/change?a=1&apiKey=REDACTED&b=2",
		},
		{
			name: "no apiKey",
			uri:  "/stream/v1/ws/flag/change?a=1&b=2",
			want: "GET /stream/v1/ws/flag/change?a=1&b=2",
		},
		{
			name: "percent-encoded apiKey name",
			uri:  "/ws/v1/flag/change?api%4Bey=secret",
			want: "GET /ws/v1/flag/change?api%4Bey=REDACTED",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, tt.uri, nil)
			c := e.NewContext(req, httptest.NewRecorder())
			h := func(c *echo.Context) error {
				return c.String(http.StatusUnauthorized, "")
			}

			obs, logs := observer.New(zap.DebugLevel)
			err := middleware.ZapLogger(zap.New(obs), false)(h)(c)
			assert.NoError(t, err)
			assert.Equal(t, 1, logs.Len())
			assert.Equal(t, tt.want, logs.AllUntimed()[0].ContextMap()["request"])
		})
	}
}
