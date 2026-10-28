package s3retrieverv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/thomaspoignant/go-feature-flag/internal/folderloader"
)

var _ folderloader.Store = (*s3Store)(nil)

// s3Store gives the folder loader access to the bucket of a Retriever. It reads the lister and
// the downloader of the Retriever at each call, so the caller must hold the Retriever lock.
type s3Store struct {
	r *Retriever
}

// List returns the objects under prefix with their ETag as version. The pages are walked, and
// the delimiter "/" is used when not recursive so that only the direct children are returned.
func (s *s3Store) List(ctx context.Context, prefix string, recursive bool) ([]folderloader.Entry, error) {
	input := &s3.ListObjectsV2Input{Bucket: aws.String(s.r.Bucket)}
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}
	if !recursive {
		input.Delimiter = aws.String("/")
	}
	var res []folderloader.Entry
	paginator := s3.NewListObjectsV2Paginator(s.r.lister, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			res = append(res, folderloader.Entry{Key: aws.ToString(obj.Key), Version: aws.ToString(obj.ETag)})
		}
	}
	return res, nil
}

// Get downloads an object and returns its content with the ETag of the downloaded version.
func (s *s3Store) Get(ctx context.Context, key string) ([]byte, string, error) {
	return s.r.downloadWithETag(ctx, key)
}
