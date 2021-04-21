package ffexporter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/exporter"
	"github.com/thomaspoignant/go-feature-flag/testutils"
)

func TestS3_Export(t *testing.T) {
	hostname, _ := os.Hostname()
	type fields struct {
		Bucket      string
		AwsConfig   *aws.Config
		Format      string
		S3Path      string
		Filename    string
		CsvTemplate string
	}

	tests := []struct {
		name         string
		fields       fields
		events       []exporter.FeatureEvent
		wantErr      bool
		expectedFile string
		expectedName string
	}{
		{
			name: "All default test",
			fields: fields{
				Bucket: "test",
			},
			events: []exporter.FeatureEvent{
				{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false},
			},
			expectedFile: "../testdata/ffexporter/s3/all_default.json",
			expectedName: "^/flag-variation-" + hostname + "-[0-9]*\\.json$",
		},
		{
			name: "With S3 Path",
			fields: fields{
				S3Path: "random/path",
				Bucket: "test",
			},
			events: []exporter.FeatureEvent{
				{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false},
			},
			expectedFile: "../testdata/ffexporter/s3/all_default.json",
			expectedName: "^random/path/flag-variation-" + hostname + "-[0-9]*\\.json$",
		},
		{
			name: "All default CSV",
			fields: fields{
				Format: "csv",
				Bucket: "test",
			},
			events: []exporter.FeatureEvent{
				{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false},
			},
			expectedFile: "../testdata/ffexporter/s3/all_default.csv",
			expectedName: "^/flag-variation-" + hostname + "-[0-9]*\\.csv$",
		},
		{
			name: "Custom CSV",
			fields: fields{
				Format:      "csv",
				CsvTemplate: "{{ .Kind}};{{ .ContextKind}}\n",
				Bucket:      "test",
			},
			events: []exporter.FeatureEvent{
				{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false},
			},
			expectedFile: "../testdata/ffexporter/s3/custom_csv_format.csv",
			expectedName: "^/flag-variation-" + hostname + "-[0-9]*\\.csv$",
		},
		{
			name: "Custom FileName",
			fields: fields{
				Format:   "json",
				Filename: "{{ .Format}}-test-{{ .Timestamp}}",
				Bucket:   "test",
			},
			events: []exporter.FeatureEvent{
				{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false},
			},
			expectedFile: "../testdata/ffexporter/s3/all_default.json",
			expectedName: "^/json-test-[0-9]*$",
		},
		{
			name: "Invalid format",
			fields: fields{
				Format: "xxx",
				Bucket: "test",
			},
			events: []exporter.FeatureEvent{
				{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false},
			},
			expectedFile: "../testdata/ffexporter/s3/all_default.json",
			expectedName: "^/flag-variation-" + hostname + "-[0-9]*\\.xxx$",
		},
		{
			name: "Empty Bucket",
			fields: fields{
				Format: "xxx",
			},
			events: []exporter.FeatureEvent{
				{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false},
			},
			wantErr: true,
		},
		{
			name: "Invalid filename template",
			fields: fields{
				Filename: "{{ .InvalidField}}",
				Bucket:   "test",
			},
			events: []exporter.FeatureEvent{
				{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false},
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
				{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s3ManagerMock := testutils.S3ManagerMock{}
			f := &S3{
				Bucket:      tt.fields.Bucket,
				AwsConfig:   tt.fields.AwsConfig,
				Format:      tt.fields.Format,
				S3Path:      tt.fields.S3Path,
				Filename:    tt.fields.Filename,
				CsvTemplate: tt.fields.CsvTemplate,
				s3Uploader:  &s3ManagerMock,
			}
			err := f.Export(log.New(os.Stdout, "", 0), tt.events)
			if tt.wantErr {
				assert.Error(t, err, "Export should error")
				return
			}

			assert.NoError(t, err, "Export should not error")
			assert.Equal(t, 1, len(s3ManagerMock.S3ManagerMockFileSystem), "we should have 1 file in our mock")
			expectedContent, _ := ioutil.ReadFile(tt.expectedFile)
			for k, v := range s3ManagerMock.S3ManagerMockFileSystem {
				assert.Equal(t, string(expectedContent), v, "invalid file content")
				assert.Regexp(t, tt.expectedName, k, "invalid file name")
			}
		})
	}
}

func Test_errSDK(t *testing.T) {
	f := &S3{
		Bucket:    "empty",
		AwsConfig: &aws.Config{},
	}
	err := f.Export(log.New(os.Stdout, "", 0), []exporter.FeatureEvent{})
	assert.Error(t, err, "Empty AWS config should failed")
}

func TestS3_IsBulk(t *testing.T) {
	exporter := S3{}
	assert.True(t, exporter.IsBulk(), "S3 exporter is not a bulk exporter")
}
