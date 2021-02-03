package ffclient

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"net/http"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/retriever"
)

// Config is the configuration of go-feature-flag.
// PollInterval is the interval in seconds where we gonna read the file to update the cache.
// You should also have a retriever to specify where to read the flags file.
type Config struct {
	PollInterval int // Poll every X seconds
	Logger       *log.Logger
	Context      context.Context // default is context.Background()
	Retriever    Retriever
	Webhooks     []WebhookConfig // webhooks we should call when a flag create/update/delete
}

// GetRetriever returns a retriever.FlagRetriever configure with the retriever available in the config.
func (c *Config) GetRetriever() (retriever.FlagRetriever, error) {
	if c.Retriever == nil {
		return nil, errors.New("no retriever in the configuration, impossible to get the flags")
	}
	return c.Retriever.getFlagRetriever()
}

type Retriever interface {
	getFlagRetriever() (retriever.FlagRetriever, error)
}

// FileRetriever is a configuration struct for a local flat file.
type FileRetriever struct {
	Path string
}

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

func (r *HTTPRetriever) getFlagRetriever() (retriever.FlagRetriever, error) {
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return retriever.NewHTTPRetriever(
		&http.Client{
			Timeout: timeout,
		},
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

// WebhookConfig is the configuration of your webhook.
// we will call this URL with a POST request with the following format
//
//   {
//    "meta":{
//        "hostname": "server01"
//    },
//    "flags":{
//        "deleted": {
//            "test-flag": {
//                "rule": "key eq \"random-key\"",
//                "percentage": 100,
//                "true": true,
//                "false": false,
//                "default": false
//            }
//        },
//        "added": {
//            "test-flag3": {
//                "percentage": 5,
//                "true": "test",
//                "false": "false",
//                "default": "default"
//            }
//        },
//        "updated": {
//            "test-flag2": {
//                "old_value": {
//                    "rule": "key eq \"not-a-key\"",
//                    "percentage": 100,
//                    "true": true,
//                    "false": false,
//                    "default": false
//                },
//                "new_value": {
//                    "disable": true,
//                    "rule": "key eq \"not-a-key\"",
//                    "percentage": 100,
//                    "true": true,
//                    "false": false,
//                    "default": false
//                }
//            }
//        }
//    }
//   }
type WebhookConfig struct {
	PayloadURL string            // PayloadURL of your webhook
	Secret     string            // Secret used to sign your request body.
	Meta       map[string]string // Meta information that you want to send to your webhook (not mandatory)
}
