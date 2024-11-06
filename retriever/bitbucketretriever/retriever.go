package bitbucketretriever

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/retriever/shared"
)

type Retriever struct {
	RepositorySlug     string
	FilePath           string
	Branch             string
	BitBucketToken     string
	Timeout            time.Duration
	httpClient         internal.HTTPClient
	rateLimitRemaining int
	rateLimitNearLimit bool
	rateLimitReset     time.Time
}

func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	if r.FilePath == "" || r.RepositorySlug == "" {
		return nil, fmt.Errorf("missing mandatory information filePath=%s, repositorySlug=%s", r.FilePath, r.RepositorySlug)
	}

	header := http.Header{}
	header.Add("Accept", "application/json")

	branch := r.Branch
	if branch == "" {
		branch = "main"
	}

	if r.BitBucketToken != "" {
		header.Add("Authorization", fmt.Sprintf("Bearer %s", r.BitBucketToken))
	}

	if (r.rateLimitRemaining <= 0) && time.Now().Before(r.rateLimitReset) {
		return nil, fmt.Errorf("rate limit exceeded. Next call will be after %s", r.rateLimitReset)
	}

	URL := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/src/%s/%s",
		r.RepositorySlug,
		branch,
		r.FilePath)

	resp, err := shared.CallHTTPAPI(ctx, URL, http.MethodGet, "", r.Timeout, header, r.httpClient)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	r.updateRateLimit(resp.Header)

	if resp.StatusCode > 399 {
		// Collect the headers to add in the error message
		bitbucketHeaders := map[string]string{}
		for name := range resp.Header {
			if strings.HasPrefix(name, "X-") {
				bitbucketHeaders[name] = resp.Header.Get(name)
			}
		}

		return nil, fmt.Errorf("request to %s failed with code %d."+
			" Bitbucket Headers: %v", URL, resp.StatusCode, bitbucketHeaders)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (r *Retriever) SetHTTPClient(client internal.HTTPClient) {
	r.httpClient = client
}

func (r *Retriever) updateRateLimit(headers http.Header) {
	if remaining := headers.Get("X-RateLimit-Limit"); remaining != "" {
		if remainingInt, err := strconv.Atoi(remaining); err == nil {
			r.rateLimitRemaining = remainingInt
		}
	}

	if nearLimit := headers.Get("X-RateLimit-NearLimit"); nearLimit != "" {
		if nearLimitBool, err := strconv.ParseBool(nearLimit); err == nil {
			r.rateLimitNearLimit = nearLimitBool
		}
	}

	if reset := headers.Get("X-RateLimit-Reset"); reset != "" {
		if resetInt, err := strconv.ParseInt(reset, 10, 64); err == nil {
			r.rateLimitReset = time.Unix(resetInt, 0)
		}
	}
}
