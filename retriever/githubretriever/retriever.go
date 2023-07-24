package githubretriever

import (
	"context"
	"fmt"
	"net/http"
	"time"

	httpretriever "github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"

	"github.com/thomaspoignant/go-feature-flag/internal"
)

// Retriever is a configuration struct for a GitHub retriever.
type Retriever struct {
	RepositorySlug string
	Branch         string // default is main
	FilePath       string
	GithubToken    string
	Timeout        time.Duration // default is 10 seconds

	// httpClient is the http.Client if you want to override it.
	httpClient internal.HTTPClient
}

func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	if r.FilePath == "" || r.RepositorySlug == "" {
		return nil, fmt.Errorf("missing mandatory information filePath=%s, repositorySlug=%s", r.FilePath, r.RepositorySlug)
	}

	// default branch is main
	branch := r.Branch
	if branch == "" {
		branch = "main"
	}

	header := http.Header{}
	header.Add("Accept", "application/vnd.github.raw")
	header.Add("X-GitHub-Api-Version", "2022-11-28")
	// add header for GitHub Token if specified
	if r.GithubToken != "" {
		header.Add("Authorization", fmt.Sprintf("Bearer %s", r.GithubToken))
	}

	URL := fmt.Sprintf(
		"https://api.github.com/repos/%s/contents/%s?ref=%s",
		r.RepositorySlug,
		r.FilePath,
		branch)

	httpRetriever := httpretriever.Retriever{
		URL:     URL,
		Method:  http.MethodGet,
		Header:  header,
		Timeout: r.Timeout,
	}

	if r.httpClient != nil {
		httpRetriever.SetHTTPClient(r.httpClient)
	}

	return httpRetriever.Retrieve(ctx)
}

// SetHTTPClient is here if you want to override the default http.Client we are using.
// It is also used for the tests.
func (r *Retriever) SetHTTPClient(client internal.HTTPClient) {
	r.httpClient = client
}
