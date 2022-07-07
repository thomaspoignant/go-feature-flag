package fileexporter_test

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"

	"github.com/stretchr/testify/assert"

	"github.com/thomaspoignant/go-feature-flag/exporter"
)

func TestFile_Export(t *testing.T) {
	hostname, _ := os.Hostname()
	type fields struct {
		Format      string
		Filename    string
		CsvTemplate string
		OutputDir   string
	}
	type args struct {
		logger        *log.Logger
		featureEvents []exporter.FeatureEvent
	}
	type expected struct {
		fileNameRegex string
		content       string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		expected expected
	}{
		{
			name:    "all default json",
			wantErr: false,
			fields:  fields{},
			args: args{
				featureEvents: []exporter.FeatureEvent{
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false,
					},
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false, Version: 127,
					},
				},
			},
			expected: expected{
				fileNameRegex: "^flag-variation-" + hostname + "-[0-9]*\\.json$",
				content:       "./testdata/all_default.json",
			},
		},
		{
			name:    "all default csv",
			wantErr: false,
			fields: fields{
				Format: "csv",
			},
			args: args{
				featureEvents: []exporter.FeatureEvent{
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false,
					},
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false,
					},
				},
			},
			expected: expected{
				fileNameRegex: "^flag-variation-" + hostname + "-[0-9]*\\.csv",
				content:       "./testdata/all_default.csv",
			},
		},
		{
			name:    "custom CSV format",
			wantErr: false,
			fields: fields{
				Format:      "csv",
				CsvTemplate: "{{ .Kind}};{{ .ContextKind}}\n",
			},
			args: args{
				featureEvents: []exporter.FeatureEvent{
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false,
					},
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false,
					},
				},
			},
			expected: expected{
				fileNameRegex: "^flag-variation-" + hostname + "-[0-9]*\\.csv",
				content:       "./testdata/custom_csv_format.csv",
			},
		},
		{
			name:    "custom file name",
			wantErr: false,
			fields: fields{
				Format:   "json",
				Filename: "{{ .Format}}-test-{{ .Timestamp}}",
			},
			args: args{
				featureEvents: []exporter.FeatureEvent{
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false,
					},
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false,
					},
				},
			},
			expected: expected{
				fileNameRegex: "^json-test-[0-9]*$",
				content:       "./testdata/custom_file_name.json",
			},
		},
		{
			name:    "invalid format",
			wantErr: false,
			fields: fields{
				Format: "xxx",
			},
			args: args{
				featureEvents: []exporter.FeatureEvent{
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false,
					},
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false, Version: 127,
					},
				},
			},
			expected: expected{
				fileNameRegex: "^flag-variation-" + hostname + "-[0-9]*\\.xxx$",
				content:       "./testdata/all_default.json",
			},
		},
		{
			name:    "invalid outputdir",
			wantErr: true,
			fields: fields{
				OutputDir: "/tmp/foo/bar/",
			},
			args: args{
				featureEvents: []exporter.FeatureEvent{
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false,
					},
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false,
					},
				},
			},
		},
		{
			name:    "invalid filename template",
			wantErr: true,
			fields: fields{
				Filename: "{{ .InvalidField}}",
			},
			args: args{
				featureEvents: []exporter.FeatureEvent{
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false,
					},
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false,
					},
				},
			},
		},
		{
			name:    "invalid csv formatter",
			wantErr: true,
			fields: fields{
				Format:      "csv",
				CsvTemplate: "{{ .Foo}}",
			},
			args: args{
				featureEvents: []exporter.FeatureEvent{
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false,
					},
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputDir := tt.fields.OutputDir
			if tt.fields.OutputDir == "" {
				outputDir, _ = ioutil.TempDir("", "fileExporter")
				defer os.Remove(outputDir)
			}

			f := &fileexporter.Exporter{
				Format:      tt.fields.Format,
				OutputDir:   outputDir,
				Filename:    tt.fields.Filename,
				CsvTemplate: tt.fields.CsvTemplate,
			}
			err := f.Export(context.Background(), tt.args.logger, tt.args.featureEvents)
			if tt.wantErr {
				assert.Error(t, err, "export method should error")
				return
			}

			files, _ := ioutil.ReadDir(outputDir)
			assert.Equal(t, 1, len(files), "Directory %s should have only one file", outputDir)
			assert.Regexp(t, tt.expected.fileNameRegex, files[0].Name(), "Invalid file name")

			expectedContent, _ := ioutil.ReadFile(tt.expected.content)
			gotContent, _ := ioutil.ReadFile(outputDir + "/" + files[0].Name())
			assert.Equal(t, string(expectedContent), string(gotContent), "Wrong content in the output file")
		})
	}
}

func TestFile_IsBulk(t *testing.T) {
	exporter := fileexporter.Exporter{}
	assert.True(t, exporter.IsBulk(), "Exporter exporter is a bulk exporter")
}
