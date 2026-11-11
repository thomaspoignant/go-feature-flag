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
		DatasetID               string
		Topic                   string
		StreamName              string
		Container               string
		AccountName             string
		Database                string
		ClusterIdentifier       string
		WorkgroupName           string
		TableName               string
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
			name: "kind kinesis without stream",
			fields: fields{
				Kind: "kinesis",
			},
			wantErr:  true,
			errValue: "invalid exporter: no \"streamArn\" or \"streamName\" property found for kind \"kinesis\"",
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
			name: "kind azureBlobStorage without container",
			fields: fields{
				Kind:        "azureBlobStorage",
				Container:   "",
				AccountName: "MyAccount",
			},
			wantErr:  true,
			errValue: "invalid exporter: no \"container\" property found for kind \"azureBlobStorage\"",
		}, {
			name: "kind azureBlobStorage without accountName",
			fields: fields{
				Kind:        "azureBlobStorage",
				Container:   "MyContainer",
				AccountName: "",
			},
			wantErr:  true,
			errValue: "invalid exporter: no \"accountName\" property found for kind \"azureBlobStorage\"",
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
			name: "kind Kinesis valid",
			fields: fields{
				Kind:       "kinesis",
				StreamName: "test-stream",
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
		{
			name: "kind BigQuery valid",
			fields: fields{
				Kind:      "bigquery",
				ProjectID: "fake-project-id",
				DatasetID: "fake-dataset",
			},
			wantErr: false,
		},
		{
			name: "kind BigQuery without project id",
			fields: fields{
				Kind:      "bigquery",
				DatasetID: "fake-dataset",
			},
			wantErr:  true,
			errValue: "invalid exporter: \"projectID\" and \"datasetID\" are required for kind \"bigquery\"",
		},
		{
			name: "kind BigQuery without dataset id",
			fields: fields{
				Kind:      "bigquery",
				ProjectID: "fake-project-id",
			},
			wantErr:  true,
			errValue: "invalid exporter: \"projectID\" and \"datasetID\" are required for kind \"bigquery\"",
		},
		{
			name: "kind Redshift valid with cluster",
			fields: fields{
				Kind:              "redshift",
				Database:          "analytics",
				ClusterIdentifier: "my-cluster",
			},
			wantErr: false,
		},
		{
			name: "kind Redshift valid with workgroup",
			fields: fields{
				Kind:          "redshift",
				Database:      "analytics",
				WorkgroupName: "my-workgroup",
			},
			wantErr: false,
		},
		{
			name: "kind Redshift valid with schema and table name",
			fields: fields{
				Kind:              "redshift",
				Database:          "analytics",
				ClusterIdentifier: "my-cluster",
				TableName:         "goff.feature_flag_evaluations",
			},
			wantErr: false,
		},
		{
			name: "kind Redshift with invalid table name",
			fields: fields{
				Kind:              "redshift",
				Database:          "analytics",
				ClusterIdentifier: "my-cluster",
				TableName:         "events; DROP TABLE users",
			},
			wantErr:  true,
			errValue: "invalid exporter: \"tableName\" \"events; DROP TABLE users\" is not a valid table name for kind \"redshift\"",
		},
		{
			name: "kind Redshift without database",
			fields: fields{
				Kind:              "redshift",
				ClusterIdentifier: "my-cluster",
			},
			wantErr:  true,
			errValue: "invalid exporter: no \"database\" property found for kind \"redshift\"",
		},
		{
			name: "kind Redshift without cluster or workgroup",
			fields: fields{
				Kind:     "redshift",
				Database: "analytics",
			},
			wantErr:  true,
			errValue: "invalid exporter: exactly one of \"clusterIdentifier\" or \"workgroupName\" is required for kind \"redshift\"",
		},
		{
			name: "kind Redshift with both cluster and workgroup",
			fields: fields{
				Kind:              "redshift",
				Database:          "analytics",
				ClusterIdentifier: "my-cluster",
				WorkgroupName:     "my-workgroup",
			},
			wantErr:  true,
			errValue: "invalid exporter: exactly one of \"clusterIdentifier\" or \"workgroupName\" is required for kind \"redshift\"",
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
				DatasetID:               tt.fields.DatasetID,
				Topic:                   tt.fields.Topic,
				StreamName:              tt.fields.StreamName,
				AccountName:             tt.fields.AccountName,
				Container:               tt.fields.Container,
				Database:                tt.fields.Database,
				ClusterIdentifier:       tt.fields.ClusterIdentifier,
				WorkgroupName:           tt.fields.WorkgroupName,
				TableName:               tt.fields.TableName,
			}
			err := c.IsValid()
			assert.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				assert.Equal(t, tt.errValue, err.Error())
			}
		})
	}
}
