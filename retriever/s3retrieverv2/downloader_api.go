package s3retrieverv2

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var _ DownloaderAPI = (*manager.Downloader)(nil)

// DownloaderAPI provides methods to manage downloads to an S3 bucket.
type DownloaderAPI interface {
	Download(
		ctx context.Context,
		w io.WriterAt,
		input *s3.GetObjectInput,
		options ...func(*manager.Downloader),
	) (
		n int64, err error)
}
