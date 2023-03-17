package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gcstorageretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retriever"
	"net/http"
	"testing"
	"time"
)

func Test_initRetriever(t *testing.T) {
	tests := []struct {
		name    string
		conf    *config.RetrieverConf
		want    retriever.Retriever
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Convert Github Retriever",
			wantErr: assert.NoError,
			conf: &config.RetrieverConf{
				Kind:           "github",
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Path:           "testdata/flag-config.yaml",
			},
			want: &githubretriever.Retriever{
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Branch:         "main",
				FilePath:       "testdata/flag-config.yaml",
				GithubToken:    "",
				Timeout:        10 * time.Second,
			},
		},
		{
			name:    "Convert File Retriever",
			wantErr: assert.NoError,
			conf: &config.RetrieverConf{
				Kind: "file",
				Path: "testdata/flag-config.yaml",
			},
			want: &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
		},
		{
			name:    "Convert S3 Retriever",
			wantErr: assert.NoError,
			conf: &config.RetrieverConf{
				Kind:   "s3",
				Bucket: "my-bucket-name",
				Item:   "testdata/flag-config.yaml",
			},
			want: &s3retriever.Retriever{
				Bucket: "my-bucket-name",
				Item:   "testdata/flag-config.yaml",
			},
		},
		{
			name:    "Convert HTTP Retriever",
			wantErr: assert.NoError,
			conf: &config.RetrieverConf{
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
		}, {
			name:    "Convert Google storage Retriever",
			wantErr: assert.NoError,
			conf: &config.RetrieverConf{
				Kind:   "googleStorage",
				Bucket: "my-bucket-name",
				Object: "testdata/flag-config.yaml",
			},
			want: &gcstorageretriever.Retriever{
				Bucket:  "my-bucket-name",
				Object:  "testdata/flag-config.yaml",
				Options: nil,
			},
		},
		{
			name:    "Convert unknown Retriever",
			wantErr: assert.Error,
			conf: &config.RetrieverConf{
				Kind: "unknown",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := initRetriever(tt.conf)
			tt.wantErr(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
