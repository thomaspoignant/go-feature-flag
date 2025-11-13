package init

import (
	"context"
	"fmt"
	"time"

	awsConf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/redis/go-redis/v9"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	azblobretriever "github.com/thomaspoignant/go-feature-flag/retriever/azblobstorageretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/bitbucketretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gcstorageretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gitlabretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/k8sretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/mongodbretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/postgresqlretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/redisretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retrieverv2"
	"k8s.io/client-go/rest"
)

// retrieverFactory defines the signature for retriever factory functions
type retrieverFactory func(*retrieverconf.RetrieverConf, time.Duration) (retriever.Retriever, error)

// retrieverFactories maps retriever kinds to their factory functions
var retrieverFactories = map[retrieverconf.RetrieverKind]retrieverFactory{
	retrieverconf.GitHubRetriever:        createGitHubRetriever,
	retrieverconf.GitlabRetriever:        createGitlabRetriever,
	retrieverconf.BitbucketRetriever:     createBitbucketRetriever,
	retrieverconf.FileRetriever:          createFileRetriever,
	retrieverconf.S3Retriever:            createS3Retriever,
	retrieverconf.HTTPRetriever:          createHTTPRetriever,
	retrieverconf.GoogleStorageRetriever: createGoogleStorageRetriever,
	retrieverconf.KubernetesRetriever:    createKubernetesRetriever,
	retrieverconf.MongoDBRetriever:       createMongoDBRetriever,
	retrieverconf.RedisRetriever:         createRedisRetriever,
	retrieverconf.AzBlobStorageRetriever: createAzBlobStorageRetriever,
	retrieverconf.PostgreSQLRetriever:    createPostgreSQLRetriever,
}

// InitRetriever initialize the retriever based on the configuration
func InitRetriever(
	c *retrieverconf.RetrieverConf) (retriever.Retriever, error) {
	retrieverTimeout := retrieverconf.DefaultRetrieverConfig.Timeout
	if c.Timeout != 0 {
		retrieverTimeout = time.Duration(c.Timeout) * time.Millisecond
	}
	retrieverFactory, exists := retrieverFactories[c.Kind]
	if !exists {
		return nil, fmt.Errorf("invalid retriever: kind \"%s\" is not supported", c.Kind)
	}
	return retrieverFactory(c, retrieverTimeout)
}

// Factory functions for each retriever type
func createGitHubRetriever(
	c *retrieverconf.RetrieverConf, timeout time.Duration) (retriever.Retriever, error) {
	token := c.AuthToken
	if token == "" && c.GithubToken != "" { // nolint: staticcheck
		token = c.GithubToken // nolint: staticcheck
	}
	return &githubretriever.Retriever{
		RepositorySlug: c.RepositorySlug,
		Branch: func() string {
			if c.Branch == "" {
				return retrieverconf.DefaultRetrieverConfig.GitBranch
			}
			return c.Branch
		}(),
		FilePath:    c.Path,
		GithubToken: token,
		Timeout:     timeout,
	}, nil
}

func createGitlabRetriever(
	c *retrieverconf.RetrieverConf, timeout time.Duration) (retriever.Retriever, error) {
	return &gitlabretriever.Retriever{
		BaseURL: c.BaseURL,
		Branch: func() string {
			if c.Branch == "" {
				return retrieverconf.DefaultRetrieverConfig.GitBranch
			}
			return c.Branch
		}(),
		FilePath:       c.Path,
		GitlabToken:    c.AuthToken,
		RepositorySlug: c.RepositorySlug,
		Timeout:        timeout,
	}, nil
}

func createBitbucketRetriever(
	c *retrieverconf.RetrieverConf, timeout time.Duration) (retriever.Retriever, error) {
	return &bitbucketretriever.Retriever{
		RepositorySlug: c.RepositorySlug,
		Branch: func() string {
			if c.Branch == "" {
				return retrieverconf.DefaultRetrieverConfig.GitBranch
			}
			return c.Branch
		}(),
		FilePath:       c.Path,
		BitBucketToken: c.AuthToken,
		BaseURL:        c.BaseURL,
		Timeout:        timeout,
	}, nil
}

func createFileRetriever(c *retrieverconf.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	return &fileretriever.Retriever{Path: c.Path}, nil
}

func createS3Retriever(c *retrieverconf.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	awsConfig, err := awsConf.LoadDefaultConfig(context.Background())
	return &s3retrieverv2.Retriever{Bucket: c.Bucket, Item: c.Item, AwsConfig: &awsConfig}, err
}

func createHTTPRetriever(
	c *retrieverconf.RetrieverConf, timeout time.Duration) (retriever.Retriever, error) {
	return &httpretriever.Retriever{
		URL: c.URL,
		Method: func() string {
			if c.HTTPMethod == "" {
				return retrieverconf.DefaultRetrieverConfig.HTTPMethod
			}
			return c.HTTPMethod
		}(), Body: c.HTTPBody, Header: c.HTTPHeaders, Timeout: timeout}, nil
}

func createGoogleStorageRetriever(
	c *retrieverconf.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	return &gcstorageretriever.Retriever{Bucket: c.Bucket, Object: c.Object}, nil
}

func createKubernetesRetriever(
	c *retrieverconf.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	client, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return &k8sretriever.Retriever{
		Namespace:     c.Namespace,
		ConfigMapName: c.ConfigMap,
		Key:           c.Key,
		ClientConfig:  *client,
	}, nil
}

func createMongoDBRetriever(c *retrieverconf.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	return &mongodbretriever.Retriever{Database: c.Database, URI: c.URI, Collection: c.Collection}, nil
}

func createRedisRetriever(c *retrieverconf.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	var options *redis.Options
	if c.RedisOptions != nil {
		options = c.RedisOptions.ToRedisOptions()
	}
	return &redisretriever.Retriever{Options: options, Prefix: c.RedisPrefix}, nil
}

func createAzBlobStorageRetriever(
	c *retrieverconf.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	return &azblobretriever.Retriever{
		Container:   c.Container,
		Object:      c.Object,
		AccountName: c.AccountName,
		AccountKey:  c.AccountKey,
	}, nil
}

func createPostgreSQLRetriever(
	c *retrieverconf.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	return &postgresqlretriever.Retriever{URI: c.URI, Table: c.Table, Columns: c.Columns}, nil
}
