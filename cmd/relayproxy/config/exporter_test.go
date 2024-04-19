package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

func TestExporterConf_IsValid(t *testing.T) {
	type fields struct {
		Kind                    string
		OutputDir               string
		Format                  string
		Filename                string
		CsvTemplate             string
		Bucket                  string
		Path                    string
		EndpointURL             string
		Secret                  string
		Meta                    map[string]string
		ParquetCompressionCodec string
		QueueURL                string
		ProjectID               string
		Topic                   string
	}
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		errValue string
	}{
		{
			name:     "no fields",
			fields:   fields{},
			wantErr:  true,
			errValue: "invalid exporter: kind \"\" is not supported",
		},
		{
			name: "invalid kind",
			fields: fields{
				Kind: "invalid",
			},
			wantErr:  true,
			errValue: "invalid exporter: kind \"invalid\" is not supported",
		},
		{
			name: "kind file without outputDir",
			fields: fields{
				Kind: "file",
			},
			wantErr:  true,
			errValue: "invalid exporter: no \"outputDir\" property found for kind \"file\"",
		},
		{
			name: "kind s3 without bucket",
			fields: fields{
				Kind: "s3",
			},
			wantErr:  true,
			errValue: "invalid exporter: no \"bucket\" property found for kind \"s3\"",
		},
		{
			name: "kind googleStorage without bucket",
			fields: fields{
				Kind: "googleStorage",
			},
			wantErr:  true,
			errValue: "invalid exporter: no \"bucket\" property found for kind \"googleStorage\"",
		},
		{
			name: "kind webhook without bucket",
			fields: fields{
				Kind: "webhook",
			},
			wantErr:  true,
			errValue: "invalid exporter: no \"endpointUrl\" property found for kind \"webhook\"",
		},
		{
			name: "kind webhook valid",
			fields: fields{
				Kind:        "webhook",
				EndpointURL: "http://testingwebhook.com/test/",
				Secret:      "secret-for-signing",
				Meta: map[string]string{
					"extraInfo": "info",
				},
			},
			wantErr: false,
		},
		{
			name: "kind file valid",
			fields: fields{
				Kind:        "file",
				OutputDir:   "/tmp/",
				Format:      "json",
				Filename:    "flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}-custom-file",
				CsvTemplate: "{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}}\n",
			},
			wantErr: false,
		},
		{
			name: "kind log valid",
			fields: fields{
				Kind:   "log",
				Format: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", value=\"{{ .Value}}\"",
			},
			wantErr: false,
		},
		{
			name: "kind s3 valid",
			fields: fields{
				Kind:        "s3",
				Bucket:      "testbucket",
				CsvTemplate: "{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}}\n",
				Filename:    "flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}-custom-file",
				Format:      "json",
				Path:        "/here/",
			},
			wantErr: false,
		},
		{
			name: "kind googleStorage valid",
			fields: fields{
				Kind:        "googleStorage",
				Bucket:      "testbucket",
				CsvTemplate: "{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}}\n",
				Filename:    "flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}-custom-file",
				Format:      "json",
				Path:        "/here/",
			},
			wantErr: false,
		},
		{
			name: "invalid parquetCompressionCodec",
			fields: fields{
				Kind:                    "googleStorage",
				Bucket:                  "testbucket",
				Format:                  "parquet",
				ParquetCompressionCodec: "invalid",
			},
			wantErr:  true,
			errValue: "invalid exporter: \"parquetCompressionCodec\" err: not a valid CompressionCodec string",
		},
		{
			name: "kind SQS valid",
			fields: fields{
				Kind:     "sqs",
				QueueURL: "https://sqs.eu-west-1.amazonaws.com/XXX/test-queue",
			},
			wantErr: false,
		},
		{
			name: "kind SQS with queueURL",
			fields: fields{
				Kind:     "sqs",
				QueueURL: "",
			},
			wantErr:  true,
			errValue: "invalid exporter: no \"queueUrl\" property found for kind \"sqs\"",
		},
		{
			name: "kind PubSub valid",
			fields: fields{
				Kind:      "pubsub",
				ProjectID: "fake-project-id",
				Topic:     "fake-topic",
			},
			wantErr: false,
		},
		{
			name: "kind PubSub without project id",
			fields: fields{
				Kind:  "pubsub",
				Topic: "fake-topic",
			},
			wantErr:  true,
			errValue: "invalid exporter: \"projectID\" and \"topic\" are required for kind \"pubsub\"",
		},
		{
			name: "kind PubSub without topic",
			fields: fields{
				Kind:      "pubsub",
				ProjectID: "fake-project-id",
			},
			wantErr:  true,
			errValue: "invalid exporter: \"projectID\" and \"topic\" are required for kind \"pubsub\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config.ExporterConf{
				Kind:                    config.ExporterKind(tt.fields.Kind),
				OutputDir:               tt.fields.OutputDir,
				Format:                  tt.fields.Format,
				Filename:                tt.fields.Filename,
				CsvTemplate:             tt.fields.CsvTemplate,
				Bucket:                  tt.fields.Bucket,
				Path:                    tt.fields.Path,
				EndpointURL:             tt.fields.EndpointURL,
				Secret:                  tt.fields.Secret,
				Meta:                    tt.fields.Meta,
				ParquetCompressionCodec: tt.fields.ParquetCompressionCodec,
				QueueURL:                tt.fields.QueueURL,
				ProjectID:               tt.fields.ProjectID,
				Topic:                   tt.fields.Topic,
			}
			err := c.IsValid()
			assert.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				assert.Equal(t, tt.errValue, err.Error())
			}
		})
	}
}
