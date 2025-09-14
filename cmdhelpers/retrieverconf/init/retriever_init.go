package init

import (
	"context"
	"fmt"
	"time"

	awsConf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gcstorageretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/redisretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retrieverv2"
)

type DefaultRetrieverConfig struct {
	Timeout    time.Duration
	HTTPMethod string
	GitBranch  string
}

// InitRetriever initialize the retriever based on the configuration
func InitRetriever(
	c *retrieverconf.RetrieverConf, defaultRetrieverConfig DefaultRetrieverConfig) (retriever.Retriever, error) {
	var retrieverTimeout = defaultRetrieverConfig.Timeout
	if c.Timeout != 0 {
		retrieverTimeout = time.Duration(c.Timeout) * time.Millisecond
	}
	switch c.Kind {
	case retrieverconf.GitHubRetriever:
		return initGithubRetriever(c, retrieverTimeout, defaultRetrieverConfig.GitBranch), nil
	case retrieverconf.GitlabRetriever:
		return initGitlabRetriever(c, retrieverTimeout, defaultRetrieverConfig.GitBranch), nil
	case retrieverconf.BitbucketRetriever:
		return initBitbucketRetriever(c, retrieverTimeout, defaultRetrieverConfig.GitBranch), nil
	case retrieverconf.FileRetriever:
		return &fileretriever.Retriever{Path: c.Path}, nil
	case retrieverconf.S3Retriever:
		awsConfig, err := awsConf.LoadDefaultConfig(context.Background())
		return &s3retrieverv2.Retriever{Bucket: c.Bucket, Item: c.Item, AwsConfig: &awsConfig}, err
	case retrieverconf.HTTPRetriever:
		return initHTTPRetriever(c, retrieverTimeout, defaultRetrieverConfig.HTTPMethod), nil
	case retrieverconf.GoogleStorageRetriever:
		return &gcstorageretriever.Retriever{Bucket: c.Bucket, Object: c.Object}, nil
	case retrieverconf.KubernetesRetriever:
		return initK8sRetriever(c)
	case retrieverconf.MongoDBRetriever:
		return initMongoRetriever(c), nil
	case retrieverconf.RedisRetriever:
		return &redisretriever.Retriever{Options: c.RedisOptions, Prefix: c.RedisPrefix}, nil
	case retrieverconf.AzBlobStorageRetriever:
		return initAzBlobRetriever(c), nil
	default:
		return nil, fmt.Errorf("invalid retriever: kind \"%s\" "+
			"is not supported", c.Kind)
	}
}
