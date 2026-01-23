package s3exporterv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var _ UploaderAPI = (*manager.Uploader)(nil)

// UploaderAPI provides methods to manage uploads to an S3 bucket.
type UploaderAPI interface {
	// Upload provides a method to upload objects to S3.
	Upload(
		ctx context.Context,
		input *s3.PutObjectInput,
		opts ...func(uploader *manager.Uploader),
	) (
		*manager.UploadOutput, error)
}
