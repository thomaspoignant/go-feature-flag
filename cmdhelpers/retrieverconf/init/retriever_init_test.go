package init

import (
	"net/http"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/bitbucketretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gcstorageretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gitlabretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/postgresqlretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/redisretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retrieverv2"
)

func Test_InitRetriever(t *testing.T) {
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
		{
			name:    "Convert Postgres Retriever",
			wantErr: assert.NoError,
			conf: &retrieverconf.RetrieverConf{
				Kind:    "postgresql",
				URI:     "postgresql://user:password@localhost:5432/database",
				Table:   "flags",
				Columns: map[string]string{"flagset": "settings"},
			},
			want: &postgresqlretriever.Retriever{
				URI:     "postgresql://user:password@localhost:5432/database",
				Table:   "flags",
				Columns: map[string]string{"flagset": "settings"},
			},
			wantType: &postgresqlretriever.Retriever{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitRetriever(tt.conf)
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

func Test_InitRetriever_Redis(t *testing.T) {
	tests := []struct {
		name              string
		conf              *retrieverconf.RetrieverConf
		wantErr           bool
		validateRetriever func(t *testing.T, r retriever.Retriever)
	}{
		{
			name: "Redis with RedisOptions",
			conf: &retrieverconf.RetrieverConf{
				Kind: retrieverconf.RedisRetriever,
				RedisOptions: &retrieverconf.SerializableRedisOptions{
					Addr:     "localhost:6379",
					Password: "secret",
					DB:       1,
				},
				RedisPrefix: "test:",
			},
			wantErr: false,
			validateRetriever: func(t *testing.T, r retriever.Retriever) {
				redisRet, ok := r.(*redisretriever.Retriever)
				require.True(t, ok, "expected *redisretriever.Retriever")
				assert.Equal(t, "localhost:6379", redisRet.Options.Addr)
				assert.Equal(t, "secret", redisRet.Options.Password)
				assert.Equal(t, 1, redisRet.Options.DB)
				assert.Equal(t, "test:", redisRet.Prefix)
			},
		},
		{
			name: "Redis with SerializableRedisOptions with username",
			conf: &retrieverconf.RetrieverConf{
				Kind: retrieverconf.RedisRetriever,
				RedisOptions: &retrieverconf.SerializableRedisOptions{
					Addr:     "redis.example.com:6380",
					Password: "newsecret",
					DB:       2,
					Username: "admin",
				},
				RedisPrefix: "flags:",
			},
			wantErr: false,
			validateRetriever: func(t *testing.T, r retriever.Retriever) {
				redisRet, ok := r.(*redisretriever.Retriever)
				require.True(t, ok, "expected *redisretriever.Retriever")
				assert.Equal(t, "redis.example.com:6380", redisRet.Options.Addr)
				assert.Equal(t, "newsecret", redisRet.Options.Password)
				assert.Equal(t, 2, redisRet.Options.DB)
				assert.Equal(t, "admin", redisRet.Options.Username)
				assert.Equal(t, "flags:", redisRet.Prefix)
			},
		},
		{
			name: "Redis with RedisOptions password",
			conf: &retrieverconf.RetrieverConf{
				Kind:        retrieverconf.RedisRetriever,
				RedisOptions: &retrieverconf.SerializableRedisOptions{
					Addr:     "new:6380",
					Password: "new-secret",
				},
				RedisPrefix: "test:",
			},
			wantErr: false,
			validateRetriever: func(t *testing.T, r retriever.Retriever) {
				redisRet, ok := r.(*redisretriever.Retriever)
				require.True(t, ok, "expected *redisretriever.Retriever")
				assert.Equal(t, "new:6380", redisRet.Options.Addr)
				assert.Equal(t, "new-secret", redisRet.Options.Password)
				assert.Equal(t, "test:", redisRet.Prefix)
			},
		},
		{
			name: "Redis with full SerializableRedisOptions",
			conf: &retrieverconf.RetrieverConf{
				Kind: retrieverconf.RedisRetriever,
				RedisOptions: &retrieverconf.SerializableRedisOptions{
					Addr:              "redis:6379",
					Password:          "pass",
					DB:                3,
					MaxRetries:        5,
					DialTimeoutMs:     10000,
					ReadTimeoutMs:     5000,
					PoolSize:          20,
					MinIdleConns:      5,
					ClientName:        "test-client",
					ContextTimeoutEnabled: true,
				},
			},
			wantErr: false,
			validateRetriever: func(t *testing.T, r retriever.Retriever) {
				redisRet, ok := r.(*redisretriever.Retriever)
				require.True(t, ok, "expected *redisretriever.Retriever")
				assert.Equal(t, "redis:6379", redisRet.Options.Addr)
				assert.Equal(t, "pass", redisRet.Options.Password)
				assert.Equal(t, 3, redisRet.Options.DB)
				assert.Equal(t, 5, redisRet.Options.MaxRetries)
				assert.Equal(t, 10000*time.Millisecond, redisRet.Options.DialTimeout)
				assert.Equal(t, 5000*time.Millisecond, redisRet.Options.ReadTimeout)
				assert.Equal(t, 20, redisRet.Options.PoolSize)
				assert.Equal(t, 5, redisRet.Options.MinIdleConns)
				assert.Equal(t, "test-client", redisRet.Options.ClientName)
				assert.True(t, redisRet.Options.ContextTimeoutEnabled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitRetriever(tt.conf)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.IsType(t, &redisretriever.Retriever{}, got)
				if tt.validateRetriever != nil {
					tt.validateRetriever(t, got)
				}
			}
		})
	}
}
