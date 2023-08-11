package s3retrieverv2

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"os"
	"sync"
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
	init       sync.Once
}

func (s *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	if s.downloader == nil {
		initErr := s.initializeDownloader(ctx)
		if initErr != nil {
			return nil, initErr
		}
	}

	// Download the item from the bucket.
	// If an error occurs, log it and exit.
	// Otherwise, notify the user that the download succeeded.
	file, err := os.CreateTemp("", "go_feature_flag")
	if err != nil {
		return nil, err
	}

	s3Req := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Item),
	}
	_, err = s.downloader.Download(ctx, file, s3Req)
	if err != nil {
		return nil, fmt.Errorf("unable to download item from S3 %q, %v", s.Item, err)
	}
	// Read file content
	content, err := os.ReadFile(file.Name())
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (s *Retriever) initializeDownloader(ctx context.Context) error {
	var initErr error
	s.init.Do(func() {
		if s.AwsConfig == nil {
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				initErr = fmt.Errorf("impossible to init S3 retriever: %v", err)
				return
			}
			s.AwsConfig = &cfg
		}
		client := s3.NewFromConfig(*s.AwsConfig)
		s.downloader = manager.NewDownloader(client)
	})
	return initErr
}
