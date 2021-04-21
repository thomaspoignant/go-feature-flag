package testutils

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
)

type HTTPClientMock struct {
	ForceError bool
	StatusCode int
	Body       string
	Signature  string
}

func (h *HTTPClientMock) Do(req *http.Request) (*http.Response, error) {
	if h.ForceError {
		return nil, errors.New("random error")
	}

	b, _ := ioutil.ReadAll(req.Body)
	h.Body = string(b)
	h.Signature = req.Header.Get("X-Hub-Signature-256")
	resp := &http.Response{
		Body: ioutil.NopCloser(bytes.NewReader([]byte(""))),
	}
	resp.StatusCode = h.StatusCode
	return resp, nil
}
