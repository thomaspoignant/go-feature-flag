package controller_test

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestDeleteFlag_Handler(t *testing.T) {
	tests := []struct {
		name       string
		deleteFunc func(ctx context.Context, flagKey string, ifMatch *string) error
		wantStatus int
	}{
		{
			name:       "successful delete",
			deleteFunc: func(_ context.Context, _ string, _ *string) error { return nil },
			wantStatus: http.StatusNoContent,
		},
		{
			name: "not found",
			deleteFunc: func(_ context.Context, _ string, _ *string) error {
				return retriever.ErrFlagNotFound
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "etag mismatch",
			deleteFunc: func(_ context.Context, _ string, _ *string) error {
				return retriever.ErrETagMismatch
			},
			wantStatus: http.StatusPreconditionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockR := mockretriever.NewWritableRetriever("pg")
			mockR.DeleteFlagFunc = tt.deleteFunc
			client, err := ffclient.New(ffclient.Config{
				PollingInterval: 60 * time.Second,
				Retriever:       mockR,
			})
			assert.NoError(t, err)
			t.Cleanup(client.Close)

			fm := &fakeFlagsetManager{flagset: client}
			ctrl := controller.NewDeleteFlag(fm, metric.Metrics{})

			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/v1/flags/my-flag", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("flag_key")
			c.SetParamValues("my-flag")

			_ = ctrl.Handler(c)
			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantStatus != http.StatusNoContent {
				var errResp model.ErrorResponse
				assert.NoError(t, decodeJSON(rec.Body.Bytes(), &errResp))
				assert.NotEmpty(t, errResp.ErrorCode)
			}
		})
	}
}
