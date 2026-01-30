package datadogexporter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// httpClientMock is a mock HTTP client for testing.
type httpClientMock struct {
	forceError bool
	statusCode int
	body       string
	headers    http.Header
}

func (h *httpClientMock) Do(req *http.Request) (*http.Response, error) {
	if h.forceError {
		return nil, errors.New("mock http error")
	}

	b, _ := io.ReadAll(req.Body)
	h.body = string(b)
	h.headers = req.Header.Clone()

	resp := &http.Response{
		Body:       io.NopCloser(bytes.NewReader([]byte(""))),
		StatusCode: h.statusCode,
	}
	return resp, nil
}

func TestDatadog_IsBulk(t *testing.T) {
	exp := Exporter{}
	assert.True(t, exp.IsBulk(), "Datadog exporter should be a bulk exporter")
}

func TestDatadog_Export(t *testing.T) {
	logger := &fflog.FFLogger{LeveledLogger: slog.Default()}

	type fields struct {
		APIKey     string
		Site       string
		Source     string
		Service    string
		Tags       []string
		httpClient *httpClientMock
	}
	type args struct {
		logger        *fflog.FFLogger
		featureEvents []exporter.ExportableEvent
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantErr        bool
		wantErrContain string
		checkFunc      func(t *testing.T, mock *httpClientMock)
	}{
		{
			name: "missing API key should return error",
			fields: fields{
				APIKey:     "",
				httpClient: &httpClientMock{statusCode: 200},
			},
			args: args{
				logger: logger,
				featureEvents: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "user-123",
						CreationDate: 1617970547, Key: "my-flag", Variation: "enabled",
						Value: true, Default: false, Source: "SERVER",
					},
				},
			},
			wantErr:        true,
			wantErrContain: "API key is required",
		},
		{
			name: "successful export with default settings",
			fields: fields{
				APIKey:     "test-api-key",
				httpClient: &httpClientMock{statusCode: 202},
			},
			args: args{
				logger: logger,
				featureEvents: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "user-123",
						CreationDate: 1617970547, Key: "my-flag", Variation: "enabled",
						Value: true, Default: false, Source: "SERVER",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, mock *httpClientMock) {
				assert.Equal(t, "test-api-key", mock.headers.Get("DD-API-KEY"))
				assert.Equal(t, "application/json", mock.headers.Get("Content-Type"))

				var entries []datadogLogEntry
				err := json.Unmarshal([]byte(mock.body), &entries)
				require.NoError(t, err)
				require.Len(t, entries, 1)
				assert.Equal(t, "go-feature-flag", entries[0].DDSource)
				assert.Equal(t, "go-feature-flag", entries[0].Service)
				assert.Equal(t, "my-flag", entries[0].FeatureFlag.Key)
				assert.Equal(t, true, entries[0].FeatureFlag.Value)
				assert.Equal(t, "enabled", entries[0].FeatureFlag.Variation)
				assert.Equal(t, "user-123", entries[0].User.ID)
			},
		},
		{
			name: "successful export with custom settings",
			fields: fields{
				APIKey:     "custom-api-key",
				Site:       "datadoghq.eu",
				Source:     "custom-source",
				Service:    "custom-service",
				Tags:       []string{"env:production", "team:platform"},
				httpClient: &httpClientMock{statusCode: 202},
			},
			args: args{
				logger: logger,
				featureEvents: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "user", UserKey: "user-456",
						CreationDate: 1617970701, Key: "feature-x", Variation: "variant-a",
						Value: "test-value", Default: false, Version: "v1.0.0", Source: "SERVER",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, mock *httpClientMock) {
				assert.Equal(t, "custom-api-key", mock.headers.Get("DD-API-KEY"))

				var entries []datadogLogEntry
				err := json.Unmarshal([]byte(mock.body), &entries)
				require.NoError(t, err)
				require.Len(t, entries, 1)
				assert.Equal(t, "custom-source", entries[0].DDSource)
				assert.Equal(t, "custom-service", entries[0].Service)
				assert.Equal(t, "env:production,team:platform", entries[0].DDTags)
				assert.Equal(t, "feature-x", entries[0].FeatureFlag.Key)
				assert.Equal(t, "test-value", entries[0].FeatureFlag.Value)
				assert.Equal(t, "v1.0.0", entries[0].FeatureFlag.Version)
				assert.Equal(t, "user-456", entries[0].User.ID)
			},
		},
		{
			name: "multiple events should be sent in single request",
			fields: fields{
				APIKey:     "test-api-key",
				httpClient: &httpClientMock{statusCode: 202},
			},
			args: args{
				logger: logger,
				featureEvents: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "user", UserKey: "user-1",
						CreationDate: 1617970547, Key: "flag-1", Variation: "on",
						Value: true, Default: false, Source: "SERVER",
					},
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "user", UserKey: "user-2",
						CreationDate: 1617970548, Key: "flag-2", Variation: "off",
						Value: false, Default: false, Source: "SERVER",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, mock *httpClientMock) {
				var entries []datadogLogEntry
				err := json.Unmarshal([]byte(mock.body), &entries)
				require.NoError(t, err)
				assert.Len(t, entries, 2)
				assert.Equal(t, "flag-1", entries[0].FeatureFlag.Key)
				assert.Equal(t, "flag-2", entries[1].FeatureFlag.Key)
			},
		},
		{
			name: "HTTP error should be returned",
			fields: fields{
				APIKey:     "test-api-key",
				httpClient: &httpClientMock{forceError: true},
			},
			args: args{
				logger: logger,
				featureEvents: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "user", UserKey: "user-123",
						CreationDate: 1617970547, Key: "my-flag", Variation: "enabled",
						Value: true, Default: false, Source: "SERVER",
					},
				},
			},
			wantErr:        true,
			wantErrContain: "failed to send request",
		},
		{
			name: "HTTP 4xx error should be returned",
			fields: fields{
				APIKey:     "test-api-key",
				httpClient: &httpClientMock{statusCode: 403},
			},
			args: args{
				logger: logger,
				featureEvents: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "user", UserKey: "user-123",
						CreationDate: 1617970547, Key: "my-flag", Variation: "enabled",
						Value: true, Default: false, Source: "SERVER",
					},
				},
			},
			wantErr:        true,
			wantErrContain: "received HTTP 403",
		},
		{
			name: "empty events should not make request",
			fields: fields{
				APIKey:     "test-api-key",
				httpClient: &httpClientMock{statusCode: 202},
			},
			args: args{
				logger:        logger,
				featureEvents: []exporter.ExportableEvent{},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, mock *httpClientMock) {
				// No request should have been made
				assert.Empty(t, mock.body)
			},
		},
		{
			name: "default value flag should be indicated",
			fields: fields{
				APIKey:     "test-api-key",
				httpClient: &httpClientMock{statusCode: 202},
			},
			args: args{
				logger: logger,
				featureEvents: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "user", UserKey: "user-123",
						CreationDate: 1617970547, Key: "my-flag", Variation: "SdkDefault",
						Value: "default-value", Default: true, Source: "SERVER",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, mock *httpClientMock) {
				var entries []datadogLogEntry
				err := json.Unmarshal([]byte(mock.body), &entries)
				require.NoError(t, err)
				require.Len(t, entries, 1)
				assert.True(t, entries[0].FeatureFlag.Default)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := &Exporter{
				APIKey:     tt.fields.APIKey,
				Site:       tt.fields.Site,
				Source:     tt.fields.Source,
				Service:    tt.fields.Service,
				Tags:       tt.fields.Tags,
				httpClient: tt.fields.httpClient,
			}

			err := exp.Export(context.Background(), tt.args.logger, tt.args.featureEvents)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrContain != "" {
					assert.Contains(t, err.Error(), tt.wantErrContain)
				}
				return
			}

			assert.NoError(t, err)
			if tt.checkFunc != nil {
				tt.checkFunc(t, tt.fields.httpClient)
			}
		})
	}
}

func TestDatadog_buildTags(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{
			name: "empty tags",
			tags: nil,
			want: "",
		},
		{
			name: "single tag",
			tags: []string{"env:production"},
			want: "env:production",
		},
		{
			name: "multiple tags",
			tags: []string{"env:production", "team:backend", "service:api"},
			want: "env:production,team:backend,service:api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Exporter{Tags: tt.tags}
			result := e.buildTags()
			assert.Equal(t, tt.want, result)
		})
	}
}
