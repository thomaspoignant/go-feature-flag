package azureexporter_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/azureexporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

func TestAzureBlobStorage_Export(t *testing.T) {
	hostname, _ := os.Hostname()
	type fields struct {
		Container   string
		AccountName string
		AccountKey  string
		Format      string
		Path        string
		Filename    string
		CsvTemplate string
	}

	tests := []struct {
		name         string
		fields       fields
		events       []exporter.FeatureEvent
		wantErr      bool
		expectedName string
	}{
		{
			name: "All default test",
			fields: fields{
				Container: "test",
			},
			events: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			expectedName: "^flag-variation-" + hostname + "-[0-9]*\\.json$",
		},
		{
			name: "With AzBlob Path",
			fields: fields{
				Path:      "random/path",
				Container: "test",
			},
			events: []exporter.FeatureEvent{
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
				Format:    "csv",
				Container: "test",
			},
			events: []exporter.FeatureEvent{
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
				Container:   "test",
			},
			events: []exporter.FeatureEvent{
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
				Format:    "json",
				Filename:  "{{ .Format}}-test-{{ .Timestamp}}",
				Container: "test",
			},
			events: []exporter.FeatureEvent{
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
				Format:    "xxx",
				Container: "test",
			},
			events: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			expectedName: "^flag-variation-" + hostname + "-[0-9]*\\.xxx$",
		},
		{
			name: "Empty Container",
			fields: fields{
				Format: "xxx",
			},
			events: []exporter.FeatureEvent{
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
				Filename:  "{{ .InvalidField}}",
				Container: "test",
			},
			events: []exporter.FeatureEvent{
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
			events: []exporter.FeatureEvent{
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
			f := azureexporter.Exporter{
				Container:   tt.fields.Container,
				AccountName: tt.fields.AccountName,
				AccountKey:  tt.fields.AccountKey,
				Format:      tt.fields.Format,
				Path:        tt.fields.Path,
				Filename:    tt.fields.Filename,
				CsvTemplate: tt.fields.CsvTemplate,
			}

			err := f.Export(context.Background(), &fflog.FFLogger{LeveledLogger: slog.Default()}, tt.events)
			if tt.wantErr {
				assert.Error(t, err, "Export should error")
				return
			}
			assert.NoError(t, err, "Export should not error")
		})
	}
}

func TestAzureBlobStorage_IsBulk(t *testing.T) {
	exporter := azureexporter.Exporter{}
	assert.True(t, exporter.IsBulk(), "exporter is a bulk exporter")
}
