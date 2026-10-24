package bitbucketretriever_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/bitbucketretriever"
)

var _ retriever.ChangeInfoProvider = (*bitbucketretriever.Retriever)(nil)

func TestRetriever_LastChange(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantErr    bool
		wantNil    bool
		wantAuthor string
		wantEmail  string
	}{
		{
			name:   "display name preferred, email from the raw author",
			status: http.StatusOK,
			body: `{"values":[{"hash":"abc123","date":"2026-01-02T03:04:05+00:00",
				"author":{"raw":"Jane Doe <jane@x.io>","user":{"display_name":"Jane D."}}}]}`,
			wantAuthor: "Jane D.", wantEmail: "jane@x.io",
		},
		{
			name:   "raw author only",
			status: http.StatusOK,
			body: `{"values":[{"hash":"abc123","date":"2026-01-02T03:04:05+00:00",
				"author":{"raw":"Jane Doe <jane@x.io>"}}]}`,
			wantAuthor: "Jane Doe", wantEmail: "jane@x.io",
		},
		{
			name:       "unparsable raw author is kept as is",
			status:     http.StatusOK,
			body:       `{"values":[{"hash":"abc123","author":{"raw":"ci-bot"}}]}`,
			wantAuthor: "ci-bot",
		},
		{name: "no commit", status: http.StatusOK, body: `{"values":[]}`, wantNil: true},
		{name: "api error", status: http.StatusNotFound, body: `{}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/2.0/repositories/ws/repo/commits/dev", r.URL.Path)
				assert.Equal(t, "flags.yaml", r.URL.Query().Get("path"))
				assert.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			r := &bitbucketretriever.Retriever{
				RepositorySlug: "ws/repo", FilePath: "flags.yaml", Branch: "dev",
				BitBucketToken: "tok", BaseURL: srv.URL,
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
			assert.Equal(t, tt.wantEmail, got.Email)
			assert.Equal(t, "abc123", got.Reference)
		})
	}
}
