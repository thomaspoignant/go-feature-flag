package service

import (
	"fmt"
	"time"

	"github.com/thomaspoignant/go-feature-flag/exporter/kafkaexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/s3exporterv2"
	"github.com/thomaspoignant/go-feature-flag/exporter/sqsexporter"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retrieverv2"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/gcstorageexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/logsexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/webhookexporter"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/notifier/slacknotifier"
	"github.com/thomaspoignant/go-feature-flag/notifier/webhooknotifier"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gcstorageretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gitlabretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/k8sretriever"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"k8s.io/client-go/rest"

	awsConf "github.com/aws/aws-sdk-go-v2/config"
)

func NewGoFeatureFlagClient(
	proxyConf *config.Config,
	logger *zap.Logger,
	notifiers []notifier.Notifier,
) (*ffclient.GoFeatureFlag, error) {
	var mainRetriever retriever.Retriever
	var err error
	if proxyConf.Retriever != nil {
		mainRetriever, err = initRetriever(proxyConf.Retriever)
		if err != nil {
			return nil, err
		}
	}

	// Manage if we have more than 1 retriver
	retrievers := make([]retriever.Retriever, 0)
	if proxyConf.Retrievers != nil {
		for _, r := range *proxyConf.Retrievers {
			r := r
			currentRetriever, err := initRetriever(&r)
			if err != nil {
				return nil, err
			}
			retrievers = append(retrievers, currentRetriever)
		}
	}

	var exp ffclient.DataExporter
	if proxyConf.Exporter != nil {
		exp, err = initExporter(proxyConf.Exporter)
		if err != nil {
			return nil, err
		}
	}

	var notif []notifier.Notifier
	if proxyConf.Notifiers != nil {
		notif, err = initNotifier(proxyConf.Notifiers)
		if err != nil {
			return nil, err
		}
	}
	notif = append(notif, notifiers...)

	f := ffclient.Config{
		PollingInterval:             time.Duration(proxyConf.PollingInterval) * time.Millisecond,
		Logger:                      zap.NewStdLog(logger),
		Context:                     context.Background(),
		Retriever:                   mainRetriever,
		Retrievers:                  retrievers,
		Notifiers:                   notif,
		FileFormat:                  proxyConf.FileFormat,
		DataExporter:                exp,
		StartWithRetrieverError:     proxyConf.StartWithRetrieverError,
		EnablePollingJitter:         proxyConf.EnablePollingJitter,
		EvaluationContextEnrichment: proxyConf.EvaluationContextEnrichment,
	}

	return ffclient.New(f)
}

func initRetriever(c *config.RetrieverConf) (retriever.Retriever, error) {
	retrieverTimeout := config.DefaultRetriever.Timeout
	if c.Timeout != 0 {
		retrieverTimeout = time.Duration(c.Timeout) * time.Millisecond
	}

	// Conversions
	switch c.Kind {
	case config.GitHubRetriever:
		token := c.AuthToken
		if token == "" && c.GithubToken != "" { // nolint: staticcheck
			token = c.GithubToken // nolint: staticcheck
		}
		return &githubretriever.Retriever{
			RepositorySlug: c.RepositorySlug,
			Branch: func() string {
				if c.Branch == "" {
					return config.DefaultRetriever.GitBranch
				}
				return c.Branch
			}(),
			FilePath:    c.Path,
			GithubToken: token,
			Timeout:     retrieverTimeout,
		}, nil
	case config.GitlabRetriever:
		return &gitlabretriever.Retriever{
			BaseURL: c.BaseURL,
			Branch: func() string {
				if c.Branch == "" {
					return config.DefaultRetriever.GitBranch
				}
				return c.Branch
			}(),
			FilePath:       c.Path,
			GitlabToken:    c.AuthToken,
			RepositorySlug: c.RepositorySlug,
			Timeout:        retrieverTimeout,
		}, nil
	case config.FileRetriever:
		return &fileretriever.Retriever{Path: c.Path}, nil
	case config.S3Retriever:
		awsConfig, err := awsConf.LoadDefaultConfig(context.Background())
		return &s3retrieverv2.Retriever{Bucket: c.Bucket, Item: c.Item, AwsConfig: &awsConfig}, err
	case config.HTTPRetriever:
		return &httpretriever.Retriever{
			URL: c.URL,
			Method: func() string {
				if c.HTTPMethod == "" {
					return config.DefaultRetriever.HTTPMethod
				}
				return c.HTTPMethod
			}(), Body: c.HTTPBody, Header: c.HTTPHeaders, Timeout: retrieverTimeout}, nil
	case config.GoogleStorageRetriever:
		return &gcstorageretriever.Retriever{Bucket: c.Bucket, Object: c.Object}, nil
	case config.KubernetesRetriever:
		client, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return &k8sretriever.Retriever{Namespace: c.Namespace, ConfigMapName: c.ConfigMap, Key: c.Key,
			ClientConfig: *client}, nil
	default:
		return nil, fmt.Errorf("invalid retriever: kind \"%s\" "+
			"is not supported, accepted kind: [googleStorage, http, s3, file, github]", c.Kind)
	}
}

func initExporter(c *config.ExporterConf) (ffclient.DataExporter, error) {
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

	switch c.Kind {
	case config.WebhookExporter:
		dataExp.Exporter = &webhookexporter.Exporter{
			EndpointURL: c.EndpointURL, Secret: c.Secret, Meta: c.Meta, Headers: c.Headers}
		return dataExp, nil
	case config.FileExporter:
		dataExp.Exporter = &fileexporter.Exporter{Format: format, OutputDir: c.OutputDir, Filename: filename,
			CsvTemplate: csvTemplate, ParquetCompressionCodec: parquetCompressionCodec}
		return dataExp, nil
	case config.LogExporter:
		dataExp.Exporter = &logsexporter.Exporter{
			LogFormat: func() string {
				if c.LogFormat != "" {
					return c.LogFormat
				}
				return config.DefaultExporter.LogFormat
			}(),
		}
		return dataExp, nil
	case config.S3Exporter:
		awsConfig, err := awsConf.LoadDefaultConfig(context.Background())
		if err != nil {
			return dataExp, err
		}
		dataExp.Exporter = &s3exporterv2.Exporter{
			Bucket:                  c.Bucket,
			Format:                  format,
			S3Path:                  c.Path,
			Filename:                filename,
			CsvTemplate:             csvTemplate,
			ParquetCompressionCodec: parquetCompressionCodec,
			AwsConfig:               &awsConfig,
		}
		return dataExp, nil
	case config.GoogleStorageExporter:
		dataExp.Exporter = &gcstorageexporter.Exporter{
			Bucket:                  c.Bucket,
			Format:                  format,
			Path:                    c.Path,
			Filename:                filename,
			CsvTemplate:             csvTemplate,
			ParquetCompressionCodec: parquetCompressionCodec,
		}
		return dataExp, nil
	case config.SQSExporter:
		awsConfig, err := awsConf.LoadDefaultConfig(context.Background())
		if err != nil {
			return dataExp, err
		}
		dataExp.Exporter = &sqsexporter.Exporter{QueueURL: c.QueueURL, AwsConfig: &awsConfig}
		return dataExp, nil
	case config.KafkaExporter:
		dataExp.Exporter = &kafkaexporter.Exporter{
			Format:   format,
			Settings: c.Kafka,
		}

		return dataExp, nil
	default:
		return ffclient.DataExporter{}, fmt.Errorf("invalid exporter: kind \"%s\" is not supported", c.Kind)
	}
}

func initNotifier(c []config.NotifierConf) ([]notifier.Notifier, error) {
	var notifiers []notifier.Notifier

	for _, cNotif := range c {
		switch cNotif.Kind {
		case config.SlackNotifier:
			notifiers = append(notifiers, &slacknotifier.Notifier{SlackWebhookURL: cNotif.SlackWebhookURL})

		case config.WebhookNotifier:
			notifiers = append(notifiers,
				&webhooknotifier.Notifier{
					Secret:      cNotif.Secret,
					EndpointURL: cNotif.EndpointURL,
					Meta:        cNotif.Meta,
					Headers:     cNotif.Headers,
				},
			)

		default:
			return nil, fmt.Errorf("invalid notifier: kind \"%s\" is not supported", cNotif.Kind)
		}
	}
	return notifiers, nil
}
