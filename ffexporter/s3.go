package ffexporter

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/internal/exporter"
	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
)

type S3 struct {
	// Bucket is the name of your S3 Bucket.
	Bucket string

	// AwsConfig is the AWS SDK configuration object we will use to
	// upload your exported data files.
	AwsConfig *aws.Config

	// Format is the output format you want in your exported file.
	// Available format are JSON and CSV.
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

	s3Uploader s3manageriface.UploaderAPI
	init       sync.Once
}

// Export is saving a collection of events in a file.
func (f *S3) Export(ctx context.Context, logger *log.Logger, featureEvents []exporter.FeatureEvent) error {
	// init the s3 uploader
	if f.s3Uploader == nil {
		var initErr error
		f.init.Do(func() {
			var sess *session.Session
			sess, initErr = session.NewSession(f.AwsConfig)
			f.s3Uploader = s3manager.NewUploader(sess)
		})
		// Check that we don't have error in the init.Do()
		if initErr != nil {
			return initErr
		}
	}

	// Create a temp directory to store the file we will produce
	outputDir, err := ioutil.TempDir("", "go_feature_flag_s3_export")
	if err != nil {
		return err
	}
	defer os.Remove(outputDir)

	// We call the File data exporter to get the file in the right format.
	// Files will be put in the temp directory, so we will be able to upload them to S3 from there.
	fileExporter := File{
		Format:      f.Format,
		OutputDir:   outputDir,
		Filename:    f.Filename,
		CsvTemplate: f.CsvTemplate,
	}
	err = fileExporter.Export(ctx, logger, featureEvents)
	if err != nil {
		return err
	}

	// Upload all the files in the folder to S3
	files, err := ioutil.ReadDir(outputDir)
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

		result, err := f.s3Uploader.UploadWithContext(
			ctx,
			&s3manager.UploadInput{
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

func (f *S3) IsBulk() bool {
	return true
}
