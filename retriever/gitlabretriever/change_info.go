package gitlabretriever

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/retriever/shared"
)

type gitlabCommit struct {
	ID            string    `json:"id"`
	AuthorName    string    `json:"author_name"`
	AuthorEmail   string    `json:"author_email"`
	CommittedDate time.Time `json:"committed_date"`
}

// LastChange returns the author of the last commit that touched the flag file.
func (r *Retriever) LastChange(ctx context.Context) (*notifier.ChangeInfo, error) {
	if r.FilePath == "" || r.RepositorySlug == "" {
		return nil, fmt.Errorf(
			"missing mandatory information filePath=%s, repositorySlug=%s", r.FilePath, r.RepositorySlug)
	}
	branch := r.Branch
	if branch == "" {
		branch = "main"
	}
	baseURL := r.BaseURL
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	parsed, err := url.Parse(strings.Join([]string{
		baseURL, "api/v4/projects", escapePathSegment(r.RepositorySlug), "repository/commits"}, "/"))
	if err != nil {
		return nil, fmt.Errorf("impossible to parse the url %w", err)
	}
	q := parsed.Query()
	q.Set("ref_name", branch)
	q.Set("path", r.FilePath)
	q.Set("per_page", "1")
	parsed.RawQuery = q.Encode()

	header := http.Header{}
	if r.GitlabToken != "" {
		header.Add("PRIVATE-TOKEN", r.GitlabToken)
	}
	var commits []gitlabCommit
	if _, err := shared.GetJSON(ctx, parsed.String(), r.Timeout, header, r.httpClient, &commits); err != nil {
		return nil, err
	}
	if len(commits) == 0 {
		return nil, nil
	}
	c := commits[0]
	return &notifier.ChangeInfo{
		Author:    c.AuthorName,
		Email:     c.AuthorEmail,
		Reference: c.ID,
		ChangedAt: c.CommittedDate,
	}, nil
}
