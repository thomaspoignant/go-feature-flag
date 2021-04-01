package retriever_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/retriever"
)

type mockHTTP struct {
	req http.Request
}

func (m *mockHTTP) Do(req *http.Request) (*http.Response, error) {
	m.req = *req
	success := &http.Response{
		Status:     "OK",
		StatusCode: http.StatusOK,
		Body: ioutil.NopCloser(bytes.NewReader([]byte(`test-flag:
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
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
	}

	if strings.HasSuffix(req.URL.String(), "error") {
		return nil, errors.New("http error")
	} else if strings.HasSuffix(req.URL.String(), "httpError") {
		return error, nil
	}

	return success, nil
}

func Test_httpRetriever_Retrieve(t *testing.T) {
	type fields struct {
		httpClient *mockHTTP
		url        string
		method     string
		body       string
		header     http.Header
		context    context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				httpClient: &mockHTTP{},
				url:        "http://localhost.example/file",
				method:     http.MethodGet,
				body:       "",
				header:     nil,
			},
			want: []byte(`test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false
`),
			wantErr: false,
		},
		{
			name: "Success with context",
			fields: fields{
				httpClient: &mockHTTP{},
				url:        "http://localhost.example/file",
				method:     http.MethodGet,
				body:       "",
				header:     nil,
				context:    context.Background(),
			},
			want: []byte(`test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false
`),
			wantErr: false,
		},
		{
			name: "Success with default method",
			fields: fields{
				httpClient: &mockHTTP{},
				url:        "http://localhost.example/file",
				body:       "",
				header:     nil,
			},
			want: []byte(`test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false
`),
			wantErr: false,
		},
		{
			name: "HTTP Error",
			fields: fields{
				httpClient: &mockHTTP{},
				url:        "http://localhost.example/httpError",
				method:     http.MethodPost,
				body:       "",
				header:     map[string][]string{"Content-Type": {"application/json"}},
			},
			wantErr: true,
		},
		{
			name: "Error",
			fields: fields{
				httpClient: &mockHTTP{},
				url:        "http://localhost.example/error",
				method:     http.MethodPost,
				body:       "",
				header:     map[string][]string{"Content-Type": {"application/json"}},
			},
			wantErr: true,
		},
		{
			name: "No URL",
			fields: fields{
				httpClient: &mockHTTP{},
				url:        "",
				method:     http.MethodPost,
				body:       "",
				header:     map[string][]string{"Content-Type": {"application/json"}},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := retriever.NewHTTPRetriever(
				tt.fields.httpClient,
				tt.fields.url,
				tt.fields.method,
				tt.fields.body,
				tt.fields.header,
			)
			got, err := h.Retrieve(tt.fields.context)
			assert.Equal(t, tt.wantErr, err != nil, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)

			if tt.fields.method == "" {
				assert.Equal(t, http.MethodGet, tt.fields.httpClient.req.Method)
			}

			if !t.Failed() {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
