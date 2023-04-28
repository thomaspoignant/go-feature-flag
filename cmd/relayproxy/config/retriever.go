package config

import "fmt"

// RetrieverConf contains all the field to configure a retriever
type RetrieverConf struct {
	Kind           RetrieverKind       `mapstructure:"kind"`
	RepositorySlug string              `mapstructure:"repositorySlug"`
	Branch         string              `mapstructure:"branch"`
	Path           string              `mapstructure:"path"`
	GithubToken    string              `mapstructure:"githubToken"`
	URL            string              `mapstructure:"url"`
	Timeout        int64               `mapstructure:"timeout"`
	HTTPMethod     string              `mapstructure:"method"`
	HTTPBody       string              `mapstructure:"body"`
	HTTPHeaders    map[string][]string `mapstructure:"headers"`
	Bucket         string              `mapstructure:"bucket"`
	Object         string              `mapstructure:"object"`
	Item           string              `mapstructure:"item"`
	Namespace      string              `mapstructure:"namespace"`
	ConfigMap      string              `mapstructure:"configmap"`
	Key            string              `mapstructure:"key"`
	BaseURL        string              `mapstructure:"baseUrl"`
	AuthToken      string              `mapstructure:"token"`
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
	if c.Kind == GitlabRetriever && c.URL == "" {
		return fmt.Errorf("invalid retriever: no \"URL\" property found for kind \"%s\"", c.Kind)
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
)

// IsValid is checking if the value is part of the enum
func (r RetrieverKind) IsValid() error {
	switch r {
	case HTTPRetriever, GitHubRetriever, GitlabRetriever, S3Retriever,
		FileRetriever, GoogleStorageRetriever, KubernetesRetriever:
		return nil
	}
	return fmt.Errorf("invalid retriever: kind \"%s\" is not supported", r)
}
