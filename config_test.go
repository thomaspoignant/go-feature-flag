package ffclient_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ffClient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gitlabretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retrieverv2"
)

func TestConfig_Initialize(t *testing.T) {
	t.Run("provided context", func(t *testing.T) {
		ctx := context.WithValue(t.Context(), "type", "custom")

		c := ffClient.Config{
			Context: ctx,
		}
		c.Initialize()

		assert.Equal(t, ctx, c.Context)
	})

	t.Run("default context", func(t *testing.T) {
		c := ffClient.Config{}
		c.Initialize()

		assert.Equal(t, context.Background(), c.Context)
	})

	t.Run("adjust polling interval", func(t *testing.T) {
		tests := []struct {
			name            string
			pollingInterval time.Duration
			want            time.Duration
		}{
			{
				name:            "empty",
				pollingInterval: 0,
				want:            60 * time.Second,
			},
			{
				name:            "lower than minimum",
				pollingInterval: 1 * time.Millisecond,
				want:            1 * time.Second,
			},
			{
				name:            "valid",
				pollingInterval: 42 * time.Minute,
				want:            42 * time.Minute,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := ffClient.Config{
					PollingInterval: tt.pollingInterval,
				}
				c.Initialize()

				assert.Equal(t, tt.want, c.PollingInterval)
			})
		}
	})
}

func TestConfig_GetRetrievers(t *testing.T) {
	type fields struct {
		PollingInterval time.Duration
		Retriever       retriever.Retriever
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "File retriever",
			fields: fields{
				PollingInterval: 3 * time.Second,
				Retriever:       &fileretriever.Retriever{Path: "file-example.yaml"},
			},
			want:    "*fileretriever.Retriever",
			wantErr: false,
		},
		{
			name: "S3 retriever",
			fields: fields{
				PollingInterval: 3 * time.Second,
				Retriever: &s3retrieverv2.Retriever{
					Bucket: "tpoi-test",
					Item:   "flag-config.yaml",
				},
			},
			want:    "*s3retrieverv2.Retriever",
			wantErr: false,
		},
		{
			name: "HTTP retriever",
			fields: fields{
				PollingInterval: 3 * time.Second,
				Retriever: &httpretriever.Retriever{
					URL:    "http://example.com/flag-config.yaml",
					Method: http.MethodGet,
				},
			},
			want:    "*httpretriever.Retriever",
			wantErr: false,
		},
		{
			name: "Github retriever",
			fields: fields{
				PollingInterval: 3 * time.Second,
				Retriever: &githubretriever.Retriever{
					RepositorySlug: "thomaspoignant/go-feature-flag",
					FilePath:       "testdata/flag-config.yaml",
					GithubToken:    "XXX",
				},
			},
			want:    "*githubretriever.Retriever",
			wantErr: false,
		},
		{
			name: "Gitlab retriever",
			fields: fields{
				PollingInterval: 3 * time.Second,
				Retriever: &gitlabretriever.Retriever{
					BaseURL:        "https://gitlab.com/",
					RepositorySlug: "ruairi2/go-feature-flags-config",
					FilePath:       "flag-config.yaml",
					GitlabToken:    "XXX",
				},
			},
			want:    "*gitlabretriever.Retriever",
			wantErr: false,
		},
		{
			name: "No retriever",
			fields: fields{
				PollingInterval: 3 * time.Second,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ffClient.Config{
				PollingInterval: tt.fields.PollingInterval,
				Retriever:       tt.fields.Retriever,
			}
			got, err := c.GetRetrievers()
			assert.Equal(t, tt.wantErr, err != nil)
			if err == nil {
				assert.Equal(t, tt.want, reflect.ValueOf(got[0]).Type().String())
			}
		})
	}
}

func TestOfflineConfig(t *testing.T) {
	c := ffClient.Config{
		Offline: true,
	}
	assert.True(t, c.IsOffline())
	c.SetOffline(false)
	assert.False(t, c.IsOffline())
}

func TestConfig_GetDataExporters(t *testing.T) {
	type fields struct {
		DataExporter  ffClient.DataExporter
		DataExporters []ffClient.DataExporter
	}
	tests := []struct {
		name   string
		fields fields
		want   []ffClient.DataExporter
	}{
		{
			name:   "No data exporter",
			fields: fields{},
			want:   []ffClient.DataExporter{},
		},
		{
			name: "Single data exporter",
			fields: fields{
				DataExporter: ffClient.DataExporter{
					FlushInterval:    10 * time.Second,
					MaxEventInMemory: 100,
				},
			},
			want: []ffClient.DataExporter{
				{
					FlushInterval:    10 * time.Second,
					MaxEventInMemory: 100,
				},
			},
		},
		{
			name: "Multiple data exporters",
			fields: fields{
				DataExporters: []ffClient.DataExporter{
					{
						FlushInterval:    20 * time.Second,
						MaxEventInMemory: 200,
					},
					{
						FlushInterval:    30 * time.Second,
						MaxEventInMemory: 300,
					},
				},
			},
			want: []ffClient.DataExporter{
				{
					FlushInterval:    20 * time.Second,
					MaxEventInMemory: 200,
				},
				{
					FlushInterval:    30 * time.Second,
					MaxEventInMemory: 300,
				},
			},
		},
		{
			name: "Both single and multiple data exporters",
			fields: fields{
				DataExporter: ffClient.DataExporter{
					FlushInterval:    10 * time.Second,
					MaxEventInMemory: 100,
				},
				DataExporters: []ffClient.DataExporter{
					{
						FlushInterval:    20 * time.Second,
						MaxEventInMemory: 200,
					},
					{
						FlushInterval:    30 * time.Second,
						MaxEventInMemory: 300,
					},
				},
			},
			want: []ffClient.DataExporter{
				{
					FlushInterval:    10 * time.Second,
					MaxEventInMemory: 100,
				},
				{
					FlushInterval:    20 * time.Second,
					MaxEventInMemory: 200,
				},
				{
					FlushInterval:    30 * time.Second,
					MaxEventInMemory: 300,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ffClient.Config{
				DataExporter:  tt.fields.DataExporter,
				DataExporters: tt.fields.DataExporters,
			}
			got := c.GetDataExporters()
			assert.Equal(t, tt.want, got)
		})
	}
}
