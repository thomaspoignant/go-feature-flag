package controller_test

import (
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
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock/mockretriever"
)

func TestGetFlag_Handler(t *testing.T) {
	tests := []struct {
		name           string
		flagKey        string
		useNonWritable bool
		wantStatus     int
		wantEditable   bool
	}{
		{
			name:         "existing flag on a writable flagset",
			flagKey:      "test-flag",
			wantStatus:   http.StatusOK,
			wantEditable: true,
		},
		{
			name:       "missing flag",
			flagKey:    "does-not-exist",
			wantStatus: http.StatusNotFound,
		},
		{
			name:           "existing flag on a non-writable flagset is still readable",
			flagKey:        "array-flag",
			useNonWritable: true,
			wantStatus:     http.StatusOK,
			wantEditable:   false,
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
				client, err = ffclient.New(ffclient.Config{
					PollingInterval: 60 * time.Second,
					Retriever:       mockretriever.NewWritableRetriever("pg"),
				})
			}
			assert.NoError(t, err)
			t.Cleanup(client.Close)

			fm := &fakeFlagsetManager{flagset: client}
			ctrl := controller.NewGetFlag(fm, metric.Metrics{})

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/v1/flags/"+tt.flagKey, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("flag_key")
			c.SetParamValues(tt.flagKey)

			_ = ctrl.Handler(c)
			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantStatus == http.StatusOK {
				var resp model.FlagResponse
				assert.NoError(t, decodeJSON(rec.Body.Bytes(), &resp))
				assert.Equal(t, tt.flagKey, resp.Key)
				assert.Equal(t, tt.wantEditable, resp.Editable)
				assert.NotEmpty(t, resp.ETag)
				assert.Equal(t, resp.ETag, rec.Header().Get("ETag"))
			} else {
				var errResp model.ErrorResponse
				assert.NoError(t, decodeJSON(rec.Body.Bytes(), &errResp))
				assert.NotEmpty(t, errResp.ErrorCode)
			}
		})
	}
}
