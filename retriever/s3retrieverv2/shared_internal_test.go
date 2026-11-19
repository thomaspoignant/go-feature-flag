package s3retrieverv2

// Whitebox: reaches the defensive paths that the public API can't trigger. Not parallel because
// of t.Setenv.

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
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
	g.Retriever("a/registered.yaml", 0)
	// the key is not registered and out of the listed prefix, so the loop gives up.
	_, err = g.etagOf(t.Context(), "b/unregistered.yaml")
	assert.ErrorContains(t, err, "not registered")
}

func TestSharedItems_listingPrefix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		keys []string
		want string
	}{
		{name: "no key", want: ""},
		{name: "single key", keys: []string{"a/b/flags.yaml"}, want: "a/b/"},
		{name: "same prefix", keys: []string{"a/b/1.yaml", "a/b/2.yaml"}, want: "a/b/"},
		{name: "cut at the last slash", keys: []string{"a/b/flags1.yaml", "a/b/flags2.yaml"}, want: "a/b/"},
		{name: "sibling prefixes", keys: []string{"project-1/dev/f.yaml", "project-1/prod/f.yaml"}, want: "project-1/"},
		{name: "partial segment is not a prefix", keys: []string{"proj-a/f.yaml", "proj-b/f.yaml"}, want: ""},
		{name: "no common prefix", keys: []string{"a/f.yaml", "b/f.yaml"}, want: ""},
		{name: "root level item", keys: []string{"flags.yaml"}, want: ""},
		{name: "opaque dot key", keys: []string{"./a", "./b"}, want: "./"},
		{name: "opaque leading slash", keys: []string{"/a", "/b"}, want: "/"},
		{name: "opaque parent segment", keys: []string{"dir/../a", "dir/../b"}, want: "dir/../"},
		{name: "never cuts a multi-byte character", keys: []string{"é/é1.yaml", "é/é2.yaml", "è/f.yaml"}, want: ""},
		{name: "multi-byte prefix", keys: []string{"日本/a.yaml", "日本/b.yaml"}, want: "日本/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, listingPrefix(tt.keys))
		})
	}
}

// unrelatedBucket lists many unrelated objects besides the registered one.
type unrelatedBucket struct {
	emptyBucket
	n int
}

func (u unrelatedBucket) ListObjectsV2(
	context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options),
) (*s3.ListObjectsV2Output, error) {
	out := &s3.ListObjectsV2Output{
		Contents: []s3types.Object{{Key: aws.String("d/flags.yaml"), ETag: aws.String(`"x"`)}},
	}
	for i := range u.n {
		out.Contents = append(out.Contents, s3types.Object{Key: aws.String(fmt.Sprintf("d/other-%d", i))})
	}
	return out, nil
}

func TestSharedItems_listingKeepsRegisteredKeysOnly(t *testing.T) {
	t.Parallel()
	g, err := NewSharedItems(t.Context(), SharedItemsOptions{
		Bucket: "b", Lister: unrelatedBucket{n: 5000}, Downloader: emptyBucket{},
	})
	require.NoError(t, err)
	_, err = g.etagOf(t.Context(), g.Retriever("d/flags.yaml", 0).item)
	require.NoError(t, err)
	g.mu.Lock()
	defer g.mu.Unlock()
	assert.Equal(t, map[string]string{"d/flags.yaml": "x"}, g.listing)
}

// tokenLister returns truncated pages following a token policy.
type tokenLister struct {
	calls atomic.Int64
	// next gives the NextContinuationToken of the page number n.
	next func(n int64) string
}

func (l *tokenLister) ListObjectsV2(
	context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options),
) (*s3.ListObjectsV2Output, error) {
	n := l.calls.Add(1)
	return &s3.ListObjectsV2Output{
		IsTruncated:           aws.Bool(true),
		NextContinuationToken: aws.String(l.next(n)),
	}, nil
}

func TestSharedItems_list(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		next      func(n int64) string
		wantErr   string
		wantCalls int64
	}{
		{
			name:      "a repeated token ends the listing with an error",
			next:      func(int64) string { return "same" },
			wantErr:   "incomplete",
			wantCalls: 2,
		},
		{
			name:      "a truncated page without token is an error",
			next:      func(int64) string { return "" },
			wantErr:   "incomplete",
			wantCalls: 1,
		},
		{
			name:      "the number of pages is capped",
			next:      func(n int64) string { return strconv.FormatInt(n, 10) },
			wantErr:   "more than",
			wantCalls: maxSharedListPages,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := &tokenLister{next: tt.next}
			g := &SharedItems{bucket: "b", lister: l}
			_, err := g.list(t.Context(), "p/", nil)
			assert.ErrorContains(t, err, tt.wantErr)
			assert.Equal(t, tt.wantCalls, l.calls.Load())
		})
	}
}
