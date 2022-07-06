package httpretriever

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal"
)

// Retriever is a configuration struct for an HTTP endpoint retriever.
type Retriever struct {
	// URL of your endpoint
	URL string

	// HTTP Method we should use (default: GET)
	Method string

	// Body of the request if needed (default: empty body)
	Body string

	// Header added to the request
	Header http.Header

	// Timeout we should wait before failing (default: 10 seconds)
	Timeout time.Duration

	httpClient internal.HTTPClient
}

// SetHTTPClient is here if you want to override the default http.Client we are using.
// It is also used for the tests.
func (r *Retriever) SetHTTPClient(client internal.HTTPClient) {
	r.httpClient = client
}

func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	if r.URL == "" {
		return nil, errors.New("URL is a mandatory parameter when using httpretriever.Retriever")
	}

	method := r.Method
	if method == "" {
		method = http.MethodGet
	}

	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, method, r.URL, strings.NewReader(r.Body))
	if err != nil {
		return nil, err
	}

	// Add header if some are passed
	if len(r.Header) > 0 {
		req.Header = r.Header
	}

	if r.httpClient == nil {
		r.httpClient = internal.HTTPClientWithTimeout(timeout)
	}

	// API call
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Error if http code is more that 399
	if resp.StatusCode > 399 {
		return nil, fmt.Errorf("request to %s failed with code %d", r.URL, resp.StatusCode)
	}

	// read content of the URL.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
