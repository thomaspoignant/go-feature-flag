package mock

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
)

type HTTP struct {
	Req http.Request
}

func (m *HTTP) Do(req *http.Request) (*http.Response, error) {
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

	if strings.Contains(req.URL.String(), "error") {
		return nil, errors.New("http error")
	} else if strings.HasSuffix(req.URL.String(), "httpError") {
		return error, nil
	}

	return success, nil
}
