package githubretriever

import (
	"context"
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/retriever/shared"
	"io"
	"net/http"
	"strings"
	"time"

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

	resp, err := shared.CallHTTPAPI(ctx, URL, http.MethodGet, "", r.Timeout, header, r.httpClient)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode > 399 {
		// Collect the headers to add in the error message
		ghHeaders := map[string]string{}
		for name := range resp.Header {
			if strings.HasPrefix(name, "X-") {
				ghHeaders[name] = resp.Header.Get(name)
			}
		}

		return nil, fmt.Errorf("request to %s failed with code %d."+
			" GitHub Headers: %v", URL, resp.StatusCode, ghHeaders)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// SetHTTPClient is here if you want to override the default http.Client we are using.
// It is also used for the tests.
func (r *Retriever) SetHTTPClient(client internal.HTTPClient) {
	r.httpClient = client
}
