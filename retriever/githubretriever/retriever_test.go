package githubretriever_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"

	"github.com/stretchr/testify/assert"
)

func Test_github_Retrieve(t *testing.T) {
	type fields struct {
		httpClient     mock.HTTP
		context        context.Context
		repositorySlug string
		filePath       string
		githubToken    string
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
				repositorySlug: "thomaspoignant/go-feature-flag",
				filePath:       "testdata/flag-config.yaml",
			},
			want: []byte(`test-flag:
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
`),
			wantErr: false,
		},
		{
			name: "Success with context",
			fields: fields{
				httpClient:     mock.HTTP{},
				repositorySlug: "thomaspoignant/go-feature-flag",
				filePath:       "testdata/flag-config.yaml",
				context:        context.Background(),
			},
			want: []byte(`test-flag:
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
`),
			wantErr: false,
		},
		{
			name: "Success with default method",
			fields: fields{
				httpClient:     mock.HTTP{},
				repositorySlug: "thomaspoignant/go-feature-flag",
				filePath:       "testdata/flag-config.yaml",
			},
			want: []byte(`test-flag:
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
`),
			wantErr: false,
		},
		{
			name: "HTTP Error",
			fields: fields{
				httpClient:     mock.HTTP{},
				repositorySlug: "thomaspoignant/go-feature-flag",
				filePath:       "testdata/error",
			},
			wantErr: true,
		},
		{
			name: "Error missing slug",
			fields: fields{
				httpClient:     mock.HTTP{},
				repositorySlug: "",
				filePath:       "testdata/flag-config.yaml",
			},
			wantErr: true,
		},
		{
			name: "Error missing file path",
			fields: fields{
				httpClient:     mock.HTTP{},
				repositorySlug: "thomaspoignant/go-feature-flag",
				filePath:       "",
			},
			wantErr: true,
		},
		{
			name: "Use GitHub token",
			fields: fields{
				httpClient:     mock.HTTP{},
				repositorySlug: "thomaspoignant/go-feature-flag",
				filePath:       "testdata/flag-config.yaml",
				githubToken:    "XXX_GH_TOKEN",
			},
			want: []byte(`test-flag:
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
`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := githubretriever.Retriever{
				RepositorySlug: tt.fields.repositorySlug,
				FilePath:       tt.fields.filePath,
				GithubToken:    tt.fields.githubToken,
			}

			h.SetHTTPClient(&tt.fields.httpClient)
			got, err := h.Retrieve(tt.fields.context)
			assert.Equal(t, tt.wantErr, err != nil, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)
			if !tt.wantErr {
				assert.Equal(t, http.MethodGet, tt.fields.httpClient.Req.Method)
				assert.Equal(t, tt.want, got)
				if tt.fields.githubToken != "" {
					assert.Equal(t, "Bearer "+tt.fields.githubToken, tt.fields.httpClient.Req.Header.Get("Authorization"))
				}
			}
		})
	}
}
