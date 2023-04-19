package gitlabretriever

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	httpretriever "github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"

	"github.com/thomaspoignant/go-feature-flag/internal"
)

// Retriever is a configuration struct for a GitHub retriever.
type Retriever struct {
	Branch         string // default is main
	FilePath       string
	GitlabToken    string
	RepositorySlug string
	URL            string        // https://gitlab.com
	Timeout        time.Duration // default is 10 seconds

	// httpClient is the http.Client if you want to override it.
	httpClient internal.HTTPClient
}

func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	if r.FilePath == "" || r.URL == "" {
		return nil, fmt.Errorf("missing mandatory information URL=%s, FilePath=%s", r.URL, r.FilePath)
	}

	// default branch is main
	branch := r.Branch
	if branch == "" {
		branch = "main"
	}

	URL := r.URL
	if URL == "" {
		URL = "https://gitlab.com"
	}

	// add header for Gitlab Token if specified
	header := http.Header{}
	if r.GitlabToken != "" {
		header.Add("PRIVATE-TOKEN", r.GitlabToken)
	}

	URL = strings.Trim(r.URL, "/")
	slug := strings.Trim(r.RepositorySlug, "/")
	path := strings.Trim(r.FilePath, "/")
	reqURL := fmt.Sprintf("%s/api/v4/projects/%s/repository/files/%s/raw?ref=%s", URL, url.QueryEscape(slug), url.QueryEscape(path), r.Branch)

	httpRetriever := httpretriever.Retriever{
		URL:     reqURL,
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
