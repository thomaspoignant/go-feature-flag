package s3exporterv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
)

var _ UploaderAPI = (*transfermanager.Client)(nil)

// UploaderAPI provides methods to manage uploads to an S3 bucket.
type UploaderAPI interface {
	// UploadObject uploads an object to S3.
	UploadObject(
		ctx context.Context,
		input *transfermanager.UploadObjectInput,
		opts ...func(*transfermanager.Options),
	) (*transfermanager.UploadObjectOutput, error)
}
