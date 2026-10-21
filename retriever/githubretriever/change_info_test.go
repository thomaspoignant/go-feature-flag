package githubretriever_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
)

var _ retriever.ChangeInfoProvider = (*githubretriever.Retriever)(nil)

func TestRetriever_LastChange(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantErr    bool
		wantAuthor string
		wantNil    bool
	}{
		{
			name:   "login is preferred over the commit name",
			status: http.StatusOK,
			body: `[{"sha":"abc123","commit":{"author":{"name":"Jane Doe","email":"jane@x.io",
				"date":"2026-01-02T03:04:05Z"}},"author":{"login":"jdoe"}}]`,
			wantAuthor: "jdoe",
		},
		{
			name:   "commit name when the author has no github account",
			status: http.StatusOK,
			body: `[{"sha":"abc123","commit":{"author":{"name":"Jane Doe","email":"jane@x.io",
				"date":"2026-01-02T03:04:05Z"}},"author":null}]`,
			wantAuthor: "Jane Doe",
		},
		{name: "no commit", status: http.StatusOK, body: `[]`, wantNil: true},
		{name: "api error", status: http.StatusForbidden, body: `{}`, wantErr: true},
		{name: "invalid json", status: http.StatusOK, body: `not json`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/repos/org/repo/commits", r.URL.Path)
				assert.Equal(t, "flags/flags.yaml", r.URL.Query().Get("path"))
				assert.Equal(t, "develop", r.URL.Query().Get("sha"))
				assert.Equal(t, "1", r.URL.Query().Get("per_page"))
				assert.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			r := &githubretriever.Retriever{
				RepositorySlug: "org/repo", FilePath: "flags/flags.yaml", Branch: "develop",
				GithubToken: "tok", BaseURL: srv.URL,
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
			assert.Equal(t, tt.wantAuthor, got.Author)
			assert.Equal(t, "jane@x.io", got.Email)
			assert.Equal(t, "abc123", got.Reference)
			assert.Equal(t, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), got.ChangedAt)
		})
	}
}

func TestRetriever_LastChange_pathTraversal(t *testing.T) {
	r := &githubretriever.Retriever{RepositorySlug: "org/repo", FilePath: "../secret"}
	_, err := r.LastChange(context.Background())
	assert.Error(t, err)
}
