package ffexporter_test

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffexporter"
	"google.golang.org/api/option"
)

func TestGoogleStorage_Export(t *testing.T) {
	hostname, _ := os.Hostname()
	type fields struct {
		Bucket      string
		AwsConfig   *aws.Config
		Format      string
		Path        string
		Filename    string
		CsvTemplate string
	}

	tests := []struct {
		name         string
		fields       fields
		events       []ffexporter.FeatureEvent
		wantErr      bool
		expectedName string
	}{
		{
			name: "All default test",
			fields: fields{
				Bucket: "test",
			},
			events: []ffexporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			expectedName: "^flag-variation-" + hostname + "-[0-9]*\\.json$",
		},
		{
			name: "With S3 Path",
			fields: fields{
				Path:   "random/path",
				Bucket: "test",
			},
			events: []ffexporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			expectedName: "^random/path/flag-variation-" + hostname + "-[0-9]*\\.json$",
		},
		{
			name: "All default CSV",
			fields: fields{
				Format: "csv",
				Bucket: "test",
			},
			events: []ffexporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			expectedName: "^flag-variation-" + hostname + "-[0-9]*\\.csv$",
		},
		{
			name: "Custom CSV",
			fields: fields{
				Format:      "csv",
				CsvTemplate: "{{ .Kind}};{{ .ContextKind}}\n",
				Bucket:      "test",
			},
			events: []ffexporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			expectedName: "^flag-variation-" + hostname + "-[0-9]*\\.csv$",
		},
		{
			name: "Custom FileName",
			fields: fields{
				Format:   "json",
				Filename: "{{ .Format}}-test-{{ .Timestamp}}",
				Bucket:   "test",
			},
			events: []ffexporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			expectedName: "^json-test-[0-9]*$",
		},
		{
			name: "Invalid format",
			fields: fields{
				Format: "xxx",
				Bucket: "test",
			},
			events: []ffexporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			expectedName: "^flag-variation-" + hostname + "-[0-9]*\\.xxx$",
		},
		{
			name: "Empty Bucket",
			fields: fields{
				Format: "xxx",
			},
			events: []ffexporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid filename template",
			fields: fields{
				Filename: "{{ .InvalidField}}",
				Bucket:   "test",
			},
			events: []ffexporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid csv formatter",
			fields: fields{
				Format:      "csv",
				CsvTemplate: "{{ .Foo}}",
			},
			events: []ffexporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init mock httpclient
			handler := GoogleCloudStorageHandler{}

			serv := httptest.NewTLSServer(handler.handler())
			httpclient := http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true, // nolint: gosec
					},
				},
			}

			// init GoogleCloudStorage
			f := ffexporter.GoogleCloudStorage{
				Bucket: tt.fields.Bucket,
				Options: []option.ClientOption{
					option.WithEndpoint(serv.URL),
					option.WithoutAuthentication(),
					option.WithHTTPClient(&httpclient),
				},
				Format:      tt.fields.Format,
				Path:        tt.fields.Path,
				Filename:    tt.fields.Filename,
				CsvTemplate: tt.fields.CsvTemplate,
			}

			err := f.Export(context.Background(), log.New(os.Stdout, "", 0), tt.events)
			if tt.wantErr {
				assert.Error(t, err, "Export should error")
				return
			}

			assert.NoError(t, err, "Export should not error")
			extractedFileName := handler.r.URL.Query().Get("name")
			assert.Regexp(t, tt.expectedName, extractedFileName)

			// check that the bucket name is in the URL
			assert.Equal(t, "/upload/storage/v1/b/"+tt.fields.Bucket+"/o", handler.r.URL.Path)
		})
	}
}

type GoogleCloudStorageHandler struct {
	r *http.Request
}

func (g *GoogleCloudStorageHandler) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.r = r
		w.WriteHeader(http.StatusOK)
	})
}

func TestGoogleCloudStorage_IsBulk(t *testing.T) {
	exporter := ffexporter.GoogleCloudStorage{}
	assert.True(t, exporter.IsBulk(), "S3 exporter is not a bulk exporter")
}
