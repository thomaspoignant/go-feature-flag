package githubretriever

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/retriever/shared"
)

// Retriever is a configuration struct for a GitHub retriever.
type Retriever struct {
	RepositorySlug string
	Branch         string // default is main
	FilePath       string
	GithubToken    string
	Timeout        time.Duration // default is 10 seconds

	// BaseURL is the base URL for the GitHub API.
	// If not specified, it defaults to "https://api.github.com" for GitHub.com.
	// For GitHub Enterprise instances, specify your instance URL (e.g., "https://github.acme.com/api/v3").
	BaseURL string

	// GitHub App authentication fields.
	// When set, these take precedence over GithubToken.
	// AppID and InstallationID are required together with either
	// PrivateKey (PEM string) or PrivateKeyFile (path to a PEM file).
	// Mutually exclusive with GithubToken / AuthToken.
	AppID          int64
	InstallationID int64
	PrivateKey     string // PEM-encoded RSA private key
	PrivateKeyFile string // path to a PEM file; takes precedence over PrivateKey if both are set

	// httpClient is the http.Client if you want to override it.
	httpClient internal.HTTPClient

	// rate limit fields
	rateLimitRemaining int
	rateLimitReset     time.Time

	// cached installation token state (GitHub App auth only)
	mu             sync.Mutex
	cachedToken    string
	tokenExpiresAt time.Time
}

// installationTokenResponse is the JSON body returned by the access_tokens endpoint.
type installationTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Retrieve is the function in charge of fetching the flag configuration.
func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	if r.FilePath == "" || r.RepositorySlug == "" {
		return nil, fmt.Errorf(
			"missing mandatory information filePath=%s, repositorySlug=%s",
			r.FilePath,
			r.RepositorySlug,
		)
	}

	// default branch is main
	branch := r.Branch
	if branch == "" {
		branch = "main"
	}

	token, err := r.resolveToken(ctx)
	if err != nil {
		return nil, err
	}

	header := r.buildHeaders(token)

	if err := r.checkRateLimit(); err != nil {
		return nil, err
	}

	URL, err := r.buildURL(branch)
	if err != nil {
		return nil, err
	}

	resp, err := shared.CallHTTPAPI(ctx, URL, http.MethodGet, "", r.Timeout, header, r.httpClient)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	r.updateRateLimit(resp.Header)

	// On a 401 with App auth, the cached token may have just expired.
	// Invalidate it and retry once with a fresh token.
	if resp.StatusCode == http.StatusUnauthorized && r.isAppAuth() {
		_ = resp.Body.Close()
		r.invalidateCachedToken()

		token, err = r.resolveToken(ctx)
		if err != nil {
			return nil, err
		}
		header = r.buildHeaders(token)

		resp, err = shared.CallHTTPAPI(ctx, URL, http.MethodGet, "", r.Timeout, header, r.httpClient)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()
		r.updateRateLimit(resp.Header)
	}

	if err := r.checkResponseError(resp, URL); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// isAppAuth reports whether GitHub App auth fields are configured.
func (r *Retriever) isAppAuth() bool {
	return r.AppID != 0 && r.InstallationID != 0 && (r.PrivateKey != "" || r.PrivateKeyFile != "")
}

// resolveToken returns the bearer token to use for the current request.
// For PAT auth it returns GithubToken directly.
// For App auth it returns a cached installation token, refreshing it when needed.
func (r *Retriever) resolveToken(ctx context.Context) (string, error) {
	if !r.isAppAuth() {
		return r.GithubToken, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Refresh ~5 minutes before expiry so we never send an about-to-expire token.
	if r.cachedToken != "" && time.Now().Before(r.tokenExpiresAt.Add(-5*time.Minute)) {
		return r.cachedToken, nil
	}

	token, expiresAt, err := r.mintInstallationToken(ctx)
	if err != nil {
		return "", fmt.Errorf("github app: failed to mint installation token: %w", err)
	}
	r.cachedToken = token
	r.tokenExpiresAt = expiresAt
	return token, nil
}

// invalidateCachedToken clears the cached installation token so the next call to
// resolveToken mints a fresh one.
func (r *Retriever) invalidateCachedToken() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cachedToken = ""
	r.tokenExpiresAt = time.Time{}
}

// mintInstallationToken signs a JWT, exchanges it for an installation token, and
// returns the token string and its expiry time.
func (r *Retriever) mintInstallationToken(ctx context.Context) (string, time.Time, error) {
	key, err := r.loadPrivateKey()
	if err != nil {
		return "", time.Time{}, err
	}

	jwt, err := signAppJWT(r.AppID, key)
	if err != nil {
		return "", time.Time{}, err
	}

	baseURL := r.BaseURL
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path.Join(u.Path, "app", "installations", strconv.FormatInt(r.InstallationID, 10), "access_tokens")

	header := http.Header{}
	header.Set("Accept", "application/vnd.github+json")
	header.Set("X-GitHub-Api-Version", "2022-11-28")
	header.Set("Authorization", "Bearer "+jwt)

	resp, err := shared.CallHTTPAPI(ctx, u.String(), http.MethodPost, "", r.Timeout, header, r.httpClient)
	if err != nil {
		return "", time.Time{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return "", time.Time{}, fmt.Errorf(
			"github app: access_tokens endpoint returned %d for installation %d",
			resp.StatusCode, r.InstallationID,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, err
	}

	var result installationTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", time.Time{}, fmt.Errorf("github app: failed to parse token response: %w", err)
	}
	if result.Token == "" {
		return "", time.Time{}, fmt.Errorf("github app: token response contained no token")
	}
	return result.Token, result.ExpiresAt, nil
}

// loadPrivateKey returns the RSA private key from PrivateKeyFile (preferred) or PrivateKey.
func (r *Retriever) loadPrivateKey() (*rsa.PrivateKey, error) {
	var pemBytes []byte
	var err error

	if r.PrivateKeyFile != "" {
		pemBytes, err = os.ReadFile(r.PrivateKeyFile) //nolint:gosec // path is operator-supplied config
		if err != nil {
			return nil, fmt.Errorf("github app: failed to read private key file %s: %w", r.PrivateKeyFile, err)
		}
	} else {
		pemBytes = []byte(r.PrivateKey)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("github app: failed to decode PEM block from private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("github app: failed to parse RSA private key: %w", err)
	}
	return key, nil
}

// buildHeaders constructs the HTTP headers for the GitHub API request.
func (r *Retriever) buildHeaders(token string) http.Header {
	header := http.Header{}
	header.Add("Accept", "application/vnd.github.raw")
	header.Add("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	return header
}

// checkRateLimit checks if the rate limit has been exceeded.
func (r *Retriever) checkRateLimit() error {
	if r.rateLimitRemaining <= 0 && time.Now().Before(r.rateLimitReset) {
		return fmt.Errorf("rate limit exceeded. Next call will be after %s", r.rateLimitReset)
	}
	return nil
}

// buildURL constructs the GitHub API URL for retrieving the file.
func (r *Retriever) buildURL(branch string) (string, error) {
	// Validate inputs to prevent path traversal attacks
	if strings.Contains(r.FilePath, "..") {
		return "", fmt.Errorf("filepath must not contain '..'")
	}
	if strings.Contains(r.RepositorySlug, "..") {
		return "", fmt.Errorf("repository slug must not contain '..'")
	}

	baseURL := r.BaseURL
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path.Join(u.Path, "repos", r.RepositorySlug, "contents", r.FilePath)

	q := u.Query()
	q.Set("ref", branch)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// checkResponseError checks if the response indicates an error and returns an appropriate error message.
func (r *Retriever) checkResponseError(resp *http.Response, url string) error {
	if resp.StatusCode > 399 {
		// Collect the headers to add in the error message
		ghHeaders := map[string]string{}
		for name := range resp.Header {
			if strings.HasPrefix(name, "X-") {
				ghHeaders[name] = resp.Header.Get(name)
			}
		}

		return fmt.Errorf("request to %s failed with code %d."+
			" GitHub Headers: %v", url, resp.StatusCode, ghHeaders)
	}
	return nil
}

// SetHTTPClient is here if you want to override the default http.Client we are using.
// It is also used for the tests.
func (r *Retriever) SetHTTPClient(client internal.HTTPClient) {
	r.httpClient = client
}

func (r *Retriever) updateRateLimit(headers http.Header) {
	if remaining := headers.Get("X-RateLimit-Remaining"); remaining != "" {
		if remainingInt, err := strconv.Atoi(remaining); err == nil {
			r.rateLimitRemaining = remainingInt
		}
	}

	if reset := headers.Get("X-RateLimit-Reset"); reset != "" {
		if resetInt, err := strconv.ParseInt(reset, 10, 64); err == nil {
			r.rateLimitReset = time.Unix(resetInt, 0)
		}
	}
}
