package retriever

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io/ioutil"
)

func NewS3Retriever(bucket string, item string, awsConfig aws.Config) FlagRetriever {
	return &s3Retriever{bucket, item, awsConfig}
}

type s3Retriever struct {
	bucket    string
	item      string
	awsConfig aws.Config
}

func (s *s3Retriever) Retrieve() ([]byte, error) {
	// Create an AWS session
	sess, err := session.NewSession(&s.awsConfig)
	if err != nil {
		return nil, err
	}

	// Create a new AWS S3 downloader
	downloader := s3manager.NewDownloader(sess)

	// Download the item from the bucket.
	// If an error occurs, log it and exit.
	// Otherwise, notify the user that the download succeeded.
	file, err := ioutil.TempFile("", "go_feature_flag")
	if err != nil {
		return nil, err
	}

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(s.item),
		})

	if err != nil {
		return nil, fmt.Errorf("unable to download item from S3 %q, %v", s.item, err)
	}

	// Read file content
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return content, nil
}
