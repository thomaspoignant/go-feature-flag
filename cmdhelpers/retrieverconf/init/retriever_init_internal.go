package init

import (
	"time"

	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
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
	c *retrieverconf.RetrieverConf,
	retrieverTimeout time.Duration,
	defaultGitBranch string,
) *githubretriever.Retriever {
	token := c.AuthToken
	if token == "" && c.GithubToken != "" { // nolint: staticcheck
		token = c.GithubToken // nolint: staticcheck
	}
	return &githubretriever.Retriever{
		RepositorySlug: c.RepositorySlug,
		Branch: func() string {
			if c.Branch == "" {
				return defaultGitBranch
			}
			return c.Branch
		}(),
		FilePath:    c.Path,
		GithubToken: token,
		Timeout:     retrieverTimeout,
	}
}

func initGitlabRetriever(
	c *retrieverconf.RetrieverConf,
	retrieverTimeout time.Duration,
	defaultGitBranch string,
) *gitlabretriever.Retriever {
	return &gitlabretriever.Retriever{
		BaseURL: c.BaseURL,
		Branch: func() string {
			if c.Branch == "" {
				return defaultGitBranch
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
	c *retrieverconf.RetrieverConf,
	retrieverTimeout time.Duration,
	defaultGitBranch string,
) *bitbucketretriever.Retriever {
	return &bitbucketretriever.Retriever{
		RepositorySlug: c.RepositorySlug,
		Branch: func() string {
			if c.Branch == "" {
				return defaultGitBranch
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
	c *retrieverconf.RetrieverConf,
	retrieverTimeout time.Duration,
	defaultHttpMethod string,
) *httpretriever.Retriever {
	return &httpretriever.Retriever{
		URL: c.URL,
		Method: func() string {
			if c.HTTPMethod == "" {
				return defaultHttpMethod
			}
			return c.HTTPMethod
		}(), Body: c.HTTPBody, Header: c.HTTPHeaders, Timeout: retrieverTimeout}
}

func initK8sRetriever(c *retrieverconf.RetrieverConf) (*k8sretriever.Retriever, error) {
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

func initMongoRetriever(c *retrieverconf.RetrieverConf) *mongodbretriever.Retriever {
	return &mongodbretriever.Retriever{
		Database:   c.Database,
		URI:        c.URI,
		Collection: c.Collection,
	}
}

func initAzBlobRetriever(c *retrieverconf.RetrieverConf) *azblobretriever.Retriever {
	return &azblobretriever.Retriever{
		Container:   c.Container,
		Object:      c.Object,
		AccountName: c.AccountName,
		AccountKey:  c.AccountKey,
	}
}
