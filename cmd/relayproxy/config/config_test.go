package config_test

import (
	"io"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"go.uber.org/zap"
)

func TestParseConfig_fileFromPflag(t *testing.T) {
	tests := []struct {
		name         string
		want         *config.Config
		fileLocation string
		wantErr      assert.ErrorAssertionFunc
	}{
		{
			name:         "Valid file",
			fileLocation: "../testdata/config/valid-file.yaml",
			want: &config.Config{
				ListenPort:      1031,
				PollingInterval: 1000,
				FileFormat:      "yaml",
				Host:            "localhost",
				Retriever: &config.RetrieverConf{
					Kind: "http",
					URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.yaml",
				},
				Exporter: &config.ExporterConf{
					Kind: "log",
				},
				StartWithRetrieverError: false,
				RestAPITimeout:          5000,
				Version:                 "1.X.X",
				EnableSwagger:           true,
				APIKeys: map[string]bool{
					"apikey1": true,
					"apikey2": true,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:         "All default",
			fileLocation: "../testdata/config/all-default.yaml",
			want: &config.Config{
				ListenPort:              1031,
				PollingInterval:         60000,
				FileFormat:              "yaml",
				Host:                    "localhost",
				StartWithRetrieverError: false,
				RestAPITimeout:          5000,
				Version:                 "1.X.X",
			},
			wantErr: assert.NoError,
		},
		{
			name:         "Invalid yaml",
			fileLocation: "../testdata/config/invalid-yaml.yaml",
			wantErr:      assert.Error,
		},
		{
			name:         "File does not exists",
			fileLocation: "../testdata/config/invalid-filename.yaml",
			wantErr:      assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("config", tt.fileLocation)
			got, err := config.ParseConfig(zap.L(), "1.X.X")
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got, "Config not matching")
			viper.Reset()
		})
	}
}

func TestParseConfig_fileFromFolder(t *testing.T) {
	tests := []struct {
		name         string
		want         *config.Config
		fileLocation string
		wantErr      assert.ErrorAssertionFunc
	}{
		{
			name:         "Valid file",
			fileLocation: "../testdata/config/valid-file.yaml",
			want: &config.Config{
				ListenPort:      1031,
				PollingInterval: 1000,
				FileFormat:      "yaml",
				Host:            "localhost",
				Retriever: &config.RetrieverConf{
					Kind: "http",
					URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.yaml",
				},
				Exporter: &config.ExporterConf{
					Kind: "log",
				},
				StartWithRetrieverError: false,
				RestAPITimeout:          5000,
				Version:                 "1.X.X",
				EnableSwagger:           true,
				APIKeys: map[string]bool{
					"apikey1": true,
					"apikey2": true,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:         "All default",
			fileLocation: "../testdata/config/all-default.yaml",
			want: &config.Config{
				ListenPort:              1031,
				PollingInterval:         60000,
				FileFormat:              "yaml",
				Host:                    "localhost",
				StartWithRetrieverError: false,
				RestAPITimeout:          5000,
				Version:                 "1.X.X",
			},
			wantErr: assert.NoError,
		},
		{
			name:         "Invalid yaml",
			fileLocation: "../testdata/config/invalid-yaml.yaml",
			wantErr:      assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Remove("./goff-proxy.yaml")
			source, _ := os.Open(tt.fileLocation)
			destination, _ := os.Create("./goff-proxy.yaml")
			defer destination.Close()
			defer source.Close()
			defer os.Remove("./goff-proxy.yaml")
			_, _ = io.Copy(destination, source)

			got, err := config.ParseConfig(zap.L(), "1.X.X")
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got, "Config not matching")
		})
	}
}

func TestConfig_IsValid(t *testing.T) {
	type fields struct {
		ListenPort              int
		HideBanner              bool
		EnableSwagger           bool
		Host                    string
		Debug                   bool
		PollingInterval         int
		FileFormat              string
		StartWithRetrieverError bool
		Retriever               *config.RetrieverConf
		Retrievers              *[]config.RetrieverConf
		Exporter                *config.ExporterConf
		Notifiers               []config.NotifierConf
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "invalid port",
			fields:  fields{ListenPort: 0},
			wantErr: assert.Error,
		},
		{
			name: "no retriever",
			fields: fields{
				ListenPort: 8080,
				Notifiers: []config.NotifierConf{
					{
						Kind:        "webhook",
						EndpointURL: "https://hooktest.com/",
						Secret:      "xxxx",
					},
					{
						Kind:            "slack",
						SlackWebhookURL: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "valid configuration",
			fields: fields{
				ListenPort: 8080,
				Retriever: &config.RetrieverConf{
					Kind: "file",
					Path: "../testdata/config/valid-file.yaml",
				},
				Exporter: &config.ExporterConf{
					Kind:        "webhook",
					EndpointURL: "http://testingwebhook.com/test/",
					Secret:      "secret-for-signing",
					Meta: map[string]string{
						"extraInfo": "info",
					},
				},
				Notifiers: []config.NotifierConf{
					{
						Kind:        "webhook",
						EndpointURL: "https://hooktest.com/",
						Secret:      "xxxx",
					},
					{
						Kind:            "slack",
						SlackWebhookURL: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "invalid retriever",
			fields: fields{
				ListenPort: 8080,
				Retriever: &config.RetrieverConf{
					Kind: "file",
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "1 invalid retriever in the list of retrievers",
			fields: fields{
				ListenPort: 8080,
				Retrievers: &[]config.RetrieverConf{
					{
						Kind: "file",
						Path: "../testdata/config/valid-file.yaml",
					},
					{
						Kind: "file",
					},
					{
						Kind: "file",
						Path: "../testdata/config/valid-file.yaml",
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid exporter",
			fields: fields{
				ListenPort: 8080,
				Retriever: &config.RetrieverConf{
					Kind: "file",
					Path: "../testdata/config/valid-file.yaml",
				},
				Exporter: &config.ExporterConf{
					Kind: "webhook",
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid notifier",
			fields: fields{
				ListenPort: 8080,
				Retriever: &config.RetrieverConf{
					Kind: "file",
					Path: "../testdata/config/valid-file.yaml",
				},
				Notifiers: []config.NotifierConf{
					{
						Kind: "webhook",
					},
				},
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config.Config{
				ListenPort:              tt.fields.ListenPort,
				HideBanner:              tt.fields.HideBanner,
				EnableSwagger:           tt.fields.EnableSwagger,
				Host:                    tt.fields.Host,
				Debug:                   tt.fields.Debug,
				PollingInterval:         tt.fields.PollingInterval,
				FileFormat:              tt.fields.FileFormat,
				StartWithRetrieverError: tt.fields.StartWithRetrieverError,
				Retriever:               tt.fields.Retriever,
				Exporter:                tt.fields.Exporter,
				Notifiers:               tt.fields.Notifiers,
				Retrievers:              tt.fields.Retrievers,
			}
			tt.wantErr(t, c.IsValid(), "invalid configuration")
		})
	}
}
