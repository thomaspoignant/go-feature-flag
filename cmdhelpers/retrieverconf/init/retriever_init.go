package init

import (
	"fmt"
	"time"

	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/retriever"
)

type DefaultRetrieverConfig struct {
	Timeout    time.Duration
	HTTPMethod string
	GitBranch  string
}

// InitRetriever initialize the retriever based on the configuration
func InitRetriever(
	c *retrieverconf.RetrieverConf, defaultRetrieverConfig DefaultRetrieverConfig) (retriever.Retriever, error) {
	if c.Timeout != 0 {
		defaultRetrieverConfig.Timeout = time.Duration(c.Timeout) * time.Millisecond
	}
	switch c.Kind {
	case retrieverconf.GitHubRetriever:
		return createGitHubRetriever(c, &defaultRetrieverConfig)
	case retrieverconf.GitlabRetriever:
		return createGitlabRetriever(c, &defaultRetrieverConfig)
	case retrieverconf.BitbucketRetriever:
		return createBitbucketRetriever(c, &defaultRetrieverConfig)
	case retrieverconf.FileRetriever:
		return createFileRetriever(c, nil)
	case retrieverconf.S3Retriever:
		return createS3Retriever(c, nil)
	case retrieverconf.HTTPRetriever:
		return createHTTPRetriever(c, &defaultRetrieverConfig)
	case retrieverconf.GoogleStorageRetriever:
		return createGoogleStorageRetriever(c, nil)
	case retrieverconf.KubernetesRetriever:
		return createKubernetesRetriever(c, nil)
	case retrieverconf.MongoDBRetriever:
		return createMongoDBRetriever(c, nil)
	case retrieverconf.RedisRetriever:
		return createRedisRetriever(c, nil)
	case retrieverconf.AzBlobStorageRetriever:
		return createAzBlobStorageRetriever(c, nil)
	case retrieverconf.PostgreSQLRetriever:
		return createPostgreSQLRetriever(c, nil)
	default:
		return nil, fmt.Errorf("invalid retriever: kind \"%s\" "+
			"is not supported", c.Kind)
	}
}
