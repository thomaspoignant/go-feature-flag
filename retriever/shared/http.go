package shared

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal"
)

// CallHTTPAPI is the default function to make an HTTP call in the retrievers.
func CallHTTPAPI(
	ctx context.Context,
	url string, method string,
	body string,
	timeout time.Duration,
	header http.Header,
	httpClient internal.HTTPClient) (*http.Response, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	if url == "" {
		return nil, errors.New("URL is a mandatory parameter when using httpretriever.Retriever")
	}

	if method == "" {
		method = http.MethodGet
	}

	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Add header if some are passed
	if len(header) > 0 {
		req.Header = header
	}

	if httpClient == nil {
		httpClient = internal.HTTPClientWithTimeout(timeout)
	}

	// API call
	return httpClient.Do(req)
}
