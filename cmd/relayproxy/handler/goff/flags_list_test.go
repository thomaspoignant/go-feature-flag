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
	"github.com/thomaspoignant/go-feature-flag/testutils/mock/mockretriever"
)

func TestListFlags_Handler(t *testing.T) {
	retrieverFunc := func(_ context.Context) ([]byte, error) {
		return []byte(`{
			"flag-a": {"variations": {"true": true, "false": false}, "defaultRule": {"variation": "false"}},
			"flag-b": {"variations": {"true": true, "false": false}, "defaultRule": {"variation": "false"}, "disable": true},
			"other":  {"variations": {"true": true, "false": false}, "defaultRule": {"variation": "false"}}
		}`), nil
	}

	tests := []struct {
		name      string
		query     string
		wantTotal int
		wantKeys  []string
	}{
		{name: "no filter", wantTotal: 3, wantKeys: []string{"flag-a", "flag-b", "other"}},
		{name: "filter by q", query: "?q=flag", wantTotal: 2, wantKeys: []string{"flag-a", "flag-b"}},
		{name: "filter by disabled=true", query: "?disabled=true", wantTotal: 1, wantKeys: []string{"flag-b"}},
		{name: "filter by disabled=false", query: "?disabled=false", wantTotal: 2, wantKeys: []string{"flag-a", "other"}},
		{name: "pagination", query: "?offset=1&limit=1", wantTotal: 3, wantKeys: []string{"flag-b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retriever := mockretriever.NewWritableRetriever("pg")
			retriever.RetrieveFunc = retrieverFunc
			client, err := ffclient.New(ffclient.Config{
				PollingInterval: 60 * time.Second,
				Retriever:       retriever,
			})
			assert.NoError(t, err)
			t.Cleanup(client.Close)

			fm := &fakeFlagsetManager{flagset: client}
			ctrl := controller.NewListFlags(fm, metric.Metrics{})

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/v1/flags"+tt.query, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			assert.NoError(t, ctrl.Handler(c))
			assert.Equal(t, http.StatusOK, rec.Code)

			var resp model.FlagListResponse
			assert.NoError(t, decodeJSON(rec.Body.Bytes(), &resp))
			assert.Equal(t, tt.wantTotal, resp.Meta.Total)

			gotKeys := make([]string, 0, len(resp.Flags))
			for _, f := range resp.Flags {
				gotKeys = append(gotKeys, f.Key)
				assert.True(t, f.Editable)
				assert.Equal(t, "mock:pg", f.Source)
			}
			assert.Equal(t, tt.wantKeys, gotKeys)
			assert.NotEmpty(t, rec.Header().Get("ETag"))
		})
	}
}
