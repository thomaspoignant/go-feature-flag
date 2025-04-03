package service

import (
	"time"

	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	azblobretriever "github.com/thomaspoignant/go-feature-flag/retriever/azblobstorageretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/bitbucketretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gitlabretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/k8sretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/mongodbretriever"
	"k8s.io/client-go/rest"
)

func initGithubRetriever(
	c *config.RetrieverConf,
	retrieverTimeout time.Duration,
) *githubretriever.Retriever {
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
		Timeout:     retrieverTimeout,
	}
}

func initGitlabRetriever(
	c *config.RetrieverConf,
	retrieverTimeout time.Duration,
) *gitlabretriever.Retriever {
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
		Timeout:        retrieverTimeout,
	}
}

func initBitbucketRetriever(
	c *config.RetrieverConf,
	retrieverTimeout time.Duration,
) *bitbucketretriever.Retriever {
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
		Timeout:        retrieverTimeout,
	}
}

func initHTTPRetriever(
	c *config.RetrieverConf,
	retrieverTimeout time.Duration,
) *httpretriever.Retriever {
	return &httpretriever.Retriever{
		URL: c.URL,
		Method: func() string {
			if c.HTTPMethod == "" {
				return config.DefaultRetriever.HTTPMethod
			}
			return c.HTTPMethod
		}(), Body: c.HTTPBody, Header: c.HTTPHeaders, Timeout: retrieverTimeout}
}

func initK8sRetriever(c *config.RetrieverConf) (*k8sretriever.Retriever, error) {
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

func initMongoRetriever(c *config.RetrieverConf) *mongodbretriever.Retriever {
	return &mongodbretriever.Retriever{
		Database:   c.Database,
		URI:        c.URI,
		Collection: c.Collection,
	}
}

func initAzBlobRetriever(c *config.RetrieverConf) *azblobretriever.Retriever {
	return &azblobretriever.Retriever{
		Container:   c.Container,
		Object:      c.Object,
		AccountName: c.AccountName,
		AccountKey:  c.AccountKey,
	}
}
