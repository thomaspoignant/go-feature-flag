package internal

import (
	"net/http"
	"time"
)

var _ HTTPClient = (*http.Client)(nil)

// HTTPClient is an interface over http.Client to make mock easier.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func DefaultHTTPClient() HTTPClient {
	return HTTPClientWithTimeout(10 * time.Second)
}

func HTTPClientWithTimeout(timeout time.Duration) HTTPClient {
	httpClient := http.DefaultClient
	httpClient.Timeout = timeout
	return httpClient
}
