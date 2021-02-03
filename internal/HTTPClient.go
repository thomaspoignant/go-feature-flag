package internal

import "net/http"

// HTTPClient is an interface over http.Client to make mock easier.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
