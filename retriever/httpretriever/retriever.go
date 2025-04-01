package httpretriever

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/retriever/shared"
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

// Retrieve is the function in charge of fetching the flag configuration.
func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	resp, err := shared.CallHTTPAPI(ctx, r.URL, r.Method, r.Body, r.Timeout, r.Header, r.httpClient)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode > 399 {
		return nil, fmt.Errorf("request to %s failed with code %d", r.URL, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
