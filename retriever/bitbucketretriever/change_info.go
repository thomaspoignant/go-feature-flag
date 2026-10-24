package bitbucketretriever

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"time"

	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/retriever/shared"
)

type bitbucketCommits struct {
	Values []struct {
		Hash   string    `json:"hash"`
		Date   time.Time `json:"date"`
		Author struct {
			Raw  string `json:"raw"`
			User struct {
				DisplayName string `json:"display_name"`
			} `json:"user"`
		} `json:"author"`
	} `json:"values"`
}

// LastChange returns the author of the last commit that touched the flag file.
func (r *Retriever) LastChange(ctx context.Context) (*notifier.ChangeInfo, error) {
	if r.FilePath == "" || r.RepositorySlug == "" {
		return nil, fmt.Errorf(
			"missing mandatory information filePath=%s, repositorySlug=%s", r.FilePath, r.RepositorySlug)
	}
	if r.rateLimitRemaining <= 0 && time.Now().Before(r.rateLimitReset) {
		return nil, fmt.Errorf("rate limit exceeded. Next call will be after %s", r.rateLimitReset)
	}
	branch := r.Branch
	if branch == "" {
		branch = "main"
	}
	baseURL := r.BaseURL
	if baseURL == "" {
		baseURL = "https://api.bitbucket.org"
	}
	q := url.Values{}
	q.Set("path", r.FilePath)
	q.Set("pagelen", "1")
	commitsURL := fmt.Sprintf("%s/2.0/repositories/%s/commits/%s?%s",
		baseURL, r.RepositorySlug, url.PathEscape(branch), q.Encode())

	header := http.Header{}
	header.Add("Accept", "application/json")
	if r.BitBucketToken != "" {
		header.Add("Authorization", fmt.Sprintf("Bearer %s", r.BitBucketToken))
	}
	var commits bitbucketCommits
	headers, err := shared.GetJSON(ctx, commitsURL, r.Timeout, header, r.httpClient, &commits)
	if headers != nil {
		r.updateRateLimit(headers)
	}
	if err != nil {
		return nil, err
	}
	if len(commits.Values) == 0 {
		return nil, nil
	}

	c := commits.Values[0]
	info := &notifier.ChangeInfo{Author: c.Author.Raw, Reference: c.Hash, ChangedAt: c.Date}
	if addr, err := mail.ParseAddress(c.Author.Raw); err == nil {
		info.Author, info.Email = addr.Name, addr.Address
	}
	if c.Author.User.DisplayName != "" {
		info.Author = c.Author.User.DisplayName
	}
	return info, nil
}
