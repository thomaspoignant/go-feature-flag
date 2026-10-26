package s3retrieverv2

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/thomaspoignant/go-feature-flag/internal/folderloader"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// retrieverMode is how the retriever reads the bucket, it is derived from the configuration by Init.
type retrieverMode string

const (
	// singleFileMode downloads the file set in Item. It is the zero value so a Retriever
	// used without calling Init keeps its original behavior.
	singleFileMode retrieverMode = ""
	// folderMode merges the files selected by Prefix, Pattern and Recursive (Item is not set).
	folderMode retrieverMode = "folder"
)

// Retriever is a configuration struct for a S3 retriever.
type Retriever struct {
	// Bucket is the name of your S3 Bucket.
	Bucket string

	// Item is the path to your flag file in your bucket.
	// If it is not set, the retriever is in folder mode: all the files selected by Prefix, Pattern
	// and Recursive are merged into a single flag set.
	Item string

	// Prefix is a key prefix (ex: "flags/") restricting the files loaded in folder mode.
	// Empty (or "/") is the root of the bucket. Leading "/" are trimmed and a trailing "/" is added
	// if missing. Can't be used with Item.
	Prefix string

	// Pattern is an optional glob (see path.Match) matched against the file name only (the last
	// element of the key, not the whole path) of each object found in folder mode (ex: "*.yaml").
	// It can't contain a "/". If empty, all the objects are used. Can't be used with Item.
	Pattern string

	// Recursive makes the folder mode also look into the "sub folders" of Prefix.
	// If false only the direct children of Prefix are used. Can't be used with Item.
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
	// mode is set once by Init, from Item.
	mode retrieverMode

	// loader is the folder mode engine, built by the first Init in folder mode and kept after.
	// Like mode, it is never reassigned once set, so OutputFormat reads both without the lock.
	loader *folderloader.Loader

	// mu protects downloader, lister and loader, and serializes Retrieve calls.
	mu sync.Mutex
}

// Init is initializing the retriever to start fetching the flags configuration.
func (s *Retriever) Init(ctx context.Context, logger *fflog.FFLogger) error {
	if ctx == nil {
		ctx = context.Background()
	}
	s.status = retriever.RetrieverNotReady
	if err := folderloader.Validate("item", s.Item, s.Prefix, s.Pattern, s.Recursive); err != nil {
		s.status = retriever.RetrieverError
		return fmt.Errorf("impossible to init S3 retriever v2: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger = logger
	s.mode = singleFileMode
	if s.Item == "" {
		s.mode = folderMode
		s.initLoader(logger)
	}
	// Single file mode only needs a downloader, folder mode needs a lister too.
	if err := s.initClients(ctx, s.mode == folderMode); err != nil {
		s.status = retriever.RetrieverError
		return fmt.Errorf("impossible to init S3 retriever v2: %v", err)
	}
	s.status = retriever.RetrieverReady
	return nil
}

// initLoader creates the folder loader, once. The caller must hold s.mu.
func (s *Retriever) initLoader(logger *fflog.FFLogger) {
	if s.loader != nil {
		return
	}
	s.loader = folderloader.New(&s3Store{r: s}, folderloader.Options{
		Prefix: s.Prefix, Pattern: s.Pattern, Recursive: s.Recursive,
		Logger: logger, Location: s.Bucket, LocationKey: "bucket", Provider: "S3",
	})
}

// initClients builds the S3 clients that were not provided (see SetDownloader and SetLister).
// The caller must hold s.mu.
func (s *Retriever) initClients(ctx context.Context, needLister bool) error {
	if s.downloader != nil && (!needLister || s.lister != nil) {
		return nil
	}
	if s.AwsConfig == nil {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return err
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
	return nil
}

// Shutdown gracefully shutdown the provider and set the status as not ready.
func (s *Retriever) Shutdown(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = retriever.RetrieverNotReady
	s.downloader = nil
	s.lister = nil
	if s.loader != nil {
		s.loader.Reset()
	}
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

	if s.mode == folderMode {
		if s.lister == nil {
			s.status = retriever.RetrieverError
			return nil, errors.New("lister is not initialized")
		}
		return s.loader.Load(ctx)
	}

	if s.Item == "" {
		return nil, errors.New("S3 retriever not initialized: call Init (Item is empty means folder mode)")
	}

	content, _, err := s.downloadWithETag(ctx, s.Item)
	if err != nil {
		return nil, fmt.Errorf("unable to download item from S3 %q, %v", s.Item, err)
	}
	return content, nil
}

// downloadWithETag fetches the content of a single object, with the ETag of the downloaded version.
func (s *Retriever) downloadWithETag(ctx context.Context, key string) ([]byte, string, error) {
	writerAt := types.NewWriteAtBuffer([]byte{})
	out, err := s.downloader.DownloadObject(ctx, &transfermanager.DownloadObjectInput{
		Bucket:   aws.String(s.Bucket),
		Key:      aws.String(key),
		WriterAt: writerAt,
	})
	if err != nil {
		return nil, "", err
	}
	return writerAt.Bytes(), aws.ToString(out.ETag), nil
}

// OutputFormat declares the format of the bytes returned by Retrieve.
// In single file mode it is the format of the file, unknown here, so we return an empty string to
// let the global file format apply.
// In folder mode the merged flag set is encoded in the format of the files found by the last
// Retrieve ("json" for an empty flag set). Before the first Retrieve it is guessed from Pattern.
func (s *Retriever) OutputFormat() string {
	if s.mode == folderMode {
		return s.loader.OutputFormat()
	}
	return ""
}

func (s *Retriever) SetDownloader(downloader DownloaderAPI) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.downloader = downloader
}

// SetLister overrides the component used to list the objects of the bucket (folder mode).
func (s *Retriever) SetLister(lister ListerAPI) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lister = lister
}
