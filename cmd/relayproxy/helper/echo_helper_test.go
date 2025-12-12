package helper_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/helper"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
)

func TestAPIKey(t *testing.T) {
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
			name:           "Bearer token with leading/trailing spaces",
			authorization:  "Bearer   my-api-key-with-spaces  ",
			expectedAPIKey: "my-api-key-with-spaces",
		},
		{
			name:           "Bearer token with only leading spaces",
			authorization:  "Bearer   my-api-key-leading",
			expectedAPIKey: "my-api-key-leading",
		},
		{
			name:           "Bearer token with only trailing spaces",
			authorization:  "Bearer my-api-key-trailing   ",
			expectedAPIKey: "my-api-key-trailing",
		},
		{
			name:           "Bearer token with tabs and newlines",
			authorization:  "Bearer \t\nmy-api-key-with-tabs\n\t",
			expectedAPIKey: "my-api-key-with-tabs",
		},
		{
			name:           "Bearer token with mixed whitespace",
			authorization:  "Bearer \t \n my-api-key-mixed \t \n ",
			expectedAPIKey: "my-api-key-mixed",
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
			name:           "Bearer with single space",
			authorization:  "Bearer ",
			expectedAPIKey: "",
		},
		{
			name:           "Bearer with multiple spaces only",
			authorization:  "Bearer   ",
			expectedAPIKey: "",
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
		{
			name:           "Bearer with spaces around special characters",
			authorization:  "Bearer   my-api-key!@#$%^&*()  ",
			expectedAPIKey: "my-api-key!@#$%^&*()",
		},
		{
			name:           "Case insensitive Bearer prefix",
			authorization:  "bearer my-api-key-lowercase",
			expectedAPIKey: "my-api-key-lowercase",
		},
		{
			name:           "Mixed case Bearer prefix",
			authorization:  "BeArEr my-api-key-mixed-case",
			expectedAPIKey: "my-api-key-mixed-case",
		},
		{
			name:           "Bearer with only whitespace token",
			authorization:  "Bearer \t\n\r ",
			expectedAPIKey: "",
		},
		{
			name:           "Bearer with token containing only spaces",
			authorization:  "Bearer   \t  \n  ",
			expectedAPIKey: "",
		},
		{
			name:           "Bearer with token that has internal spaces",
			authorization:  "Bearer my api key with spaces",
			expectedAPIKey: "my api key with spaces",
		},
		{
			name:           "Bearer with token that has internal spaces and trimming",
			authorization:  "Bearer   my api key with spaces  ",
			expectedAPIKey: "my api key with spaces",
		},
		{
			name:           "Non-Bearer scheme with spaces",
			authorization:  "  CustomScheme my-custom-key  ",
			expectedAPIKey: "  CustomScheme my-custom-key  ",
		},
		{
			name:           "Non-Bearer scheme without spaces",
			authorization:  "CustomScheme my-custom-key",
			expectedAPIKey: "CustomScheme my-custom-key",
		},
		{
			name:           "Authorization header with only spaces",
			authorization:  "   ",
			expectedAPIKey: "   ",
		},
		{
			name:           "Authorization header with tabs and newlines",
			authorization:  "\t\n\r",
			expectedAPIKey: "\t\n\r",
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
			result := helper.APIKey(c)

			// Assert the result
			assert.Equal(t, tt.expectedAPIKey, result, "GetAPIKey() = %v, want %v", result, tt.expectedAPIKey)
		})
	}
}

func TestFlagSet(t *testing.T) {
	tests := []struct {
		name           string
		flagsetManager service.FlagsetManager
		apiKey         string
		wantError      bool
		wantMsg        string
	}{
		{
			name:           "nil flagset manager should return internal server error",
			flagsetManager: nil,
			apiKey:         "test-api-key",
			wantError:      true,
			wantMsg:        "flagset manager is not initialized",
		},
		{
			name:           "flagset manager error should return bad request error",
			flagsetManager: &MockFlagsetManager{err: errors.New("flagset not found for API key")},
			apiKey:         "invalid-api-key",
			wantError:      true,
			wantMsg:        "error while getting flagset: flagset not found for API key",
		},
		{
			name:           "successful flagset retrieval should return flagset",
			flagsetManager: &MockFlagsetManager{flagset: &ffclient.GoFeatureFlag{}},
			apiKey:         "valid-api-key",
			wantError:      false,
		},
		{
			name:           "empty api key with error should return bad request error",
			flagsetManager: &MockFlagsetManager{err: errors.New("no API key provided")},
			apiKey:         "",
			wantError:      true,
			wantMsg:        "error while getting flagset: no API key provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagset, err := helper.FlagSet(tt.flagsetManager, tt.apiKey)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, flagset)
				if tt.wantMsg != "" {
					assert.Contains(t, err.Message, tt.wantMsg)
				}
			} else {
				if err != nil || flagset == nil {
					t.Logf("DEBUG: err=%v, flagset=%v", err, flagset)
				}
				assert.Nil(t, err)
				assert.NotNil(t, flagset)
			}
		})
	}
}

func TestFlagSet_Integration(t *testing.T) {
	t.Run("should handle specific error messages correctly", func(t *testing.T) {
		mockManager := &MockFlagsetManager{err: errors.New("test error")}
		flagset, err := helper.FlagSet(mockManager, "test-key")
		assert.Error(t, err)
		assert.Nil(t, flagset)
		assert.Contains(t, err.Message, "error while getting flagset:")
	})

	t.Run("should return flagset when manager returns valid flagset", func(t *testing.T) {
		mockManager := &MockFlagsetManager{flagset: &ffclient.GoFeatureFlag{}}
		flagset, err := helper.FlagSet(mockManager, "valid-key")
		if err != nil || flagset == nil {
			t.Logf("DEBUG: err=%v, flagset=%v", err, flagset)
		}
		assert.Nil(t, err)
		assert.NotNil(t, flagset)
	})
}

// MockFlagsetManager is a simple mock implementation of service.FlagsetManager
type MockFlagsetManager struct {
	flagset *ffclient.GoFeatureFlag
	err     error
}

func (m *MockFlagsetManager) FlagSet(apiKey string) (*ffclient.GoFeatureFlag, error) {
	return m.flagset, m.err
}

func (m *MockFlagsetManager) FlagSetName(apiKey string) (string, error) {
	return "", nil
}

func (m *MockFlagsetManager) AllFlagSets() (map[string]*ffclient.GoFeatureFlag, error) {
	return nil, nil
}

func (m *MockFlagsetManager) Default() *ffclient.GoFeatureFlag {
	return nil
}

func (m *MockFlagsetManager) IsDefaultFlagSet() bool {
	return false
}

func (m *MockFlagsetManager) Close() {
	// nothing to do
}
