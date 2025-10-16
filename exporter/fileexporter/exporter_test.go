package fileexporter_test

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/reader"
)

func TestFile_Export(t *testing.T) {
	// Create a temporary directory for test file operations
	tempDir, err := os.MkdirTemp("", "fileexporter-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after tests

	hostname, _ := os.Hostname()
	type fields struct {
		Format                  string
		Filename                string
		CsvTemplate             string
		OutputDir               string
		ParquetCompressionCodec string
		EventType               string
	}
	type args struct {
		logger *fflog.FFLogger
		events []exporter.ExportableEvent
	}
	type expected struct {
		fileNameRegex  string
		content        string
		featureEvents  []exporter.FeatureEvent
		trackingEvents []exporter.TrackingEvent
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		expected expected
		setup    func(t *testing.T, fields fields)
		teardown func(t *testing.T, fields fields)
	}{
		{
			name:    "all default json",
			wantErr: false,
			fields:  fields{},
			args: args{
				events: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false, Source: "SERVER",
					},
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false, Version: "127", Source: "SERVER",
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
				events: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false, Source: "SERVER",
					},
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false, Source: "SERVER",
					},
				},
			},
			expected: expected{
				fileNameRegex: "^flag-variation-" + hostname + "-[0-9]*\\.csv",
				content:       "./testdata/all_default.csv",
			},
		},
		{
			name:    "all default parquet",
			wantErr: false,
			fields: fields{
				Format:                  "parquet",
				ParquetCompressionCodec: parquet.CompressionCodec_SNAPPY.String(),
			},
			args: args{
				events: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false, Source: "SERVER", Metadata: map[string]interface{}{"test": "test"},
					},
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false, Version: "127", Source: "SERVER",
					},
				},
			},
			expected: expected{
				fileNameRegex: "^flag-variation-" + hostname + "-[0-9]*\\.parquet$",
				featureEvents: []exporter.FeatureEvent{
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: `"YO"`, Default: false, Source: "SERVER", Metadata: map[string]interface{}{"test": "test"},
					},
					{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: `"YO2"`, Default: false, Version: "127", Source: "SERVER", Metadata: map[string]interface{}{},
					},
				},
			},
		},
		{
			name:    "all default parquet tracking events",
			wantErr: false,
			fields: fields{
				Format:                  "parquet",
				ParquetCompressionCodec: parquet.CompressionCodec_SNAPPY.String(),
				Filename:                "tracking-{{ .Hostname}}-{{ .Timestamp}}.parquet",
				EventType:               "tracking",
			},
			args: args{

				events: []exporter.ExportableEvent{
					exporter.TrackingEvent{
						Kind:              "feature",
						ContextKind:       "anonymous",
						UserKey:           "xxx",
						CreationDate:      1617970547,
						Key:               "what-ever-you-want",
						EvaluationContext: ffcontext.NewEvaluationContext("xxx-xxx-xxx").ToMap(),
						TrackingDetails:   map[string]interface{}{"foo": "bar"},
					},
				},
			},
			expected: expected{
				fileNameRegex: "^tracking-" + hostname + "-[0-9]*\\.parquet$",
				trackingEvents: []exporter.TrackingEvent{
					{
						Kind:              "feature",
						ContextKind:       "anonymous",
						UserKey:           "xxx",
						CreationDate:      1617970547,
						Key:               "what-ever-you-want",
						EvaluationContext: ffcontext.NewEvaluationContext("xxx-xxx-xxx").ToMap(),
						TrackingDetails:   map[string]interface{}{"foo": "bar"},
					},
				},
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
				events: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false, Source: "SERVER",
					},
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false, Source: "SERVER",
					},
				},
			},
			expected: expected{
				fileNameRegex: "^flag-variation-" + hostname + "-[0-9]*\\.csv",
				content:       "./testdata/custom_csv_format.csv",
			},
		},
		{
			name:    "complex parquet value",
			wantErr: false,
			fields: fields{
				Format:                  "parquet",
				ParquetCompressionCodec: parquet.CompressionCodec_UNCOMPRESSED.String(),
			},
			args: args{
				events: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind:         "feature",
						ContextKind:  "anonymousUser",
						UserKey:      "ABCD",
						CreationDate: 1617970547,
						Key:          "random-key",
						Variation:    "Default",
						Value: map[string]interface{}{
							"string": "string",
							"bool":   true,
							"float":  1.23,
							"int":    1,
						},
						Default:  false,
						Source:   "SERVER",
						Metadata: map[string]interface{}{"test": "test"},
					},
				},
			},
			expected: expected{
				fileNameRegex: "^flag-variation-" + hostname + "-[0-9]*\\.parquet$",
				featureEvents: []exporter.FeatureEvent{
					{
						Kind:         "feature",
						ContextKind:  "anonymousUser",
						UserKey:      "ABCD",
						CreationDate: 1617970547,
						Key:          "random-key",
						Variation:    "Default",
						Value:        `{"bool":true,"float":1.23,"int":1,"string":"string"}`,
						Default:      false,
						Source:       "SERVER",
						Metadata:     map[string]interface{}{"test": "test"},
					},
				},
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
				events: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false, Source: "SERVER",
					},
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false, Source: "SERVER",
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
				events: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false, Source: "SERVER",
					},
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false, Version: "127", Source: "SERVER",
					},
				},
			},
			expected: expected{
				fileNameRegex: "^flag-variation-" + hostname + "-[0-9]*\\.xxx$",
				content:       "./testdata/all_default.json",
			},
		},
		{
			name:    "non-existent outputdir",
			wantErr: false,
			fields: fields{
				OutputDir: filepath.Join(tempDir, "non-existent-dir"),
			},
			args: args{
				events: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false, Source: "SERVER",
					},
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false, Version: "127", Source: "SERVER",
					},
				},
			},
			expected: expected{
				fileNameRegex: "^flag-variation-" + hostname + "-[0-9]*\\.json$",
				content:       "./testdata/all_default.json",
			},
		},
		{
			name:    "invalid filename template",
			wantErr: true,
			fields: fields{
				Filename: "{{ .InvalidField}}",
			},
			args: args{
				events: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false, Source: "SERVER",
					},
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false, Source: "SERVER",
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
				events: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false, Source: "SERVER",
					},
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false, Source: "SERVER",
					},
				},
			},
		},
		{
			name:    "outputdir with invalid permissions",
			wantErr: true,
			fields: fields{
				Format:    "parquet",
				OutputDir: filepath.Join(tempDir, "invalid-permissions-dir"),
			},
			args: args{
				events: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false, Source: "SERVER",
					},
				},
			},
			setup: func(t *testing.T, fields fields) {
				err := os.MkdirAll(fields.OutputDir, 0755)
				assert.NoError(t, err)
				err = os.Chmod(fields.OutputDir, 0000) // Remove all permissions
				assert.NoError(t, err)
			},
			teardown: func(t *testing.T, fields fields) {
				err := os.Chmod(fields.OutputDir, 0755) // Restore permissions for cleanup
				assert.NoError(t, err)
			},
		},
		{
			name:    "outputdir with invalid parent folder",
			wantErr: true,
			fields: fields{
				Format:    "parquet",
				OutputDir: filepath.Join(tempDir, "invalid-parent-dir"),
			},
			args: args{
				events: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false, Source: "SERVER",
					},
				},
			},
			setup: func(t *testing.T, fields fields) {
				err := os.MkdirAll(tempDir, 0755)
				assert.NoError(t, err)
				err = os.Chmod(tempDir, 0000) // Remove all permissions
				assert.NoError(t, err)
			},
			teardown: func(t *testing.T, fields fields) {
				err := os.Chmod(tempDir, 0755) // Restore permissions for cleanup
				assert.NoError(t, err)
			},
		},
		{
			name:    "outputdir with trailing slash",
			wantErr: false,
			fields: fields{
				Format:    "json",
				OutputDir: filepath.Join(tempDir, "dir-with-trailing-slash") + "/",
			},
			args: args{
				events: []exporter.ExportableEvent{
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false, Source: "SERVER",
					},
					exporter.FeatureEvent{
						Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false, Version: "127", Source: "SERVER",
					},
				},
			},
			expected: expected{
				fileNameRegex: "^flag-variation-" + hostname + "-[0-9]*\\.json$",
				content:       "./testdata/all_default.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputDir := tt.fields.OutputDir
			if tt.fields.OutputDir == "" {
				outputDir, _ = os.MkdirTemp("", "fileExporter")
				defer os.Remove(outputDir)
			}

			if tt.setup != nil {
				tt.setup(t, tt.fields)
			}

			if tt.teardown != nil {
				defer tt.teardown(t, tt.fields)
			}

			f := &fileexporter.Exporter{
				Format:                  tt.fields.Format,
				OutputDir:               outputDir,
				Filename:                tt.fields.Filename,
				CsvTemplate:             tt.fields.CsvTemplate,
				ParquetCompressionCodec: tt.fields.ParquetCompressionCodec,
			}
			err := f.Export(context.Background(), tt.args.logger, tt.args.events)
			if tt.wantErr {
				assert.Error(t, err, "export method should error")
				return
			}

			assert.NoError(t, err)

			// Check if the directory was created
			_, err = os.Stat(outputDir)
			assert.NoError(t, err, "Output directory should exist")

			files, _ := os.ReadDir(outputDir)
			assert.Equal(t, 1, len(files), "Directory %s should have only one file", outputDir)
			assert.Regexp(t, tt.expected.fileNameRegex, files[0].Name(), "Invalid file name")

			if tt.fields.Format == "parquet" {
				switch tt.fields.EventType {
				case "tracking":
					fr, err := local.NewLocalFileReader(outputDir + "/" + files[0].Name())
					assert.NoError(t, err)
					defer fr.Close()
					pr, err := reader.NewParquetReader(
						fr,
						new(exporter.TrackingEvent),
						int64(runtime.NumCPU()),
					)
					assert.NoError(t, err)
					defer pr.ReadStop()
					gotFeatureEvents := make([]exporter.TrackingEvent, pr.GetNumRows())
					err = pr.Read(&gotFeatureEvents)
					assert.NoError(t, err)
					assert.ElementsMatch(t, tt.expected.trackingEvents, gotFeatureEvents)
					return
				default:
					fr, err := local.NewLocalFileReader(outputDir + "/" + files[0].Name())
					assert.NoError(t, err)
					defer fr.Close()
					pr, err := reader.NewParquetReader(
						fr,
						new(exporter.FeatureEvent),
						int64(runtime.NumCPU()),
					)
					assert.NoError(t, err)
					defer pr.ReadStop()
					gotFeatureEvents := make([]exporter.FeatureEvent, pr.GetNumRows())
					err = pr.Read(&gotFeatureEvents)
					assert.NoError(t, err)
					assert.ElementsMatch(t, tt.expected.featureEvents, gotFeatureEvents)
					return
				}
			}

			expectedContent, _ := os.ReadFile(tt.expected.content)
			gotContent, _ := os.ReadFile(outputDir + "/" + files[0].Name())
			assert.Equal(
				t,
				string(expectedContent),
				string(gotContent),
				"Wrong content in the output file",
			)
		})
	}
}

func TestFile_IsBulk(t *testing.T) {
	e := fileexporter.Exporter{}
	assert.True(t, e.IsBulk(), "DeprecatedExporterV1 is a bulk exporter")
}

func TestExportWithoutOutputDir(t *testing.T) {
	featureEvents := []exporter.ExportableEvent{
		exporter.FeatureEvent{
			Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
			Variation: "Default", Value: "YO", Default: false, Source: "SERVER",
		}}

	filePrefix := "test-flag-variation-EXAMPLE-"
	e := fileexporter.Exporter{
		Format:   "json",
		Filename: filePrefix + "{{ .Timestamp}}.{{ .Format}}",
	}
	err := e.Export(context.Background(), nil, featureEvents)
	require.NoError(t, err)

	// check that a file exist
	files, err := os.ReadDir("./")
	require.NoError(t, err)

	countFileWithPrefix := 0
	for _, file := range files {
		if strings.HasPrefix(file.Name(), filePrefix) {
			countFileWithPrefix++
			err := os.Remove(path.Clean(file.Name()))
			require.NoError(t, err)
		}
	}
	assert.True(t, countFileWithPrefix > 0, "At least one file should have been created")
}
