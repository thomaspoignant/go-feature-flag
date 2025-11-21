package retrieverconf

import (
	"fmt"
	"net/http"
	"time"

	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/err"
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
	Kind           RetrieverKind `mapstructure:"kind"             koanf:"kind"`
	RepositorySlug string        `mapstructure:"repositorySlug"   koanf:"repositoryslug"`
	Branch         string        `mapstructure:"branch"           koanf:"branch"`
	Path           string        `mapstructure:"path"             koanf:"path"`
	// Deprecated: Please use AuthToken instead
	GithubToken string              `mapstructure:"githubToken"   koanf:"githubtoken"`
	URL         string              `mapstructure:"url"           koanf:"url"`
	Timeout     int64               `mapstructure:"timeout"       koanf:"timeout"`
	HTTPMethod  string              `mapstructure:"method"        koanf:"method"`
	HTTPBody    string              `mapstructure:"body"          koanf:"body"`
	HTTPHeaders map[string][]string `mapstructure:"headers"       koanf:"headers"`
	Bucket      string              `mapstructure:"bucket"        koanf:"bucket"`
	Object      string              `mapstructure:"object"        koanf:"object"`
	Item        string              `mapstructure:"item"          koanf:"item"`
	Namespace   string              `mapstructure:"namespace"     koanf:"namespace"`
	ConfigMap   string              `mapstructure:"configmap"     koanf:"configmap"`
	Key         string              `mapstructure:"key"           koanf:"key"`
	BaseURL     string              `mapstructure:"baseUrl"       koanf:"baseurl"`
	AuthToken   string              `mapstructure:"token"         koanf:"token"`

	// URI is used by
	// - the postgresql retriever
	// - the mongodb retriever
	URI string `mapstructure:"uri"  koanf:"uri"`

	// Table is used by
	// - the postgresql retriever
	Table string `mapstructure:"table"  koanf:"table"`

	// Columns is used by
	// - the postgresql retriever (it allows to use custom column names)
	Columns    map[string]string `mapstructure:"columns"        koanf:"columns"`
	Database   string            `mapstructure:"database"       koanf:"database"`
	Collection string            `mapstructure:"collection"     koanf:"collection"`

	// RedisOptions is the serializable redis configuration that can be used in JSON/YAML files
	RedisOptions *SerializableRedisOptions `mapstructure:"redisOptions"   koanf:"redisOptions"`

	RedisPrefix string `mapstructure:"redisPrefix"    koanf:"redisPrefix"`
	AccountName string `mapstructure:"accountName"    koanf:"accountname"`
	AccountKey  string `mapstructure:"accountKey"     koanf:"accountkey"`
	Container   string `mapstructure:"container"      koanf:"container"`
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
		return err.NewRetrieverConfError("item", string(c.Kind))
	}
	if c.Kind == HTTPRetriever && c.URL == "" {
		return err.NewRetrieverConfError("url", string(c.Kind))
	}
	if c.Kind == GoogleStorageRetriever && c.Object == "" {
		return err.NewRetrieverConfError("object", string(c.Kind))
	}
	if c.Kind == FileRetriever && c.Path == "" {
		return err.NewRetrieverConfError("path", string(c.Kind))
	}
	if (c.Kind == S3Retriever || c.Kind == GoogleStorageRetriever) && c.Bucket == "" {
		return err.NewRetrieverConfError("bucket", string(c.Kind))
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
		return err.NewRetrieverConfError("uri", string(c.Kind))
	}
	if c.Table == "" {
		return err.NewRetrieverConfError("table", string(c.Kind))
	}
	return nil
}

func (c *RetrieverConf) validateGitRetriever() error {
	if c.RepositorySlug == "" {
		return err.NewRetrieverConfError("repositorySlug", string(c.Kind))
	}
	if c.Path == "" {
		return err.NewRetrieverConfError("path", string(c.Kind))
	}
	return nil
}

func (c *RetrieverConf) validateKubernetesRetriever() error {
	if c.ConfigMap == "" {
		return err.NewRetrieverConfError("configmap", string(c.Kind))
	}
	if c.Namespace == "" {
		return err.NewRetrieverConfError("namespace", string(c.Kind))
	}
	if c.Key == "" {
		return err.NewRetrieverConfError("key", string(c.Kind))
	}
	return nil
}

func (c *RetrieverConf) validateMongoDBRetriever() error {
	if c.Collection == "" {
		return err.NewRetrieverConfError("collection", string(c.Kind))
	}
	if c.Database == "" {
		return err.NewRetrieverConfError("database", string(c.Kind))
	}
	if c.URI == "" {
		return err.NewRetrieverConfError("uri", string(c.Kind))
	}
	return nil
}

func (c *RetrieverConf) validateRedisRetriever() error {
	if c.RedisOptions == nil {
		return err.NewRetrieverConfError("redisOptions", string(c.Kind))
	}
	if c.RedisOptions.Addr == "" {
		return err.NewRetrieverConfError("redisOptions.addr", string(c.Kind))
	}
	return nil
}

func (c *RetrieverConf) validateAzBlobStorageRetriever() error {
	if c.AccountName == "" {
		return err.NewRetrieverConfError("accountName", string(c.Kind))
	}
	if c.Container == "" {
		return err.NewRetrieverConfError("container", string(c.Kind))
	}
	if c.Object == "" {
		return err.NewRetrieverConfError("object", string(c.Kind))
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
