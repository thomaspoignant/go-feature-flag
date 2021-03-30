package ffclient

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"net/http"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/internal/retriever"
)

type Retriever interface {
	getFlagRetriever() (retriever.FlagRetriever, error)
}

// FileRetriever is a configuration struct for a local flat file.
type FileRetriever struct {
	Path string
}

// nolint: unused
func (r *FileRetriever) getFlagRetriever() (retriever.FlagRetriever, error) { // nolint: unparam
	return retriever.NewLocalRetriever(r.Path), nil
}

// HTTPRetriever is a configuration struct for an HTTP endpoint retriever.
type HTTPRetriever struct {
	URL     string
	Method  string
	Body    string
	Header  http.Header
	Timeout time.Duration
}

// nolint: unused
func (r *HTTPRetriever) getFlagRetriever() (retriever.FlagRetriever, error) {
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return retriever.NewHTTPRetriever(
		internal.HTTPClientWithTimeout(timeout),
		r.URL,
		r.Method,
		r.Body,
		r.Header,
	), nil
}

// S3Retriever is a configuration struct for a S3 retriever.
type S3Retriever struct {
	Bucket    string
	Item      string
	AwsConfig aws.Config
}

// nolint: unused
func (r *S3Retriever) getFlagRetriever() (retriever.FlagRetriever, error) {
	// Create an AWS session
	sess, err := session.NewSession(&r.AwsConfig)
	if err != nil {
		return nil, err
	}

	// Create a new AWS S3 downloader
	downloader := s3manager.NewDownloader(sess)
	return retriever.NewS3Retriever(
		downloader,
		r.Bucket,
		r.Item,
	), nil
}

// GithubRetriever is a configuration struct for a GitHub retriever.
type GithubRetriever struct {
	RepositorySlug string
	Branch         string // default is main
	FilePath       string
	GithubToken    string
	Timeout        time.Duration // default is 10 seconds
}
// nolint: unused
func (r *GithubRetriever) getFlagRetriever() (retriever.FlagRetriever, error) {
	// default branch is main
	branch := r.Branch
	if branch == "" {
		branch = "main"
	}

	// add header for Github Token if specified
	header := http.Header{}
	if r.GithubToken != "" {
		header.Add("Authorization", fmt.Sprintf("token %s", r.GithubToken))
	}

	URL := fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/%s",
		r.RepositorySlug,
		branch,
		r.FilePath)

	httpRetriever := HTTPRetriever{
		URL:     URL,
		Method:  http.MethodGet,
		Header:  header,
		Timeout: r.Timeout,
	}

	return httpRetriever.getFlagRetriever()
}
