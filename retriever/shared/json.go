package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal"
)

// GetJSON calls the API with a GET and decodes the JSON response in out.
// It returns the response headers, available even when the call failed with an HTTP status,
// to let the caller track the rate limit.
func GetJSON(
	ctx context.Context,
	url string,
	timeout time.Duration,
	header http.Header,
	httpClient internal.HTTPClient,
	out any,
) (http.Header, error) {
	resp, err := CallHTTPAPI(ctx, url, http.MethodGet, "", timeout, header, httpClient)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode > 399 {
		return resp.Header, fmt.Errorf("request to %s failed with code %d", url, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return resp.Header, fmt.Errorf("invalid response from %s: %w", url, err)
	}
	return resp.Header, nil
}
