package service

import (
	"context"
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
	retrieverInit "github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf/init"
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
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils"
	"go.uber.org/zap"
)

func NewGoFeatureFlagClient(
	cFlagSet *config.FlagSet,
	logger *zap.Logger,
	notifiers []notifier.Notifier,
) (*ffclient.GoFeatureFlag, error) {
	var err error
	if cFlagSet == nil {
		return nil, fmt.Errorf("proxy config is empty")
	}

	retrievers, err := initRetrievers(cFlagSet)
	if err != nil {
		return nil, err
	}

	exporters, err := initDataExporters(cFlagSet)
	if err != nil {
		return nil, err
	}

	notif := make([]notifier.Notifier, 0)
	if cFlagSet.Notifiers != nil {
		notif, err = initNotifier(cFlagSet.Notifiers)
		if err != nil {
			return nil, err
		}
	}

	// backward compatibility for the notifier field, it was called "notifier" instead of "notifiers"
	// fixed in version v1.46.0
	if len(notif) == 0 && cFlagSet.FixNotifiers != nil { // nolint: staticcheck
		notif, err = initNotifier(cFlagSet.FixNotifiers) // nolint: staticcheck
		if err != nil {
			return nil, err
		}
	}
	// end of backward compatibility for the notifier field in version v1.66.0
	notif = append(notif, notifiers...)

	f := ffclient.Config{
		PollingInterval: time.Duration(
			cFlagSet.PollingInterval,
		) * time.Millisecond,
		LeveledLogger:                   initLeveledLogger(cFlagSet, logger),
		Context:                         context.Background(),
		Retrievers:                      retrievers,
		Notifiers:                       notif,
		FileFormat:                      cFlagSet.FileFormat,
		DataExporters:                   exporters,
		StartWithRetrieverError:         cFlagSet.StartWithRetrieverError,
		EnablePollingJitter:             cFlagSet.EnablePollingJitter,
		DisableNotifierOnInit:           cFlagSet.DisableNotifierOnInit,
		EvaluationContextEnrichment:     cFlagSet.EvaluationContextEnrichment,
		PersistentFlagConfigurationFile: cFlagSet.PersistentFlagConfigurationFile,
		Name:                            &cFlagSet.Name,
	}
	client, err := ffclient.New(f)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// initLeveledLogger initializes the leveled logger
// it will add the flagset name as a contextual attribute if it is not the default flagset
func initLeveledLogger(c *config.FlagSet, logger *zap.Logger) *slog.Logger {
	baseHandler := slogzap.Option{Level: slog.LevelDebug, Logger: logger}.NewZapHandler()
	if c.Name != "" && c.Name != utils.DefaultFlagSetName {
		attrs := []slog.Attr{slog.String("flagset", c.Name)}
		baseHandler = baseHandler.WithAttrs(attrs)
	}
	return slog.New(baseHandler)
}

// initRetrievers initialize the retrievers based on the configuration
// it handles both the `retriever` and `retrievers` fields
func initRetrievers(proxyConf *config.FlagSet) ([]retriever.Retriever, error) {
	retrievers := make([]retriever.Retriever, 0)
	// if the retriever is set, we add it to the retrievers
	if proxyConf.Retriever != nil {
		currentRetriever, err := retrieverInit.InitRetriever(proxyConf.Retriever)
		if err != nil {
			return nil, err
		}
		retrievers = append(retrievers, currentRetriever)
	}
	// if the retrievers are set, we add them to the retrievers
	if proxyConf.Retrievers != nil {
		for _, r := range *proxyConf.Retrievers {
			currentRetriever, err := retrieverInit.InitRetriever(&r)
			if err != nil {
				return nil, err
			}
			retrievers = append(retrievers, currentRetriever)
		}
	}
	return retrievers, nil
}

// initDataExporters initialize the exporters based on the configuration
// it handles both the `exporter` and `exporters` fields.
func initDataExporters(proxyConf *config.FlagSet) ([]ffclient.DataExporter, error) {
	exporters := make([]ffclient.DataExporter, 0)
	if proxyConf.Exporter != nil {
		currentExporter, err := initDataExporter(proxyConf.Exporter)
		if err != nil {
			return nil, err
		}
		exporters = append(exporters, currentExporter)
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
	exporterEventType := c.ExporterEventType
	if exporterEventType == "" {
		exporterEventType = config.DefaultExporter.ExporterEventType
	}
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
		ExporterEventType: exporterEventType,
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
		return &pubsubexporterv2.Exporter{
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
	c := kafkaexporter.Settings{
		Topic:     k.Topic,
		Addresses: k.Addresses,
	}

	if k.Config == nil {
		return c, nil
	}
	saramaConfig := sarama.NewConfig()
	err := mergo.Merge(saramaConfig, k.Config)
	if err != nil {
		return kafkaexporter.Settings{}, err
	}
	saramaConfig.Producer.Return.Errors = true

	switch saramaConfig.Net.SASL.Mechanism {
	case sarama.SASLTypeSCRAMSHA256:
		saramaConfig.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &kafka.XDGSCRAMClient{HashGeneratorFcn: kafka.SHA256}
		}
	case sarama.SASLTypeSCRAMSHA512:
		saramaConfig.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &kafka.XDGSCRAMClient{HashGeneratorFcn: kafka.SHA512}
		}
	}
	c.Config = saramaConfig
	return c, nil
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
