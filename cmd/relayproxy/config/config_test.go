package config_test

import (
	"fmt"
	"github.com/spf13/pflag"
	"io"
	"os"
	"testing"

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
				APIKeys: []string{
					"apikey1",
					"apikey2",
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
			os.Unsetenv("version")
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
				APIKeys: []string{
					"apikey1",
					"apikey2",
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
			_ = os.Unsetenv("version")

			f := pflag.NewFlagSet("config", pflag.ContinueOnError)
			f.String("config", "", "Location of your config file")
			_ = f.Parse([]string{fmt.Sprintf("--config=%s", source.Name())})
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
				APIKeys: []string{},
			},
			apiKey: "49b67ab9-20fc-42ac-ac53-b36e29834c7",
			want:   false,
		},
		{
			name:   "key exists in a list of keys",
			apiKey: "49b67ab9-20fc-42ac-ac53-b36e29834c7",
			config: config.Config{
				APIKeys: []string{
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
			want: true,
		},
		{
			name: "no api key passed in the function",
			config: config.Config{
				APIKeys: []string{
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
			want: false,
		},
		{
			name:   "empty key passed in the function",
			apiKey: "",
			config: config.Config{
				APIKeys: []string{
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
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.config.APIKeyExists(tt.apiKey), "APIKeyExists(%v)", tt.apiKey)
		})
	}
}
