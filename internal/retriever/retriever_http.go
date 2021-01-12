package retriever

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// HTTPClient is an interface over http.Client to make mock easier.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewHTTPRetriever return a new HTTPRetriever to get the file from an HTTP endpoint.
func NewHTTPRetriever(httpClient HTTPClient, url string, method string, body string, header http.Header) FlagRetriever {
	return &httpRetriever{
		httpClient,
		url,
		method,
		body,
		header,
	}
}

type httpRetriever struct {
	httpClient HTTPClient
	url        string
	method     string
	body       string
	header     http.Header
}

func (h *httpRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	if h.url == "" {
		return nil, errors.New("URL is a mandatory parameter when using HTTPRetriever")
	}

	method := h.method
	if method == "" {
		method = http.MethodGet
	}

	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, method, h.url, strings.NewReader(h.body))
	if err != nil {
		return nil, err
	}

	// Add header if some are passed
	if len(h.header) > 0 {
		req.Header = h.header
	}

	// API call
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Error if http code is more that 399
	if resp.StatusCode > 399 {
		return nil, fmt.Errorf("request to %s failed with code %d", h.url, resp.StatusCode)
	}

	// read content of the URL.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
