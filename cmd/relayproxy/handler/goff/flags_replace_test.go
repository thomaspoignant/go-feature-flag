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

const validDefinitionBody = `{
	"variations": {"enabled": true, "disabled": false},
	"defaultRule": {"variation": "disabled"}
}`

func TestReplaceFlag_Handler(t *testing.T) {
	tests := []struct {
		name        string
		upsertFunc  func(ctx context.Context, flagKey string, definition []byte, ifMatch *string) (bool, string, error)
		ifMatch     string
		wantStatus  int
		wantCreated bool
	}{
		{
			name: "creates when absent",
			upsertFunc: func(_ context.Context, _ string, _ []byte, _ *string) (bool, string, error) {
				return true, `"etag-1"`, nil
			},
			wantStatus:  http.StatusCreated,
			wantCreated: true,
		},
		{
			name: "replaces when present",
			upsertFunc: func(_ context.Context, _ string, _ []byte, _ *string) (bool, string, error) {
				return false, `"etag-2"`, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "etag mismatch",
			ifMatch: `"stale"`,
			upsertFunc: func(_ context.Context, _ string, _ []byte, _ *string) (bool, string, error) {
				return false, "", retriever.ErrETagMismatch
			},
			wantStatus: http.StatusPreconditionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockR := mockretriever.NewWritableRetriever("pg")
			mockR.UpsertFlagFunc = tt.upsertFunc
			client, err := ffclient.New(ffclient.Config{
				PollingInterval: 60 * time.Second,
				Retriever:       mockR,
			})
			assert.NoError(t, err)
			t.Cleanup(client.Close)

			fm := &fakeFlagsetManager{flagset: client}
			ctrl := controller.NewReplaceFlag(fm, metric.Metrics{})

			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/v1/flags/my-flag", strings.NewReader(validDefinitionBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			if tt.ifMatch != "" {
				req.Header.Set("If-Match", tt.ifMatch)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("flag_key")
			c.SetParamValues("my-flag")

			_ = ctrl.Handler(c)
			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantStatus == http.StatusOK || tt.wantStatus == http.StatusCreated {
				var resp model.FlagResponse
				assert.NoError(t, decodeJSON(rec.Body.Bytes(), &resp))
				assert.Equal(t, "my-flag", resp.Key)
				if tt.wantCreated {
					assert.NotEmpty(t, rec.Header().Get("Location"))
				}
			} else {
				var errResp model.ErrorResponse
				assert.NoError(t, decodeJSON(rec.Body.Bytes(), &errResp))
				assert.NotEmpty(t, errResp.ErrorCode)
			}
		})
	}
}
