package init

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/bitbucketretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gcstorageretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gitlabretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retrieverv2"
)

var defaultRetrieverConfig = DefaultRetrieverConfig{
	Timeout:    10 * time.Second,
	HTTPMethod: http.MethodGet,
	GitBranch:  "main",
}

func Test_initRetriever(t *testing.T) {
	tests := []struct {
		name                   string
		conf                   *retrieverconf.RetrieverConf
		want                   retriever.Retriever
		wantErr                assert.ErrorAssertionFunc
		wantType               retriever.Retriever
		skipCompleteValidation bool
	}{
		{
			name:    "Convert Github Retriever",
			wantErr: assert.NoError,
			conf: &retrieverconf.RetrieverConf{
				Kind:           "github",
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Path:           "testdata/flag-config.yaml",
				Timeout:        20,
			},
			want: &githubretriever.Retriever{
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Branch:         "main",
				FilePath:       "testdata/flag-config.yaml",
				GithubToken:    "",
				Timeout:        20 * time.Millisecond,
			},
			wantType: &githubretriever.Retriever{},
		},
		{
			name:    "Convert Github Retriever with token",
			wantErr: assert.NoError,
			conf: &retrieverconf.RetrieverConf{
				Kind:           "github",
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Path:           "testdata/flag-config.yaml",
				Timeout:        20,
				AuthToken:      "xxx",
			},
			want: &githubretriever.Retriever{
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Branch:         "main",
				FilePath:       "testdata/flag-config.yaml",
				GithubToken:    "xxx",
				Timeout:        20 * time.Millisecond,
			},
			wantType: &githubretriever.Retriever{},
		},
		{
			name:    "Convert Github Retriever with deprecated token",
			wantErr: assert.NoError,
			conf: &retrieverconf.RetrieverConf{
				Kind:           "github",
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Path:           "testdata/flag-config.yaml",
				Timeout:        20,
				GithubToken:    "xxx",
			},
			want: &githubretriever.Retriever{
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Branch:         "main",
				FilePath:       "testdata/flag-config.yaml",
				GithubToken:    "xxx",
				Timeout:        20 * time.Millisecond,
			},
			wantType: &githubretriever.Retriever{},
		},
		{
			name:    "Convert Gitlab Retriever",
			wantErr: assert.NoError,
			conf: &retrieverconf.RetrieverConf{
				Kind:           "gitlab",
				BaseURL:        "http://localhost",
				Path:           "flag-config.yaml",
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Timeout:        20,
			},
			want: &gitlabretriever.Retriever{
				BaseURL:        "http://localhost",
				Branch:         "main",
				FilePath:       "flag-config.yaml",
				RepositorySlug: "thomaspoignant/go-feature-flag",
				GitlabToken:    "",
				Timeout:        20 * time.Millisecond,
			},
			wantType: &gitlabretriever.Retriever{},
		},
		{
			name:    "Convert File Retriever",
			wantErr: assert.NoError,
			conf: &retrieverconf.RetrieverConf{
				Kind: "file",
				Path: "testdata/flag-config.yaml",
			},
			want:     &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
			wantType: &fileretriever.Retriever{},
		},
		{
			name:    "Convert S3 Retriever",
			wantErr: assert.NoError,
			conf: &retrieverconf.RetrieverConf{
				Kind:   "s3",
				Bucket: "my-bucket-name",
				Item:   "testdata/flag-config.yaml",
			},
			want: &s3retrieverv2.Retriever{
				Bucket: "my-bucket-name",
				Item:   "testdata/flag-config.yaml",
			},
			wantType:               &s3retrieverv2.Retriever{},
			skipCompleteValidation: true,
		},
		{
			name:    "Convert HTTP Retriever",
			wantErr: assert.NoError,
			conf: &retrieverconf.RetrieverConf{
				Kind: "http",
				URL:  "https://gofeatureflag.org/my-flag-test.yaml",
			},
			want: &httpretriever.Retriever{
				URL:     "https://gofeatureflag.org/my-flag-test.yaml",
				Method:  http.MethodGet,
				Body:    "",
				Header:  nil,
				Timeout: 10000000000,
			},
			wantType: &httpretriever.Retriever{},
		},
		{
			name:    "Convert Google storage Retriever",
			wantErr: assert.NoError,
			conf: &retrieverconf.RetrieverConf{
				Kind:   "googleStorage",
				Bucket: "my-bucket-name",
				Object: "testdata/flag-config.yaml",
			},
			want: &gcstorageretriever.Retriever{
				Bucket:  "my-bucket-name",
				Object:  "testdata/flag-config.yaml",
				Options: nil,
			},
			wantType: &gcstorageretriever.Retriever{},
		},
		{
			name:    "Convert unknown Retriever",
			wantErr: assert.Error,
			conf: &retrieverconf.RetrieverConf{
				Kind: "unknown",
			},
		},
		{
			name:    "Convert Bitbucket Retriever default branch",
			wantErr: assert.NoError,
			conf: &retrieverconf.RetrieverConf{
				Kind:           "bitbucket",
				RepositorySlug: "gofeatureflag/config-repo",
				Path:           "flags/config.goff.yaml",
				AuthToken:      "XXX_BITBUCKET_TOKEN",
				BaseURL:        "https://api.bitbucket.goff.org",
			},
			want: &bitbucketretriever.Retriever{
				RepositorySlug: "gofeatureflag/config-repo",
				Branch:         "main",
				FilePath:       "flags/config.goff.yaml",
				BitBucketToken: "XXX_BITBUCKET_TOKEN",
				BaseURL:        "https://api.bitbucket.goff.org",
				Timeout:        10000000000,
			},
			wantType: &bitbucketretriever.Retriever{},
		},
		{
			name:    "Convert Bitbucket Retriever branch specified",
			wantErr: assert.NoError,
			conf: &retrieverconf.RetrieverConf{
				Kind:           "bitbucket",
				Branch:         "develop",
				RepositorySlug: "gofeatureflag/config-repo",
				Path:           "flags/config.goff.yaml",
				AuthToken:      "XXX_BITBUCKET_TOKEN",
				BaseURL:        "https://api.bitbucket.goff.org",
			},
			want: &bitbucketretriever.Retriever{
				RepositorySlug: "gofeatureflag/config-repo",
				Branch:         "develop",
				FilePath:       "flags/config.goff.yaml",
				BitBucketToken: "XXX_BITBUCKET_TOKEN",
				BaseURL:        "https://api.bitbucket.goff.org",
				Timeout:        10000000000,
			},
			wantType: &bitbucketretriever.Retriever{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitRetriever(tt.conf, defaultRetrieverConfig)
			tt.wantErr(t, err)
			if err == nil {
				assert.IsType(t, tt.wantType, got)
				if !tt.skipCompleteValidation {
					assert.Equal(t, tt.want, got)
				}
			}
		})
	}
}
