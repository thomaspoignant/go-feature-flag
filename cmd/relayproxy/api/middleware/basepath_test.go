package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestBasePathMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		basePath     string
		requestPath  string
		expectedPath string
	}{
		{
			name:         "Strip single level base path",
			basePath:     "/api",
			requestPath:  "/api/health",
			expectedPath: "/health",
		},
		{
			name:         "Strip multi-level base path",
			basePath:     "/api/feature-flags",
			requestPath:  "/api/feature-flags/health",
			expectedPath: "/health",
		},
		{
			name:         "Strip base path with nested endpoint",
			basePath:     "/api/feature-flags",
			requestPath:  "/api/feature-flags/v1/allflags",
			expectedPath: "/v1/allflags",
		},
		{
			name:         "No stripping when path doesn't match base path",
			basePath:     "/api",
			requestPath:  "/health",
			expectedPath: "/health",
		},
		{
			name:         "Empty base path should not modify path",
			basePath:     "",
			requestPath:  "/health",
			expectedPath: "/health",
		},
		{
			name:         "Base path without leading slash should work",
			basePath:     "api",
			requestPath:  "/api/health",
			expectedPath: "/health",
		},
		{
			name:         "Base path with trailing slash should work",
			basePath:     "/api/",
			requestPath:  "/api/health",
			expectedPath: "/health",
		},
		{
			name:         "Root path after stripping base path",
			basePath:     "/api",
			requestPath:  "/api",
			expectedPath: "/",
		},
		{
			name:         "Root path after stripping base path with trailing slash",
			basePath:     "/api",
			requestPath:  "/api/",
			expectedPath: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			// Create a handler that captures the processed path
			var processedPath string
			handler := func(c echo.Context) error {
				processedPath = c.Request().URL.Path
				return c.String(http.StatusOK, "OK")
			}

			// Create the middleware
			middleware := BasePathMiddleware(BasePathConfig{
				BasePath: tt.basePath,
			})

			// Create request
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute middleware + handler
			err := middleware(handler)(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPath, processedPath)
		})
	}
}

func TestBasePathMiddleware_Skipper(t *testing.T) {
	e := echo.New()

	var processedPath string
	handler := func(c echo.Context) error {
		processedPath = c.Request().URL.Path
		return c.String(http.StatusOK, "OK")
	}

	// Create middleware with skipper that skips all requests
	middleware := BasePathMiddleware(BasePathConfig{
		BasePath: "/api",
		Skipper: func(c echo.Context) bool {
			return true // Skip all requests
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := middleware(handler)(c)

	assert.NoError(t, err)
	// Path should not be modified when skipped
	assert.Equal(t, "/api/health", processedPath)
}
