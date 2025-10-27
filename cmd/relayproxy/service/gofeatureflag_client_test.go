package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/azureexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/gcstorageexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/kafkaexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/kinesisexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/logsexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/pubsubexporterv2"
	"github.com/thomaspoignant/go-feature-flag/exporter/s3exporterv2"
	"github.com/thomaspoignant/go-feature-flag/exporter/sqsexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/webhookexporter"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/notifier/discordnotifier"
	"github.com/thomaspoignant/go-feature-flag/notifier/microsoftteamsnotifier"
	"github.com/thomaspoignant/go-feature-flag/notifier/slacknotifier"
	"github.com/thomaspoignant/go-feature-flag/notifier/webhooknotifier"
	"github.com/thomaspoignant/go-feature-flag/utils"
	"github.com/xitongsys/parquet-go/parquet"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/proxy"
)

func Test_initRetrievers(t *testing.T) {
	tests := []struct {
		name       string
		retrievers *[]retrieverconf.RetrieverConf
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name:    "both retriever and retrievers",
			wantErr: assert.NoError,
			retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind:           "bitbucket",
					Branch:         "develop",
					RepositorySlug: "gofeatureflag/config-repo",
					Path:           "flags/config.goff.yaml",
					AuthToken:      "XXX_BITBUCKET_TOKEN",
					BaseURL:        "https://api.bitbucket.goff.org",
				},
				{
					Kind:           "bitbucket",
					Branch:         "main",
					RepositorySlug: "gofeatureflag/config-repo",
					Path:           "flags/config.goff.yaml",
					AuthToken:      "XXX_BITBUCKET_TOKEN",
					BaseURL:        "https://api.bitbucket.goff.org",
				},
			},
		},
		{
			name:    "should error with invalid retriever",
			wantErr: assert.Error,
			retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind:           "bitbucket",
					Branch:         "develop",
					RepositorySlug: "gofeatureflag/config-repo",
					Path:           "flags/config.goff.yaml",
					AuthToken:      "XXX_BITBUCKET_TOKEN",
					BaseURL:        "https://api.bitbucket.goff.org",
				},
				{
					Kind: "unknown",
				},
			},
		},
		{
			name:    "should error with invalid retriever",
			wantErr: assert.Error,
			retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "unknown",
				},
			},
		},
		{
			name:    "only retrievers",
			wantErr: assert.NoError,
			retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind:           "bitbucket",
					Branch:         "develop",
					RepositorySlug: "gofeatureflag/config-repo",
					Path:           "flags/config.goff.yaml",
					AuthToken:      "XXX_BITBUCKET_TOKEN",
					BaseURL:        "https://api.bitbucket.goff.org",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyConf := config.FlagSet{
				CommonFlagSet: config.CommonFlagSet{
					Retrievers: tt.retrievers,
				},
			}
			r, err := initRetrievers(&proxyConf)
			tt.wantErr(t, err)
			if r != nil {
				nbRetriever := 0
				if tt.retrievers != nil {
					nbRetriever += len(*tt.retrievers)
				}
				assert.Len(t, r, nbRetriever)
			}
		})
	}
}

func Test_initExporters(t *testing.T) {
	tests := []struct {
		name      string
		exporters *[]config.ExporterConf
		exporter  *config.ExporterConf
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			name:    "both exporter and exporters",
			wantErr: assert.NoError,
			exporters: &[]config.ExporterConf{
				{
					Kind:        "webhook",
					EndpointURL: "https://gofeatureflag.org/webhook-example",
					Secret:      "1234",
				},
				{
					Kind:        "webhook",
					EndpointURL: "https://gofeatureflag.org/webhook-example",
					Secret:      "1234",
				},
			},
		},
		{
			name:    "exporters only",
			wantErr: assert.NoError,
			exporters: &[]config.ExporterConf{
				{
					Kind:        "webhook",
					EndpointURL: "https://gofeatureflag.org/webhook-example",
					Secret:      "1234",
				},
			},
		},
		{
			name:    "invalid exporter",
			wantErr: assert.Error,
			exporters: &[]config.ExporterConf{
				{
					Kind: "invalid",
				},
				{
					Kind:        "webhook",
					EndpointURL: "https://gofeatureflag.org/webhook-example",
					Secret:      "1234",
				},
			},
		},
		{
			name:    "invalid exporters",
			wantErr: assert.Error,
			exporters: &[]config.ExporterConf{
				{

					Kind:        "webhook",
					EndpointURL: "https://gofeatureflag.org/webhook-example",
					Secret:      "1234",
				},
				{
					Kind: "invalid",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyConf := config.FlagSet{
				CommonFlagSet: config.CommonFlagSet{
					Exporters: tt.exporters,
				},
			}
			r, err := initDataExporters(&proxyConf)
			tt.wantErr(t, err)
			if r != nil {
				nbExp := 0
				if tt.exporters != nil {
					nbExp += len(*tt.exporters)
				}
				if tt.exporter != nil {
					nbExp++
				}
				assert.Len(t, r, nbExp)
			}
		})
	}
}

func Test_initExporter(t *testing.T) {
	tests := []struct {
		name                   string
		conf                   *config.ExporterConf
		want                   ffclient.DataExporter
		wantErr                assert.ErrorAssertionFunc
		wantType               exporter.CommonExporter
		skipCompleteValidation bool
	}{
		{
			name:    "Convert unknown Exporter",
			wantErr: assert.Error,
			conf: &config.ExporterConf{
				Kind: "unknown",
			},
		},
		{
			name:    "Convert WebhookExporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind:        "webhook",
				EndpointURL: "https://gofeatureflag.org/webhook-example",
				Secret:      "1234",
			},
			want: ffclient.DataExporter{
				FlushInterval:    config.DefaultExporter.FlushInterval,
				MaxEventInMemory: config.DefaultExporter.MaxEventInMemory,
				Exporter: &webhookexporter.Exporter{
					EndpointURL: "https://gofeatureflag.org/webhook-example",
					Secret:      "1234",
					Meta:        nil,
				},
				ExporterEventType: ffclient.FeatureEventExporter,
			},
			wantType: &webhookexporter.Exporter{},
		},
		{
			name:    "Convert FileExporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind:                    "file",
				OutputDir:               "/outputfolder/",
				ParquetCompressionCodec: parquet.CompressionCodec_UNCOMPRESSED.String(),
			},
			want: ffclient.DataExporter{
				FlushInterval:    config.DefaultExporter.FlushInterval,
				MaxEventInMemory: config.DefaultExporter.MaxEventInMemory,
				Exporter: &fileexporter.Exporter{
					Format:                  config.DefaultExporter.Format,
					OutputDir:               "/outputfolder/",
					Filename:                config.DefaultExporter.FileName,
					CsvTemplate:             config.DefaultExporter.CsvFormat,
					ParquetCompressionCodec: parquet.CompressionCodec_UNCOMPRESSED.String(),
				},
				ExporterEventType: ffclient.FeatureEventExporter,
			},
			wantType: &fileexporter.Exporter{},
		},
		{
			name:    "Convert LogExporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind: "log",
			},
			want: ffclient.DataExporter{
				FlushInterval:    config.DefaultExporter.FlushInterval,
				MaxEventInMemory: config.DefaultExporter.MaxEventInMemory,
				Exporter: &logsexporter.Exporter{
					LogFormat: config.DefaultExporter.LogFormat,
				},
				ExporterEventType: ffclient.FeatureEventExporter,
			},
			wantType: &logsexporter.Exporter{},
		},
		{
			name:    "Convert S3Exporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind:          "s3",
				Bucket:        "my-bucket",
				Path:          "/my-path/",
				FlushInterval: 10,
			},
			want: ffclient.DataExporter{
				FlushInterval:    10 * time.Millisecond,
				MaxEventInMemory: config.DefaultExporter.MaxEventInMemory,
				Exporter: &s3exporterv2.Exporter{
					Bucket:                  "my-bucket",
					Format:                  config.DefaultExporter.Format,
					S3Path:                  "/my-path/",
					Filename:                config.DefaultExporter.FileName,
					CsvTemplate:             config.DefaultExporter.CsvFormat,
					ParquetCompressionCodec: config.DefaultExporter.ParquetCompressionCodec,
				},
			},
			wantType:               &s3exporterv2.Exporter{},
			skipCompleteValidation: true,
		},
		{
			name:    "Convert SQSExporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind:          "sqs",
				QueueURL:      "https://sqs.eu-west-1.amazonaws.com/XXX/test-queue",
				FlushInterval: 10,
			},
			want: ffclient.DataExporter{
				FlushInterval:    10 * time.Millisecond,
				MaxEventInMemory: config.DefaultExporter.MaxEventInMemory,
				Exporter: &sqsexporter.Exporter{
					QueueURL: "https://sqs.eu-west-1.amazonaws.com/XXX/test-queue",
				},
			},
			wantType:               &sqsexporter.Exporter{},
			skipCompleteValidation: true,
		},
		{
			name:    "Convert PubSubExporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind:      "pubsub",
				ProjectID: "fake-project-id",
				Topic:     "fake-topic",
			},
			want: ffclient.DataExporter{
				Exporter: &pubsubexporterv2.Exporter{
					ProjectID: "fake-project-id",
					Topic:     "fake-topic",
				},
			},
			wantType:               &pubsubexporterv2.Exporter{},
			skipCompleteValidation: true,
		},
		{
			name:    "Convert GoogleStorageExporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind:             "googleStorage",
				Bucket:           "my-bucket",
				Path:             "/my-path/",
				MaxEventInMemory: 1990,
			},
			want: ffclient.DataExporter{
				FlushInterval:    config.DefaultExporter.FlushInterval,
				MaxEventInMemory: 1990,
				Exporter: &gcstorageexporter.Exporter{
					Bucket:                  "my-bucket",
					Format:                  config.DefaultExporter.Format,
					Path:                    "/my-path/",
					Filename:                config.DefaultExporter.FileName,
					CsvTemplate:             config.DefaultExporter.CsvFormat,
					ParquetCompressionCodec: config.DefaultExporter.ParquetCompressionCodec,
				},
				ExporterEventType: ffclient.FeatureEventExporter,
			},
			wantType: &gcstorageexporter.Exporter{},
		},
		{
			name:    "Convert KafkaExporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind:             "kafka",
				MaxEventInMemory: 1990,
				Kafka: kafkaexporter.Settings{
					Topic:     "example-topic",
					Addresses: []string{"addr1", "addr2"},
				},
				ExporterEventType: ffclient.FeatureEventExporter,
			},
			want: ffclient.DataExporter{
				FlushInterval:    config.DefaultExporter.FlushInterval,
				MaxEventInMemory: 1990,
				Exporter: &kafkaexporter.Exporter{
					Format: config.DefaultExporter.Format,
					Settings: kafkaexporter.Settings{
						Topic:     "example-topic",
						Addresses: []string{"addr1", "addr2"},
					},
				},
				ExporterEventType: ffclient.FeatureEventExporter,
			},
			wantType: &kafkaexporter.Exporter{},
		},
		{
			name:    "AWS Kinesis Exporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind:       "kinesis",
				StreamName: "my-stream",
			},
			want: ffclient.DataExporter{
				FlushInterval:    10 * time.Millisecond,
				MaxEventInMemory: config.DefaultExporter.MaxEventInMemory,
				Exporter: &kinesisexporter.Exporter{
					Format: config.DefaultExporter.Format,
					Settings: kinesisexporter.NewSettings(
						kinesisexporter.WithStreamArn("my-stream"),
					),
				},
			},
			wantType:               &kinesisexporter.Exporter{},
			skipCompleteValidation: true,
		},
		{
			name:    "Azure Blob Storage Exporter",
			wantErr: assert.NoError,
			conf: &config.ExporterConf{
				Kind:             "azureBlobStorage",
				Container:        "my-container",
				Path:             "/my-path/",
				MaxEventInMemory: 1990,
			},
			want: ffclient.DataExporter{
				FlushInterval:    config.DefaultExporter.FlushInterval,
				MaxEventInMemory: 1990,
				Exporter: &azureexporter.Exporter{
					Container:               "my-container",
					Format:                  config.DefaultExporter.Format,
					Path:                    "/my-path/",
					Filename:                config.DefaultExporter.FileName,
					CsvTemplate:             config.DefaultExporter.CsvFormat,
					ParquetCompressionCodec: config.DefaultExporter.ParquetCompressionCodec,
				},
				ExporterEventType: ffclient.FeatureEventExporter,
			},
			wantType: &azureexporter.Exporter{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := initDataExporter(tt.conf)
			tt.wantErr(t, err)
			assert.IsType(t, tt.wantType, got.Exporter)
			if err == nil && !tt.skipCompleteValidation {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_initNotifier(t *testing.T) {
	type args struct {
		c []config.NotifierConf
	}
	tests := []struct {
		name    string
		args    args
		want    []notifier.Notifier
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid",
			args: args{
				c: []config.NotifierConf{
					{
						Kind:       config.SlackNotifier,
						WebhookURL: "http:xxxx.xxx",
					},
					{
						Kind:        config.WebhookNotifier,
						EndpointURL: "http:yyyy.yyy",
					},
					{
						Kind:       config.MicrosoftTeamsNotifier,
						WebhookURL: "http:zzzz.zzz",
					},
					{
						Kind:       config.DiscordNotifier,
						WebhookURL: "http:aaaa.aaa",
					},
				},
			},
			want: []notifier.Notifier{
				&slacknotifier.Notifier{SlackWebhookURL: "http:xxxx.xxx"},
				&webhooknotifier.Notifier{EndpointURL: "http:yyyy.yyy"},
				&microsoftteamsnotifier.Notifier{MicrosoftTeamsWebhookURL: "http:zzzz.zzz"},
				&discordnotifier.Notifier{DiscordWebhookURL: "http:aaaa.aaa"},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := initNotifier(tt.args.c)
			if !tt.wantErr(t, err, fmt.Sprintf("initNotifier(%v)", tt.args.c)) {
				return
			}
			assert.Equalf(t, tt.want, got, "initNotifier(%v)", tt.args.c)
		})
	}
}

func TestNewGoFeatureFlagClient_ProxyConfNil(t *testing.T) {
	// Create a logger for testing
	logger := zap.NewNop()

	// Call NewGoFeatureFlagClient with nil proxyConf
	goff, err := NewGoFeatureFlagClient(nil, logger, nil)

	// Assert that the function returns nil and an error
	assert.Nil(t, goff, "Expected GoFeatureFlag client to be nil when proxyConf is nil")
	assert.EqualError(
		t,
		err,
		"proxy config is empty",
		"Expected error message to indicate empty proxy config",
	)
}

func TestSetKafkaConfig(t *testing.T) {
	t.Run(
		"should have a SCRAMClientGeneratorFunc if SCRAM is enabled and use SCRAM-SHA-512 mechanism",
		func(t *testing.T) {
			settings := kafkaexporter.Settings{
				Topic:     "my-kafka-topic",
				Addresses: []string{"addr1", "addr2"},
				Config: &sarama.Config{
					Net: struct {
						MaxOpenRequests                  int
						DialTimeout                      time.Duration
						ReadTimeout                      time.Duration
						WriteTimeout                     time.Duration
						ResolveCanonicalBootstrapServers bool
						TLS                              struct {
							Enable bool
							Config *tls.Config
						}
						SASL struct {
							Enable                   bool
							Mechanism                sarama.SASLMechanism
							Version                  int16
							Handshake                bool
							AuthIdentity             string
							User                     string
							Password                 string
							SCRAMAuthzID             string
							SCRAMClientGeneratorFunc func() sarama.SCRAMClient
							TokenProvider            sarama.AccessTokenProvider
							GSSAPI                   sarama.GSSAPIConfig
						}
						KeepAlive time.Duration
						LocalAddr net.Addr
						Proxy     struct {
							Enable bool
							Dialer proxy.Dialer
						}
					}{SASL: struct {
						Enable                   bool
						Mechanism                sarama.SASLMechanism
						Version                  int16
						Handshake                bool
						AuthIdentity             string
						User                     string
						Password                 string
						SCRAMAuthzID             string
						SCRAMClientGeneratorFunc func() sarama.SCRAMClient
						TokenProvider            sarama.AccessTokenProvider
						GSSAPI                   sarama.GSSAPIConfig
					}{Enable: true, Mechanism: "SCRAM-SHA-512", User: "TODO", Password: "TODO"}},
				},
			}
			kafkaConfig, err := setKafkaConfig(settings)
			assert.NoError(t, err)

			assert.NotNil(t, kafkaConfig.Net.SASL.SCRAMClientGeneratorFunc)
		},
	)

	t.Run(
		"should have a SCRAMClientGeneratorFunc if SCRAM is enabled and use SCRAM-SHA-256 mechanism",
		func(t *testing.T) {
			settings := kafkaexporter.Settings{
				Topic:     "my-kafka-topic",
				Addresses: []string{"addr1", "addr2"},
				Config: &sarama.Config{
					Net: struct {
						MaxOpenRequests                  int
						DialTimeout                      time.Duration
						ReadTimeout                      time.Duration
						WriteTimeout                     time.Duration
						ResolveCanonicalBootstrapServers bool
						TLS                              struct {
							Enable bool
							Config *tls.Config
						}
						SASL struct {
							Enable                   bool
							Mechanism                sarama.SASLMechanism
							Version                  int16
							Handshake                bool
							AuthIdentity             string
							User                     string
							Password                 string
							SCRAMAuthzID             string
							SCRAMClientGeneratorFunc func() sarama.SCRAMClient
							TokenProvider            sarama.AccessTokenProvider
							GSSAPI                   sarama.GSSAPIConfig
						}
						KeepAlive time.Duration
						LocalAddr net.Addr
						Proxy     struct {
							Enable bool
							Dialer proxy.Dialer
						}
					}{SASL: struct {
						Enable                   bool
						Mechanism                sarama.SASLMechanism
						Version                  int16
						Handshake                bool
						AuthIdentity             string
						User                     string
						Password                 string
						SCRAMAuthzID             string
						SCRAMClientGeneratorFunc func() sarama.SCRAMClient
						TokenProvider            sarama.AccessTokenProvider
						GSSAPI                   sarama.GSSAPIConfig
					}{Enable: true, Mechanism: "SCRAM-SHA-256", User: "TODO", Password: "TODO"}},
				},
			}
			kafkaConfig, err := setKafkaConfig(settings)
			assert.NoError(t, err)

			assert.NotNil(t, kafkaConfig.Net.SASL.SCRAMClientGeneratorFunc)
		},
	)

	t.Run(
		"should not have a SCRAMClientGeneratorFunc if SCRAM is enabled and use an unknown mechanism",
		func(t *testing.T) {
			settings := kafkaexporter.Settings{
				Topic:     "my-kafka-topic",
				Addresses: []string{"addr1", "addr2"},
				Config: &sarama.Config{
					Net: struct {
						MaxOpenRequests                  int
						DialTimeout                      time.Duration
						ReadTimeout                      time.Duration
						WriteTimeout                     time.Duration
						ResolveCanonicalBootstrapServers bool
						TLS                              struct {
							Enable bool
							Config *tls.Config
						}
						SASL struct {
							Enable                   bool
							Mechanism                sarama.SASLMechanism
							Version                  int16
							Handshake                bool
							AuthIdentity             string
							User                     string
							Password                 string
							SCRAMAuthzID             string
							SCRAMClientGeneratorFunc func() sarama.SCRAMClient
							TokenProvider            sarama.AccessTokenProvider
							GSSAPI                   sarama.GSSAPIConfig
						}
						KeepAlive time.Duration
						LocalAddr net.Addr
						Proxy     struct {
							Enable bool
							Dialer proxy.Dialer
						}
					}{SASL: struct {
						Enable                   bool
						Mechanism                sarama.SASLMechanism
						Version                  int16
						Handshake                bool
						AuthIdentity             string
						User                     string
						Password                 string
						SCRAMAuthzID             string
						SCRAMClientGeneratorFunc func() sarama.SCRAMClient
						TokenProvider            sarama.AccessTokenProvider
						GSSAPI                   sarama.GSSAPIConfig
					}{Enable: true, Mechanism: "UNKNONW-MECHANISM", User: "TODO", Password: "TODO"}},
				},
			}
			kafkaConfig, err := setKafkaConfig(settings)
			assert.NoError(t, err)
			assert.Nil(t, kafkaConfig.Net.SASL.SCRAMClientGeneratorFunc)
		},
	)

	t.Run("should return a valid sarama.Config if settings are nil", func(t *testing.T) {
		settings := kafkaexporter.Settings{
			Topic:     "my-kafka-topic",
			Addresses: []string{"addr1", "addr2"},
			Config: &sarama.Config{
				Version: sarama.V2_1_0_0,
			},
		}
		// We expect am error because the settings are nil
		assert.Error(t, settings.Config.Validate())

		kafkaConfig, err := setKafkaConfig(settings)
		assert.NoError(t, err)

		// after calling setKafkaConfig, the settings should be valid because they are merged with the default config.
		assert.NoError(t, kafkaConfig.Config.Validate())
	})

	t.Run("should return a valid sarama.Config with specific settings set and value should be"+
		" merged with default kafka config", func(t *testing.T) {
		settings := kafkaexporter.Settings{
			Topic:     "my-kafka-topic",
			Addresses: []string{"addr1", "addr2"},
			Config: &sarama.Config{
				Net: struct {
					MaxOpenRequests                  int
					DialTimeout                      time.Duration
					ReadTimeout                      time.Duration
					WriteTimeout                     time.Duration
					ResolveCanonicalBootstrapServers bool
					TLS                              struct {
						Enable bool
						Config *tls.Config
					}
					SASL struct {
						Enable                   bool
						Mechanism                sarama.SASLMechanism
						Version                  int16
						Handshake                bool
						AuthIdentity             string
						User                     string
						Password                 string
						SCRAMAuthzID             string
						SCRAMClientGeneratorFunc func() sarama.SCRAMClient
						TokenProvider            sarama.AccessTokenProvider
						GSSAPI                   sarama.GSSAPIConfig
					}
					KeepAlive time.Duration
					LocalAddr net.Addr
					Proxy     struct {
						Enable bool
						Dialer proxy.Dialer
					}
				}{SASL: struct {
					Enable                   bool
					Mechanism                sarama.SASLMechanism
					Version                  int16
					Handshake                bool
					AuthIdentity             string
					User                     string
					Password                 string
					SCRAMAuthzID             string
					SCRAMClientGeneratorFunc func() sarama.SCRAMClient
					TokenProvider            sarama.AccessTokenProvider
					GSSAPI                   sarama.GSSAPIConfig
				}{Enable: true, Mechanism: "SCRAM-SHA-512", User: "TODO", Password: "TODO"}},
			},
		}
		// We expect am error because the settings are nil
		assert.Error(t, settings.Config.Validate())

		kafkaConfig, err := setKafkaConfig(settings)
		assert.NoError(t, err)

		// after calling setKafkaConfig, the settings should be valid because they are merged with the default config.
		assert.NoError(t, kafkaConfig.Config.Validate())
		assert.Equal(t, true, settings.Config.Net.SASL.Enable)
		assert.Equal(t, "TODO", settings.Config.Net.SASL.User)
		assert.Equal(t, "TODO", settings.Config.Net.SASL.Password)
	})

	t.Run("should return a nil config", func(t *testing.T) {
		settings := kafkaexporter.Settings{
			Topic:     "my-kafka-topic",
			Addresses: []string{"addr1", "addr2"},
		}
		kafkaConfig, err := setKafkaConfig(settings)
		assert.NoError(t, err)
		assert.Nil(t, kafkaConfig.Config)
	})
}

func Test_initLeveledLogger_FlagsetAttribute(t *testing.T) {
	t.Run("verify flagset attribute logic", func(t *testing.T) {
		// Test the logic directly by checking the conditions
		tests := []struct {
			name          string
			flagsetName   string
			shouldAddAttr bool
		}{
			{
				name:          "default flagset name should not add attribute",
				flagsetName:   utils.DefaultFlagSetName,
				shouldAddAttr: false,
			},
			{
				name:          "empty flagset name should not add attribute",
				flagsetName:   "",
				shouldAddAttr: false,
			},
			{
				name:          "custom flagset name should add attribute",
				flagsetName:   "my-custom-flagset",
				shouldAddAttr: true,
			},
			{
				name:          "another custom flagset name should add attribute",
				flagsetName:   "production-flags",
				shouldAddAttr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				flagset := &config.FlagSet{
					Name: tt.flagsetName,
				}
				// Create a buffer to capture log output
				var logBuffer bytes.Buffer

				// Create a zap logger that writes to our buffer
				encoderCfg := zap.NewProductionEncoderConfig()
				encoderCfg.TimeKey = ""
				encoderCfg.LevelKey = ""
				encoderCfg.CallerKey = ""
				core := zapcore.NewCore(
					zapcore.NewJSONEncoder(encoderCfg),
					zapcore.AddSync(&logBuffer),
					zapcore.InfoLevel,
				)
				zapLogger := zap.New(core)

				slogLogger := initLeveledLogger(flagset, zapLogger)
				assert.NotNil(t, slogLogger, "initLeveledLogger should return a non-nil logger")

				slogLogger.Info("test message")
				_ = zapLogger.Sync()
				logOutput := logBuffer.String()

				if tt.shouldAddAttr {
					assert.Contains(t, logOutput, `"flagset"`, "log output should contain flagset attribute")
					assert.Contains(t, logOutput, tt.flagsetName, "log output should contain the flagset name")
				} else {
					assert.NotContains(t, logOutput, `"flagset"`, "log output should not contain flagset attribute for default flagset")
				}
			})
		}
	})
}
