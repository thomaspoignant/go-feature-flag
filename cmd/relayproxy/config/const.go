package config

import (
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/xitongsys/parquet-go/parquet"
)

const OtelTracerName = "go-feature-flag"
const DefaultLogLevel = "info"

var DefaultExporter = struct {
	Format                  string
	LogFormat               string
	FileName                string
	CsvFormat               string
	FlushInterval           time.Duration
	MaxEventInMemory        int64
	ParquetCompressionCodec string
	LogLevel                string
	ExporterEventType       ffclient.ExporterEventType
}{
	Format:    "JSON",
	LogFormat: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", value=\"{{ .Value}}\"",
	FileName:  "flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}",
	CsvFormat: "{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};" +
		"{{ .Value}};{{ .Default}};{{ .Source}}\n",
	FlushInterval:           60000 * time.Millisecond,
	MaxEventInMemory:        100000,
	ParquetCompressionCodec: parquet.CompressionCodec_SNAPPY.String(),
	LogLevel:                DefaultLogLevel,
	ExporterEventType:       ffclient.FeatureEventExporter,
}
