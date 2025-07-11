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

func TestGetFlagSet(t *testing.T) {
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
			flagset, err := helper.GetFlagSet(tt.flagsetManager, tt.apiKey)
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

func TestGetFlagSet_Integration(t *testing.T) {
	t.Run("should handle specific error messages correctly", func(t *testing.T) {
		mockManager := &MockFlagsetManager{err: errors.New("test error")}
		flagset, err := helper.GetFlagSet(mockManager, "test-key")
		assert.Error(t, err)
		assert.Nil(t, flagset)
		assert.Contains(t, err.Message, "error while getting flagset:")
	})

	t.Run("should return flagset when manager returns valid flagset", func(t *testing.T) {
		mockManager := &MockFlagsetManager{flagset: &ffclient.GoFeatureFlag{}}
		flagset, err := helper.GetFlagSet(mockManager, "valid-key")
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

func (m *MockFlagsetManager) GetFlagSet(apiKey string) (*ffclient.GoFeatureFlag, error) {
	return m.flagset, m.err
}

func (m *MockFlagsetManager) GetFlagSetName(apiKey string) (string, error) {
	return "", nil
}

func (m *MockFlagsetManager) GetFlagSets() (map[string]*ffclient.GoFeatureFlag, error) {
	return nil, nil
}

func (m *MockFlagsetManager) GetDefaultFlagSet() *ffclient.GoFeatureFlag {
	return nil
}

func (m *MockFlagsetManager) IsDefaultFlagSet() bool {
	return false
}

func (m *MockFlagsetManager) Close() {
	// nothing to do
}
