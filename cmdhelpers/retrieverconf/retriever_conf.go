package retrieverconf

import (
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

var DefaultRetrieverConfig = struct {
	Timeout    time.Duration
	HTTPMethod string
	GitBranch  string
}{
	Timeout:    10 * time.Second,
	HTTPMethod: http.MethodGet,
	GitBranch:  "main",
}

// RetrieverConf contains all the field to configure a retriever
type RetrieverConf struct {
	Kind           RetrieverKind `mapstructure:"kind"             koanf:"kind" json:"Kind"`
	RepositorySlug string        `mapstructure:"repositorySlug"   koanf:"repositoryslug" json:"RepositorySlug"`
	Branch         string        `mapstructure:"branch"           koanf:"branch" json:"Branch"`
	Path           string        `mapstructure:"path"             koanf:"path" json:"Path"`
	// Deprecated: Please use AuthToken instead
	GithubToken string              `mapstructure:"githubToken"   koanf:"githubtoken" json:"GithubToken"`
	URL         string              `mapstructure:"url"           koanf:"url" json:"URL"`
	Timeout     int64               `mapstructure:"timeout"       koanf:"timeout" json:"Timeout"`
	HTTPMethod  string              `mapstructure:"method"        koanf:"method" json:"HTTPMethod"`
	HTTPBody    string              `mapstructure:"body"          koanf:"body" json:"HTTPBody"`
	HTTPHeaders map[string][]string `mapstructure:"headers"       koanf:"headers" json:"HTTPHeaders"`
	Bucket      string              `mapstructure:"bucket"        koanf:"bucket" json:"Bucket"`
	Object      string              `mapstructure:"object"        koanf:"object" json:"Object"`
	Item        string              `mapstructure:"item"          koanf:"item" json:"Item"`
	Namespace   string              `mapstructure:"namespace"     koanf:"namespace" json:"Namespace"`
	ConfigMap   string              `mapstructure:"configmap"     koanf:"configmap" json:"ConfigMap"`
	Key         string              `mapstructure:"key"           koanf:"key" json:"Key"`
	BaseURL     string              `mapstructure:"baseUrl"       koanf:"baseurl" json:"BaseURL"`
	AuthToken   string              `mapstructure:"token"         koanf:"token" json:"AuthToken"`

	// URI is used by
	// - the postgresql retriever
	// - the mongodb retriever
	URI string `mapstructure:"uri"  koanf:"uri"  json:"URI"`

	// Table is used by
	// - the postgresql retriever
	Table string `mapstructure:"table"  koanf:"table"  json:"Table"`

	// Columns is used by
	// - the postgresql retriever (it allows to use custom column names)
	Columns      map[string]string `mapstructure:"columns"        koanf:"columns"        json:"Columns"`
	Database     string            `mapstructure:"database"       koanf:"database"       json:"Database"`
	Collection   string            `mapstructure:"collection"     koanf:"collection"     json:"Collection"`
	RedisOptions *redis.Options    `mapstructure:"redisOptions"   koanf:"redisOptions"   json:"RedisOptions"`
	RedisPrefix  string            `mapstructure:"redisPrefix"    koanf:"redisPrefix"    json:"RedisPrefix"`
	AccountName  string            `mapstructure:"accountName"    koanf:"accountname"    json:"AccountName"`
	AccountKey   string            `mapstructure:"accountKey"     koanf:"accountkey"     json:"AccountKey"`
	Container    string            `mapstructure:"container"      koanf:"container"      json:"Container"`
}

// IsValid validate the configuration of the retriever
// nolint:gocognit
func (c *RetrieverConf) IsValid() error {
	if err := c.Kind.IsValid(); err != nil {
		return err
	}
	if c.Kind == PostgreSQLRetriever {
		return c.validatePostgreSQLRetriever()
	}
	if c.Kind == GitHubRetriever || c.Kind == GitlabRetriever || c.Kind == BitbucketRetriever {
		return c.validateGitRetriever()
	}
	if c.Kind == S3Retriever && c.Item == "" {
		return fmt.Errorf("invalid retriever: no \"item\" property found for kind \"%s\"", c.Kind)
	}
	if c.Kind == HTTPRetriever && c.URL == "" {
		return fmt.Errorf("invalid retriever: no \"url\" property found for kind \"%s\"", c.Kind)
	}
	if c.Kind == GoogleStorageRetriever && c.Object == "" {
		return fmt.Errorf("invalid retriever: no \"object\" property found for kind \"%s\"", c.Kind)
	}
	if c.Kind == FileRetriever && c.Path == "" {
		return fmt.Errorf("invalid retriever: no \"path\" property found for kind \"%s\"", c.Kind)
	}
	if (c.Kind == S3Retriever || c.Kind == GoogleStorageRetriever) && c.Bucket == "" {
		return fmt.Errorf("invalid retriever: no \"bucket\" property found for kind \"%s\"", c.Kind)
	}
	if c.Kind == KubernetesRetriever {
		return c.validateKubernetesRetriever()
	}
	if c.Kind == MongoDBRetriever {
		return c.validateMongoDBRetriever()
	}
	if c.Kind == RedisRetriever {
		return c.validateRedisRetriever()
	}
	if c.Kind == AzBlobStorageRetriever {
		return c.validateAzBlobStorageRetriever()
	}
	return nil
}

// validatePostgreSQLRetriever validates the configuration of the postgresql retriever
func (c *RetrieverConf) validatePostgreSQLRetriever() error {
	if c.URI == "" {
		return fmt.Errorf("invalid retriever: no \"uri\" property found for kind \"%s\"", c.Kind)
	}
	if c.Table == "" {
		return fmt.Errorf("invalid retriever: no \"table\" property found for kind \"%s\"", c.Kind)
	}
	return nil
}

func (c *RetrieverConf) validateGitRetriever() error {
	if c.RepositorySlug == "" {
		return fmt.Errorf(
			"invalid retriever: no \"repositorySlug\" property found for kind \"%s\"",
			c.Kind,
		)
	}
	if c.Path == "" {
		return fmt.Errorf("invalid retriever: no \"path\" property found for kind \"%s\"", c.Kind)
	}
	return nil
}

func (c *RetrieverConf) validateKubernetesRetriever() error {
	if c.ConfigMap == "" {
		return fmt.Errorf(
			"invalid retriever: no \"configmap\" property found for kind \"%s\"",
			c.Kind,
		)
	}
	if c.Namespace == "" {
		return fmt.Errorf(
			"invalid retriever: no \"namespace\" property found for kind \"%s\"",
			c.Kind,
		)
	}
	if c.Key == "" {
		return fmt.Errorf("invalid retriever: no \"key\" property found for kind \"%s\"", c.Kind)
	}
	return nil
}

func (c *RetrieverConf) validateMongoDBRetriever() error {
	if c.Collection == "" {
		return fmt.Errorf(
			"invalid retriever: no \"collection\" property found for kind \"%s\"",
			c.Kind,
		)
	}
	if c.Database == "" {
		return fmt.Errorf(
			"invalid retriever: no \"database\" property found for kind \"%s\"",
			c.Kind,
		)
	}
	if c.URI == "" {
		return fmt.Errorf("invalid retriever: no \"uri\" property found for kind \"%s\"", c.Kind)
	}
	return nil
}

func (c *RetrieverConf) validateRedisRetriever() error {
	if c.RedisOptions == nil {
		return fmt.Errorf(
			"invalid retriever: no \"redisOptions\" property found for kind \"%s\"",
			c.Kind,
		)
	}
	return nil
}

func (c *RetrieverConf) validateAzBlobStorageRetriever() error {
	if c.AccountName == "" {
		return fmt.Errorf(
			"invalid retriever: no \"accountName\" property found for kind \"%s\"",
			c.Kind,
		)
	}
	if c.Container == "" {
		return fmt.Errorf(
			"invalid retriever: no \"container\" property found for kind \"%s\"",
			c.Kind,
		)
	}
	if c.Object == "" {
		return fmt.Errorf("invalid retriever: no \"object\" property found for kind \"%s\"", c.Kind)
	}
	return nil
}

// RetrieverKind is an enum containing all accepted Retriever kind
type RetrieverKind string

const (
	HTTPRetriever          RetrieverKind = "http"
	GitHubRetriever        RetrieverKind = "github"
	GitlabRetriever        RetrieverKind = "gitlab"
	S3Retriever            RetrieverKind = "s3"
	FileRetriever          RetrieverKind = "file"
	GoogleStorageRetriever RetrieverKind = "googleStorage"
	KubernetesRetriever    RetrieverKind = "configmap"
	MongoDBRetriever       RetrieverKind = "mongodb"
	RedisRetriever         RetrieverKind = "redis"
	BitbucketRetriever     RetrieverKind = "bitbucket"
	AzBlobStorageRetriever RetrieverKind = "azureBlobStorage"
	PostgreSQLRetriever    RetrieverKind = "postgresql"
)

// IsValid is checking if the value is part of the enum
func (r RetrieverKind) IsValid() error {
	switch r {
	case HTTPRetriever, GitHubRetriever, GitlabRetriever, S3Retriever, RedisRetriever,
		FileRetriever, GoogleStorageRetriever, KubernetesRetriever, MongoDBRetriever,
		BitbucketRetriever, AzBlobStorageRetriever, PostgreSQLRetriever:
		return nil
	}
	return fmt.Errorf("invalid retriever: kind \"%s\" is not supported", r)
}
