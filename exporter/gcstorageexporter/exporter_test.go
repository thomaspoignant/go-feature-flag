package gcstorageexporter_test

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/gcstorageexporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"google.golang.org/api/option"
)

func TestGoogleStorage_Export(t *testing.T) {
	hostname, _ := os.Hostname()
	type fields struct {
		Bucket      string
		Format      string
		Path        string
		Filename    string
		CsvTemplate string
	}

	tests := []struct {
		name         string
		fields       fields
		events       []exporter.ExportableEvent
		wantErr      bool
		expectedName string
	}{
		{
			name: "All default test",
			fields: fields{
				Bucket: "test",
			},
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
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
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
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
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
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
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
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
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
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
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
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
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
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
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
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
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
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

			// init DeprecatedExporterV1
			f := gcstorageexporter.Exporter{
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

			err := f.Export(
				context.Background(),
				&fflog.FFLogger{LeveledLogger: slog.Default()},
				tt.events,
			)
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
		// Return a minimal but valid GCS object response so the client's writer
		// can decode it on Close() without reporting an upload error.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"test-object","bucket":"test"}`))
	})
}

// TestGoogleStorage_Export_UploadError ensures that a failed upload is reported
// instead of being silently swallowed. The GCS writer buffers the payload and
// flushes it on Close, so the server error surfaces from wc.Close() rather than
// io.Copy; Export must still return an error.
func TestGoogleStorage_Export_UploadError(t *testing.T) {
	serv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"code":500,"message":"boom"}}`))
	}))
	httpclient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // nolint: gosec
			},
		},
	}

	f := gcstorageexporter.Exporter{
		Bucket: "test",
		Options: []option.ClientOption{
			option.WithEndpoint(serv.URL),
			option.WithoutAuthentication(),
			option.WithHTTPClient(&httpclient),
		},
	}

	err := f.Export(
		context.Background(),
		&fflog.FFLogger{LeveledLogger: slog.Default()},
		[]exporter.ExportableEvent{
			exporter.FeatureEvent{
				Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
				Variation: "Default", Value: "YO", Default: false,
			},
		},
	)
	assert.Error(t, err, "Export should error when the upload fails")
}

func TestGoogleCloudStorage_IsBulk(t *testing.T) {
	exporter := gcstorageexporter.Exporter{}
	assert.True(t, exporter.IsBulk(), "exporter is a bulk exporter")
}
