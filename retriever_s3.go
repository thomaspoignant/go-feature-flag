package ffclient

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"io/ioutil"
)

// S3Retriever is a configuration struct for a S3 retriever.
type S3Retriever struct {
	// Bucket is the name of your S3 Bucket.
	Bucket string

	// Item is the path to your flag file in your bucket.
	Item string

	// AwsConfig is the AWS SDK configuration object we will use to
	// download your feature flag configuration file.
	AwsConfig aws.Config

	// downloader is an internal field, it is the downloader use by the AWS-SDK
	downloader s3manageriface.DownloaderAPI
}

func (s *S3Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	// Download the item from the bucket.
	// If an error occurs, log it and exit.
	// Otherwise, notify the user that the download succeeded.
	file, err := ioutil.TempFile("", "go_feature_flag")
	if err != nil {
		return nil, err
	}

	// Create an AWS session
	sess, err := session.NewSession(&s.AwsConfig)
	if err != nil {
		return nil, err
	}

	// Create a new AWS S3 downloader
	if s.downloader == nil {
		s.downloader = s3manager.NewDownloader(sess)
	}

	s3Req := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Item),
	}

	if ctx == nil {
		_, err = s.downloader.Download(file, s3Req)
	} else {
		_, err = s.downloader.DownloadWithContext(ctx, file, s3Req)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to download item from S3 %q, %v", s.Item, err)
	}

	// Read file content
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return content, nil
}
