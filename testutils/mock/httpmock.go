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
 rule: key eq "random-key"
 percentage: 100
 true: true
 false: false
 default: false
`))),
	}

	error := &http.Response{
		Status:     "KO",
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewReader([]byte(""))),
	}

	if strings.HasSuffix(req.URL.String(), "error") {
		return nil, errors.New("http error")
	} else if strings.HasSuffix(req.URL.String(), "httpError") {
		return error, nil
	}

	return success, nil
}
