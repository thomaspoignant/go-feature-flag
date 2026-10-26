package s3retrieverv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var _ ListerAPI = (*s3.Client)(nil)

// ListerAPI provides methods to list the objects of an S3 bucket.
type ListerAPI interface {
	// ListObjectsV2 returns some or all (up to 1,000) of the objects in a bucket.
	ListObjectsV2(
		ctx context.Context,
		params *s3.ListObjectsV2Input,
		optFns ...func(*s3.Options),
	) (*s3.ListObjectsV2Output, error)
}
