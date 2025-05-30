package bitbucketretriever_test

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/retriever/bitbucketretriever"
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

func Test_bitbucket_Retrieve(t *testing.T) {
	endRatelimit := time.Now().Add(1 * time.Hour)
	type fields struct {
		httpClient     mock.HTTP
		context        context.Context
		repositorySlug string
		filePath       string
		bitbucketToken string
		branch         string
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
				repositorySlug: "gofeatureflag/config-repo",
				filePath:       "flags/config.goff.yaml",
			},
			want:    []byte(sampleText()),
			wantErr: false,
		},
		{
			name: "Success with context",
			fields: fields{
				httpClient:     mock.HTTP{},
				repositorySlug: "gofeatureflag/config-repo",
				filePath:       "flags/config.goff.yaml",
				context:        context.Background(),
			},
			want:    []byte(sampleText()),
			wantErr: false,
		},
		{
			name: "HTTP Error",
			fields: fields{
				httpClient:     mock.HTTP{},
				repositorySlug: "gofeatureflag/config-repo",
				filePath:       "flags/error",
			},
			wantErr: true,
		},
		{
			name: "Error missing slug",
			fields: fields{
				httpClient: mock.HTTP{},
				filePath:   "tests/__init__.py",
				branch:     "main",
			},
			wantErr: true,
		},
		{
			name: "Error missing file path",
			fields: fields{
				httpClient:     mock.HTTP{},
				repositorySlug: "gofeatureflag/config-repo",
				filePath:       "",
			},
			wantErr: true,
		},
		{
			name: "Rate limiting",
			fields: fields{
				httpClient:     mock.HTTP{RateLimit: true, EndRatelimit: endRatelimit},
				repositorySlug: "gofeatureflag/config-repo",
				filePath:       "flags/config.goff.yaml",
			},
			wantErr: true,
			errMsg: "request to https://api.bitbucket.org/2.0/repositories/gofeatureflag/config-repo/src/main/flags/config.goff.yaml failed with code 429. Bitbucket Headers: map[X-Content-Type-Options:nosniff X-Frame-Options:deny X-Github-Media-Type:github.v3; format=json X-Github-Request-Id:F82D:37B98C:232EF263:235C93BD:6650BDC6 X-Ratelimit-Limit:60 X-Ratelimit-Remaining:0 X-Ratelimit-Reset:" + strconv.FormatInt(
				endRatelimit.Unix(),
				10,
			) + " X-Ratelimit-Resource:core X-Ratelimit-Used:60 X-Xss-Protection:1; mode=block]",
		},
		{
			name: "Use Bitbucket token",
			fields: fields{
				httpClient:     mock.HTTP{},
				repositorySlug: "gofeatureflag/config-repo",
				filePath:       "flags/config.goff.yaml",
				bitbucketToken: "XXX_BITBUCKET_TOKEN",
			},
			want:    []byte(sampleText()),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := bitbucketretriever.Retriever{
				RepositorySlug: tt.fields.repositorySlug,
				Branch:         tt.fields.branch,
				FilePath:       tt.fields.filePath,
				BitBucketToken: tt.fields.bitbucketToken,
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
				"retrieve() error = %v wantErr %v",
				err,
				tt.wantErr,
			)
			if !tt.wantErr {
				assert.Equal(t, http.MethodGet, tt.fields.httpClient.Req.Method)
				assert.Equal(t, strings.TrimSpace(string(tt.want)), strings.TrimSpace(string(got)))
				if tt.fields.bitbucketToken != "" {
					assert.Equal(
						t,
						"Bearer "+tt.fields.bitbucketToken,
						tt.fields.httpClient.Req.Header.Get("Authorization"),
					)
				}
			}
		})
	}
}
