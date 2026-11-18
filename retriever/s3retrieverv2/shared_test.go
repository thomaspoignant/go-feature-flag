package s3retrieverv2_test

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retrieverv2"
)

// fakeBucket is a fake S3 bucket counting the requests it receives.
type fakeBucket struct {
	mu      sync.Mutex
	objects map[string]fakeObject
	listErr error
	getErr  error
	// lastPrefix is the prefix of the last listing.
	lastPrefix string
	// pageSize forces the listing to be paginated when > 0.
	pageSize int

	lists atomic.Int32
	gets  map[string]int
}

type fakeObject struct {
	etag string
	data string
}

func newFakeBucket(objects map[string]fakeObject) *fakeBucket {
	return &fakeBucket{objects: objects, gets: map[string]int{}}
}

func (b *fakeBucket) put(key, etag, data string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.objects[key] = fakeObject{etag: etag, data: data}
}

func (b *fakeBucket) getCount(key string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.gets[key]
}

func (b *fakeBucket) totalGets() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	total := 0
	for _, n := range b.gets {
		total += n
	}
	return total
}

func (b *fakeBucket) ListObjectsV2(
	_ context.Context, in *s3.ListObjectsV2Input, _ ...func(*s3.Options),
) (*s3.ListObjectsV2Output, error) {
	b.lists.Add(1)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lastPrefix = aws.ToString(in.Prefix)
	if b.listErr != nil {
		return nil, b.listErr
	}
	var keys []string
	for k := range b.objects {
		if strings.HasPrefix(k, aws.ToString(in.Prefix)) && k > aws.ToString(in.ContinuationToken) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	out := &s3.ListObjectsV2Output{}
	if b.pageSize > 0 && len(keys) > b.pageSize {
		keys = keys[:b.pageSize]
		out.IsTruncated = aws.Bool(true)
		out.NextContinuationToken = aws.String(keys[len(keys)-1])
	}
	for _, k := range keys {
		out.Contents = append(out.Contents, s3types.Object{Key: aws.String(k), ETag: aws.String(b.objects[k].etag)})
	}
	return out, nil
}

func (b *fakeBucket) DownloadObject(
	_ context.Context, in *transfermanager.DownloadObjectInput, _ ...func(*transfermanager.Options),
) (*transfermanager.DownloadObjectOutput, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := aws.ToString(in.Key)
	b.gets[key]++
	if b.getErr != nil {
		return nil, b.getErr
	}
	obj, ok := b.objects[key]
	if !ok {
		return nil, errors.New("NoSuchKey")
	}
	if _, err := in.WriterAt.WriteAt([]byte(obj.data), 0); err != nil {
		return nil, err
	}
	return &transfermanager.DownloadObjectOutput{ETag: aws.String(obj.etag)}, nil
}

func newGroup(t *testing.T, b *fakeBucket) *s3retrieverv2.SharedItems {
	t.Helper()
	g, err := s3retrieverv2.NewSharedItems(t.Context(), s3retrieverv2.SharedItemsOptions{
		Bucket: "flags", Lister: b, Downloader: b,
	})
	require.NoError(t, err)
	return g
}

func initRetriever(t *testing.T, g *s3retrieverv2.SharedItems, item string, polling time.Duration) *s3retrieverv2.SharedItemRetriever {
	t.Helper()
	r := g.Retriever(item, polling)
	require.NoError(t, r.Init(t.Context(), nil))
	return r
}

func TestNewSharedItems(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		opts    s3retrieverv2.SharedItemsOptions
		wantErr string
	}{
		{name: "no bucket", opts: s3retrieverv2.SharedItemsOptions{}, wantErr: "bucket is empty"},
		{
			name: "explicit AWS config builds the clients",
			opts: s3retrieverv2.SharedItemsOptions{Bucket: "b", AwsConfig: &aws.Config{Region: "us-east-1"}},
		},
		{
			name: "fake clients",
			opts: s3retrieverv2.SharedItemsOptions{Bucket: "b", Lister: newFakeBucket(nil), Downloader: newFakeBucket(nil)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g, err := s3retrieverv2.NewSharedItems(t.Context(), tt.opts)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.opts.Bucket, g.Bucket())
		})
	}
}

func TestSharedItems_Bucket(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "flags", newGroup(t, newFakeBucket(nil)).Bucket())
}

func TestSharedItems_Retriever(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		item       string
		polling    time.Duration
		wantErrMsg string
	}{
		{name: "valid item is initializable", item: "a/flags.yaml", polling: time.Second},
		{name: "default polling interval", item: "a/flags.yaml"},
		{name: "empty item fails Init", item: "", wantErrMsg: "item is empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := newGroup(t, newFakeBucket(nil)).Retriever(tt.item, tt.polling)
			assert.Equal(t, retriever.RetrieverNotReady, r.Status())
			err := r.Init(t.Context(), nil)
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
				assert.Equal(t, retriever.RetrieverError, r.Status())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestSharedItemRetriever_Retrieve(t *testing.T) {
	t.Parallel()
	base := func() map[string]fakeObject {
		return map[string]fakeObject{
			"p1/dev/flags.yaml":  {etag: `"e1"`, data: "dev"},
			"p1/prod/flags.yaml": {etag: `W/"e2"`, data: "prod"},
			"p2/flags.yaml":      {etag: `"e3"`, data: "p2"},
		}
	}

	t.Run("not initialized", func(t *testing.T) {
		t.Parallel()
		r := newGroup(t, newFakeBucket(base())).Retriever("p1/dev/flags.yaml", 0)
		_, err := r.Retrieve(t.Context())
		assert.ErrorContains(t, err, "not initialized")
	})

	t.Run("N facades in one window do one list and no extra get", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			b := newFakeBucket(base())
			b.pageSize = 1
			g := newGroup(t, b)
			dev := initRetriever(t, g, "p1/dev/flags.yaml", 10*time.Second)
			prod := initRetriever(t, g, "p1/prod/flags.yaml", 10*time.Second)

			for _, tc := range []struct {
				r    *s3retrieverv2.SharedItemRetriever
				want string
			}{{dev, "dev"}, {prod, "prod"}} {
				got, err := tc.r.Retrieve(t.Context())
				require.NoError(t, err)
				assert.Equal(t, tc.want, string(got))
			}
			listsAfterFirst := b.lists.Load()
			assert.Equal(t, 2, b.totalGets())

			// same window, nothing changed: no List and no Get
			for range 3 {
				_, err := dev.Retrieve(t.Context())
				require.NoError(t, err)
				_, err = prod.Retrieve(t.Context())
				require.NoError(t, err)
			}
			assert.Equal(t, listsAfterFirst, b.lists.Load())
			assert.Equal(t, 2, b.totalGets())
		})
	})

	t.Run("TTL expiry lists again and only the changed item is downloaded", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			b := newFakeBucket(base())
			g := newGroup(t, b)
			// TTL = max(1s, 20s/10) = 2s
			dev := initRetriever(t, g, "p1/dev/flags.yaml", 20*time.Second)
			prod := initRetriever(t, g, "p1/prod/flags.yaml", time.Minute)
			_, err := dev.Retrieve(t.Context())
			require.NoError(t, err)
			_, err = prod.Retrieve(t.Context())
			require.NoError(t, err)
			require.EqualValues(t, 1, b.lists.Load())

			b.put("p1/dev/flags.yaml", `"e1b"`, "dev2")
			time.Sleep(time.Second)
			got, err := dev.Retrieve(t.Context())
			require.NoError(t, err)
			assert.Equal(t, "dev", string(got), "still inside the TTL window")
			assert.EqualValues(t, 1, b.lists.Load())

			time.Sleep(2 * time.Second)
			got, err = dev.Retrieve(t.Context())
			require.NoError(t, err)
			assert.Equal(t, "dev2", string(got))
			got, err = prod.Retrieve(t.Context())
			require.NoError(t, err)
			assert.Equal(t, "prod", string(got))
			assert.EqualValues(t, 2, b.lists.Load())
			assert.Equal(t, 2, b.getCount("p1/dev/flags.yaml"))
			assert.Equal(t, 1, b.getCount("p1/prod/flags.yaml"))
		})
	})

	t.Run("short polling interval uses the 1s minimum TTL", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			b := newFakeBucket(base())
			r := initRetriever(t, newGroup(t, b), "p2/flags.yaml", time.Second)
			_, err := r.Retrieve(t.Context())
			require.NoError(t, err)
			time.Sleep(1100 * time.Millisecond)
			_, err = r.Retrieve(t.Context())
			require.NoError(t, err)
			assert.EqualValues(t, 2, b.lists.Load())
			assert.Equal(t, 1, b.getCount("p2/flags.yaml"))
		})
	})

	t.Run("missing item", func(t *testing.T) {
		t.Parallel()
		r := initRetriever(t, newGroup(t, newFakeBucket(base())), "p1/nope.yaml", 0)
		_, err := r.Retrieve(t.Context())
		assert.ErrorContains(t, err, `"p1/nope.yaml" not found in S3 bucket "flags"`)
	})

	t.Run("list error is returned and retried", func(t *testing.T) {
		t.Parallel()
		b := newFakeBucket(base())
		b.listErr = errors.New("boom")
		r := initRetriever(t, newGroup(t, b), "p2/flags.yaml", 0)
		_, err := r.Retrieve(t.Context())
		assert.ErrorContains(t, err, "boom")
		b.mu.Lock()
		b.listErr = nil
		b.mu.Unlock()
		got, err := r.Retrieve(t.Context())
		require.NoError(t, err)
		assert.Equal(t, "p2", string(got))
	})

	t.Run("download error is returned", func(t *testing.T) {
		t.Parallel()
		b := newFakeBucket(base())
		b.getErr = errors.New("denied")
		r := initRetriever(t, newGroup(t, b), "p2/flags.yaml", 0)
		_, err := r.Retrieve(t.Context())
		assert.ErrorContains(t, err, "denied")
	})

	t.Run("item registered later widens the listing", func(t *testing.T) {
		t.Parallel()
		b := newFakeBucket(base())
		g := newGroup(t, b)
		dev := initRetriever(t, g, "p1/dev/flags.yaml", 0)
		_, err := dev.Retrieve(t.Context())
		require.NoError(t, err)
		other := initRetriever(t, g, "p2/flags.yaml", 0)
		got, err := other.Retrieve(t.Context())
		require.NoError(t, err)
		assert.Equal(t, "p2", string(got))
		assert.EqualValues(t, 2, b.lists.Load())
	})

	t.Run("concurrent retrieves coalesce the listing", func(t *testing.T) {
		t.Parallel()
		b := newFakeBucket(base())
		g := newGroup(t, b)
		r1 := initRetriever(t, g, "p1/dev/flags.yaml", 0)
		r2 := initRetriever(t, g, "p1/prod/flags.yaml", 0)
		var wg sync.WaitGroup
		for i := range 20 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				r := r1
				if i%2 == 0 {
					r = r2
				}
				_, err := r.Retrieve(t.Context())
				assert.NoError(t, err)
			}()
		}
		wg.Wait()
		assert.EqualValues(t, 1, b.lists.Load())
		assert.Equal(t, 1, b.getCount("p1/dev/flags.yaml"))
		assert.Equal(t, 1, b.getCount("p1/prod/flags.yaml"))
	})

	t.Run("returned bytes can be modified by the caller", func(t *testing.T) {
		t.Parallel()
		r := initRetriever(t, newGroup(t, newFakeBucket(base())), "p2/flags.yaml", 0)
		got, err := r.Retrieve(t.Context())
		require.NoError(t, err)
		got[0] = 'X'
		again, err := r.Retrieve(t.Context())
		require.NoError(t, err)
		assert.Equal(t, "p2", string(again))
	})
}

func TestSharedItemRetriever_Init(t *testing.T) {
	t.Parallel()
	g := newGroup(t, newFakeBucket(nil))
	r := g.Retriever("a.yaml", 0)
	require.NoError(t, r.Init(t.Context(), nil))
	assert.Equal(t, retriever.RetrieverReady, r.Status())
}

func TestSharedItemRetriever_Shutdown(t *testing.T) {
	t.Parallel()
	t.Run("shut down item is deregistered", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			b := newFakeBucket(map[string]fakeObject{
				"a/1.yaml": {etag: "1", data: "1"},
				"b/2.yaml": {etag: "2", data: "2"},
			})
			g := newGroup(t, b)
			r1 := initRetriever(t, g, "a/1.yaml", time.Second)
			r2 := initRetriever(t, g, "b/2.yaml", time.Second)
			_, err := r1.Retrieve(t.Context())
			require.NoError(t, err)
			assert.Empty(t, b.lastPrefix, "listing covers both items")

			require.NoError(t, r2.Shutdown(t.Context()))
			assert.Equal(t, retriever.RetrieverNotReady, r2.Status())
			_, err = r2.Retrieve(t.Context())
			assert.Error(t, err)

			time.Sleep(2 * time.Second)
			_, err = r1.Retrieve(t.Context())
			require.NoError(t, err)
			assert.Equal(t, "a/1.yaml", b.lastPrefix, "listing only covers the remaining item")
		})
	})

	t.Run("re-Init joins the group again", func(t *testing.T) {
		t.Parallel()
		b := newFakeBucket(map[string]fakeObject{"a/1.yaml": {etag: "1", data: "1"}})
		r := initRetriever(t, newGroup(t, b), "a/1.yaml", 0)
		require.NoError(t, r.Shutdown(t.Context()))
		require.NoError(t, r.Init(t.Context(), nil))
		got, err := r.Retrieve(t.Context())
		require.NoError(t, err)
		assert.Equal(t, "1", string(got))
	})
}

func TestSharedItemRetriever_Status(t *testing.T) {
	t.Parallel()
	r := newGroup(t, newFakeBucket(nil)).Retriever("a.yaml", 0)
	assert.Equal(t, retriever.RetrieverNotReady, r.Status())
	require.NoError(t, r.Init(t.Context(), nil))
	assert.Equal(t, retriever.RetrieverReady, r.Status())
	require.NoError(t, r.Shutdown(t.Context()))
	assert.Equal(t, retriever.RetrieverNotReady, r.Status())
}

func TestSharedItemRetriever_OutputFormat(t *testing.T) {
	t.Parallel()
	r := newGroup(t, newFakeBucket(nil)).Retriever("a.yaml", 0)
	assert.Empty(t, r.OutputFormat())
}
