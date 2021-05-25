package ffclient

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// GithubRetriever is a configuration struct for a GitHub retriever.
type GithubRetriever struct {
	RepositorySlug string
	Branch         string // default is main
	FilePath       string
	GithubToken    string
	Timeout        time.Duration // default is 10 seconds
}

func (r *GithubRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	// default branch is main
	branch := r.Branch
	if branch == "" {
		branch = "main"
	}

	// add header for Github Token if specified
	header := http.Header{}
	if r.GithubToken != "" {
		header.Add("Authorization", fmt.Sprintf("token %s", r.GithubToken))
	}

	URL := fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/%s",
		r.RepositorySlug,
		branch,
		r.FilePath)

	httpRetriever := HTTPRetriever{
		URL:     URL,
		Method:  http.MethodGet,
		Header:  header,
		Timeout: r.Timeout,
	}

	return httpRetriever.Retrieve(ctx)
}
