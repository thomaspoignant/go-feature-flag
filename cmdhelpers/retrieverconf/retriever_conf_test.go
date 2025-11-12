package retrieverconf_test

import (
	"net/http"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
)

func TestRetrieverConf_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		fields   retrieverconf.RetrieverConf
		wantErr  bool
		errValue string
	}{
		{
			name:     "no fields",
			fields:   retrieverconf.RetrieverConf{},
			wantErr:  true,
			errValue: "invalid retriever: kind \"\" is not supported",
		},
		{
			name: "invalid kind",
			fields: retrieverconf.RetrieverConf{
				Kind: "invalid",
			},
			wantErr:  true,
			errValue: "invalid retriever: kind \"invalid\" is not supported",
		},
		{
			name: "kind GitHubRetriever without repo slug",
			fields: retrieverconf.RetrieverConf{
				Kind: "github",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"repositorySlug\" property found for kind \"github\"",
		},
		{
			name: "kind GitlabRetriever, with URL but without path",
			fields: retrieverconf.RetrieverConf{
				Kind: "gitlab",
				URL:  "XXX",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"repositorySlug\" property found for kind \"gitlab\"",
		},
		{
			name: "kind GitlabRetriever, with URL but without path",
			fields: retrieverconf.RetrieverConf{
				Kind:           "gitlab",
				RepositorySlug: "aaa/bbb",
				URL:            "XXX",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"path\" property found for kind \"gitlab\"",
		}, {
			name: "kind BitbucketRetriever without repo slug",
			fields: retrieverconf.RetrieverConf{
				Kind: "bitbucket",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"repositorySlug\" property found for kind \"bitbucket\"",
		},
		{
			name: "kind S3Retriever without item",
			fields: retrieverconf.RetrieverConf{
				Kind: "s3",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"item\" property found for kind \"s3\"",
		},
		{
			name: "kind HTTPRetriever without URL",
			fields: retrieverconf.RetrieverConf{
				Kind: "http",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"url\" property found for kind \"http\"",
		},
		{
			name: "kind GCP without Object",
			fields: retrieverconf.RetrieverConf{
				Kind: "googleStorage",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"object\" property found for kind \"googleStorage\"",
		},
		{
			name: "kind GitHubRetriever without path",
			fields: retrieverconf.RetrieverConf{
				Kind:           "github",
				RepositorySlug: "thomaspoignant/go-feature-flag",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"path\" property found for kind \"github\"",
		},
		{
			name: "kind file without path",
			fields: retrieverconf.RetrieverConf{
				Kind: "file",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"path\" property found for kind \"file\"",
		},
		{
			name: "kind s3 without bucket",
			fields: retrieverconf.RetrieverConf{
				Kind: "s3",
				Item: "test.yaml",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"bucket\" property found for kind \"s3\"",
		},
		{
			name: "kind google storage without bucket",
			fields: retrieverconf.RetrieverConf{
				Kind:   "googleStorage",
				Object: "test.yaml",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"bucket\" property found for kind \"googleStorage\"",
		},
		{
			name: "kind azureBlobStorage without object",
			fields: retrieverconf.RetrieverConf{
				Kind:        "azureBlobStorage",
				Container:   "testcontainer",
				AccountName: "devstoreaccount1",
				AccountKey:  "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"object\" property found for kind \"azureBlobStorage\"",
		},
		{
			name: "kind azureBlobStorage without accountName",
			fields: retrieverconf.RetrieverConf{
				Kind:       "azureBlobStorage",
				Container:  "testcontainer",
				Object:     "flag-config.yaml",
				AccountKey: "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"accountName\" property found for kind \"azureBlobStorage\"",
		},
		{
			name: "kind azureBlobStorage without container",
			fields: retrieverconf.RetrieverConf{
				Kind:        "azureBlobStorage",
				Object:      "flag-config.yaml",
				AccountName: "devstoreaccount1",
				AccountKey:  "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"container\" property found for kind \"azureBlobStorage\"",
		},
		{
			name: "valid s3",
			fields: retrieverconf.RetrieverConf{
				Kind:   "s3",
				Item:   "test.yaml",
				Bucket: "testBucket",
			},
		},
		{
			name: "valid googleStorage",
			fields: retrieverconf.RetrieverConf{
				Kind:   "googleStorage",
				Object: "test.yaml",
				Bucket: "testBucket",
			},
		},
		{
			name: "valid github",
			fields: retrieverconf.RetrieverConf{
				Kind:           "github",
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Branch:         "main",
				Path:           "testdata/config.yaml",
				GithubToken:    "XXX",
				Timeout:        5000,
			},
		},
		{
			name: "valid file",
			fields: retrieverconf.RetrieverConf{
				Kind: "file",
				Path: "testdata/config.yaml",
			},
		},
		{
			name: "valid http",
			fields: retrieverconf.RetrieverConf{
				Kind:       "http",
				URL:        "http://perdu.com/flags",
				HTTPMethod: http.MethodGet,
				HTTPBody:   `{"yo"": "yo"}`,
				HTTPHeaders: map[string][]string{
					"Test": {"Val1"},
				},
				Timeout: 5000,
			},
		},
		{
			name: "kind k8s configmap without namespace",
			fields: retrieverconf.RetrieverConf{
				Kind:      "configmap",
				Namespace: "",
				Key:       "xxx",
				ConfigMap: "xxx",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"namespace\" property found for kind \"configmap\"",
		},
		{
			name: "kind k8s configmap without key",
			fields: retrieverconf.RetrieverConf{
				Kind:      "configmap",
				Namespace: "xxx",
				ConfigMap: "xxx",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"key\" property found for kind \"configmap\"",
		},
		{
			name: "kind k8s configmap without ConfigMap",
			fields: retrieverconf.RetrieverConf{
				Kind:      "configmap",
				Namespace: "xxx",
				Key:       "xxx",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"configmap\" property found for kind \"configmap\"",
		},
		{
			name: "kind k8s valid",
			fields: retrieverconf.RetrieverConf{
				Kind:      "configmap",
				Namespace: "xxx",
				Key:       "xxx",
				ConfigMap: "xxx",
			},
		},
		{
			name: "kind mongoDB without URI",
			fields: retrieverconf.RetrieverConf{
				Kind:       "mongodb",
				URI:        "",
				Collection: "xxx",
				Database:   "xxx",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"uri\" property found for kind \"mongodb\"",
		},
		{
			name: "kind redis without options (old RedisOptions)",
			fields: retrieverconf.RetrieverConf{
				Kind:        "redis",
				RedisPrefix: "xxx",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"redis\" property found for kind \"redis\"",
		},
		{
			name: "kind redis with old RedisOptions (backward compatibility)",
			fields: retrieverconf.RetrieverConf{
				Kind: "redis",
				RedisOptions: &redis.Options{
					Addr: "localhost:6379",
				},
				RedisPrefix: "xxx",
			},
			wantErr: false,
		},
		{
			name: "kind redis with new SerializableRedisOptions",
			fields: retrieverconf.RetrieverConf{
				Kind: "redis",
				Redis: &retrieverconf.SerializableRedisOptions{
					Addr: "localhost:6379",
				},
				RedisPrefix: "xxx",
			},
			wantErr: false,
		},
		{
			name: "kind redis with both options (new takes priority)",
			fields: retrieverconf.RetrieverConf{
				Kind: "redis",
				RedisOptions: &redis.Options{
					Addr: "old:6379",
				},
				Redis: &retrieverconf.SerializableRedisOptions{
					Addr: "new:6379",
				},
				RedisPrefix: "xxx",
			},
			wantErr: false,
		},
		{
			name: "kind redis with RedisOptions but empty Addr",
			fields: retrieverconf.RetrieverConf{
				Kind: "redis",
				RedisOptions: &redis.Options{
					Addr: "",
				},
				RedisPrefix: "xxx",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"redis.addr\" property found for kind \"redis\"",
		},
		{
			name: "kind redis with new Redis but empty Addr",
			fields: retrieverconf.RetrieverConf{
				Kind: "redis",
				Redis: &retrieverconf.SerializableRedisOptions{
					Addr: "",
				},
				RedisPrefix: "xxx",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"redis.addr\" property found for kind \"redis\"",
		},
		{
			name: "kind mongoDB without Collection",
			fields: retrieverconf.RetrieverConf{
				Kind:       "mongodb",
				URI:        "xxx",
				Collection: "",
				Database:   "xxx",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"collection\" property found for kind \"mongodb\"",
		},
		{
			name: "kind mongoDB without database",
			fields: retrieverconf.RetrieverConf{
				Kind:       "mongodb",
				URI:        "xxx",
				Collection: "xxx",
				Database:   "",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"database\" property found for kind \"mongodb\"",
		},
		{
			name: "kind mongoDB valid",
			fields: retrieverconf.RetrieverConf{
				Kind:       "mongodb",
				URI:        "xxx",
				Collection: "xxx",
				Database:   "xxx",
			},
		},
		{
			name: "kind postgresql valid",
			fields: retrieverconf.RetrieverConf{
				Kind:  "postgresql",
				URI:   "xxx",
				Table: "xxx",
			},
		},
		{
			name: "kind postgresql invalid without URI",
			fields: retrieverconf.RetrieverConf{
				Kind:  "postgresql",
				Table: "xxx",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"uri\" property found for kind \"postgresql\"",
		},
		{
			name: "kind postgresql invalid without Table",
			fields: retrieverconf.RetrieverConf{
				Kind: "postgresql",
				URI:  "xxx",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"table\" property found for kind \"postgresql\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields
			err := c.IsValid()
			assert.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				assert.Equal(t, tt.errValue, err.Error())
			}
		})
	}
}
