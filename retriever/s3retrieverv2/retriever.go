package s3retrieverv2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
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

	// S3ClientOptions is a list of functional options to configure the S3 client.
	// Provide additional functional options to further configure the behavior of the client,
	// such as changing the client's endpoint or adding custom middleware behavior.
	// For more information about the options, please check:
	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3#Options
	S3ClientOptions []func(*s3.Options)

	// downloader is an internal field, it is the downloader use by the AWS-SDK
	downloader DownloaderAPI
	status     retriever.Status
}

// Init is initializing the retriever to start fetching the flags configuration.
func (s *Retriever) Init(ctx context.Context, _ *fflog.FFLogger) error {
	if ctx == nil {
		ctx = context.Background()
	}
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
		client := s3.NewFromConfig(*s.AwsConfig, s.S3ClientOptions...)
		s.downloader = transfermanager.New(client)
	}
	s.status = retriever.RetrieverReady
	return nil
}

// Shutdown gracefully shutdown the provider and set the status as not ready.
func (s *Retriever) Shutdown(_ context.Context) error {
	s.status = retriever.RetrieverNotReady
	s.downloader = nil
	return nil
}

// Status is the function returning the internal state of the retriever.
func (s *Retriever) Status() retriever.Status {
	return s.status
}

// Retrieve is the function in charge of fetching the flag configuration.
func (s *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if s.downloader == nil {
		s.status = retriever.RetrieverError
		return nil, fmt.Errorf("downloader is not initialized")
	}

	// Download the item from the bucket.
	writerAt := types.NewWriteAtBuffer([]byte{})

	_, err := s.downloader.DownloadObject(ctx, &transfermanager.DownloadObjectInput{
		Bucket:   aws.String(s.Bucket),
		Key:      aws.String(s.Item),
		WriterAt: writerAt,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to download item from S3 %q, %v", s.Item, err)
	}

	return writerAt.Bytes(), nil
}

func (s *Retriever) SetDownloader(downloader DownloaderAPI) {
	s.downloader = downloader
}
