package githubretriever_test

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"
)

func Test_github_Retrieve(t *testing.T) {
	endRatelimit := time.Now().Add(1 * time.Hour)
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
		errMsg  string
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
			name: "Ratelimiting",
			fields: fields{
				httpClient:     mock.HTTP{RateLimit: true, EndRatelimit: endRatelimit},
				repositorySlug: "thomaspoignant/go-feature-flag",
				filePath:       "testdata/flag-config.yaml",
			},
			errMsg: "request to https://api.github.com/repos/thomaspoignant/go-feature-flag/contents/testdata/flag-config.yaml?ref=main failed with code 429. GitHub Headers: map[X-Content-Type-Options:nosniff X-Frame-Options:deny X-Github-Media-Type:github.v3; format=json X-Github-Request-Id:F82D:37B98C:232EF263:235C93BD:6650BDC6 X-Ratelimit-Limit:60 X-Ratelimit-Remaining:0 X-Ratelimit-Reset:" + strconv.FormatInt(
				endRatelimit.Unix(),
				10,
			) + " X-Ratelimit-Resource:core X-Ratelimit-Used:60 X-Xss-Protection:1; mode=block]",
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
			if tt.errMsg != "" {
				assert.EqualError(t, err, tt.errMsg)
			}
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
				assert.Equal(t, tt.want, got)
				if tt.fields.githubToken != "" {
					assert.Equal(
						t,
						"Bearer "+tt.fields.githubToken,
						tt.fields.httpClient.Req.Header.Get("Authorization"),
					)
				}
			}
		})
	}
}

func Test_github_Retrieve_BaseURL(t *testing.T) {
	type fields struct {
		httpClient     mock.HTTP
		context        context.Context
		repositorySlug string
		filePath       string
		githubToken    string
		baseURL        string
	}
	tests := []struct {
		name        string
		fields      fields
		want        []byte
		wantErr     bool
		expectedURL string
	}{
		{
			name: "GitHub Enterprise base URL",
			fields: fields{
				httpClient:     mock.HTTP{},
				repositorySlug: "myorg/myrepo",
				filePath:       "config/flags.yaml",
				githubToken:    "ghp_token",
				baseURL:        "https://github.acme.com/api/v3",
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
			wantErr:     false,
			expectedURL: "https://github.acme.com/api/v3/repos/myorg/myrepo/contents/config/flags.yaml?ref=main",
		},
		{
			name: "Default GitHub API URL (empty BaseURL)",
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
			wantErr:     false,
			expectedURL: "https://api.github.com/repos/thomaspoignant/go-feature-flag/contents/testdata/flag-config.yaml?ref=main",
		},
		{
			name: "GitHub Enterprise with custom branch",
			fields: fields{
				httpClient:     mock.HTTP{},
				repositorySlug: "myorg/myrepo",
				filePath:       "config/flags.yaml",
				baseURL:        "https://github.enterprise.com/api/v3",
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
			wantErr:     false,
			expectedURL: "https://github.enterprise.com/api/v3/repos/myorg/myrepo/contents/config/flags.yaml?ref=main",
		},
		{
			name: "GitHub Enterprise with trailing slash in BaseURL",
			fields: fields{
				httpClient:     mock.HTTP{},
				repositorySlug: "myorg/myrepo",
				filePath:       "config/flags.yaml",
				baseURL:        "https://github.acme.com/api/v3/",
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
			wantErr:     false,
			expectedURL: "https://github.acme.com/api/v3/repos/myorg/myrepo/contents/config/flags.yaml?ref=main",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := githubretriever.Retriever{
				RepositorySlug: tt.fields.repositorySlug,
				FilePath:       tt.fields.filePath,
				GithubToken:    tt.fields.githubToken,
				BaseURL:        tt.fields.baseURL,
			}

			h.SetHTTPClient(&tt.fields.httpClient)
			got, err := h.Retrieve(tt.fields.context)
			assert.Equal(t, tt.wantErr, err != nil, "retrieve() error = %v, wantErr %v", err, tt.wantErr)
			if !tt.wantErr {
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.expectedURL, tt.fields.httpClient.Req.URL.String())
				if tt.fields.githubToken != "" {
					assert.Equal(
						t,
						"Bearer "+tt.fields.githubToken,
						tt.fields.httpClient.Req.Header.Get("Authorization"),
					)
				}
			}
		})
	}
}

func TestRateLimiting(t *testing.T) {
	h := githubretriever.Retriever{
		RepositorySlug: "thomaspoignant/go-feature-flag",
		FilePath:       "testdata/flag-config.yaml",
	}

	httpClient := &mock.HTTP{}
	h.SetHTTPClient(httpClient)
	_, err := h.Retrieve(context.TODO())
	assert.NoError(t, err)
	assert.True(t, httpClient.HasBeenCalled)

	httpClient = &mock.HTTP{RateLimit: true}
	h.SetHTTPClient(httpClient)
	_, err = h.Retrieve(context.TODO())
	assert.Error(t, err)
	assert.True(t, httpClient.HasBeenCalled)

	httpClient = &mock.HTTP{}
	h.SetHTTPClient(httpClient)
	_, err = h.Retrieve(context.TODO())
	assert.Error(t, err)
	assert.False(t, httpClient.HasBeenCalled)
}
