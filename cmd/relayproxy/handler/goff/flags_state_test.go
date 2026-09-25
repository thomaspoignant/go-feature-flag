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
	"github.com/thomaspoignant/go-feature-flag/testutils/mock/mockretriever"
)

func TestSetFlagState_Handler(t *testing.T) {
	baseDefinition := []byte(`{
		"variations": {"enabled": true, "disabled": false},
		"defaultRule": {"variation": "disabled"}
	}`)

	tests := []struct {
		name       string
		body       string
		getFunc    func(ctx context.Context, flagKey string) ([]byte, string, error)
		wantStatus int
	}{
		{
			name: "disables an existing flag",
			body: `{"disable": true}`,
			getFunc: func(_ context.Context, _ string) ([]byte, string, error) {
				return baseDefinition, `"etag-1"`, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "flag not found",
			body: `{"disable": true}`,
			getFunc: func(_ context.Context, _ string) ([]byte, string, error) {
				return nil, "", retriever.ErrFlagNotFound
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid body",
			body:       `not-json`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockR := mockretriever.NewWritableRetriever("pg")
			mockR.GetFlagFunc = tt.getFunc
			mockR.UpsertFlagFunc = func(_ context.Context, _ string, definition []byte, _ *string) (bool, string, error) {
				return false, `"etag-2"`, nil
			}
			client, err := ffclient.New(ffclient.Config{
				PollingInterval: 60 * time.Second,
				Retriever:       mockR,
			})
			assert.NoError(t, err)
			t.Cleanup(client.Close)

			fm := &fakeFlagsetManager{flagset: client}
			ctrl := controller.NewSetFlagState(fm, metric.Metrics{})

			e := echo.New()
			req := httptest.NewRequest(http.MethodPatch, "/v1/flags/my-flag/state", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("flag_key")
			c.SetParamValues("my-flag")

			_ = ctrl.Handler(c)
			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantStatus == http.StatusOK {
				var resp model.FlagResponse
				assert.NoError(t, decodeJSON(rec.Body.Bytes(), &resp))
				assert.Equal(t, "my-flag", resp.Key)
				assert.NotNil(t, resp.Definition.Disable)
				assert.True(t, *resp.Definition.Disable)
			} else {
				var errResp model.ErrorResponse
				assert.NoError(t, decodeJSON(rec.Body.Bytes(), &errResp))
				assert.NotEmpty(t, errResp.ErrorCode)
			}
		})
	}
}
