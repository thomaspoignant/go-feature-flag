package s3exporterv2

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"log"
	"os"
	"sync"
)

type Exporter struct {
	// Bucket is the name of your Exporter Bucket.
	Bucket string

	// AwsConfig is the AWS SDK configuration object we will use to
	// upload your exported data files.
	AwsConfig *aws.Config

	// Format is the output format you want in your exported file.
	// Available format are JSON, CSV and Parquet.
	// Default: JSON
	Format string

	// S3Path allows you to specify in which directory you want to export your data.
	S3Path string

	// Filename is the name of your output file
	// You can use a templated config to define the name of your export files.
	// Available replacement are {{ .Hostname}}, {{ .Timestamp}} and {{ .Format}}
	// Default: "flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}"
	Filename string

	// CsvTemplate is used if your output format is CSV.
	// This field will be ignored if you are using another format than CSV.
	// You can decide which fields you want in your CSV line with a go-template syntax,
	// please check internal/exporter/feature_event.go to see what are the fields available.
	// Default:
	// {{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}}\n
	CsvTemplate string

	// ParquetCompressionCodec is the parquet compression codec for better space efficiency.
	// Available options https://github.com/apache/parquet-format/blob/master/Compression.md
	// Default: SNAPPY
	ParquetCompressionCodec string

	s3Uploader UploaderAPI
	init       sync.Once
}

func (f *Exporter) initializeUploader(ctx context.Context) error {
	var initErr error
	f.init.Do(func() {
		if f.AwsConfig == nil {
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				initErr = fmt.Errorf("impossible to init S3 exporter: %v", err)
				return
			}
			f.AwsConfig = &cfg
		}

		client := s3.NewFromConfig(*f.AwsConfig)
		f.s3Uploader = manager.NewUploader(client)
	})
	return initErr
}

// Export is saving a collection of events in a file.
func (f *Exporter) Export(ctx context.Context, logger *log.Logger, featureEvents []exporter.FeatureEvent) error {
	if f.s3Uploader == nil {
		initErr := f.initializeUploader(ctx)
		if initErr != nil {
			return initErr
		}
	}

	// Create a temp directory to store the file we will produce
	outputDir, err := os.MkdirTemp("", "go_feature_flag_s3_export")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(outputDir) }()

	// We call the File data exporter to get the file in the right format.
	// Files will be put in the temp directory, so we will be able to upload them to Exporter from there.
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

	// Upload all the files in the folder to Exporter
	files, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		// read file
		of, err := os.Open(outputDir + "/" + file.Name())
		if err != nil {
			fflog.Printf(logger, "error: [S3Exporter] impossible to open the file %s/%s", outputDir, file.Name())
			continue
		}

		result, err := f.s3Uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket: aws.String(f.Bucket),
			Key:    aws.String(f.S3Path + "/" + file.Name()),
			Body:   of,
		})

		if err != nil {
			return err
		}

		fflog.Printf(logger, "info: [S3Exporter] file %s uploaded.", result.Location)
	}
	return nil
}

func (f *Exporter) IsBulk() bool {
	return true
}
