package s3retrieverv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
)

var _ DownloaderAPI = (*transfermanager.Client)(nil)

// DownloaderAPI provides methods to manage downloads from an S3 bucket.
type DownloaderAPI interface {
	// DownloadObject downloads an object from S3.
	DownloadObject(
		ctx context.Context,
		input *transfermanager.DownloadObjectInput,
		opts ...func(*transfermanager.Options),
	) (*transfermanager.DownloadObjectOutput, error)
}
