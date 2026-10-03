package gitlabretriever_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/retriever/gitlabretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"
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
			assert.Equal(
				t,
				tt.wantErr,
				err != nil,
				"retrieve() error = %v, wantErr %v",
				err,
				tt.wantErr,
			)
			if !tt.wantErr {
				assert.Equal(t, http.MethodGet, tt.fields.httpClient.Req.Method)
				assert.Equal(t, strings.TrimSpace(string(tt.want)), strings.TrimSpace(string(got)))
				if tt.fields.gitlabToken != "" {
					assert.Equal(
						t,
						tt.fields.gitlabToken,
						tt.fields.httpClient.Req.Header.Get("PRIVATE-TOKEN"),
					)
				}
			}
		})
	}
}

func Test_gitlab_RetrievePathEscaping(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
	}{
		{
			name:     "simple filename",
			filePath: "flags.yaml",
		},
		{
			name:     "nested path",
			filePath: "config/flags.yaml",
		},
		{
			name:     "spaces in directory and filename",
			filePath: "flag configs/my flags.yaml",
		},
		{
			name:     "literal plus sign",
			filePath: "flags/my+flags.yaml",
		},
		{
			name:     "literal percent encoding",
			filePath: "flags/my%20flags.yaml",
		},
		{
			name:     "query and fragment delimiters",
			filePath: "flags/my?#flags.yaml",
		},
		{
			name:     "non-ASCII filename",
			filePath: "flags/caf\u00e9.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mock.HTTP{}
			r := gitlabretriever.Retriever{
				RepositorySlug: "team/subgroup/config",
				FilePath:       tt.filePath,
				Branch:         "feature/add+flags",
			}
			r.SetHTTPClient(client)

			_, err := r.Retrieve(context.Background())
			if !assert.NoError(t, err) {
				return
			}
			// The project slug and file path must each occupy a single URL path segment.
			assert.Len(t, strings.Split(client.Req.URL.EscapedPath(), "/"), 9)
			assert.Equal(t,
				"/api/v4/projects/team/subgroup/config/repository/files/"+tt.filePath+"/raw",
				client.Req.URL.Path)
			assert.Equal(t, "ref=feature%2Fadd%2Bflags", client.Req.URL.RawQuery)
			assert.Empty(t, client.Req.URL.Fragment)
		})
	}
}
