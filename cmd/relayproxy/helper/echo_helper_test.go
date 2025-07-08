package helper_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/helper"
)

func TestGetAPIKey(t *testing.T) {
	tests := []struct {
		name           string
		authorization  string
		expectedAPIKey string
	}{
		{
			name:           "Bearer token",
			authorization:  "Bearer my-api-key-123",
			expectedAPIKey: "my-api-key-123",
		},
		{
			name:           "Bearer token with spaces",
			authorization:  "Bearer   my-api-key-with-spaces  ",
			expectedAPIKey: "  my-api-key-with-spaces  ",
		},
		{
			name:           "Basic auth",
			authorization:  "Basic dXNlcjpwYXNz",
			expectedAPIKey: "Basic dXNlcjpwYXNz",
		},
		{
			name:           "Custom auth scheme",
			authorization:  "CustomScheme my-custom-key",
			expectedAPIKey: "CustomScheme my-custom-key",
		},
		{
			name:           "Empty authorization header",
			authorization:  "",
			expectedAPIKey: "",
		},
		{
			name:           "Bearer prefix only",
			authorization:  "Bearer",
			expectedAPIKey: "Bearer",
		},
		{
			name:           "Bearer with empty token",
			authorization:  "Bearer ",
			expectedAPIKey: "Bearer ", // Function only strips "Bearer " if length > 7
		},
		{
			name:           "Short bearer token",
			authorization:  "Bearer abc",
			expectedAPIKey: "abc",
		},
		{
			name:           "No auth scheme",
			authorization:  "my-api-key-123",
			expectedAPIKey: "my-api-key-123",
		},
		{
			name:           "Bearer with special characters",
			authorization:  "Bearer my-api-key!@#$%^&*()",
			expectedAPIKey: "my-api-key!@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Echo context
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Set the Authorization header
			if tt.authorization != "" {
				c.Request().Header.Set("Authorization", tt.authorization)
			}

			// Call the function
			result := helper.GetAPIKey(c)

			// Assert the result
			assert.Equal(t, tt.expectedAPIKey, result, "GetAPIKey() = %v, want %v", result, tt.expectedAPIKey)
		})
	}
}
