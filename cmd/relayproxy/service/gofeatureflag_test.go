package service

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/gcstorageexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/logsexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/s3exporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/webhookexporter"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gcstorageretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gitlabretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retriever"
	"github.com/xitongsys/parquet-go/parquet"
)

func Test_initRetriever(t *testing.T) {
	tests := []struct {
		name    string
		conf    *config.RetrieverConf
		want    retriever.Retriever
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Convert Github Retriever",
			wantErr: assert.NoError,
			conf: &config.RetrieverConf{
				Kind:           "github",
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Path:           "testdata/flag-config.yaml",
				Timeout:        20,
			},
			want: &githubretriever.Retriever{
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Branch:         "main",
				FilePath:       "testdata/flag-config.yaml",
				GithubToken:    "",
				Timeout:        20 * time.Millisecond,
			},
		},
		{
			name:    "Convert Github Retriever with token",
			wantErr: assert.NoError,
			conf: &config.RetrieverConf{
				Kind:           "github",
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Path:           "testdata/flag-config.yaml",
				Timeout:        20,
				AuthToken:      "xxx",
			},
			want: &githubretriever.Retriever{
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Branch:         "main",
				FilePath:       "testdata/flag-config.yaml",
				GithubToken:    "xxx",
				Timeout:        20 * time.Millisecond,
			},
		},
		{
			name:    "Convert Github Retriever with deprecated token",
			wantErr: assert.NoError,
			conf: &config.RetrieverConf{
				Kind:           "github",
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Path:           "testdata/flag-config.yaml",
				Timeout:        20,
				GithubToken:    "xxx",
			},
			want: &githubretriever.Retriever{
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Branch:         "main",
				FilePath:       "testdata/flag-config.yaml",
				GithubToken:    "xxx",
				Timeout:        20 * time.Millisecond,
			},
		},
		{
			name:    "Convert Gitlab Retriever",
			wantErr: assert.NoError,
			conf: &config.RetrieverConf{
				Kind:           "gitlab",
				BaseURL:        "http://localhost",
				Path:           "flag-config.yaml",
				RepositorySlug: "thomaspoignant/go-feature-flag",
				Timeout:        20,
			},
			want: &gitlabretriever.Retriever{
				BaseURL:        "http://localhost",
				Branch:         "main",
				FilePath:       "flag-config.yaml",
				RepositorySlug: "thomaspoignant/go-feature-flag",
				GitlabToken:    "",
				Timeout:        20 * time.Millisecond,
			},
		},
		{
			name:    "Convert File Retriever",
			wantErr: assert.NoError,
			conf: &config.RetrieverConf{
				Kind: "file",
				Path: "testdata/flag-config.yaml",
			},
			want: &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
		},
		{
			name:    "Convert S3 Retriever",
			wantErr: assert.NoError,
			conf: &config.RetrieverConf{
				Kind:   "s3",
				Bucket: "my-bucket-name",
				Item:   "testdata/flag-config.yaml",
			},
			want: &s3retriever.Retriever{
				Bucket: "my-bucket-name",
				Item:   "testdata/flag-config.yaml",
			},
		},
		{
			name:    "Convert HTTP Retriever",
			wantErr: assert.NoError,
			conf: &config.RetrieverConf{
				Kind: "http",
				URL:  "https://gofeatureflag.org/my-flag-test.yaml",
			},
			want: &httpretriever.Retriever{
				URL:     "https://gofeatureflag.org/my-flag-test.yaml",
				Method:  http.MethodGet,
				Body:    "",
				Header:  nil,
				Timeout: 10000000000,
			},
		}, {
			name:    "Convert Google storage Retriever",
			wantErr: assert.NoError,
			conf: &config.RetrieverConf{
				Kind:   "googleStorage",
				Bucket: "my-bucket-name",
				Object: "testdata/flag-config.yaml",
			},
			want: &gcstorageretriever.Retriever{
				Bucket:  "my-bucket-name",
				Object:  "testdata/flag-config.yaml",
				Options: nil,
			},
		},
		{
			name:    "Convert unknown Retriever",
			wantErr: assert.Error,
			conf: &config.RetrieverConf{
				Kind: "unknown",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := initRetriever(tt.conf)
			tt.wantErr(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_initExporter(t *testing.T) {
	tests := []struct {
		name    string
		conf    *config.ExporterConf
		want    ffclient.DataExporter
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Convert unknown Exporter",
			wantErr: assert.Error,
			conf: &config.ExporterConf{
				Kind: "unknown",
			},
		},
		{
			name:    "Convert WebhookExporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind:        "webhook",
				EndpointURL: "https://gofeatureflag.org/webhook-example",
				Secret:      "1234",
			},
			want: ffclient.DataExporter{
				FlushInterval:    config.DefaultExporter.FlushInterval,
				MaxEventInMemory: config.DefaultExporter.MaxEventInMemory,
				Exporter: &webhookexporter.Exporter{
					EndpointURL: "https://gofeatureflag.org/webhook-example",
					Secret:      "1234",
					Meta:        nil,
				},
			},
		},
		{
			name:    "Convert FileExporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind:                    "file",
				OutputDir:               "/outputfolder/",
				ParquetCompressionCodec: parquet.CompressionCodec_UNCOMPRESSED.String(),
			},
			want: ffclient.DataExporter{
				FlushInterval:    config.DefaultExporter.FlushInterval,
				MaxEventInMemory: config.DefaultExporter.MaxEventInMemory,
				Exporter: &fileexporter.Exporter{
					Format:                  config.DefaultExporter.Format,
					OutputDir:               "/outputfolder/",
					Filename:                config.DefaultExporter.FileName,
					CsvTemplate:             config.DefaultExporter.CsvFormat,
					ParquetCompressionCodec: parquet.CompressionCodec_UNCOMPRESSED.String(),
				},
			},
		},
		{
			name:    "Convert LogExporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind: "log",
			},
			want: ffclient.DataExporter{
				FlushInterval:    config.DefaultExporter.FlushInterval,
				MaxEventInMemory: config.DefaultExporter.MaxEventInMemory,
				Exporter: &logsexporter.Exporter{
					LogFormat: config.DefaultExporter.LogFormat,
				},
			},
		},
		{
			name:    "Convert S3Exporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind:          "s3",
				Bucket:        "my-bucket",
				Path:          "/my-path/",
				FlushInterval: 10,
			},
			want: ffclient.DataExporter{
				FlushInterval:    10 * time.Millisecond,
				MaxEventInMemory: config.DefaultExporter.MaxEventInMemory,
				Exporter: &s3exporter.Exporter{
					Bucket:                  "my-bucket",
					Format:                  config.DefaultExporter.Format,
					S3Path:                  "/my-path/",
					Filename:                config.DefaultExporter.FileName,
					CsvTemplate:             config.DefaultExporter.CsvFormat,
					ParquetCompressionCodec: config.DefaultExporter.ParquetCompressionCodec,
				},
			},
		},
		{
			name:    "Convert GoogleStorageExporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind:             "googleStorage",
				Bucket:           "my-bucket",
				Path:             "/my-path/",
				MaxEventInMemory: 1990,
			},
			want: ffclient.DataExporter{
				FlushInterval:    config.DefaultExporter.FlushInterval,
				MaxEventInMemory: 1990,
				Exporter: &gcstorageexporter.Exporter{
					Bucket:                  "my-bucket",
					Format:                  config.DefaultExporter.Format,
					Path:                    "/my-path/",
					Filename:                config.DefaultExporter.FileName,
					CsvTemplate:             config.DefaultExporter.CsvFormat,
					ParquetCompressionCodec: config.DefaultExporter.ParquetCompressionCodec,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := initExporter(tt.conf)
			tt.wantErr(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
