package service

import (
	"time"

	awsConf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
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
	"golang.org/x/net/context"
	"k8s.io/client-go/rest"
)

// retrieverFactory defines the signature for retriever factory functions
type retrieverFactory func(*config.RetrieverConf, time.Duration) (retriever.Retriever, error)

// retrieverFactories maps retriever kinds to their factory functions
var retrieverFactories = map[config.RetrieverKind]retrieverFactory{
	config.GitHubRetriever:        createGitHubRetriever,
	config.GitlabRetriever:        createGitlabRetriever,
	config.BitbucketRetriever:     createBitbucketRetriever,
	config.FileRetriever:          createFileRetriever,
	config.S3Retriever:            createS3Retriever,
	config.HTTPRetriever:          createHTTPRetriever,
	config.GoogleStorageRetriever: createGoogleStorageRetriever,
	config.KubernetesRetriever:    createKubernetesRetriever,
	config.MongoDBRetriever:       createMongoDBRetriever,
	config.RedisRetriever:         createRedisRetriever,
	config.AzBlobStorageRetriever: createAzBlobStorageRetriever,
	config.PostgreSQLRetriever:    createPostgreSQLRetriever,
}

// Factory functions for each retriever type
func createGitHubRetriever(c *config.RetrieverConf, timeout time.Duration) (retriever.Retriever, error) {
	token := c.AuthToken
	if token == "" && c.GithubToken != "" { // nolint: staticcheck
		token = c.GithubToken // nolint: staticcheck
	}
	return &githubretriever.Retriever{
		RepositorySlug: c.RepositorySlug,
		Branch: func() string {
			if c.Branch == "" {
				return config.DefaultRetriever.GitBranch
			}
			return c.Branch
		}(),
		FilePath:    c.Path,
		GithubToken: token,
		Timeout:     timeout,
	}, nil
}

func createGitlabRetriever(c *config.RetrieverConf, timeout time.Duration) (retriever.Retriever, error) {
	return &gitlabretriever.Retriever{
		BaseURL: c.BaseURL,
		Branch: func() string {
			if c.Branch == "" {
				return config.DefaultRetriever.GitBranch
			}
			return c.Branch
		}(),
		FilePath:       c.Path,
		GitlabToken:    c.AuthToken,
		RepositorySlug: c.RepositorySlug,
		Timeout:        timeout,
	}, nil
}

func createBitbucketRetriever(c *config.RetrieverConf, timeout time.Duration) (retriever.Retriever, error) {
	return &bitbucketretriever.Retriever{
		RepositorySlug: c.RepositorySlug,
		Branch: func() string {
			if c.Branch == "" {
				return config.DefaultRetriever.GitBranch
			}
			return c.Branch
		}(),
		FilePath:       c.Path,
		BitBucketToken: c.AuthToken,
		BaseURL:        c.BaseURL,
		Timeout:        timeout,
	}, nil
}

func createFileRetriever(c *config.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	return &fileretriever.Retriever{Path: c.Path}, nil
}

func createS3Retriever(c *config.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	awsConfig, err := awsConf.LoadDefaultConfig(context.Background())
	return &s3retrieverv2.Retriever{Bucket: c.Bucket, Item: c.Item, AwsConfig: &awsConfig}, err
}

func createHTTPRetriever(c *config.RetrieverConf, timeout time.Duration) (retriever.Retriever, error) {
	return &httpretriever.Retriever{
		URL: c.URL,
		Method: func() string {
			if c.HTTPMethod == "" {
				return config.DefaultRetriever.HTTPMethod
			}
			return c.HTTPMethod
		}(), Body: c.HTTPBody, Header: c.HTTPHeaders, Timeout: timeout}, nil
}

func createGoogleStorageRetriever(c *config.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	return &gcstorageretriever.Retriever{Bucket: c.Bucket, Object: c.Object}, nil
}

func createKubernetesRetriever(c *config.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
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

func createMongoDBRetriever(c *config.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	return &mongodbretriever.Retriever{Database: c.Database, URI: c.URI, Collection: c.Collection}, nil
}

func createRedisRetriever(c *config.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	return &redisretriever.Retriever{Options: c.RedisOptions, Prefix: c.RedisPrefix}, nil
}

func createAzBlobStorageRetriever(c *config.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	return &azblobretriever.Retriever{
		Container:   c.Container,
		Object:      c.Object,
		AccountName: c.AccountName,
		AccountKey:  c.AccountKey,
	}, nil
}

func createPostgreSQLRetriever(c *config.RetrieverConf, _ time.Duration) (retriever.Retriever, error) {
	return &postgresqlretriever.Retriever{URI: c.URI, Table: c.Table, Columns: c.Columns}, nil
}
