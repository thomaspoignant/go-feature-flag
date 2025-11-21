package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	middleware2 "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
)

func TestKeyAuthExtended(t *testing.T) {
	validKey := "valid-api-key"
	invalidKey := "invalid-api-key"

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

