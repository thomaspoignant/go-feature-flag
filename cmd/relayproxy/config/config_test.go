package config_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/exporter/kafkaexporter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestParseConfig_fileFromPflag(t *testing.T) {
	tests := []struct {
		name         string
		want         *config.Config
		fileLocation string
		wantErr      assert.ErrorAssertionFunc
	}{
		{
			name:         "Valid yaml file",
			fileLocation: "../testdata/config/valid-file.yaml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					FileFormat:      "yaml",
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "http",
						URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.goff.yaml",
					},
					Exporter: &config.ExporterConf{
						Kind: "log",
					},
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:          "localhost",
				Version:       "1.X.X",
				EnableSwagger: true,
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"apikey3",
					},
					Evaluation: []string{
						"apikey1",
						"apikey2",
					},
				},
				LogLevel: "info",
			},
			wantErr: assert.NoError,
		},
		{
			name:         "Valid yaml file with legacy key notifier (without s)",
			fileLocation: "../testdata/config/valid-file-notifier.yaml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					FileFormat:      "yaml",
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "http",
						URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.goff.yaml",
					},
					Exporter: &config.ExporterConf{
						Kind: "log",
					},
					StartWithRetrieverError: false,
					FixNotifiers: []config.NotifierConf{
						config.NotifierConf{Kind: config.DiscordNotifier, WebhookURL: "https://discord.com/api/webhooks/yyyy/xxxxxxx"},
					},
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:          "localhost",
				Version:       "1.X.X",
				EnableSwagger: true,
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"apikey3",
					},
					Evaluation: []string{
						"apikey1",
						"apikey2",
					},
				},

				LogLevel: "info",
			},
			wantErr: assert.NoError,
		},
		{
			name:         "Valid yaml file with legacy key notifier (without s)",
			fileLocation: "../testdata/config/valid-file-notifiers.yaml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					FileFormat:      "yaml",
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "http",
						URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.goff.yaml",
					},
					Exporter: &config.ExporterConf{
						Kind: "log",
					},
					StartWithRetrieverError: false,
					Notifiers: []config.NotifierConf{
						config.NotifierConf{Kind: config.DiscordNotifier, WebhookURL: "https://discord.com/api/webhooks/yyyy/xxxxxxx"},
					},
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:          "localhost",
				Version:       "1.X.X",
				EnableSwagger: true,
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"apikey3",
					},
					Evaluation: []string{
						"apikey1",
						"apikey2",
					},
				},

				LogLevel: "info",
			},
			wantErr: assert.NoError,
		},
		{
			name:         "Valid yaml file with multiple exporters",
			fileLocation: "../testdata/config/valid-yaml-multiple-exporters.yaml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					FileFormat:      "yaml",
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "http",
						URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.goff.yaml",
					},
					Exporters: &[]config.ExporterConf{
						{
							Kind: "log",
						},
						{
							Kind:      "file",
							OutputDir: "./",
						},
					},
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:          "localhost",
				Version:       "1.X.X",
				EnableSwagger: true,
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"apikey3",
					},
					Evaluation: []string{
						"apikey1",
						"apikey2",
					},
				},
				LogLevel: "info",
			},
			wantErr: assert.NoError,
		},
		{
			name:         "Valid yaml file with both exporter and exporters",
			fileLocation: "../testdata/config/valid-yaml-exporter-and-exporters.yaml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					FileFormat:      "yaml",
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "http",
						URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.goff.yaml",
					},
					Exporter: &config.ExporterConf{
						Kind: "log",
					},
					Exporters: &[]config.ExporterConf{
						{
							Kind:        "webhook",
							EndpointURL: "https://example.com/webhook",
						},
						{
							Kind:      "file",
							OutputDir: "./",
						},
					},
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},

				Host: "localhost",

				Version:       "1.X.X",
				EnableSwagger: true,
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"apikey3",
					},
					Evaluation: []string{
						"apikey1",
						"apikey2",
					},
				},
				LogLevel: "info",
			},
			wantErr: assert.NoError,
		},
		{
			name:         "Valid json file",
			fileLocation: "../testdata/config/valid-file.json",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					FileFormat:      "yaml",
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "http",
						URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.goff.yaml",
					},
					Exporter: &config.ExporterConf{
						Kind: "log",
					},
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host: "localhost",

				Version:       "1.X.X",
				EnableSwagger: true,
				APIKeys: []string{
					"apikey1",
					"apikey2",
				},
				LogLevel: "",
			},
			wantErr: assert.NoError,
		},
		{
			name:         "Valid toml file",
			fileLocation: "../testdata/config/valid-file.toml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					FileFormat:      "yaml",
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "http",
						URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.goff.yaml",
					},
					Exporter: &config.ExporterConf{
						Kind: "log",
					},
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},

				Host: "localhost",

				Version:       "1.X.X",
				EnableSwagger: true,
				APIKeys: []string{
					"apikey1",
					"apikey2",
				},
				LogLevel: config.DefaultLogLevel,
			},
			wantErr: assert.NoError,
		},
		{
			name:         "All default",
			fileLocation: "../testdata/config/all-default.yaml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval:         60000,
					FileFormat:              "yaml",
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:     "localhost",
				Version:  "1.X.X",
				LogLevel: config.DefaultLogLevel,
			},
			wantErr: assert.NoError,
		},
		{
			name:         "Invalid yaml",
			fileLocation: "../testdata/config/invalid-yaml.yaml",
			wantErr:      assert.Error,
		},
		{
			name:         "Valid YAML with OTel config",
			fileLocation: "../testdata/config/valid-otel.yaml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000,
					FileFormat:      "yaml",
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: "examples/retriever_file/flags.goff.yaml",
						},
					},
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:     "localhost",
				LogLevel: config.DefaultLogLevel,
				Version:  "1.X.X",
				OtelConfig: config.OpenTelemetryConfiguration{
					Exporter: config.OtelExporter{
						Otlp: config.OtelExporterOtlp{
							Endpoint: "http://example.com:4317",
						},
					},
					Resource: config.OtelResource{
						Attributes: map[string]string{
							"foo.bar": "baz",
							"foo.baz": "bar",
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("version", "1.X.X")
			f := pflag.NewFlagSet("config", pflag.ContinueOnError)
			f.String("config", "", "Location of your config file")
			_ = f.Parse([]string{fmt.Sprintf("--config=%s", tt.fileLocation)})

			got, err := config.New(f, zap.L(), "1.X.X")
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got, "Config not matching")
		})
	}
}

func TestParseConfig_fileFromFolder(t *testing.T) {
	tests := []struct {
		name                       string
		want                       *config.Config
		fileLocation               string
		wantErr                    assert.ErrorAssertionFunc
		disableDefaultFileCreation bool
	}{
		{
			name:         "Valid file",
			fileLocation: "../testdata/config/valid-file.yaml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					FileFormat:      "yaml",
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "http",
						URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.goff.yaml",
					},
					Exporter: &config.ExporterConf{
						Kind: "log",
					},
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:          "localhost",
				Version:       "1.X.X",
				EnableSwagger: true,
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"apikey3",
					},
					Evaluation: []string{
						"apikey1",
						"apikey2",
					},
				},
				LogLevel: "info",
			},
			wantErr: assert.NoError,
		},
		{
			name:         "All default",
			fileLocation: "../testdata/config/all-default.yaml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval:         60000,
					FileFormat:              "yaml",
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:     "localhost",
				Version:  "1.X.X",
				LogLevel: config.DefaultLogLevel,
			},
			wantErr: assert.NoError,
		},
		{
			name:         "Invalid yaml",
			fileLocation: "../testdata/config/invalid-yaml.yaml",
			wantErr:      assert.Error,
		},
		{
			name:         "Should return all default if file does not exist",
			fileLocation: "../testdata/config/file-not-exist.yaml",
			wantErr:      assert.NoError,
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval:         60000,
					FileFormat:              "yaml",
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:     "localhost",
				Version:  "1.X.X",
				LogLevel: config.DefaultLogLevel,
			},
		},
		{
			name:         "Should return all default if no file in the command line",
			fileLocation: "",
			wantErr:      assert.NoError,
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval:         60000,
					FileFormat:              "yaml",
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:     "localhost",
				Version:  "1.X.X",
				LogLevel: config.DefaultLogLevel,
			},
		},
		{
			name:         "Should return all default if no file and no default",
			fileLocation: "",
			wantErr:      assert.NoError,
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval:         60000,
					FileFormat:              "yaml",
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:     "localhost",
				Version:  "1.X.X",
				LogLevel: config.DefaultLogLevel,
			},
			disableDefaultFileCreation: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Remove("./goff-proxy.yaml")
			if !tt.disableDefaultFileCreation {
				source, _ := os.Open(tt.fileLocation)
				destination, _ := os.Create("./goff-proxy.yaml")
				defer destination.Close()
				defer source.Close()
				defer os.Remove("./goff-proxy.yaml")
				_, _ = io.Copy(destination, source)
			}
			f := pflag.NewFlagSet("config", pflag.ContinueOnError)
			f.String("config", "", "Location of your config file")
			_ = f.Parse([]string{fmt.Sprintf("--config=%s", tt.fileLocation)})
			got, err := config.New(f, zap.L(), "1.X.X")
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
		PollingInterval         int
		FileFormat              string
		StartWithRetrieverError bool
		Retriever               *retrieverconf.RetrieverConf
		Retrievers              *[]retrieverconf.RetrieverConf
		Exporter                *config.ExporterConf
		Notifiers               []config.NotifierConf
		LogLevel                string
		Debug                   bool
		LogFormat               string
		FlagSets                []config.FlagSet
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "empty config",
			fields:  fields{},
			wantErr: assert.Error,
		},
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
						Kind:       "slack",
						WebhookURL: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "valid configuration",
			fields: fields{
				ListenPort: 8080,
				Retriever: &retrieverconf.RetrieverConf{
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
						Kind:       "slack",
						WebhookURL: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
					},
				},
				LogLevel: "info",
			},
			wantErr: assert.NoError,
		},
		{
			name: "valid configuration with notifier included",
			fields: fields{
				ListenPort: 8080,
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: "../testdata/config/valid-file-notifier.yaml",
				},
				Exporter: &config.ExporterConf{
					Kind:        "webhook",
					EndpointURL: "http://testingwebhook.com/test/",
					Secret:      "secret-for-signing",
					Meta: map[string]string{
						"extraInfo": "info",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "invalid retriever",
			fields: fields{
				ListenPort: 8080,
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "1 invalid retriever in the list of retrievers",
			fields: fields{
				ListenPort: 8080,
				Retrievers: &[]retrieverconf.RetrieverConf{
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
				Retriever: &retrieverconf.RetrieverConf{
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
				Retriever: &retrieverconf.RetrieverConf{
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
		{
			name: "invalid log level",
			fields: fields{
				ListenPort: 8080,
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: "../testdata/config/valid-file.yaml",
				},
				LogLevel: "invalid",
			},
			wantErr: assert.Error,
		},
		{
			name: "log level is not set but debug is set",
			fields: fields{
				ListenPort: 8080,
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: "../testdata/config/valid-file.yaml",
				},
				LogLevel: "",
				Debug:    true,
			},
			wantErr: assert.NoError,
		},
		{
			name: "invalid logFormat",
			fields: fields{
				LogFormat:  "unknown",
				ListenPort: 8080,
				Retriever: &retrieverconf.RetrieverConf{
					Kind: "file",
					Path: "../testdata/config/valid-file.yaml",
				},
				LogLevel: "info",
			},
			wantErr: assert.Error,
		},
		{
			name: "valid flagset configuration",
			fields: fields{
				ListenPort: 8080,
				FlagSets: []config.FlagSet{
					{
						Name:    "test-flagset",
						APIKeys: []string{"test-api-key"},
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: "file",
								Path: "../testdata/config/valid-file.yaml",
							},
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "flagset without name",
			fields: fields{
				ListenPort: 8080,
				FlagSets: []config.FlagSet{
					{
						APIKeys: []string{"test-api-key"},
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: "file",
								Path: "../testdata/config/valid-file.yaml",
							},
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "flagset without API keys",
			fields: fields{
				ListenPort: 8080,
				FlagSets: []config.FlagSet{
					{
						Name: "test-flagset",
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: "file",
								Path: "../testdata/config/valid-file.yaml",
							},
						},
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "duplicate API keys across flagsets",
			fields: fields{
				ListenPort: 8080,
				FlagSets: []config.FlagSet{
					{
						Name:    "flagset-1",
						APIKeys: []string{"shared-api-key"},
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: "file",
								Path: "../testdata/config/valid-file.yaml",
							},
						},
					},
					{
						Name:    "flagset-2",
						APIKeys: []string{"shared-api-key"},
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: "file",
								Path: "../testdata/config/valid-file.yaml",
							},
						},
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "flagset with invalid retriever",
			fields: fields{
				ListenPort: 8080,
				FlagSets: []config.FlagSet{
					{
						Name:    "test-flagset",
						APIKeys: []string{"test-api-key"},
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: "file",
								// Missing Path
							},
						},
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "flagset with invalid exporter",
			fields: fields{
				ListenPort: 8080,
				FlagSets: []config.FlagSet{
					{
						Name:    "test-flagset",
						APIKeys: []string{"test-api-key"},
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: "file",
								Path: "../testdata/config/valid-file.yaml",
							},
							Exporter: &config.ExporterConf{
								Kind: "webhook",
								// Missing EndpointURL
							},
						},
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "flagset with invalid notifier",
			fields: fields{
				ListenPort: 8080,
				FlagSets: []config.FlagSet{
					{
						Name:    "test-flagset",
						APIKeys: []string{"test-api-key"},
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: "file",
								Path: "../testdata/config/valid-file.yaml",
							},
							Notifiers: []config.NotifierConf{
								{
									Kind: "webhook",
									// Missing EndpointURL
								},
							},
						},
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "multiple valid flagsets",
			fields: fields{
				ListenPort: 8080,
				FlagSets: []config.FlagSet{
					{
						Name:    "flagset-1",
						APIKeys: []string{"api-key-1"},
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: "file",
								Path: "../testdata/config/valid-file.yaml",
							},
						},
					},
					{
						Name:    "flagset-2",
						APIKeys: []string{"api-key-2"},
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &retrieverconf.RetrieverConf{
								Kind: "file",
								Path: "../testdata/config/valid-file.yaml",
							},
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval:         tt.fields.PollingInterval,
					FileFormat:              tt.fields.FileFormat,
					StartWithRetrieverError: tt.fields.StartWithRetrieverError,
					Retriever:               tt.fields.Retriever,
					Exporter:                tt.fields.Exporter,
					Notifiers:               tt.fields.Notifiers,
					Retrievers:              tt.fields.Retrievers,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: tt.fields.ListenPort,
				},
				HideBanner:    tt.fields.HideBanner,
				EnableSwagger: tt.fields.EnableSwagger,
				Host:          tt.fields.Host,
				LogLevel:      tt.fields.LogLevel,
				LogFormat:     tt.fields.LogFormat,
				FlagSets:      tt.fields.FlagSets,
			}
			if tt.name == "empty config" {
				c = nil
			}
			tt.wantErr(t, c.IsValid(), "invalid configuration")
		})
	}
}

func TestConfig_APIKeyExists(t *testing.T) {
	tests := []struct {
		name   string
		config config.Config
		apiKey string
		want   bool
	}{
		{
			name: "no key in the config",
			config: config.Config{
				AuthorizedKeys: config.APIKeys{
					Evaluation: []string{},
				},
			},
			apiKey: "49b67ab9-20fc-42ac-ac53-b36e29834c7",
			want:   false,
		},
		{
			name:   "key exists in a list of keys (legacy)",
			apiKey: "49b67ab9-20fc-42ac-ac53-b36e29834c7",
			config: config.Config{
				AuthorizedKeys: config.APIKeys{
					Evaluation: []string{
						"0359cdb3-5fb5-4d65-b25f-b8909ec3c44",
						"fb124cf9-e058-4f34-8385-ad225ff85a3",
						"d05087dd-efff-4144-b9a6-89476a14695",
						"5082a8df-cc67-48b4-aca4-26ce1425645",
						"04d9f1b7-f50c-4407-83bb-e9c4ddc5d45",
						"62507779-bd2d-4170-b715-8d93ee7110f",
						"e0dcb798-4f97-4646-a1a9-57a6c69c235",
						"6bfd6b61-f8a9-45b3-9ca8-37125438be4",
						"aecd6aea-1350-46af-a7b9-231e9a609fd",
						"49b67ab9-20fc-42ac-ac53-b36e29834c7",
					},
				},
			},
			want: true,
		},
		{
			name:   "key exists in a list of keys",
			apiKey: "49b67ab9-20fc-42ac-ac53-b36e29834c7",
			config: config.Config{
				AuthorizedKeys: config.APIKeys{
					Evaluation: []string{
						"0359cdb3-5fb5-4d65-b25f-b8909ec3c44",
						"fb124cf9-e058-4f34-8385-ad225ff85a3",
						"d05087dd-efff-4144-b9a6-89476a14695",
						"5082a8df-cc67-48b4-aca4-26ce1425645",
						"04d9f1b7-f50c-4407-83bb-e9c4ddc5d45",
						"62507779-bd2d-4170-b715-8d93ee7110f",
						"e0dcb798-4f97-4646-a1a9-57a6c69c235",
						"6bfd6b61-f8a9-45b3-9ca8-37125438be4",
						"aecd6aea-1350-46af-a7b9-231e9a609fd",
						"49b67ab9-20fc-42ac-ac53-b36e29834c7",
					},
				},
			},
			want: true,
		},
		{
			name:   "admin key works for evaluation",
			apiKey: "49b67ab9-20fc-42ac-ac53-b36e29834c7",
			config: config.Config{
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"49b67ab9-20fc-42ac-ac53-b36e29834c7",
					},
					Evaluation: []string{
						"xxx",
					},
				},
			},
			want: true,
		},
		{
			name: "no api key passed in the function",
			config: config.Config{
				AuthorizedKeys: config.APIKeys{
					Evaluation: []string{
						"0359cdb3-5fb5-4d65-b25f-b8909ec3c44",
						"fb124cf9-e058-4f34-8385-ad225ff85a3",
						"d05087dd-efff-4144-b9a6-89476a14695",
						"5082a8df-cc67-48b4-aca4-26ce1425645",
						"04d9f1b7-f50c-4407-83bb-e9c4ddc5d45",
						"62507779-bd2d-4170-b715-8d93ee7110f",
						"e0dcb798-4f97-4646-a1a9-57a6c69c235",
						"6bfd6b61-f8a9-45b3-9ca8-37125438be4",
						"aecd6aea-1350-46af-a7b9-231e9a609fd",
						"49b67ab9-20fc-42ac-ac53-b36e29834c7",
					},
				},
			},
			want: false,
		},
		{
			name:   "empty key passed in the function",
			apiKey: "",
			config: config.Config{
				AuthorizedKeys: config.APIKeys{
					Evaluation: []string{
						"0359cdb3-5fb5-4d65-b25f-b8909ec3c44",
						"fb124cf9-e058-4f34-8385-ad225ff85a3",
						"d05087dd-efff-4144-b9a6-89476a14695",
						"5082a8df-cc67-48b4-aca4-26ce1425645",
						"04d9f1b7-f50c-4407-83bb-e9c4ddc5d45",
						"62507779-bd2d-4170-b715-8d93ee7110f",
						"e0dcb798-4f97-4646-a1a9-57a6c69c235",
						"6bfd6b61-f8a9-45b3-9ca8-37125438be4",
						"aecd6aea-1350-46af-a7b9-231e9a609fd",
						"49b67ab9-20fc-42ac-ac53-b36e29834c7",
					},
				},
			},
			want: false,
		},
		{
			name:   "key exists in flag set",
			apiKey: "flag-set-api-key",
			config: config.Config{
				FlagSets: []config.FlagSet{
					{
						Name:    "test-flag-set",
						APIKeys: []string{"flag-set-api-key"},
					},
				},
			},
			want: true,
		},
		{
			name:   "key exists in multiple flag sets",
			apiKey: "flag-set-api-key-2",
			config: config.Config{
				FlagSets: []config.FlagSet{
					{
						Name:    "test-flag-set-1",
						APIKeys: []string{"flag-set-api-key-1"},
					},
					{
						Name:    "test-flag-set-2",
						APIKeys: []string{"flag-set-api-key-2"},
					},
				},
			},
			want: true,
		},
		{
			name:   "key does not exist in flag sets",
			apiKey: "non-existent-key",
			config: config.Config{
				FlagSets: []config.FlagSet{
					{
						Name:    "test-flag-set-1",
						APIKeys: []string{"flag-set-api-key-1"},
					},
					{
						Name:    "test-flag-set-2",
						APIKeys: []string{"flag-set-api-key-2"},
					},
				},
			},
			want: false,
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(
				t,
				tt.want,
				tt.config.APIKeyExists(tt.apiKey),
				"APIKeyExists(%v)",
				tt.apiKey,
			)
		})
	}
}

func TestConfig_APIAdminKeyExists(t *testing.T) {
	tests := []struct {
		name   string
		config config.Config
		apiKey string
		want   bool
	}{
		{
			name: "no key in the config",
			config: config.Config{
				AuthorizedKeys: config.APIKeys{
					Admin:      []string{},
					Evaluation: []string{},
				},
			},
			apiKey: "49b67ab9-20fc-42ac-ac53-b36e29834c7",
			want:   false,
		},
		{
			name:   "key exists in a list of keys",
			apiKey: "49b67ab9-20fc-42ac-ac53-b36e29834c7",
			config: config.Config{
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"aecd6aea-1350-46af-a7b9-231e9a609fd",
						"49b67ab9-20fc-42ac-ac53-b36e29834c7",
					},
				},
			},
			want: true,
		},
		{
			name:   "admin key works for evaluation",
			apiKey: "49b67ab9-20fc-42ac-ac53-b36e29834c7",
			config: config.Config{
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"49b67ab9-20fc-42ac-ac53-b36e29834c7",
					},
					Evaluation: []string{
						"xxx",
					},
				},
			},
			want: true,
		},
		{
			name: "no api key passed in the function",
			config: config.Config{
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"49b67ab9-20fc-42ac-ac53-b36e29834c7",
					},
					Evaluation: []string{
						"xxx",
					},
				},
			},
			want: false,
		},
		{
			name:   "empty key passed in the function",
			apiKey: "",
			config: config.Config{
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"49b67ab9-20fc-42ac-ac53-b36e29834c7",
					},
					Evaluation: []string{
						"xxx",
					},
				},
			},
			want: false,
		},
		{
			name:   "evaluation key does not work for admin",
			apiKey: "xxx",
			config: config.Config{
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"49b67ab9-20fc-42ac-ac53-b36e29834c7",
					},
					Evaluation: []string{
						"xxx",
					},
				},
			},
			want: false,
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(
				t,
				tt.want,
				tt.config.APIKeysAdminExists(tt.apiKey),
				"APIKeyExists(%v)",
				tt.apiKey,
			)
		})
	}
}

func TestMergeConfig_FromOSEnv(t *testing.T) {
	tests := []struct {
		name                       string
		want                       *config.Config
		fileLocation               string
		wantErr                    assert.ErrorAssertionFunc
		disableDefaultFileCreation bool
		envVars                    map[string]string
	}{
		{
			name:         "Valid file",
			fileLocation: "../testdata/config/validate-array-env-file.yaml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					FileFormat:      "yaml",
					Notifiers: []config.NotifierConf{
						{
							Kind:        config.WebhookNotifier,
							EndpointURL: "https://xxx.com/notifier",
						},
					},
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "http",
							URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.goff.yaml",
							HTTPHeaders: map[string][]string{
								"authorization": {
									"test",
								},
								"token": {"token"},
							},
						},
						{
							Kind: "file",
							Path: "examples/retriever_file/flags.goff.yaml",
							HTTPHeaders: map[string][]string{
								"token": {
									"11213123",
								},
								"authorization": {
									"test1",
								},
							},
						},
						{
							HTTPHeaders: map[string][]string{
								"authorization": {
									"test1",
								},
								"x-goff-custom": {
									"custom",
								},
							},
						},
					},
					Exporter: &config.ExporterConf{
						Kind: "log",
					},
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:          "localhost",
				Version:       "1.X.X",
				EnableSwagger: true,
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"apikey3",
					},
					Evaluation: []string{
						"apikey1",
						"apikey2",
					},
				},
				LogLevel: "info",
			},
			wantErr: assert.NoError,
			envVars: map[string]string{
				"RETRIEVERS_0_HEADERS_AUTHORIZATION": "test",
				"RETRIEVERS_X_HEADERS_AUTHORIZATION": "test",
				"RETRIEVERS_1_HEADERS_AUTHORIZATION": "test1",
				"RETRIEVERS_0_HEADERS_TOKEN":         "token",
				"RETRIEVERS_2_HEADERS_AUTHORIZATION": "test1",
				"RETRIEVERS_2_HEADERS_X-GOFF-CUSTOM": "custom",
				"NOTIFIERS_0_ENDPOINTURL":            "https://xxx.com/notifier",
			},
		},
		{
			name:         "Valid file (with legacy notifier without s)",
			fileLocation: "../testdata/config/validate-array-env-file-legacy.yaml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					FileFormat:      "yaml",
					FixNotifiers: []config.NotifierConf{
						{
							Kind:        config.WebhookNotifier,
							EndpointURL: "https://xxx.com/notifier",
						},
					},
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "http",
							URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.goff.yaml",
							HTTPHeaders: map[string][]string{
								"authorization": {
									"test",
								},
								"token": {"token"},
							},
						},
						{
							Kind: "file",
							Path: "examples/retriever_file/flags.goff.yaml",
							HTTPHeaders: map[string][]string{
								"token": {
									"11213123",
								},
								"authorization": {
									"test1",
								},
							},
						},
						{
							HTTPHeaders: map[string][]string{
								"authorization": {
									"test1",
								},
								"x-goff-custom": {
									"custom",
								},
							},
						},
					},
					Exporter: &config.ExporterConf{
						Kind: "log",
					},
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:          "localhost",
				Version:       "1.X.X",
				EnableSwagger: true,
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"apikey3",
					},
					Evaluation: []string{
						"apikey1",
						"apikey2",
					},
				},
				LogLevel: "info",
			},
			wantErr: assert.NoError,
			envVars: map[string]string{
				"RETRIEVERS_0_HEADERS_AUTHORIZATION": "test",
				"RETRIEVERS_X_HEADERS_AUTHORIZATION": "test",
				"RETRIEVERS_1_HEADERS_AUTHORIZATION": "test1",
				"RETRIEVERS_0_HEADERS_TOKEN":         "token",
				"RETRIEVERS_2_HEADERS_AUTHORIZATION": "test1",
				"RETRIEVERS_2_HEADERS_X-GOFF-CUSTOM": "custom",
				"NOTIFIER_0_ENDPOINTURL":             "https://xxx.com/notifier",
			},
		},
		{
			name:         "Valid file with prefix",
			fileLocation: "../testdata/config/validate-array-env-file-envprefix.yaml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					FileFormat:      "yaml",
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "http",
							URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.goff.yaml",
							HTTPHeaders: map[string][]string{
								"authorization": {
									"test",
								},
								"token": {"token"},
							},
						},
						{
							Kind: "file",
							Path: "examples/retriever_file/flags.goff.yaml",
							HTTPHeaders: map[string][]string{
								"token": {
									"11213123",
								},
								"authorization": {
									"test1",
								},
							},
						},
						{
							HTTPHeaders: map[string][]string{
								"authorization": {
									"test1",
								},
								"x-goff-custom": {
									"custom",
								},
							},
						},
					},
					Exporter: &config.ExporterConf{
						Kind: "log",
					},
					StartWithRetrieverError: false,
				},
				EnvVariablePrefix: "GOFF_",
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:          "localhost",
				Version:       "1.X.X",
				EnableSwagger: true,
				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"apikey3",
					},
					Evaluation: []string{
						"apikey1",
						"apikey2",
					},
				},
				LogLevel: "info",
			},
			wantErr: assert.NoError,
			envVars: map[string]string{
				"GOFF_RETRIEVERS_0_HEADERS_AUTHORIZATION": "test",
				"GOFF_RETRIEVERS_X_HEADERS_AUTHORIZATION": "test",
				"GOFF_RETRIEVERS_1_HEADERS_AUTHORIZATION": "test1",
				"GOFF_RETRIEVERS_0_HEADERS_TOKEN":         "token",
				"GOFF_RETRIEVERS_2_HEADERS_AUTHORIZATION": "test1",
				"GOFF_RETRIEVERS_2_HEADERS_X-GOFF-CUSTOM": "custom",
			},
		},
		{
			name:         "Change kafka exporter",
			fileLocation: "../testdata/config/validate-array-env-file.yaml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					FileFormat:      "yaml",
					Notifiers: []config.NotifierConf{
						{
							Kind:        config.WebhookNotifier,
							EndpointURL: "http://localhost:8080/webhook",
						},
					},
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "http",
							URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.goff.yaml",
						},
						{
							Kind: "file",
							Path: "examples/retriever_file/flags.goff.yaml",
							HTTPHeaders: map[string][]string{
								"token": {
									"11213123",
								},
							},
						},
					},
					Exporter: &config.ExporterConf{
						Kind: "kafka",
						Kafka: kafkaexporter.Settings{
							Addresses: []string{"localhost:19092", "localhost:19093"},
						},
					},
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host: "localhost",

				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"apikey3",
					},
					Evaluation: []string{
						"apikey1",
						"apikey2",
					},
				},
				Version:       "1.X.X",
				EnableSwagger: true,
				LogLevel:      "info",
			},
			wantErr: assert.NoError,
			envVars: map[string]string{
				"EXPORTER_KAFKA_ADDRESSES": "localhost:19092,localhost:19093",
				"EXPORTER_KIND":            "kafka",
			},
		},
		{
			name:         "Change kafka exporters",
			fileLocation: "../testdata/config/valid-env-exporters-kafka.yaml",
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 1000,
					FileFormat:      "yaml",
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "http",
							URL:  "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.goff.yaml",
						},
						{
							Kind: "file",
							Path: "examples/retriever_file/flags.goff.yaml",
							HTTPHeaders: map[string][]string{
								"token": {
									"11213123",
								},
							},
						},
					},
					Exporters: &[]config.ExporterConf{
						{
							Kind: "kafka",
							Kafka: kafkaexporter.Settings{
								Addresses: []string{"localhost:19092", "localhost:19093"},
								Topic:     "svc-goff.evaluation",
							},
						},
					},
					StartWithRetrieverError: false,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host: "localhost",

				AuthorizedKeys: config.APIKeys{
					Admin: []string{
						"apikey3",
					},
					Evaluation: []string{
						"apikey1",
						"apikey2",
					},
				},
				Version:       "1.X.X",
				EnableSwagger: true,
				LogLevel:      "info",
			},
			wantErr: assert.NoError,
			envVars: map[string]string{
				"EXPORTERS_0_KAFKA_ADDRESSES": "localhost:19092,localhost:19093",
			},
		},
		{
			name:                       "Valid YAML with OTel config",
			fileLocation:               "../testdata/config/valid-otel.yaml",
			disableDefaultFileCreation: true,
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					PollingInterval: 60000,
					FileFormat:      "yaml",
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: "examples/retriever_file/flags.goff.yaml",
						},
					},
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:     "localhost",
				LogLevel: config.DefaultLogLevel,
				Version:  "1.X.X",
				OtelConfig: config.OpenTelemetryConfiguration{
					Exporter: config.OtelExporter{
						Otlp: config.OtelExporterOtlp{
							Endpoint: "http://localhost:4317",
						},
					},
					Resource: config.OtelResource{
						Attributes: map[string]string{
							"foo.bar": "baz",
							"foo.baz": "qux",
							"foo.qux": "quux",
						},
					},
				},
			},
			wantErr: assert.NoError,
			envVars: map[string]string{
				"OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4317",
				"OTEL_RESOURCE_ATTRIBUTES":    "foo.baz=qux,foo.qux=quux,ignored.key",
			},
		},
		{
			name:                       "Configure flagset with basic environment variables",
			fileLocation:               "../testdata/config/validate-array-env-file-flagset.yaml",
			disableDefaultFileCreation: true,
			want: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					FileFormat:      "yaml",
					PollingInterval: 60000,
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: 1031,
				},
				Host:     "localhost",
				LogLevel: config.DefaultLogLevel,
				Version:  "1.X.X",
				FlagSets: []config.FlagSet{
					{
						Name:    "xxx",
						APIKeys: []string{"other-api-key", "default-api-key"},
						CommonFlagSet: config.CommonFlagSet{
							PollingInterval:         50000,
							FileFormat:              "yaml",
							StartWithRetrieverError: true,
							Notifiers: []config.NotifierConf{
								{
									Kind:        config.WebhookNotifier,
									EndpointURL: "http://localhost:8080/webhook",
								},
							},
							Exporters: &[]config.ExporterConf{
								{
									Kind: "kafka",
									Kafka: kafkaexporter.Settings{
										Addresses: []string{"localhost:19092", "localhost:19093"},
									},
								},
							},
							Retrievers: &[]retrieverconf.RetrieverConf{
								{
									Kind: "file",
									Path: "examples/retriever_file/flags.goff.yaml",
								},
							},
						},
					},
				},
			},
			wantErr: assert.NoError,
			envVars: map[string]string{
				"FLAGSETS_0_NAME":                        "xxx",
				"FLAGSETS_0_APIKEYS":                     "other-api-key,default-api-key",
				"FLAGSETS_0_FILEFORMAT":                  "yaml",
				"FLAGSETS_0_POLLINGINTERVAL":             "50000",
				"FLAGSETS_0_STARTWITHRETRIEVERERROR":     "true",
				"FLAGSETS_0_NOTIFIERS_0_KIND":            "webhook",
				"FLAGSETS_0_NOTIFIERS_0_ENDPOINTURL":     "http://localhost:8080/webhook",
				"FLAGSETS_0_EXPORTERS_0_KIND":            "kafka",
				"FLAGSETS_0_EXPORTERS_0_KAFKA_ADDRESSES": "localhost:19092,localhost:19093",
				"FLAGSETS_0_RETRIEVERS_0_KIND":           "file",
				"FLAGSETS_0_RETRIEVERS_0_PATH":           "examples/retriever_file/flags.goff.yaml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			_ = os.Remove("./goff-proxy.yaml")
			if !tt.disableDefaultFileCreation {
				source, _ := os.Open(tt.fileLocation)
				destination, _ := os.Create("./goff-proxy.yaml")
				defer destination.Close()
				defer source.Close()
				defer os.Remove("./goff-proxy.yaml")
				_, _ = io.Copy(destination, source)
			}

			f := pflag.NewFlagSet("config", pflag.ContinueOnError)
			f.String("config", "", "Location of your config file")
			_ = f.Parse([]string{fmt.Sprintf("--config=%s", tt.fileLocation)})
			got, err := config.New(f, zap.L(), "1.X.X")
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got, "Config not matching expected %v, got %v", tt.want, got)
		})
	}
}

func TestSetAPIKeysFromEnv(t *testing.T) {
	os.Setenv("AUTHORIZEDKEYS_EVALUATION", "key1,key2,key 3")
	os.Setenv("AUTHORIZEDKEYS_ADMIN", "key4,key5")

	fileLocation := "../testdata/config/valid-file.yaml"
	f := pflag.NewFlagSet("config", pflag.ContinueOnError)
	f.String("config", "", "Location of your config file")
	_ = f.Parse([]string{fmt.Sprintf("--config=%s", fileLocation)})

	got, err := config.New(f, zap.L(), "1.X.X")
	require.NoError(t, err)
	assert.Equal(t, []string{"key1", "key2", "key 3"}, got.AuthorizedKeys.Evaluation)
	assert.Equal(t, []string{"key4", "key5"}, got.AuthorizedKeys.Admin)
}

func TestConfig_LogLevel(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.Config
		wantDebug    bool
		wantLogLevel zapcore.Level
	}{
		{
			name:         "no config",
			wantDebug:    false,
			wantLogLevel: zapcore.InvalidLevel,
		},
		{
			name: "invalid log level",
			config: &config.Config{
				LogLevel: "invalid",
			},
			wantDebug:    false,
			wantLogLevel: zapcore.InvalidLevel,
		},
		{
			name: "debug level",
			config: &config.Config{
				LogLevel: "debug",
			},
			wantDebug:    true,
			wantLogLevel: zapcore.DebugLevel,
		},
		{
			name: "info level",
			config: &config.Config{
				LogLevel: "info",
			},
			wantDebug:    false,
			wantLogLevel: zapcore.InfoLevel,
		},
		{
			name: "error level",
			config: &config.Config{
				LogLevel: "error",
			},
			wantDebug:    false,
			wantLogLevel: zapcore.ErrorLevel,
		},
		{
			name: "panic level",
			config: &config.Config{
				LogLevel: "panic",
			},
			wantDebug:    false,
			wantLogLevel: zapcore.PanicLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantDebug, tt.config.IsDebugEnabled(), "IsDebugEnabled()")
			assert.Equalf(t, tt.wantLogLevel, tt.config.ZapLogLevel(), "ZapLogLevel()")
		})
	}
}

func TestConfig_IsDebugEnabled(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
		want bool
	}{
		{
			name: "Uppercase",
			cfg: config.Config{
				LogLevel: "DEBUG",
			},
			want: true,
		},
		{
			name: "Lowercase",
			cfg: config.Config{
				LogLevel: "debug",
			},
			want: true,
		},
		{
			name: "Random Case",
			cfg: config.Config{
				LogLevel: "DeBuG",
			},
			want: true,
		},
		{
			name: "Not debug",
			cfg: config.Config{
				LogLevel: "DeBu",
			},
			want: false,
		},
		{
			name: "Empty",
			cfg: config.Config{
				LogLevel: "",
			},
			want: false,
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.cfg.IsDebugEnabled(), "IsDebugEnabled()")
		})
	}
}
