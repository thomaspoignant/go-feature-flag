package httpretriever_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"

	"github.com/stretchr/testify/assert"
)

func Test_httpRetriever_Retrieve(t *testing.T) {
	type fields struct {
		httpClient mock.HTTP
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
				httpClient: mock.HTTP{},
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
				httpClient: mock.HTTP{},
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
				httpClient: mock.HTTP{},
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
				httpClient: mock.HTTP{},
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
				httpClient: mock.HTTP{},
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
				httpClient: mock.HTTP{},
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
			h := httpretriever.Retriever{
				URL:    tt.fields.url,
				Method: tt.fields.method,
				Body:   tt.fields.body,
				Header: tt.fields.header,
			}
			h.SetHTTPClient(&tt.fields.httpClient)
			got, err := h.Retrieve(tt.fields.context)
			assert.Equal(t, tt.wantErr, err != nil, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)

			if tt.fields.method == "" {
				assert.Equal(t, http.MethodGet, tt.fields.httpClient.Req.Method)
			}

			if !t.Failed() {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
