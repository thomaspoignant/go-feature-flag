package config

import "fmt"

// RetrieverConf contains all the field to configure a retriever
type RetrieverConf struct {
	Kind           RetrieverKind `mapstructure:"kind" koanf:"kind"`
	RepositorySlug string        `mapstructure:"repositorySlug" koanf:"repositoryslug"`
	Branch         string        `mapstructure:"branch" koanf:"branch"`
	Path           string        `mapstructure:"path" koanf:"path"`
	// Deprecated: Please use AuthToken instead
	GithubToken string              `mapstructure:"githubToken" koanf:"githubtoken"`
	URL         string              `mapstructure:"url" koanf:"url"`
	Timeout     int64               `mapstructure:"timeout" koanf:"timeout"`
	HTTPMethod  string              `mapstructure:"method" koanf:"method"`
	HTTPBody    string              `mapstructure:"body" koanf:"body"`
	HTTPHeaders map[string][]string `mapstructure:"headers" koanf:"headers"`
	Bucket      string              `mapstructure:"bucket" koanf:"bucket"`
	Object      string              `mapstructure:"object" koanf:"object"`
	Item        string              `mapstructure:"item" koanf:"item"`
	Namespace   string              `mapstructure:"namespace" koanf:"namespace"`
	ConfigMap   string              `mapstructure:"configmap" koanf:"configmap"`
	Key         string              `mapstructure:"key" koanf:"key"`
	BaseURL     string              `mapstructure:"baseUrl" koanf:"baseurl"`
	AuthToken   string              `mapstructure:"token" koanf:"token"`
	Uri         string              `mapstructure:"uri" koanf:"uri"`
	Database    string              `mapstructure:"database" koanf:"database"`
	Collection  string              `mapstructure:"collection" koanf:"collection"`
}

// IsValid validate the configuration of the retriever
// nolint:gocognit
func (c *RetrieverConf) IsValid() error {
	if err := c.Kind.IsValid(); err != nil {
		return err
	}
	if c.Kind == GitHubRetriever && c.RepositorySlug == "" {
		return fmt.Errorf("invalid retriever: no \"repositorySlug\" property found for kind \"%s\"", c.Kind)
	}
	if c.Kind == GitlabRetriever && c.RepositorySlug == "" {
		return fmt.Errorf("invalid retriever: no \"repositorySlug\" property found for kind \"%s\"", c.Kind)
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
	if (c.Kind == GitHubRetriever || c.Kind == FileRetriever || c.Kind == GitlabRetriever) && c.Path == "" {
		return fmt.Errorf("invalid retriever: no \"path\" property found for kind \"%s\"", c.Kind)
	}
	if (c.Kind == S3Retriever || c.Kind == GoogleStorageRetriever) && c.Bucket == "" {
		return fmt.Errorf("invalid retriever: no \"bucket\" property found for kind \"%s\"", c.Kind)
	}
	if c.Kind == KubernetesRetriever && c.ConfigMap == "" {
		return fmt.Errorf("invalid retriever: no \"configmap\" property found for kind \"%s\"", c.Kind)
	}
	if c.Kind == KubernetesRetriever && c.Namespace == "" {
		return fmt.Errorf("invalid retriever: no \"namespace\" property found for kind \"%s\"", c.Kind)
	}
	if c.Kind == KubernetesRetriever && c.Key == "" {
		return fmt.Errorf("invalid retriever: no \"key\" property found for kind \"%s\"", c.Kind)
	}
	if c.Kind == MongoDBRetriever && c.Collection == "" {
		return fmt.Errorf("invalid retriever: no \"collection\" property found for kind \"%s\"", c.Kind)
	}
	if c.Kind == MongoDBRetriever && c.Database == ""  {
		return fmt.Errorf("invalid retriever: no \"database\" property found for kind \"%s\"", c.Kind)

	}
	if c.Kind == MongoDBRetriever && c.Uri == "" {
		return fmt.Errorf("invalid retriever: no \"uri\" property found for kind \"%s\"", c.Kind)
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
)

// IsValid is checking if the value is part of the enum
func (r RetrieverKind) IsValid() error {
	switch r {
	case HTTPRetriever, GitHubRetriever, GitlabRetriever, S3Retriever,
		FileRetriever, GoogleStorageRetriever, KubernetesRetriever, MongoDBRetriever:
		return nil
	}
	return fmt.Errorf("invalid retriever: kind \"%s\" is not supported", r)
}
