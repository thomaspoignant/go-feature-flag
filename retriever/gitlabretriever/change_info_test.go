package gitlabretriever_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gitlabretriever"
)

var _ retriever.ChangeInfoProvider = (*gitlabretriever.Retriever)(nil)

func TestRetriever_LastChange(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr bool
		wantNil bool
	}{
		{
			name:   "last commit",
			status: http.StatusOK,
			body: `[{"id":"abc123","author_name":"Jane Doe","author_email":"jane@x.io",
				"committed_date":"2026-01-02T03:04:05Z"}]`,
		},
		{name: "no commit", status: http.StatusOK, body: `[]`, wantNil: true},
		{name: "api error", status: http.StatusUnauthorized, body: `{}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v4/projects/group%2Frepo/repository/commits", r.URL.EscapedPath())
				assert.Equal(t, "dev", r.URL.Query().Get("ref_name"))
				assert.Equal(t, "flags.yaml", r.URL.Query().Get("path"))
				assert.Equal(t, "tok", r.Header.Get("PRIVATE-TOKEN"))
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			r := &gitlabretriever.Retriever{
				RepositorySlug: "group/repo", FilePath: "flags.yaml", Branch: "dev",
				GitlabToken: "tok", BaseURL: srv.URL,
			}
			got, err := r.LastChange(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			assert.Equal(t, "Jane Doe", got.Author)
			assert.Equal(t, "jane@x.io", got.Email)
			assert.Equal(t, "abc123", got.Reference)
			assert.Equal(t, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), got.ChangedAt)
		})
	}
}

func TestRetriever_LastChange_missingConfig(t *testing.T) {
	_, err := (&gitlabretriever.Retriever{}).LastChange(context.Background())
	assert.Error(t, err)
}
