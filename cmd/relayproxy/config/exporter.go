package config

import (
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/exporter/kafkaexporter"
	"github.com/xitongsys/parquet-go/parquet"
)

// ExporterConf contains all the field to configure an exporter
type ExporterConf struct {
	Kind                    ExporterKind           `mapstructure:"kind" koanf:"kind"`
	OutputDir               string                 `mapstructure:"outputDir" koanf:"outputdir"`
	Format                  string                 `mapstructure:"format" koanf:"format"`
	Filename                string                 `mapstructure:"filename" koanf:"filename"`
	CsvTemplate             string                 `mapstructure:"csvTemplate" koanf:"csvtemplate"`
	Bucket                  string                 `mapstructure:"bucket" koanf:"bucket"`
	Path                    string                 `mapstructure:"path" koanf:"path"`
	EndpointURL             string                 `mapstructure:"endpointUrl" koanf:"endpointurl"`
	Secret                  string                 `mapstructure:"secret" koanf:"secret"`
	Meta                    map[string]string      `mapstructure:"meta" koanf:"meta"`
	LogFormat               string                 `mapstructure:"logFormat" koanf:"logformat"`
	FlushInterval           int64                  `mapstructure:"flushInterval" koanf:"flushinterval"`
	MaxEventInMemory        int64                  `mapstructure:"maxEventInMemory" koanf:"maxeventinmemory"`
	ParquetCompressionCodec string                 `mapstructure:"parquetCompressionCodec" koanf:"parquetcompressioncodec"`
	Headers                 map[string][]string    `mapstructure:"headers" koanf:"headers"`
	QueueURL                string                 `mapstructure:"queueUrl" koanf:"queueurl"`
	Kafka                   kafkaexporter.Settings `mapstructure:"kafka" koanf:"kafka"`
	ProjectID               string                 `mapstructure:"projectID" koanf:"projectid"`
	Topic                   string                 `mapstructure:"topic" koanf:"topic"`
}

func (c *ExporterConf) IsValid() error {
	if err := c.Kind.IsValid(); err != nil {
		return err
	}
	if c.Kind == FileExporter && c.OutputDir == "" {
		return fmt.Errorf("invalid exporter: no \"outputDir\" property found for kind \"%s\"", c.Kind)
	}
	if (c.Kind == S3Exporter || c.Kind == GoogleStorageExporter) && c.Bucket == "" {
		return fmt.Errorf("invalid exporter: no \"bucket\" property found for kind \"%s\"", c.Kind)
	}
	if c.Kind == WebhookExporter && c.EndpointURL == "" {
		return fmt.Errorf("invalid exporter: no \"endpointUrl\" property found for kind \"%s\"", c.Kind)
	}
	if len(c.ParquetCompressionCodec) > 0 {
		if _, err := parquet.CompressionCodecFromString(c.ParquetCompressionCodec); err != nil {
			return fmt.Errorf("invalid exporter: \"parquetCompressionCodec\" err: %v", err)
		}
	}
	if c.Kind == SQSExporter && c.QueueURL == "" {
		return fmt.Errorf("invalid exporter: no \"queueUrl\" property found for kind \"%s\"", c.Kind)
	}

	if c.Kind == KafkaExporter && (c.Kafka.Topic == "" || len(c.Kafka.Addresses) == 0) {
		return fmt.Errorf("invalid exporter: \"kakfa.topic\" and \"kafka.addresses\" are required for kind \"%s\"", c.Kind)
	}

	if c.Kind == PubSubExporter && (c.ProjectID == "" || c.Topic == "") {
		return fmt.Errorf("invalid exporter: \"projectID\" and \"topic\" are required for kind \"%s\"", c.Kind)
	}

	return nil
}

type ExporterKind string

const (
	FileExporter          ExporterKind = "file"
	WebhookExporter       ExporterKind = "webhook"
	LogExporter           ExporterKind = "log"
	S3Exporter            ExporterKind = "s3"
	GoogleStorageExporter ExporterKind = "googleStorage"
	SQSExporter           ExporterKind = "sqs"
	KafkaExporter         ExporterKind = "kafka"
	PubSubExporter        ExporterKind = "pubsub"
)

// IsValid is checking if the value is part of the enum
func (r ExporterKind) IsValid() error {
	switch r {
	case FileExporter, WebhookExporter, LogExporter, S3Exporter, GoogleStorageExporter, SQSExporter, KafkaExporter,
		PubSubExporter:
		return nil
	}
	return fmt.Errorf("invalid exporter: kind \"%s\" is not supported", r)
}
