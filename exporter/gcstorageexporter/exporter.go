package gcstorageexporter

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"

	"cloud.google.com/go/storage"

	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
	"google.golang.org/api/option"
)

type Exporter struct {
	// Bucket is the name of your S3 Bucket.
	Bucket string

	// Options are Google Cloud Api options to connect to Google Storage SDK
	Options []option.ClientOption

	// Format is the output format you want in your exported file.
	// Available format are JSON, CSV, and Parquet.
	// Default: JSON
	Format string

	// Path allows you to specify in which directory you want to export your data.
	Path string

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
	// {{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}}\n
	CsvTemplate string

	// ParquetCompressionCodec is the parquet compression codec for better space efficiency.
	// Available options https://github.com/apache/parquet-format/blob/master/Compression.md
	// Default: SNAPPY
	ParquetCompressionCodec string
}

func (f *Exporter) IsBulk() bool {
	return true
}

// Export is saving a collection of events in a file.
func (f *Exporter) Export(ctx context.Context, logger *log.Logger, featureEvents []exporter.FeatureEvent) error {
	// Init google storage client
	client, err := storage.NewClient(ctx, f.Options...)
	if err != nil {
		return err
	}

	if f.Bucket == "" {
		return fmt.Errorf("you should specify a bucket. %v is invalid", f.Bucket)
	}

	// Create a temp directory to store the file we will produce
	outputDir, err := os.MkdirTemp("", "go_feature_flag_GoogleCloudStorage_export")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(outputDir) }()

	// We call the File data exporter to get the file in the right format.
	// Files will be put in the temp directory, so we will be able to upload them to S3 from there.
	fileExporter := fileexporter.Exporter{
		Format:                  f.Format,
		OutputDir:               outputDir,
		Filename:                f.Filename,
		CsvTemplate:             f.CsvTemplate,
		ParquetCompressionCodec: f.ParquetCompressionCodec,
	}
	err = fileExporter.Export(ctx, logger, featureEvents)
	if err != nil {
		return err
	}

	// Upload all the files in the folder to google storage
	files, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		// read file
		of, err := os.Open(outputDir + "/" + file.Name())
		if err != nil {
			fflog.Printf(logger, "error: [Exporter] impossible to open the file %s/%s", outputDir, file.Name())
			continue
		}

		// prepend the path
		source := file.Name()
		if f.Path != "" {
			source = f.Path + "/" + file.Name()
		}

		wc := client.Bucket(f.Bucket).Object(source).NewWriter(ctx)
		_, err = io.Copy(wc, of)
		_ = wc.Close()
		if err != nil {
			return fmt.Errorf("error: [Exporter] impossible to copy the file from %s to bucket %s: %v",
				source, f.Bucket, err)
		}
		fflog.Printf(logger, "info: [Exporter] file %s uploaded.", file.Name())
	}

	return nil
}
