package service

import (
	"fmt"
	"log/slog"
	"time"

	"dario.cat/mergo"
	"github.com/IBM/sarama"
	awsConf "github.com/aws/aws-sdk-go-v2/config"
	slogzap "github.com/samber/slog-zap/v2"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config/kafka"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/azureexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/gcstorageexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/kafkaexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/kinesisexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/logsexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/pubsubexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/s3exporterv2"
	"github.com/thomaspoignant/go-feature-flag/exporter/sqsexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/webhookexporter"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/notifier/discordnotifier"
	"github.com/thomaspoignant/go-feature-flag/notifier/microsoftteamsnotifier"
	"github.com/thomaspoignant/go-feature-flag/notifier/slacknotifier"
	"github.com/thomaspoignant/go-feature-flag/notifier/webhooknotifier"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gcstorageretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/redisretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retrieverv2"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func NewGoFeatureFlagClient(
	proxyConf *config.Config,
	logger *zap.Logger,
	notifiers []notifier.Notifier,
) (*ffclient.GoFeatureFlag, error) {
	var err error
	if proxyConf == nil {
		return nil, fmt.Errorf("proxy config is empty")
	}

	retrievers, err := initRetrievers(proxyConf)
	if err != nil {
		return nil, err
	}

	exporters, err := initDataExporters(proxyConf)
	if err != nil {
		return nil, err
	}

	notif, err := initNotifier(proxyConf.Notifiers)
	if err != nil {
		return nil, err
	}
	notif = append(notif, notifiers...)

	f := ffclient.Config{
		PollingInterval: time.Duration(
			proxyConf.PollingInterval,
		) * time.Millisecond,
		LeveledLogger: slog.New(
			slogzap.Option{Level: slog.LevelDebug, Logger: logger}.NewZapHandler(),
		),
		Context:                         context.Background(),
		Retrievers:                      retrievers,
		Notifiers:                       notif,
		FileFormat:                      proxyConf.FileFormat,
		DataExporters:                   exporters,
		StartWithRetrieverError:         proxyConf.StartWithRetrieverError,
		EnablePollingJitter:             proxyConf.EnablePollingJitter,
		DisableNotifierOnInit:           proxyConf.DisableNotifierOnInit,
		EvaluationContextEnrichment:     proxyConf.EvaluationContextEnrichment,
		PersistentFlagConfigurationFile: proxyConf.PersistentFlagConfigurationFile,
	}

	return ffclient.New(f)
}

// initRetrievers initialize the retrievers based on the configuration
// it handles both the `retriever` and `retrievers` fields
func initRetrievers(proxyConf *config.Config) ([]retriever.Retriever, error) {
	retrievers := make([]retriever.Retriever, 0)
	if proxyConf.Retriever != nil {
		currentRetriever, err := initRetriever(proxyConf.Retriever)
		if err != nil {
			return nil, err
		}
		retrievers = append(retrievers, currentRetriever)
	}
	if proxyConf.Retrievers != nil {
		for _, r := range *proxyConf.Retrievers {
			currentRetriever, err := initRetriever(&r)
			if err != nil {
				return nil, err
			}
			retrievers = append(retrievers, currentRetriever)
		}
	}
	return retrievers, nil
}

// initRetriever initialize the retriever based on the configuration
func initRetriever(c *config.RetrieverConf) (retriever.Retriever, error) {
	retrieverTimeout := config.DefaultRetriever.Timeout
	if c.Timeout != 0 {
		retrieverTimeout = time.Duration(c.Timeout) * time.Millisecond
	}
	switch c.Kind {
	case config.GitHubRetriever:
		return initGithubRetriever(c, retrieverTimeout)
	case config.GitlabRetriever:
		return initGitlabRetriever(c, retrieverTimeout)
	case config.BitbucketRetriever:
		return initBitbucketRetriever(c, retrieverTimeout)
	case config.FileRetriever:
		return &fileretriever.Retriever{Path: c.Path}, nil
	case config.S3Retriever:
		awsConfig, err := awsConf.LoadDefaultConfig(context.Background())
		return &s3retrieverv2.Retriever{Bucket: c.Bucket, Item: c.Item, AwsConfig: &awsConfig}, err
	case config.HTTPRetriever:
		return initHTTPRetriever(c, retrieverTimeout)
	case config.GoogleStorageRetriever:
		return &gcstorageretriever.Retriever{Bucket: c.Bucket, Object: c.Object}, nil
	case config.KubernetesRetriever:
		return initK8sRetriever(c)
	case config.MongoDBRetriever:
		return initMongoRetriever(c)
	case config.RedisRetriever:
		return &redisretriever.Retriever{Options: c.RedisOptions, Prefix: c.RedisPrefix}, nil
	case config.AzBlobStorageRetriever:
		return initAzBlobRetriever(c)
	default:
		return nil, fmt.Errorf("invalid retriever: kind \"%s\" "+
			"is not supported", c.Kind)
	}
}

// initDataExporters initialize the exporters based on the configuration
// it handles both the `exporter` and `exporters` fields.
func initDataExporters(proxyConf *config.Config) ([]ffclient.DataExporter, error) {
	exporters := make([]ffclient.DataExporter, 0)
	if proxyConf.Exporter != nil {
		exp, err := initDataExporter(proxyConf.Exporter)
		if err != nil {
			return nil, err
		}
		exporters = append(exporters, exp)
	}
	if proxyConf.Exporters != nil {
		for _, e := range *proxyConf.Exporters {
			currentExporter, err := initDataExporter(&e)
			if err != nil {
				return nil, err
			}
			exporters = append(exporters, currentExporter)
		}
	}
	return exporters, nil
}

func initDataExporter(c *config.ExporterConf) (ffclient.DataExporter, error) {
	dataExp := ffclient.DataExporter{
		FlushInterval: func() time.Duration {
			if c.FlushInterval != 0 {
				return time.Duration(c.FlushInterval) * time.Millisecond
			}
			return config.DefaultExporter.FlushInterval
		}(),
		MaxEventInMemory: func() int64 {
			if c.MaxEventInMemory != 0 {
				return c.MaxEventInMemory
			}
			return config.DefaultExporter.MaxEventInMemory
		}(),
	}

	var err error
	dataExp.Exporter, err = createExporter(c)
	if err != nil {
		return ffclient.DataExporter{}, err
	}

	return dataExp, nil
}

// nolint: funlen
func createExporter(c *config.ExporterConf) (exporter.CommonExporter, error) {
	format := config.DefaultExporter.Format
	if c.Format != "" {
		format = c.Format
	}

	filename := config.DefaultExporter.FileName
	if c.Filename != "" {
		filename = c.Filename
	}

	csvTemplate := config.DefaultExporter.CsvFormat
	if c.CsvTemplate != "" {
		csvTemplate = c.CsvTemplate
	}

	parquetCompressionCodec := config.DefaultExporter.ParquetCompressionCodec
	if c.ParquetCompressionCodec != "" {
		parquetCompressionCodec = c.ParquetCompressionCodec
	}

	switch c.Kind {
	case config.WebhookExporter:
		return &webhookexporter.Exporter{
			EndpointURL: c.EndpointURL,
			Secret:      c.Secret,
			Meta:        c.Meta,
			Headers:     c.Headers,
		}, nil
	case config.FileExporter:
		return &fileexporter.Exporter{
			Format:                  format,
			OutputDir:               c.OutputDir,
			Filename:                filename,
			CsvTemplate:             csvTemplate,
			ParquetCompressionCodec: parquetCompressionCodec,
		}, nil
	case config.LogExporter:
		return &logsexporter.Exporter{
			LogFormat: func() string {
				if c.LogFormat != "" {
					return c.LogFormat
				}
				return config.DefaultExporter.LogFormat
			}(),
		}, nil
	case config.S3Exporter:
		awsConfig, err := awsConf.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, err
		}

		return &s3exporterv2.Exporter{
			Bucket:                  c.Bucket,
			Format:                  format,
			S3Path:                  c.Path,
			Filename:                filename,
			CsvTemplate:             csvTemplate,
			ParquetCompressionCodec: parquetCompressionCodec,
			AwsConfig:               &awsConfig,
		}, nil
	case config.KinesisExporter:
		awsConfig, err := awsConf.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, err
		}
		return &kinesisexporter.Exporter{
			Format:    format,
			AwsConfig: &awsConfig,
			Settings: kinesisexporter.NewSettings(
				kinesisexporter.WithStreamArn(c.StreamArn),
				kinesisexporter.WithStreamName(c.StreamName),
			),
		}, nil
	case config.GoogleStorageExporter:
		return &gcstorageexporter.Exporter{
			Bucket:                  c.Bucket,
			Format:                  format,
			Path:                    c.Path,
			Filename:                filename,
			CsvTemplate:             csvTemplate,
			ParquetCompressionCodec: parquetCompressionCodec,
		}, nil
	case config.SQSExporter:
		awsConfig, err := awsConf.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, err
		}
		return &sqsexporter.Exporter{
			QueueURL:  c.QueueURL,
			AwsConfig: &awsConfig,
		}, nil
	case config.KafkaExporter:
		settings, err := setKafkaConfig(c.Kafka)
		if err != nil {
			return nil, err
		}
		return &kafkaexporter.Exporter{
			Format:   format,
			Settings: settings,
		}, nil
	case config.PubSubExporter:
		return &pubsubexporter.Exporter{
			ProjectID: c.ProjectID,
			Topic:     c.Topic,
		}, nil
	case config.AzureExporter:
		return &azureexporter.Exporter{
			Container:               c.Container,
			Format:                  format,
			Path:                    c.Path,
			Filename:                filename,
			CsvTemplate:             csvTemplate,
			ParquetCompressionCodec: parquetCompressionCodec,
			AccountKey:              c.AccountKey,
			AccountName:             c.AccountName,
		}, nil
	default:
		return nil, fmt.Errorf("invalid exporter: kind \"%s\" is not supported", c.Kind)
	}
}

// setKafkaConfig set the kafka configuration based on the default configuration
// it will initialize the default configuration and merge it with the changes from the user.
func setKafkaConfig(k kafkaexporter.Settings) (kafkaexporter.Settings, error) {
	if k.Config == nil {
		return k, nil
	}

	defaultConfig := sarama.NewConfig()
	err := mergo.Merge(k.Config, defaultConfig)
	if err != nil {
		return kafkaexporter.Settings{}, err
	}

	switch k.Net.SASL.Mechanism {
	case sarama.SASLTypeSCRAMSHA256:
		k.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &kafka.XDGSCRAMClient{HashGeneratorFcn: kafka.SHA256}
		}
		break
	case sarama.SASLTypeSCRAMSHA512:
		k.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &kafka.XDGSCRAMClient{HashGeneratorFcn: kafka.SHA512}
		}
		break
	}
	return k, nil
}

func initNotifier(c []config.NotifierConf) ([]notifier.Notifier, error) {
	if c == nil {
		return nil, nil
	}
	var notifiers []notifier.Notifier

	for _, cNotif := range c {
		switch cNotif.Kind {
		case config.SlackNotifier:
			if cNotif.WebhookURL == "" && cNotif.SlackWebhookURL != "" { // nolint
				zap.L().Warn("slackWebhookURL field is deprecated, please use webhookURL instead")
				cNotif.WebhookURL = cNotif.SlackWebhookURL // nolint
			}
			notifiers = append(
				notifiers,
				&slacknotifier.Notifier{SlackWebhookURL: cNotif.WebhookURL},
			)
		case config.MicrosoftTeamsNotifier:
			notifiers = append(
				notifiers,
				&microsoftteamsnotifier.Notifier{
					MicrosoftTeamsWebhookURL: cNotif.WebhookURL,
				},
			)
		case config.WebhookNotifier:
			notifiers = append(notifiers,
				&webhooknotifier.Notifier{
					Secret:      cNotif.Secret,
					EndpointURL: cNotif.EndpointURL,
					Meta:        cNotif.Meta,
					Headers:     cNotif.Headers,
				},
			)
		case config.DiscordNotifier:
			notifiers = append(
				notifiers,
				&discordnotifier.Notifier{DiscordWebhookURL: cNotif.WebhookURL},
			)
		default:
			return nil, fmt.Errorf("invalid notifier: kind \"%s\" is not supported", cNotif.Kind)
		}
	}
	return notifiers, nil
}
