package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/api/middleware"
)

const (
	etagBody    = "Hello World"
	weakEtag    = "W/\"11-8dcfee46\""
	strongEtag  = "\"11-0a4d55a8d778e5022fab701977c5d840bbc486d0\""
	invalidEtag = "invalid"
)

// newEtagServer builds an echo instance exposing the routes used by the Etag tests.
func newEtagServer() *echo.Echo {
	e := echo.New()

	e.GET("/etag", func(c *echo.Context) error {
		return c.String(http.StatusOK, etagBody)
	}, middleware.EtagWithConfig(middleware.EtagConfig{Weak: false}))

	e.GET("/etag/weak", func(c *echo.Context) error {
		return c.String(http.StatusOK, etagBody)
	}, middleware.Etag())

	e.GET("/etag/nocontent", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	}, middleware.Etag())

	e.GET("/etag/skipped", func(c *echo.Context) error {
		return c.String(http.StatusOK, etagBody)
	}, middleware.EtagWithConfig(middleware.EtagConfig{
		Skipper: func(_ *echo.Context) bool { return true },
	}))

	return e
}

func TestEtag(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		ifNoneMatch  string
		wantStatus   int
		wantEtag     string
		wantBody     string
		wantNoEtagHd bool
	}{
		{
			name:       "strong etag is set on first request",
			path:       "/etag",
			wantStatus: http.StatusOK,
			wantEtag:   strongEtag,
			wantBody:   etagBody,
		},
		{
			name:        "strong etag returns 304 when If-None-Match matches",
			path:        "/etag",
			ifNoneMatch: strongEtag,
			wantStatus:  http.StatusNotModified,
			wantEtag:    strongEtag,
			wantBody:    "",
		},
		{
			name:        "strong etag returns 200 when If-None-Match does not match",
			path:        "/etag",
			ifNoneMatch: invalidEtag,
			wantStatus:  http.StatusOK,
			wantEtag:    strongEtag,
			wantBody:    etagBody,
		},
		{
			name:       "weak etag is set on first request",
			path:       "/etag/weak",
			wantStatus: http.StatusOK,
			wantEtag:   weakEtag,
			wantBody:   etagBody,
		},
		{
			name:        "weak etag returns 304 when If-None-Match matches",
			path:        "/etag/weak",
			ifNoneMatch: weakEtag,
			wantStatus:  http.StatusNotModified,
			wantEtag:    weakEtag,
			wantBody:    "",
		},
		{
			name:        "weak etag returns 200 when If-None-Match does not match",
			path:        "/etag/weak",
			ifNoneMatch: invalidEtag,
			wantStatus:  http.StatusOK,
			wantEtag:    weakEtag,
			wantBody:    etagBody,
		},
		{
			name:         "no content response passes through without an etag",
			path:         "/etag/nocontent",
			wantStatus:   http.StatusNoContent,
			wantBody:     "",
			wantNoEtagHd: true,
		},
		{
			name:         "skipper bypasses the middleware entirely",
			path:         "/etag/skipped",
			wantStatus:   http.StatusOK,
			wantBody:     etagBody,
			wantNoEtagHd: true,
		},
	}

	e := newEtagServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.ifNoneMatch != "" {
				req.Header.Set("If-None-Match", tt.ifNoneMatch)
			}
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
			if tt.wantNoEtagHd {
				assert.Empty(t, rec.Header().Get("Etag"))
				return
			}
			assert.Equal(t, tt.wantEtag, rec.Header().Get("Etag"))
		})
	}
}
