package config

import (
	"fmt"

	"github.com/xitongsys/parquet-go/parquet"
)

// ExporterConf contains all the field to configure an exporter
type ExporterConf struct {
	Kind                    ExporterKind        `mapstructure:"kind" koanf:"kind"`
	OutputDir               string              `mapstructure:"outputDir" koanf:"outputdir"`
	Format                  string              `mapstructure:"format" koanf:"format"`
	Filename                string              `mapstructure:"filename" koanf:"filename"`
	CsvTemplate             string              `mapstructure:"csvTemplate" koanf:"csvtemplate"`
	Bucket                  string              `mapstructure:"bucket" koanf:"bucket"`
	Path                    string              `mapstructure:"path" koanf:"path"`
	EndpointURL             string              `mapstructure:"endpointUrl" koanf:"endpointurl"`
	Secret                  string              `mapstructure:"secret" koanf:"secret"`
	Meta                    map[string]string   `mapstructure:"meta" koanf:"meta"`
	LogFormat               string              `mapstructure:"logFormat" koanf:"logformat"`
	FlushInterval           int64               `mapstructure:"flushInterval" koanf:"flushinterval"`
	MaxEventInMemory        int64               `mapstructure:"maxEventInMemory" koanf:"maxeventinmemory"`
	ParquetCompressionCodec string              `mapstructure:"parquetCompressionCodec" koanf:"parquetcompressioncodec"`
	Headers                 map[string][]string `mapstructure:"headers" koanf:"headers"`
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
	return nil
}

type ExporterKind string

const (
	FileExporter          ExporterKind = "file"
	WebhookExporter       ExporterKind = "webhook"
	LogExporter           ExporterKind = "log"
	S3Exporter            ExporterKind = "s3"
	GoogleStorageExporter ExporterKind = "googleStorage"
)

// IsValid is checking if the value is part of the enum
func (r ExporterKind) IsValid() error {
	switch r {
	case FileExporter, WebhookExporter, LogExporter, S3Exporter, GoogleStorageExporter:
		return nil
	}
	return fmt.Errorf("invalid exporter: kind \"%s\" is not supported", r)
}
