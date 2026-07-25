package controller_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	controller "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/goff"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock/mockretriever"
)

const validCreateBody = `{
	"key": "new-flag",
	"definition": {
		"variations": {"enabled": true, "disabled": false},
		"defaultRule": {"variation": "disabled"}
	}
}`

func TestCreateFlag_Handler(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		useNonWritable bool
		createFunc     func(ctx context.Context, flagKey string, definition []byte) (string, error)
		wantStatus     int
	}{
		{
			name:       "valid create",
			body:       validCreateBody,
			wantStatus: http.StatusCreated,
		},
		{
			name:           "non-writable flagset is forbidden",
			body:           validCreateBody,
			useNonWritable: true,
			wantStatus:     http.StatusForbidden,
		},
		{
			name:       "invalid key",
			body:       strings.Replace(validCreateBody, "new-flag", "not a valid key!", 1),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid definition",
			body:       `{"key":"new-flag","definition":{"variations":{},"defaultRule":{"variation":"disabled"}}}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "flag already exists",
			body: validCreateBody,
			createFunc: func(_ context.Context, _ string, _ []byte) (string, error) {
				return "", retriever.ErrFlagAlreadyExists
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "flagset not configured",
			body: validCreateBody,
			createFunc: func(_ context.Context, _ string, _ []byte) (string, error) {
				return "", retriever.ErrFlagsetNotConfigured
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *ffclient.GoFeatureFlag
			var err error
			if tt.useNonWritable {
				client, err = ffclient.New(ffclient.Config{
					PollingInterval: 60 * time.Second,
					Retriever:       &fileretriever.Retriever{Path: testdataDir + "/config_flags.yaml"},
				})
			} else {
				mockR := mockretriever.NewWritableRetriever("pg")
				if tt.createFunc != nil {
					mockR.CreateFlagFunc = tt.createFunc
				} else {
					mockR.CreateFlagFunc = func(_ context.Context, _ string, definition []byte) (string, error) {
						return `"etag-value"`, nil
					}
				}
				client, err = ffclient.New(ffclient.Config{
					PollingInterval: 60 * time.Second,
					Retriever:       mockR,
				})
			}
			assert.NoError(t, err)
			t.Cleanup(client.Close)

			fm := &fakeFlagsetManager{flagset: client}
			ctrl := controller.NewCreateFlag(fm, metric.Metrics{})

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/v1/flags", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = ctrl.Handler(c)
			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantStatus == http.StatusCreated {
				var resp model.FlagResponse
				assert.NoError(t, decodeJSON(rec.Body.Bytes(), &resp))
				assert.Equal(t, "new-flag", resp.Key)
				assert.NotEmpty(t, rec.Header().Get("Location"))
				assert.NotEmpty(t, rec.Header().Get("ETag"))
			} else {
				var errResp model.ErrorResponse
				assert.NoError(t, decodeJSON(rec.Body.Bytes(), &errResp))
				assert.NotEmpty(t, errResp.ErrorCode)
			}
		})
	}
}
