package s3retrieverv2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/thomaspoignant/go-feature-flag/retriever"
)

// Retriever is a configuration struct for a S3 retriever.
type Retriever struct {
	// Bucket is the name of your S3 Bucket.
	Bucket string

	// Item is the path to your flag file in your bucket.
	Item string

	// AwsConfig is the AWS SDK configuration object we will use to
	// download your feature flag configuration file.
	AwsConfig *aws.Config

	// downloader is an internal field, it is the downloader use by the AWS-SDK
	downloader DownloaderAPI
	status     retriever.Status
}

func (s *Retriever) Init(ctx context.Context, _ *log.Logger) error {
	s.status = retriever.RetrieverNotReady
	if s.downloader == nil {
		if s.AwsConfig == nil {
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				s.status = retriever.RetrieverError
				return fmt.Errorf("impossible to init S3 retriever v2: %v", err)
			}
			s.AwsConfig = &cfg
		}
		client := s3.NewFromConfig(*s.AwsConfig)
		s.downloader = manager.NewDownloader(client)
	}
	s.status = retriever.RetrieverReady
	return nil
}
func (s *Retriever) Shutdown(_ context.Context) error {
	s.status = retriever.RetrieverNotReady
	s.downloader = nil
	return nil
}
func (s *Retriever) Status() retriever.Status {
	return s.status
}

func (s *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	if s.downloader == nil {
		s.status = retriever.RetrieverError
		return nil, fmt.Errorf("downloader is not initialized")
	}

	// Download the item from the bucket.
	// If an error occurs, log it and exit.
	// Otherwise, notify the user that the download succeeded.
	writerAt := manager.NewWriteAtBuffer([]byte{})

	s3Req := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Item),
	}

	_, err := s.downloader.Download(ctx, writerAt, s3Req)
	if err != nil {
		return nil, fmt.Errorf("unable to download item from S3 %q, %v", s.Item, err)
	}

	return writerAt.Bytes(), nil
}
