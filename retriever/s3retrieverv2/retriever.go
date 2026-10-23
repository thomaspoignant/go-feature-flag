package s3retrieverv2

import (
	"context"
	"fmt"
	"sync"

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

	// Prefix is the key prefix (ex: "flags/") under which all the flag files are stored.
	// A trailing "/" is added if missing.
	// All the matching files are merged into a single flag set.
	// Exactly one of Item and Prefix must be set.
	Prefix string

	// Pattern is an optional glob (see path.Match) applied to the base name of each object found
	// under Prefix (ex: "*.yaml"). If empty, all the objects under Prefix are used.
	// Only used with Prefix.
	Pattern string

	// Recursive makes the retriever also look into the "sub folders" of Prefix.
	// If false only the direct children of Prefix are used.
	// Only used with Prefix.
	Recursive bool

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
	lister     ListerAPI
	status     retriever.Status
	logger     *fflog.FFLogger

	// mu protects downloader, lister and cache, and serializes Retrieve calls.
	mu    sync.Mutex
	cache map[string]cachedObject
}

// Init is initializing the retriever to start fetching the flags configuration.
func (s *Retriever) Init(ctx context.Context, logger *fflog.FFLogger) error {
	if ctx == nil {
		ctx = context.Background()
	}
	s.status = retriever.RetrieverNotReady
	if err := s.validate(); err != nil {
		s.status = retriever.RetrieverError
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger = logger
	// Item mode only needs a downloader, prefix mode needs a lister too.
	needLister := s.Prefix != ""
	if s.downloader == nil || (needLister && s.lister == nil) {
		if s.AwsConfig == nil {
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				s.status = retriever.RetrieverError
				return fmt.Errorf("impossible to init S3 retriever v2: %v", err)
			}
			s.AwsConfig = &cfg
		}
		client := s3.NewFromConfig(*s.AwsConfig, s.S3ClientOptions...)
		if s.downloader == nil {
			s.downloader = transfermanager.New(client)
		}
		if needLister && s.lister == nil {
			s.lister = client
		}
	}
	s.status = retriever.RetrieverReady
	return nil
}

// Shutdown gracefully shutdown the provider and set the status as not ready.
func (s *Retriever) Shutdown(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = retriever.RetrieverNotReady
	s.downloader = nil
	s.lister = nil
	s.cache = nil
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.downloader == nil {
		s.status = retriever.RetrieverError
		return nil, fmt.Errorf("downloader is not initialized")
	}

	if s.Prefix != "" {
		return s.retrievePrefix(ctx)
	}

	content, err := s.download(ctx, s.Item)
	if err != nil {
		return nil, fmt.Errorf("unable to download item from S3 %q, %v", s.Item, err)
	}
	return content, nil
}

// download fetches the content of a single object.
func (s *Retriever) download(ctx context.Context, key string) ([]byte, error) {
	writerAt := types.NewWriteAtBuffer([]byte{})
	_, err := s.downloader.DownloadObject(ctx, &transfermanager.DownloadObjectInput{
		Bucket:   aws.String(s.Bucket),
		Key:      aws.String(key),
		WriterAt: writerAt,
	})
	if err != nil {
		return nil, err
	}
	return writerAt.Bytes(), nil
}

// OutputFormat declares the format of the bytes returned by Retrieve.
// In prefix mode the merged flag set is always JSON, in item mode the format is the
// one of the file so we return an empty string to let the global file format apply.
func (s *Retriever) OutputFormat() string {
	if s.Prefix != "" {
		return "json"
	}
	return ""
}

func (s *Retriever) SetDownloader(downloader DownloaderAPI) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.downloader = downloader
}

// SetLister overrides the component used to list the objects of the bucket (prefix mode).
func (s *Retriever) SetLister(lister ListerAPI) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lister = lister
}
