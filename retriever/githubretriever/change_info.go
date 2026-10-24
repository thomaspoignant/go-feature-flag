package githubretriever

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/retriever/shared"
)

type githubCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Author struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	Author *struct {
		Login string `json:"login"`
	} `json:"author"`
}

// LastChange returns the author of the last commit that touched the flag file.
func (r *Retriever) LastChange(ctx context.Context) (*notifier.ChangeInfo, error) {
	if err := r.checkRateLimit(); err != nil {
		return nil, err
	}
	commitsURL, err := r.buildCommitsURL()
	if err != nil {
		return nil, err
	}

	var commits []githubCommit
	headers, err := shared.GetJSON(ctx, commitsURL, r.Timeout, r.buildHeaders(), r.httpClient, &commits)
	if headers != nil {
		r.updateRateLimit(headers)
	}
	if err != nil {
		return nil, err
	}
	if len(commits) == 0 {
		return nil, nil
	}

	c := commits[0]
	author := c.Commit.Author.Name
	if c.Author != nil && c.Author.Login != "" {
		author = c.Author.Login
	}
	return &notifier.ChangeInfo{
		Author:    author,
		Email:     c.Commit.Author.Email,
		Reference: c.SHA,
		ChangedAt: c.Commit.Author.Date,
	}, nil
}

func (r *Retriever) buildCommitsURL() (string, error) {
	if strings.Contains(r.FilePath, "..") || strings.Contains(r.RepositorySlug, "..") {
		return "", fmt.Errorf("filepath and repository slug must not contain '..'")
	}
	baseURL := r.BaseURL
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path.Join(u.Path, "repos", r.RepositorySlug, "commits")
	branch := r.Branch
	if branch == "" {
		branch = "main"
	}
	q := u.Query()
	q.Set("path", r.FilePath)
	q.Set("sha", branch)
	q.Set("per_page", "1")
	u.RawQuery = q.Encode()
	return u.String(), nil
}
