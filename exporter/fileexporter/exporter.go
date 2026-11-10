package fileexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/compress"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

var _ exporter.Exporter = &Exporter{}

type Exporter struct {
	// Format is the output format you want in your exported file.
	// Available format are JSON, CSV, and Parquet.
	// Default: JSON
	Format string

	// OutputDir is the location of the directory where to store the exported files
	// Default: the current directory
	OutputDir string

	// Filename is the name of your output file
	// You can use a templated config to define the name of your export files.
	// Available replacement are {{ .Hostname}}, {{ .Timestamp}} and {{ .Format}}
	// Default: "flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}"
	Filename string

	// CsvTemplate is used if your output format is CSV.
	// This field will be ignored if you are using another format than CSV.
	// You can decide which fields you want in your CSV line with a go-template syntax,
	// please check exporter/feature_event.go to see what are the fields available.
	// Default:
	// {{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};
	// {{ .Default}};{{ .Source}}\n
	CsvTemplate string

	// ParquetCompressionCodec is the parquet compression codec for better space efficiency.
	// Available options https://github.com/apache/parquet-format/blob/master/Compression.md
	// Default: SNAPPY
	ParquetCompressionCodec string

	csvTemplate      *template.Template
	filenameTemplate *template.Template
	initTemplates    sync.Once
}

// Export is saving a collection of events in a file.
func (f *Exporter) Export(
	_ context.Context,
	_ *fflog.FFLogger,
	events []exporter.ExportableEvent,
) error {
	// Parse the template only once
	f.initTemplates.Do(func() {
		f.csvTemplate = exporter.ParseTemplate(
			"csvFormat",
			f.CsvTemplate,
			exporter.DefaultCsvTemplate,
		)
		f.filenameTemplate = exporter.ParseTemplate(
			"filenameFormat",
			f.Filename,
			exporter.DefaultFilenameTemplate,
		)
	})

	// Default format for the output
	if f.Format == "" {
		f.Format = "json"
	}
	f.Format = strings.ToLower(f.Format)

	// Get the filename
	filename, err := exporter.ComputeFilename(f.filenameTemplate, f.Format)
	if err != nil {
		return err
	}

	// Handle empty OutputDir and remove trailing slash
	outputDir := strings.TrimRight(f.OutputDir, "/")
	var filePath string
	if outputDir == "" {
		filePath = filename
	} else {
		// Ensure OutputDir exists or create it
		// nolint:gosec
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
		filePath = filepath.Join(outputDir, filename)
	}

	if f.Format == "parquet" {
		return f.writeParquet(filePath, events)
	}
	return f.writeFile(filePath, events)
}

// IsBulk return false if we should directly send the data as soon as it is produce
// and true if we collect the data to send them in bulk.
func (f *Exporter) IsBulk() bool {
	return true
}

func (f *Exporter) writeFile(filePath string, events []exporter.ExportableEvent) error {
	//nolint:gosec
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	for _, event := range events {
		var line []byte
		var err error

		// Convert the line in the right format
		switch f.Format {
		case "csv":
			line, err = event.FormatInCSV(f.csvTemplate)
		case "json":
			line, err = event.FormatInJSON()
		default:
			line, err = event.FormatInJSON()
		}

		// Handle error and write line into the file
		if err != nil {
			return fmt.Errorf("impossible to format the event in %s: %v", f.Format, err)
		}
		_, errWrite := file.Write(line)
		if errWrite != nil {
			return fmt.Errorf("error while writing the export file: %v", errWrite)
		}
	}
	return nil
}

func (f *Exporter) writeParquet(filePath string, events []exporter.ExportableEvent) error {
	parquetFeatureEvents := make([]exporter.FeatureEvent, 0)
	parquetTrackingEvents := make([]exporter.TrackingEvent, 0)
	for _, event := range events {
		switch ev := any(event).(type) {
		case exporter.FeatureEvent:
			parquetFeatureEvents = append(parquetFeatureEvents, ev)
		case exporter.TrackingEvent:
			parquetTrackingEvents = append(parquetTrackingEvents, ev)
		default:
			// do nothing
		}
	}
	if len(parquetTrackingEvents) > 0 {
		return f.writeParquetTrackingEvent(filePath, parquetTrackingEvents)
	}
	return f.writeParquetFeatureEvent(filePath, parquetFeatureEvents)
}

// parquetFeatureEvent is the parquet row written for an exporter.FeatureEvent.
type parquetFeatureEvent struct {
	Kind         string            `parquet:"kind"`
	ContextKind  string            `parquet:"contextKind"`
	UserKey      string            `parquet:"userKey"`
	CreationDate int64             `parquet:"creationDate"`
	Key          string            `parquet:"key"`
	Variation    string            `parquet:"variation"`
	Value        string            `parquet:"value"`
	Default      bool              `parquet:"default"`
	Version      string            `parquet:"version"`
	Source       string            `parquet:"source"`
	Metadata     map[string]string `parquet:"metadata"`
}

// parquetTrackingEvent is the parquet row written for an exporter.TrackingEvent.
type parquetTrackingEvent struct {
	Kind              string            `parquet:"kind"`
	ContextKind       string            `parquet:"contextKind"`
	UserKey           string            `parquet:"userKey"`
	CreationDate      int64             `parquet:"creationDate"`
	Key               string            `parquet:"key"`
	EvaluationContext map[string]string `parquet:"evaluationContext"`
	TrackingDetails   map[string]string `parquet:"trackingEventDetails"`
}

// parquetCodecs maps the ParquetCompressionCodec names to their codec.
// LZ4 is deprecated in the parquet format, so it is written as LZ4_RAW.
var parquetCodecs = map[string]compress.Codec{
	"UNCOMPRESSED": &parquet.Uncompressed,
	"SNAPPY":       &parquet.Snappy,
	"GZIP":         &parquet.Gzip,
	"BROTLI":       &parquet.Brotli,
	"ZSTD":         &parquet.Zstd,
	"LZ4":          &parquet.Lz4Raw,
	"LZ4_RAW":      &parquet.Lz4Raw,
}

// writeParquetFeatureEvent writes the feature events in a parquet file
func (f *Exporter) writeParquetFeatureEvent(filePath string, events []exporter.FeatureEvent) error {
	rows := make([]parquetFeatureEvent, 0, len(events))
	for _, event := range events {
		eventValue, err := event.ConvertValueForParquet()
		if err != nil {
			return err
		}
		rows = append(rows, parquetFeatureEvent{
			Kind:         event.Kind,
			ContextKind:  event.ContextKind,
			UserKey:      event.UserKey,
			CreationDate: event.CreationDate,
			Key:          event.Key,
			Variation:    event.Variation,
			Value:        eventValue,
			Default:      event.Default,
			Version:      event.Version,
			Source:       event.Source,
			Metadata:     toParquetMap(event.Metadata),
		})
	}
	return writeParquetFile(filePath, rows, f.ParquetCompressionCodec)
}

// writeParquetTrackingEvent writes the tracking events in a parquet file
func (f *Exporter) writeParquetTrackingEvent(
	filePath string,
	events []exporter.TrackingEvent,
) error {
	rows := make([]parquetTrackingEvent, 0, len(events))
	for _, event := range events {
		rows = append(rows, parquetTrackingEvent{
			Kind:              event.Kind,
			ContextKind:       event.ContextKind,
			UserKey:           event.UserKey,
			CreationDate:      event.CreationDate,
			Key:               event.Key,
			EvaluationContext: toParquetMap(event.EvaluationContext),
			TrackingDetails:   toParquetMap(event.TrackingDetails),
		})
	}
	return writeParquetFile(filePath, rows, f.ParquetCompressionCodec)
}

func writeParquetFile[T any](filePath string, rows []T, codecName string) error {
	codec, ok := parquetCodecs[codecName]
	if !ok {
		codec = &parquet.Snappy
	}
	if err := parquet.WriteFile(filePath, rows, parquet.Compression(codec)); err != nil {
		return fmt.Errorf("error while writing the parquet export file: %v", err)
	}
	return nil
}

// toParquetMap converts a map to a map of strings, non-string values are JSON encoded.
func toParquetMap(m map[string]any) map[string]string {
	res := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			res[k] = s
			continue
		}
		b, err := json.Marshal(v)
		if err != nil {
			b = []byte(fmt.Sprint(v))
		}
		res[k] = string(b)
	}
	return res
}
