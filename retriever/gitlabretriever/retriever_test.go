package gitlabretriever_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/retriever/gitlabretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"

	"github.com/stretchr/testify/assert"
)

func sampleText() string {
	return `test-flag:
  variations:
    true_var: true
    false_var: false
  targeting:
    - query: key eq "random-key"
      percentage:
        true_var: 0
        false_var: 100
  defaultRule:
    variation: false_var
`
}
func Test_gitlab_Retrieve(t *testing.T) {
	type fields struct {
		httpClient mock.HTTP
		context    context.Context

		filePath       string
		gitlabToken    string
		repositorySlug string
		baseURL        string
		branch         string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				httpClient:     mock.HTTP{},
				baseURL:        "https://gitlab.com",
				repositorySlug: "aa/go-feature-flags-config",
				filePath:       "flag-config.yaml",
			},
			want:    []byte(sampleText()),
			wantErr: false,
		},
		{
			name: "Success with context",
			fields: fields{
				httpClient:     mock.HTTP{},
				baseURL:        "https://gitlab.com",
				repositorySlug: "aa/go-feature-flags-config",
				filePath:       "flag-config.yaml",
				context:        context.Background(),
			},
			want:    []byte(sampleText()),
			wantErr: false,
		},
		{
			name: "Success with default method",
			fields: fields{
				httpClient:     mock.HTTP{},
				baseURL:        "https://gitlab.com",
				repositorySlug: "aa/go-feature-flags-config",
				filePath:       "flag-config.yaml",
			},
			want:    []byte(sampleText()),
			wantErr: false,
		},
		{
			name: "HTTP Error",
			fields: fields{
				httpClient:     mock.HTTP{},
				baseURL:        "https://gitlab.com/error",
				repositorySlug: "aa/go-feature-flags-config",
				filePath:       "bad-file/file.yaml",
				branch:         "error",
			},
			wantErr: true,
		},
		{
			name: "Error missing slug",
			fields: fields{
				httpClient: mock.HTTP{},
				baseURL:    "",
				filePath:   "flag-config.yaml",
			},
			wantErr: true,
		},
		{
			name: "Error missing file path",
			fields: fields{
				httpClient: mock.HTTP{},
				baseURL:    "https://gitlab.com/",
				filePath:   "",
			},
			wantErr: true,
		},
		{
			name: "Use gitlab token",
			fields: fields{
				httpClient:     mock.HTTP{},
				baseURL:        "https://gitlab.com",
				filePath:       "flag-config.yaml",
				gitlabToken:    "XXX",
				repositorySlug: "aa/go-feature-flags-config",
			},
			want:    []byte(sampleText()),
			wantErr: false,
		},
		{
			name: "Impossible to parse URL",
			fields: fields{
				httpClient:     mock.HTTP{},
				baseURL:        "https://user:abc{DEf1=ghi@example.com:5432/",
				filePath:       "flag-config.yaml",
				gitlabToken:    "XXX",
				repositorySlug: "aa/go-feature-flags-config",
			},
			want:    []byte(sampleText()),
			wantErr: true,
		},
		{
			name: "Use default values",
			fields: fields{
				httpClient:     mock.HTTP{},
				filePath:       "flag-config.yaml",
				repositorySlug: "aa/go-feature-flags-config",
			},
			want:    []byte(sampleText()),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := gitlabretriever.Retriever{
				BaseURL:        tt.fields.baseURL,
				FilePath:       tt.fields.filePath,
				GitlabToken:    tt.fields.gitlabToken,
				Branch:         tt.fields.branch,
				RepositorySlug: tt.fields.repositorySlug,
			}
			h.SetHTTPClient(&tt.fields.httpClient)
			got, err := h.Retrieve(tt.fields.context)
			assert.Equal(t, tt.wantErr, err != nil, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)
			if !tt.wantErr {
				assert.Equal(t, http.MethodGet, tt.fields.httpClient.Req.Method)
				assert.Equal(t, strings.TrimSpace(string(tt.want)), strings.TrimSpace(string(got)))
				if tt.fields.gitlabToken != "" {
					assert.Equal(t, tt.fields.gitlabToken, tt.fields.httpClient.Req.Header.Get("PRIVATE-TOKEN"))
				}
			}
		})
	}
}
