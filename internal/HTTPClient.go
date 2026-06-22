package internal

import (
	"net/http"
	"time"
)

var _ HTTPClient = (*http.Client)(nil)

const defaultHTTPTimeout = 10 * time.Second

// HTTPClient is an interface over http.Client to make mock easier.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func DefaultHTTPClient() HTTPClient {
	return NewHTTPClient(defaultHTTPTimeout, nil)
}

func HTTPClientWithTimeout(timeout time.Duration) HTTPClient {
	return NewHTTPClient(timeout, nil)
}

func NewHTTPClient(timeout time.Duration, transport http.RoundTripper) HTTPClient {
	if timeout <= 0 {
		timeout = defaultHTTPTimeout
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}
