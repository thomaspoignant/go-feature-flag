package controller_test

import (
	"encoding/json"
	"errors"
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

const (
	jsonBodyLimit = 1 << 20
	yamlBodyLimit = 64 << 10
	tomlBodyLimit = 64 << 10
)

// failingReader is a request body that always fails to be read.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("read failure") }

// padTo pads body with the pad string so the result is exactly size bytes long.
func padTo(body, pad string, size int) string {
	return body + strings.Repeat(pad, size-len(body))
}

func TestNewValidateFlags_Handler(t *testing.T) {
	t.Parallel()
	const validJSON = `{"flag-a":{"variations":{"a":true,"b":false},"defaultRule":{"variation":"a"}}}`
	const validYAML = "flag-a:\n  variations:\n    a: true\n    b: false\n  defaultRule:\n    variation: a\n"
	const validTOML = "[flag-a]\n[flag-a.variations]\na = true\nb = false\n[flag-a.defaultRule]\nvariation = \"a\"\n"
	const invalidJSON = `{"flag-b":{"variations":{"a":true}},"flag-a":{"variations":{"a":true}}}`
	oversizedJSON := padTo(validJSON, " ", jsonBodyLimit+1)

	tests := []struct {
		name        string
		contentType string
		body        string
		// altMessage is the message returned when modules/core validates the flag without panicking.
		altMessage string
		reader     io.Reader
		chunked    bool
		wantStatus int
		wantValid  bool
		wantErrors []controller.FlagValidationError
	}{
		{name: "valid json", contentType: "application/json", body: validJSON, wantStatus: http.StatusOK, wantValid: true},
		{name: "valid json with charset", contentType: "application/json; charset=utf-8", body: validJSON, wantStatus: http.StatusOK, wantValid: true},
		{name: "valid yaml", contentType: "application/yaml", body: validYAML, wantStatus: http.StatusOK, wantValid: true},
		{name: "valid x-yaml", contentType: "application/x-yaml", body: validYAML, wantStatus: http.StatusOK, wantValid: true},
		{name: "valid text yaml", contentType: "text/yaml", body: validYAML, wantStatus: http.StatusOK, wantValid: true},
		{name: "valid toml", contentType: "application/toml", body: validTOML, wantStatus: http.StatusOK, wantValid: true},
		{name: "empty yaml body", contentType: "application/yaml", body: "", wantStatus: http.StatusOK, wantValid: true},
		{
			name: "invalid flags sorted by key", contentType: "application/json", body: invalidJSON,
			wantStatus: http.StatusUnprocessableEntity,
			wantErrors: []controller.FlagValidationError{
				{Flag: "flag-a", Message: "missing default rule"},
				{Flag: "flag-b", Message: "missing default rule"},
			},
		},
		{
			name: "panicking flag definition", contentType: "application/json",
			body:       `{"f":{"variations":{"a":true},"defaultRule":{"progressiveRollout":{}}}}`,
			altMessage: "invalid progressive rollout, initial and end steps are required",
			wantStatus: http.StatusUnprocessableEntity,
			wantErrors: []controller.FlagValidationError{{
				Flag:    "f",
				Message: "internal validation error: flag definition could not be validated",
			}},
		},
		{name: "bad json syntax", contentType: "application/json", body: `{"flag-a":`, wantStatus: http.StatusBadRequest},
		{name: "bad yaml syntax", contentType: "application/yaml", body: "a: [", wantStatus: http.StatusBadRequest},
		{name: "bad toml syntax", contentType: "application/toml", body: "a = ", wantStatus: http.StatusBadRequest},
		{name: "yaml scalar", contentType: "application/yaml", body: "just a string", wantStatus: http.StatusBadRequest},
		{name: "yaml list", contentType: "application/yaml", body: "- a\n- b\n", wantStatus: http.StatusBadRequest},
		{name: "json list", contentType: "application/json", body: "[]", wantStatus: http.StatusBadRequest},
		{name: "unreadable body", contentType: "application/json", reader: failingReader{}, wantStatus: http.StatusBadRequest},
		{name: "unsupported content type", contentType: "text/plain", body: validJSON, wantStatus: http.StatusUnsupportedMediaType},
		{name: "malformed content type", contentType: "application/json;;;=", body: validJSON, wantStatus: http.StatusUnsupportedMediaType},
		{name: "missing content type", contentType: "", body: validJSON, wantStatus: http.StatusUnsupportedMediaType},
		{
			name: "json at the size limit", contentType: "application/json",
			body: padTo(validJSON, " ", jsonBodyLimit), wantStatus: http.StatusOK, wantValid: true,
		},
		{name: "json over the size limit", contentType: "application/json", body: oversizedJSON, wantStatus: http.StatusRequestEntityTooLarge},
		{
			name: "chunked json over the size limit", contentType: "application/json", chunked: true,
			body: oversizedJSON, wantStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name: "yaml at the size limit", contentType: "application/yaml",
			body: padTo(validYAML, "\n", yamlBodyLimit), wantStatus: http.StatusOK, wantValid: true,
		},
		{
			name: "yaml over the size limit", contentType: "application/yaml",
			body: padTo(validYAML, "\n", yamlBodyLimit+1), wantStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name: "chunked yaml over the size limit", contentType: "application/yaml", chunked: true,
			body: padTo(validYAML, "\n", yamlBodyLimit+1), wantStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name: "toml at the size limit", contentType: "application/toml",
			body: padTo(validTOML, "\n", tomlBodyLimit), wantStatus: http.StatusOK, wantValid: true,
		},
		{
			name: "toml over the size limit", contentType: "application/toml",
			body: padTo(validTOML, "\n", tomlBodyLimit+1), wantStatus: http.StatusRequestEntityTooLarge,
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
			switch {
			case tt.reader != nil:
				req.Body = io.NopCloser(tt.reader)
				req.ContentLength = -1
			case tt.chunked:
				req.Body = io.NopCloser(strings.NewReader(tt.body))
				req.ContentLength = -1
			}
			if tt.contentType != "" {
				req.Header.Set(echo.HeaderContentType, tt.contentType)
			}
			c := e.NewContext(req, rec)

			err := ctrl.Handler(c)
			if tt.wantStatus != http.StatusOK && tt.wantStatus != http.StatusUnprocessableEntity {
				var httpErr *echo.HTTPError
				require.ErrorAs(t, err, &httpErr)
				assert.Equal(t, tt.wantStatus, httpErr.Code)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var got controller.FlagValidationResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
			assert.Equal(t, "v1.2.3", got.Version)
			assert.Equal(t, tt.wantValid, got.Valid)
			if tt.altMessage != "" && len(got.Errors) == 1 && got.Errors[0].Message == tt.altMessage {
				got.Errors[0].Message = tt.wantErrors[0].Message
			}
			wantErrors := tt.wantErrors
			if wantErrors == nil {
				wantErrors = []controller.FlagValidationError{}
			}
			assert.Equal(t, wantErrors, got.Errors)
		})
	}
}
