package ffclient

import (
	"github.com/aws/aws-sdk-go/aws"
	"net/http"

	"github.com/thomaspoignant/go-feature-flag/internal/retriever"
)

type Config struct {
	PollInterval  int // Poll every X seconds
	LocalFile     string
	HTTPRetriever *HTTPRetriever
	S3Retriever   *S3Retriever
}

type HTTPRetriever struct {
	URL    string
	Method string
	Body   string
	Header http.Header
}

type S3Retriever struct {
	Bucket    string
	Item      string
	AwsConfig aws.Config
}

func (c *Config) GetRetriever() retriever.FlagRetriever {
	if c.S3Retriever != nil {
		return retriever.NewS3Retriever(
			c.S3Retriever.Bucket,
			c.S3Retriever.Item,
			c.S3Retriever.AwsConfig,
		)
	}

	if c.HTTPRetriever != nil {
		return retriever.NewHTTPRetriever(
			http.DefaultClient,
			c.HTTPRetriever.URL,
			c.HTTPRetriever.Method,
			c.HTTPRetriever.Body,
			c.HTTPRetriever.Header,
		)
	}

	if c.LocalFile != "" {
		return retriever.NewLocalRetriever(c.LocalFile)
	}
	return nil
}
