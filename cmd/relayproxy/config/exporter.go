package config

import "fmt"

// ExporterConf contains all the field to configure an exporter
type ExporterConf struct {
	Kind             ExporterKind      `mapstructure:"kind"`
	OutputDir        string            `mapstructure:"outputDir"`
	Format           string            `mapstructure:"format"`
	Filename         string            `mapstructure:"filename"`
	CsvTemplate      string            `mapstructure:"csvTemplate"`
	Bucket           string            `mapstructure:"bucket"`
	Path             string            `mapstructure:"path"`
	EndpointURL      string            `mapstructure:"endpointUrl"`
	Secret           string            `mapstructure:"secret"`
	Meta             map[string]string `mapstructure:"meta"`
	LogFormat        string            `mapstructure:"logFormat"`
	FlushInterval    int64             `mapstructure:"flushInterval"`
	MaxEventInMemory int64             `mapstructure:"maxEventInMemory"`
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
