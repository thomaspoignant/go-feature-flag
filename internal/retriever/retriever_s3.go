package retriever

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"io/ioutil"
)

func NewS3Retriever(downloader s3manageriface.DownloaderAPI, bucket string, item string) FlagRetriever {
	return &s3Retriever{downloader, bucket, item}
}

type s3Retriever struct {
	downloader s3manageriface.DownloaderAPI
	bucket     string
	item       string
}

func (s *s3Retriever) Retrieve() ([]byte, error) {
	// Download the item from the bucket.
	// If an error occurs, log it and exit.
	// Otherwise, notify the user that the download succeeded.
	file, err := ioutil.TempFile("", "go_feature_flag")
	if err != nil {
		return nil, err
	}

	_, err = s.downloader.Download(file,
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
