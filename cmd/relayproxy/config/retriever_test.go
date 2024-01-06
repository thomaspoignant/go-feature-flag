package config_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

func TestRetrieverConf_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		fields   config.RetrieverConf
		wantErr  bool
		errValue string
	}{
		{
			name:     "no fields",
			fields:   config.RetrieverConf{},
			wantErr:  true,
			errValue: "invalid retriever: kind \"\" is not supported",
		},
		{
			name: "invalid kind",
			fields: config.RetrieverConf{
				Kind: "invalid",
			},
			wantErr:  true,
			errValue: "invalid retriever: kind \"invalid\" is not supported",
		},
		{
			name: "kind GitHubRetriever without repo slug",
			fields: config.RetrieverConf{
				Kind: "github",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"repositorySlug\" property found for kind \"github\"",
		},
		{
			name: "kind GitlabRetriever, with URL but without path",
			fields: config.RetrieverConf{
				Kind: "gitlab",
				URL:  "XXX",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"repositorySlug\" property found for kind \"gitlab\"",
		},
		{
			name: "kind GitlabRetriever, with URL but without path",
			fields: config.RetrieverConf{
				Kind:           "gitlab",
				RepositorySlug: "aaa/bbb",
				URL:            "XXX",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"path\" property found for kind \"gitlab\"",
		},
		{
			name: "kind S3Retriever without item",
			fields: config.RetrieverConf{
				Kind: "s3",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"item\" property found for kind \"s3\"",
		},
		{
			name: "kind HTTPRetriever without URL",
			fields: config.RetrieverConf{
				Kind: "http",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"url\" property found for kind \"http\"",
		},
		{
			name: "kind GCP without Object",
			fields: config.RetrieverConf{
				Kind: "googleStorage",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"object\" property found for kind \"googleStorage\"",
		},
		{
			name: "kind GitHubRetriever without path",
			fields: config.RetrieverConf{
				Kind:           "github",
				RepositorySlug: "thomaspoignant/go-feature-flag",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"path\" property found for kind \"github\"",
		},
		{
			name: "kind file without path",
			fields: config.RetrieverConf{
				Kind: "file",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"path\" property found for kind \"file\"",
		},
		{
			name: "kind s3 without bucket",
			fields: config.RetrieverConf{
				Kind: "s3",
				Item: "test.yaml",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"bucket\" property found for kind \"s3\"",
		},
		{
			name: "kind google storage without bucket",
			fields: config.RetrieverConf{
				Kind:   "googleStorage",
				Object: "test.yaml",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"bucket\" property found for kind \"googleStorage\"",
		},
		{
			name: "valid s3",
			fields: config.RetrieverConf{
				Kind:   "s3",
				Item:   "test.yaml",
				Bucket: "testBucket",
			},
		},
		{
			name: "valid googleStorage",
			fields: config.RetrieverConf{
				Kind:   "googleStorage",
				Object: "test.yaml",
				Bucket: "testBucket",
			},
		},
		{
			name: "valid github",
			fields: config.RetrieverConf{
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
			fields: config.RetrieverConf{
				Kind: "file",
				Path: "testdata/config.yaml",
			},
		},
		{
			name: "valid http",
			fields: config.RetrieverConf{
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
			fields: config.RetrieverConf{
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
			fields: config.RetrieverConf{
				Kind:      "configmap",
				Namespace: "xxx",
				ConfigMap: "xxx",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"key\" property found for kind \"configmap\"",
		},
		{
			name: "kind k8s configmap without ConfigMap",
			fields: config.RetrieverConf{
				Kind:      "configmap",
				Namespace: "xxx",
				Key:       "xxx",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"configmap\" property found for kind \"configmap\"",
		},
		{
			name: "kind mongoDB without URI",
			fields: config.RetrieverConf{
				Kind:       "mongodb",
				URI:        "",
				Collection: "xxx",
				Database:   "xxx",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"uri\" property found for kind \"mongodb\"",
		},
		{
			name: "kind mongoDB without Collection",
			fields: config.RetrieverConf{
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
			fields: config.RetrieverConf{
				Kind:       "mongodb",
				URI:        "xxx",
				Collection: "xxx",
				Database:   "",
			},
			wantErr:  true,
			errValue: "invalid retriever: no \"database\" property found for kind \"mongodb\"",
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
