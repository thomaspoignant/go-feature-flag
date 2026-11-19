package s3retrieverv2

// Whitebox: reaches the defensive paths that the public API can't trigger. Not parallel because
// of t.Setenv.

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSharedItems_DefaultConfigError(t *testing.T) {
	t.Setenv("AWS_PROFILE", "this-profile-does-not-exist")
	t.Setenv("AWS_CONFIG_FILE", "/nonexistent/config")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", "/nonexistent/credentials")
	//nolint:staticcheck // a nil context is tolerated on purpose
	_, err := NewSharedItems(nil, SharedItemsOptions{Bucket: "b"})
	assert.ErrorContains(t, err, "impossible to init shared S3 items")
}

// emptyBucket is a bucket without any object.
type emptyBucket struct{}

func (emptyBucket) ListObjectsV2(
	context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options),
) (*s3.ListObjectsV2Output, error) {
	return &s3.ListObjectsV2Output{}, nil
}

func (emptyBucket) DownloadObject(
	context.Context, *transfermanager.DownloadObjectInput, ...func(*transfermanager.Options),
) (*transfermanager.DownloadObjectOutput, error) {
	return &transfermanager.DownloadObjectOutput{}, nil
}

func TestSharedItems_etagOf(t *testing.T) {
	t.Parallel()
	g, err := NewSharedItems(t.Context(), SharedItemsOptions{
		Bucket: "b", Lister: emptyBucket{}, Downloader: emptyBucket{},
	})
	require.NoError(t, err)
	// the key is not registered: its directory is never listed, so the loop gives up.
	_, _, err = g.etagOf(t.Context(), "a/unregistered.yaml")
	assert.ErrorContains(t, err, "not registered")
}
