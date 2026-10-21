package githubretriever_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
)

// flagBody is what the mock contents API returns.
const flagBody = `test-flag:
  variations:
    true_var: true
    false_var: false
  defaultRule:
    variation: false_var
`

// tokenResponse builds a JSON access_tokens API response.
func tokenResponse(token string, expiresAt time.Time) []byte {
	b, _ := json.Marshal(map[string]string{
		"token":      token,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
	return b
}

// appMockHTTP is a multi-URL mock for GitHub App tests.
// It records all requests and lets tests configure per-call behaviour.
type appMockHTTP struct {
	// tokenStatus is the HTTP status returned for the access_tokens endpoint.
	// Defaults to 201 Created.
	tokenStatus int
	// tokenBody overrides the body of the access_tokens response.
	tokenBody []byte
	// contentsStatus is the HTTP status returned for the contents endpoint.
	contentsStatus int
	// first401ThenOK makes the first contents request return 401, then 200.
	first401ThenOK bool
	contentsCalls  int
	recordedReqs   []*http.Request
}

func (m *appMockHTTP) Do(req *http.Request) (*http.Response, error) {
	m.recordedReqs = append(m.recordedReqs, req)

	if strings.Contains(req.URL.Path, "access_tokens") {
		status := m.tokenStatus
		if status == 0 {
			status = http.StatusCreated
		}
		body := m.tokenBody
		if body == nil {
			body = tokenResponse("ghs_test_token", time.Now().Add(1*time.Hour))
		}
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}

	// contents endpoint
	m.contentsCalls++
	status := m.contentsStatus
	if status == 0 {
		status = http.StatusOK
	}
	if m.first401ThenOK && m.contentsCalls == 1 {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
			Header:     make(http.Header),
		}, nil
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader([]byte(flagBody))),
		Header:     make(http.Header),
	}, nil
}

// generateTestKey generates a throwaway RSA key for tests and returns its PEM encoding.
func generateTestKey(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	return string(pemBytes)
}

func TestGitHubAppAuth_Success(t *testing.T) {
	pemStr := generateTestKey(t)

	mock := &appMockHTTP{}
	r := githubretriever.Retriever{
		RepositorySlug: "myorg/myrepo",
		FilePath:       "flags/config.yaml",
		AppID:          12345,
		InstallationID: 67890,
		PrivateKey:     pemStr,
	}
	r.SetHTTPClient(mock)

	got, err := r.Retrieve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte(flagBody), got)

	// Should have made exactly 2 requests: access_tokens + contents
	assert.Len(t, mock.recordedReqs, 2)
	assert.Contains(t, mock.recordedReqs[0].URL.Path, "access_tokens")
	assert.Contains(t, mock.recordedReqs[1].URL.Path, "contents")

	// Contents request must carry the installation token as a Bearer token
	assert.Equal(t, "Bearer ghs_test_token", mock.recordedReqs[1].Header.Get("Authorization"))
}

func TestGitHubAppAuth_TokenCached(t *testing.T) {
	pemStr := generateTestKey(t)

	mock := &appMockHTTP{}
	r := githubretriever.Retriever{
		RepositorySlug: "myorg/myrepo",
		FilePath:       "flags/config.yaml",
		AppID:          12345,
		InstallationID: 67890,
		PrivateKey:     pemStr,
	}
	r.SetHTTPClient(mock)

	// Two successive calls should only mint the token once.
	_, err := r.Retrieve(context.Background())
	require.NoError(t, err)
	_, err = r.Retrieve(context.Background())
	require.NoError(t, err)

	tokenCalls := 0
	for _, req := range mock.recordedReqs {
		if strings.Contains(req.URL.Path, "access_tokens") {
			tokenCalls++
		}
	}
	assert.Equal(t, 1, tokenCalls, "installation token should be minted only once when still valid")
}

func TestGitHubAppAuth_401TriggersRetry(t *testing.T) {
	pemStr := generateTestKey(t)

	mock := &appMockHTTP{first401ThenOK: true}
	r := githubretriever.Retriever{
		RepositorySlug: "myorg/myrepo",
		FilePath:       "flags/config.yaml",
		AppID:          12345,
		InstallationID: 67890,
		PrivateKey:     pemStr,
	}
	r.SetHTTPClient(mock)

	got, err := r.Retrieve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte(flagBody), got)

	// Requests: access_tokens (first mint) → contents (401) → access_tokens (re-mint) → contents (200)
	assert.Len(t, mock.recordedReqs, 4)
}

func TestGitHubAppAuth_TokenExchangeFailure(t *testing.T) {
	pemStr := generateTestKey(t)

	mock := &appMockHTTP{tokenStatus: http.StatusUnauthorized}
	r := githubretriever.Retriever{
		RepositorySlug: "myorg/myrepo",
		FilePath:       "flags/config.yaml",
		AppID:          12345,
		InstallationID: 67890,
		PrivateKey:     pemStr,
	}
	r.SetHTTPClient(mock)

	_, err := r.Retrieve(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mint installation token")
}

func TestGitHubAppAuth_PrivateKeyFile(t *testing.T) {
	pemStr := generateTestKey(t)

	f, err := os.CreateTemp(t.TempDir(), "test-key-*.pem")
	require.NoError(t, err)
	_, err = f.WriteString(pemStr)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	mock := &appMockHTTP{}
	r := githubretriever.Retriever{
		RepositorySlug: "myorg/myrepo",
		FilePath:       "flags/config.yaml",
		AppID:          12345,
		InstallationID: 67890,
		PrivateKeyFile: f.Name(),
	}
	r.SetHTTPClient(mock)

	got, err := r.Retrieve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte(flagBody), got)
}

func TestGitHubAppAuth_InvalidPrivateKey(t *testing.T) {
	mock := &appMockHTTP{}
	r := githubretriever.Retriever{
		RepositorySlug: "myorg/myrepo",
		FilePath:       "flags/config.yaml",
		AppID:          12345,
		InstallationID: 67890,
		PrivateKey:     "not-a-pem-key",
	}
	r.SetHTTPClient(mock)

	_, err := r.Retrieve(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PEM")
}

func TestGitHubAppAuth_GHESBaseURL(t *testing.T) {
	pemStr := generateTestKey(t)

	mock := &appMockHTTP{}
	r := githubretriever.Retriever{
		RepositorySlug: "myorg/myrepo",
		FilePath:       "flags/config.yaml",
		AppID:          12345,
		InstallationID: 67890,
		PrivateKey:     pemStr,
		BaseURL:        "https://github.enterprise.example.com/api/v3",
	}
	r.SetHTTPClient(mock)

	_, err := r.Retrieve(context.Background())
	require.NoError(t, err)

	// The access_tokens call must use the GHES base URL
	assert.Contains(t, mock.recordedReqs[0].URL.Host, "github.enterprise.example.com")
	assert.Contains(t, mock.recordedReqs[0].URL.Path, "access_tokens")
}
