package controller_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	controller "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/goff"
	"github.com/thomaspoignant/go-feature-flag/internal/flaghistory"
)

type historyStoreStub struct {
	filter  flaghistory.Filter
	page    flaghistory.Page
	listErr error
}

func (s *historyStoreStub) Migrate(context.Context) error { return nil }
func (s *historyStoreStub) Record(context.Context, flaghistory.Entry) (bool, error) {
	return true, nil
}
func (s *historyStoreStub) Latest(context.Context, string, string) (*flaghistory.Entry, error) {
	return nil, nil
}
func (s *historyStoreStub) LatestPerFlag(context.Context, string) (map[string]flaghistory.Entry, error) {
	return nil, nil
}
func (s *historyStoreStub) List(_ context.Context, f flaghistory.Filter) (flaghistory.Page, error) {
	s.filter = f
	return s.page, s.listErr
}
func (s *historyStoreStub) Purge(context.Context, time.Time) (int64, error) { return 0, nil }
func (s *historyStoreStub) Ping(context.Context) error                      { return nil }
func (s *historyStoreStub) Close(context.Context)                           {}

func startHistory(t *testing.T, store flaghistory.Store, connectErr error) *flaghistory.History {
	t.Helper()
	h := flaghistory.New(flaghistory.Config{
		Connect: func(context.Context) (flaghistory.Store, error) {
			if connectErr != nil {
				return nil, connectErr
			}
			return store, nil
		},
	})
	h.Start()
	t.Cleanup(h.Close)
	if connectErr == nil {
		require.Eventually(t, func() bool { _, err := h.Store(); return err == nil },
			time.Second, 5*time.Millisecond)
	}
	return h
}

func callHistory(t *testing.T, h *flaghistory.History, url string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	ctrl := controller.NewFlagHistory(h)
	e.GET("/admin/v1/flag-history", ctrl.Handler)
	e.GET("/admin/v1/flag-history/:flagKey", ctrl.Handler)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))
	return rec
}

func TestFlagHistory_Handler(t *testing.T) {
	from := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name       string
		url        string
		wantStatus int
		wantFilter flaghistory.Filter
	}{
		{
			name:       "no filter",
			url:        "/admin/v1/flag-history",
			wantStatus: http.StatusOK,
		},
		{
			name:       "all the filters",
			url:        "/admin/v1/flag-history?flagset=fs&action=updated&author=jane&from=2026-01-02T03:04:05Z&limit=10&cursor=42",
			wantStatus: http.StatusOK,
			wantFilter: flaghistory.Filter{
				Flagset: "fs", Action: flaghistory.ActionUpdated, Author: "jane",
				From: from, Limit: 10, Cursor: 42,
			},
		},
		{
			name:       "flag key in the path",
			url:        "/admin/v1/flag-history/my-flag",
			wantStatus: http.StatusOK,
			wantFilter: flaghistory.Filter{FlagKey: "my-flag"},
		},
		{name: "invalid action", url: "/admin/v1/flag-history?action=nope", wantStatus: http.StatusBadRequest},
		{name: "invalid from", url: "/admin/v1/flag-history?from=yesterday", wantStatus: http.StatusBadRequest},
		{name: "invalid to", url: "/admin/v1/flag-history?to=yesterday", wantStatus: http.StatusBadRequest},
		{
			name:       "from after to",
			url:        "/admin/v1/flag-history?from=2026-02-01T00:00:00Z&to=2026-01-01T00:00:00Z",
			wantStatus: http.StatusBadRequest,
		},
		{name: "limit not a number", url: "/admin/v1/flag-history?limit=abc", wantStatus: http.StatusBadRequest},
		{name: "limit zero", url: "/admin/v1/flag-history?limit=0", wantStatus: http.StatusBadRequest},
		{name: "negative cursor", url: "/admin/v1/flag-history?cursor=-1", wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &historyStoreStub{page: flaghistory.Page{
				Entries:    []flaghistory.Entry{{ID: 7, FlagKey: "f", Action: flaghistory.ActionAdded, Author: "jane"}},
				NextCursor: 7,
			}}
			rec := callHistory(t, startHistory(t, store, nil), tt.url)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantStatus != http.StatusOK {
				return
			}
			assert.Equal(t, tt.wantFilter, store.filter)
			var got flaghistory.Page
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
			assert.Equal(t, int64(7), got.NextCursor)
			assert.Equal(t, "jane", got.Entries[0].Author)
		})
	}
}

func TestFlagHistory_Handler_databaseUnavailable(t *testing.T) {
	t.Run("not connected yet", func(t *testing.T) {
		h := startHistory(t, nil, errors.New("connection refused"))
		rec := callHistory(t, h, "/admin/v1/flag-history")
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})
	t.Run("query fails", func(t *testing.T) {
		store := &historyStoreStub{listErr: errors.New("boom")}
		rec := callHistory(t, startHistory(t, store, nil), "/admin/v1/flag-history")
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.NotContains(t, rec.Body.String(), "boom", "database errors must not leak")
	})
}
