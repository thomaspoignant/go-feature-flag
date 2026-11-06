package controller_test

import (
	"encoding/json"
	"io"
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
	t.Parallel()
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
		wantErrors  []controller.ValidateFlagsError
		wantErrCode int
		chunked     bool
	}{
		{name: "empty yaml body", contentType: "application/yaml", body: "", wantStatus: 200, wantValid: true},
		{name: "yaml scalar", contentType: "application/yaml", body: "just a string", wantErrCode: 400},
		{name: "yaml list", contentType: "application/yaml", body: "- a\n- b\n", wantErrCode: 400},
		{name: "json list", contentType: "application/json", body: "[]", wantErrCode: 400},
		{
			name: "chunked body too large", contentType: "application/json", chunked: true,
			body: `{"a":"` + strings.Repeat("x", 1<<20) + `"}`, wantErrCode: 413,
		},
		{
			name: "panicking flag definition", contentType: "application/json",
			body:       `{"f":{"variations":{"a":true},"defaultRule":{"progressiveRollout":{}}}}`,
			wantStatus: 422,
			wantErrors: []controller.ValidateFlagsError{{
				Flag:    "f",
				Message: "invalid flag definition: runtime error: invalid memory address or nil pointer dereference",
			}},
		},
		{name: "valid json", contentType: "application/json", body: validJSON, wantStatus: 200, wantValid: true},
		{name: "valid json with charset", contentType: "application/json; charset=utf-8", body: validJSON, wantStatus: 200, wantValid: true},
		{name: "valid yaml", contentType: "application/yaml", body: validYAML, wantStatus: 200, wantValid: true},
		{name: "valid x-yaml", contentType: "application/x-yaml", body: validYAML, wantStatus: 200, wantValid: true},
		{name: "valid text yaml", contentType: "text/yaml", body: validYAML, wantStatus: 200, wantValid: true},
		{name: "valid toml", contentType: "application/toml", body: validTOML, wantStatus: 200, wantValid: true},
		{
			name: "invalid flags sorted by key", contentType: "application/json", body: invalidJSON,
			wantStatus: 422,
			wantErrors: []controller.ValidateFlagsError{
				{Flag: "flag-a", Message: "missing default rule"},
				{Flag: "flag-b", Message: "missing default rule"},
			},
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
			if tt.chunked {
				req.Body = io.NopCloser(strings.NewReader(tt.body))
				req.ContentLength = -1
			}
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
			wantErrors := tt.wantErrors
			if wantErrors == nil {
				wantErrors = []controller.ValidateFlagsError{}
			}
			assert.Equal(t, wantErrors, got.Errors)
		})
	}
}
