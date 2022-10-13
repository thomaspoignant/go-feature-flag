package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

func TestExporterConf_IsValid(t *testing.T) {
	type fields struct {
		Kind        string
		OutputDir   string
		Format      string
		Filename    string
		CsvTemplate string
		Bucket      string
		Path        string
		EndpointURL string
		Secret      string
		Meta        map[string]string
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config.ExporterConf{
				Kind:        config.ExporterKind(tt.fields.Kind),
				OutputDir:   tt.fields.OutputDir,
				Format:      tt.fields.Format,
				Filename:    tt.fields.Filename,
				CsvTemplate: tt.fields.CsvTemplate,
				Bucket:      tt.fields.Bucket,
				Path:        tt.fields.Path,
				EndpointURL: tt.fields.EndpointURL,
				Secret:      tt.fields.Secret,
				Meta:        tt.fields.Meta,
			}
			err := c.IsValid()
			assert.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				assert.Equal(t, tt.errValue, err.Error())
			}
		})
	}
}
