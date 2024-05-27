package mock

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HTTP struct {
	Req           http.Request
	RateLimit     bool
	HasBeenCalled bool
	EndRatelimit  time.Time
}

func (m *HTTP) Do(req *http.Request) (*http.Response, error) {
	m.HasBeenCalled = true
	m.Req = *req
	success := &http.Response{
		Status:     "OK",
		StatusCode: http.StatusOK,
		Body: io.NopCloser(bytes.NewReader([]byte(`test-flag:
  variations:
    true_var: true
    false_var: false
  targeting:
    - query: key eq "random-key"
      percentage:
        true_var: 0
        false_var: 100
  defaultRule:
    variation: false_var	
`))),
	}

	error := &http.Response{
		Status:     "KO",
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewReader([]byte(""))),
	}

	ratelimitReset := m.EndRatelimit
	if ratelimitReset.IsZero() {
		ratelimitReset = time.Now().Add(1 * time.Hour)
	}

	rateLimit := &http.Response{
		Status:     "Rate Limit",
		StatusCode: http.StatusTooManyRequests,
		Body:       io.NopCloser(bytes.NewReader([]byte(""))),
		Header: map[string][]string{
			"X-Content-Type-Options": {"nosniff"},
			"X-Frame-Options":        {"deny"},
			"X-Github-Media-Type":    {"github.v3; format=json"},
			"X-Github-Request-Id":    {"F82D:37B98C:232EF263:235C93BD:6650BDC6"},
			"X-Ratelimit-Limit":      {"60"},
			"X-Ratelimit-Remaining":  {"0"},
			"X-Ratelimit-Reset":      {strconv.FormatInt(ratelimitReset.Unix(), 10)},
			"X-Ratelimit-Resource":   {"core"},
			"X-Ratelimit-Used":       {"60"},
			"X-Xss-Protection":       {"1; mode=block"},
		},
	}
	if m.RateLimit {
		return rateLimit, nil
	}

	if strings.Contains(req.URL.String(), "error") {
		return nil, errors.New("http error")
	} else if strings.HasSuffix(req.URL.String(), "httpError") {
		return error, nil
	}

	return success, nil
}
