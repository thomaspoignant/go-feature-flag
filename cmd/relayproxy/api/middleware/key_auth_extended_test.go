package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	middleware2 "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
)

func TestKeyAuthExtended(t *testing.T) {
	validKey := "valid-api-key"
	invalidKey := "invalid-api-key"

	// nolint:unparam
	validator := func(key string, _ echo.Context) (bool, error) {
		return key == validKey, nil
	}

	tests := []struct {
		name           string
		xAPIKey        string
		authorization  string
		skipper        bool
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "valid X-API-Key header",
			xAPIKey:        validKey,
			authorization:  "",
			skipper:        false,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid X-API-Key header",
			xAPIKey:        invalidKey,
			authorization:  "",
			skipper:        false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:           "valid Authorization header (fallback)",
			xAPIKey:        "",
			authorization:  "Bearer " + validKey,
			skipper:        false,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid Authorization header (fallback)",
			xAPIKey:        "",
			authorization:  "Bearer " + invalidKey,
			skipper:        false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:           "X-API-Key takes precedence over Authorization",
			xAPIKey:        validKey,
			authorization:  "Bearer " + invalidKey,
			skipper:        false,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "no headers",
			xAPIKey:        "",
			authorization:  "",
			skipper:        false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:           "skipper skips authentication",
			xAPIKey:        "",
			authorization:  "",
			skipper:        true,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			if tt.xAPIKey != "" {
				req.Header.Set("X-API-Key", tt.xAPIKey)
			}
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			skipper := func(c echo.Context) bool {
				return tt.skipper
			}

			middleware := middleware2.KeyAuthExtended(middleware2.KeyAuthExtendedConfig{
				Validator:    validator,
				ErrorHandler: middleware2.AuthMiddlewareErrHandler,
				Skipper:      skipper,
			})

			handler := middleware(func(c echo.Context) error {
				return c.String(http.StatusOK, "Authorized")
			})

			err := handler(c)

			if tt.expectedError {
				assert.Error(t, err)
				if httpErr, ok := err.(*echo.HTTPError); ok {
					assert.Equal(t, tt.expectedStatus, httpErr.Code)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestKeyAuthExtended_SetDefaults(t *testing.T) {
	validKey := "valid-api-key"

	// nolint:unparam
	validator := func(key string, _ echo.Context) (bool, error) {
		return key == validKey, nil
	}

	t.Run("panics when Validator is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "echo: key auth extended middleware requires a validator function", func() {
			middleware2.KeyAuthExtended(middleware2.KeyAuthExtendedConfig{
				Validator: nil,
			})
		})
	})

	t.Run("uses default ErrorHandler when nil", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Create middleware with nil ErrorHandler - should use default
		mw := middleware2.KeyAuthExtended(middleware2.KeyAuthExtendedConfig{
			Validator:    validator,
			ErrorHandler: nil, // Should use default
			Skipper:      nil,
			KeyLookup:    "",
		})

		handler := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "Authorized")
		})

		// Request without valid key should use default error handler
		// Default echo error handler returns 400 Bad Request
		err := handler(c)
		require.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		// Echo's default KeyAuth error handler returns 400, not 401
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	})

	t.Run("uses default Skipper when nil", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Create middleware with nil Skipper - should use default (which doesn't skip)
		mw := middleware2.KeyAuthExtended(middleware2.KeyAuthExtendedConfig{
			Validator:    validator,
			ErrorHandler: middleware2.AuthMiddlewareErrHandler,
			Skipper:      nil, // Should use default
			KeyLookup:    "",
		})

		handler := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "Authorized")
		})

		// Default skipper should not skip, so request without key should fail
		err := handler(c)
		require.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	})

	t.Run("uses default KeyLookup when empty", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validKey)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Create middleware with empty KeyLookup - should use default "header:Authorization"
		mw := middleware2.KeyAuthExtended(middleware2.KeyAuthExtendedConfig{
			Validator:    validator,
			ErrorHandler: middleware2.AuthMiddlewareErrHandler,
			Skipper:      nil,
			KeyLookup:    "", // Should use default "header:Authorization"
		})

		handler := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "Authorized")
		})

		// Default KeyLookup should work with Authorization header
		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("uses custom ErrorHandler when provided", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		customErrorHandler := func(err error, c echo.Context) error {
			return c.String(http.StatusForbidden, "Custom error")
		}

		mw := middleware2.KeyAuthExtended(middleware2.KeyAuthExtendedConfig{
			Validator:    validator,
			ErrorHandler: customErrorHandler,
			Skipper:      nil,
			KeyLookup:    "",
		})

		handler := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "Authorized")
		})

		// Should use custom error handler
		err := handler(c)
		require.NoError(t, err) // Custom handler returns nil error but sets status
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("uses custom Skipper when provided", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		customSkipper := func(c echo.Context) bool {
			return true // Always skip
		}

		mw := middleware2.KeyAuthExtended(middleware2.KeyAuthExtendedConfig{
			Validator:    validator,
			ErrorHandler: middleware2.AuthMiddlewareErrHandler,
			Skipper:      customSkipper,
			KeyLookup:    "",
		})

		handler := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "Authorized")
		})

		// Custom skipper should skip authentication
		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("uses custom KeyLookup when provided", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test?apiKey="+validKey, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Use query parameter instead of header
		mw := middleware2.KeyAuthExtended(middleware2.KeyAuthExtendedConfig{
			Validator:    validator,
			ErrorHandler: middleware2.AuthMiddlewareErrHandler,
			Skipper:      nil,
			KeyLookup:    "query:apiKey", // Custom lookup from query parameter
		})

		handler := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "Authorized")
		})

		// Custom KeyLookup should work with query parameter
		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("verifies default KeyLookup value matches echo default", func(t *testing.T) {
		// This test verifies that the default KeyLookup is "header:Authorization"
		// by checking that it matches echo's DefaultKeyAuthConfig.KeyLookup
		assert.Equal(t, middleware.DefaultKeyAuthConfig.KeyLookup, "header:Authorization")
	})
}
