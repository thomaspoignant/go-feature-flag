package gitlabretriever

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	httpretriever "github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"

	"github.com/thomaspoignant/go-feature-flag/internal"
)

// Retriever is a configuration struct for a GitHub retriever.
type Retriever struct {
	// Branch is the name of the GITLAB branch where to download the file
	// default: main
	Branch string // default is main
	// FilePath is the location of your file in the repository
	FilePath string

	// GitlabToken is the token to use when downloading the file
	GitlabToken string

	// RepositorySlug is the name of your repository in your gitlab instance
	// ex: thomaspoignant/go-feature-flag
	RepositorySlug string

	// BaseURL is the DNS of your GITLAB installation.
	// default: https://gitlab.com
	BaseURL string

	// Timeout is the time before we timeout while retrieving the flag file.
	// default: 10 seconds
	Timeout time.Duration

	// httpClient is the http.Client if you want to override it.
	httpClient internal.HTTPClient
}

func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	if r.FilePath == "" || r.RepositorySlug == "" {
		return nil, fmt.Errorf(
			"missing mandatory information filePath=%s, repositorySlug=%s", r.FilePath, r.RepositorySlug)
	}
	branch := r.Branch
	if branch == "" {
		branch = "main"
	}
	if r.BaseURL == "" {
		r.BaseURL = "https://gitlab.com"
	}

	parsedURL, err := url.Parse(r.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("impossible to parse the param baseUrl %s", err)
	}
	parsedURL.Path, err = url.JoinPath(
		parsedURL.Path,
		"api/v4/projects",
		url.QueryEscape(r.RepositorySlug),
		"repository/files",
		url.QueryEscape(r.FilePath),
	)
	if err != nil {
		return nil, fmt.Errorf("impossible to parse the param baseUrl %s", err)
	}

	// add branch as
	rawQuery := parsedURL.Query()
	rawQuery.Set("ref", branch)
	parsedURL.RawQuery = rawQuery.Encode()

	// add header for Gitlab Token if specified
	header := http.Header{}
	if r.GitlabToken != "" {
		header.Add("PRIVATE-TOKEN", r.GitlabToken)
	}
	httpRetriever := httpretriever.Retriever{
		URL:     parsedURL.String(),
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
