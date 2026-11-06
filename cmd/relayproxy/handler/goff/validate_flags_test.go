package controller_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	controller "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/goff"
)

func TestNewValidateFlags_Handler(t *testing.T) {
	const validJSON = `{"flag-a":{"variations":{"a":true,"b":false},"defaultRule":{"variation":"a"}}}`
	const validYAML = "flag-a:\n  variations:\n    a: true\n    b: false\n  defaultRule:\n    variation: a\n"
	const validTOML = "[flag-a]\n[flag-a.variations]\na = true\nb = false\n[flag-a.defaultRule]\nvariation = \"a\"\n"
	const invalidJSON = `{"flag-b":{"variations":{"a":true}},"flag-a":{"variations":{"a":true}}}`

	tests := []struct {
		name        string
		contentType string
		body        string
		wantStatus  int
		wantValid   bool
		wantFlags   []string
		wantErrCode int
	}{
		{name: "valid json", contentType: "application/json", body: validJSON, wantStatus: 200, wantValid: true},
		{name: "valid json with charset", contentType: "application/json; charset=utf-8", body: validJSON, wantStatus: 200, wantValid: true},
		{name: "valid yaml", contentType: "application/yaml", body: validYAML, wantStatus: 200, wantValid: true},
		{name: "valid x-yaml", contentType: "application/x-yaml", body: validYAML, wantStatus: 200, wantValid: true},
		{name: "valid text yaml", contentType: "text/yaml", body: validYAML, wantStatus: 200, wantValid: true},
		{name: "valid toml", contentType: "application/toml", body: validTOML, wantStatus: 200, wantValid: true},
		{
			name: "invalid flags sorted by key", contentType: "application/json", body: invalidJSON,
			wantStatus: 422, wantFlags: []string{"flag-a", "flag-b"},
		},
		{name: "bad json syntax", contentType: "application/json", body: `{"flag-a":`, wantErrCode: 400},
		{name: "bad yaml syntax", contentType: "application/yaml", body: "a: [", wantErrCode: 400},
		{name: "bad toml syntax", contentType: "application/toml", body: "a = ", wantErrCode: 400},
		{name: "unsupported content type", contentType: "text/plain", body: validJSON, wantErrCode: 400},
		{name: "malformed content type", contentType: "application/json;;;=", body: validJSON, wantErrCode: 400},
		{name: "missing content type", contentType: "", body: validJSON, wantErrCode: 400},
		{
			name: "body too large", contentType: "application/json",
			body: `{"a":"` + strings.Repeat("x", 1<<20) + `"}`, wantErrCode: 413,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := controller.NewValidateFlags("v1.2.3")
			e := echo.New()
			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(
				t.Context(), http.MethodPost, "/admin/v1/flags/validate", strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set(echo.HeaderContentType, tt.contentType)
			}
			c := e.NewContext(req, rec)

			err := ctrl.Handler(c)
			if tt.wantErrCode != 0 {
				var httpErr *echo.HTTPError
				require.ErrorAs(t, err, &httpErr)
				assert.Equal(t, tt.wantErrCode, httpErr.Code)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var got controller.ValidateFlagsResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
			assert.Equal(t, "v1.2.3", got.Version)
			assert.Equal(t, tt.wantValid, got.Valid)
			require.NotNil(t, got.Errors)
			flags := make([]string, 0, len(got.Errors))
			for _, e := range got.Errors {
				assert.NotEmpty(t, e.Message)
				flags = append(flags, e.Flag)
			}
			if tt.wantFlags == nil {
				assert.Empty(t, flags)
			} else {
				assert.Equal(t, tt.wantFlags, flags)
			}
		})
	}
}
