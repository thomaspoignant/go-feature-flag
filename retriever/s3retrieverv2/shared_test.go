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
	// hang makes the listing wait for its context.
	hang bool
	// noDownloadETag makes the downloads return no ETag.
	noDownloadETag bool
	// prefixes are the prefixes of the listings, in order.
	prefixes []string
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
	ctx context.Context, in *s3.ListObjectsV2Input, _ ...func(*s3.Options),
) (*s3.ListObjectsV2Output, error) {
	b.lists.Add(1)
	b.mu.Lock()
	hang := b.hang
	b.mu.Unlock()
	if hang {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.prefixes = append(b.prefixes, aws.ToString(in.Prefix))
	if b.listErr != nil {
		return nil, b.listErr
	}
	var keys []string
	for k := range b.objects {
		rest := strings.TrimPrefix(k, aws.ToString(in.Prefix))
		if strings.HasPrefix(k, aws.ToString(in.Prefix)) && k > aws.ToString(in.ContinuationToken) &&
			(in.Delimiter == nil || !strings.Contains(rest, aws.ToString(in.Delimiter))) {
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
	if b.noDownloadETag {
		return &transfermanager.DownloadObjectOutput{}, nil
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
		{name: "default AWS config", opts: s3retrieverv2.SharedItemsOptions{Bucket: "b"}},
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
	// every case registers the items in order, then runs. The retrievers are initialized.
	tests := []struct {
		name  string
		items map[string]time.Duration // item -> polling interval
		setup func(b *fakeBucket)
		run   func(t *testing.T, b *fakeBucket, r map[string]*s3retrieverv2.SharedItemRetriever)
	}{
		{
			name:  "N facades in one window do one list per directory and no extra get",
			items: map[string]time.Duration{"p1/dev/flags.yaml": 10 * time.Second, "p1/prod/flags.yaml": 10 * time.Second},
			setup: func(b *fakeBucket) { b.pageSize = 1 },
			run:   runSameWindow,
		},
		{
			name: "only the directories of the items are listed, not recursively",
			items: map[string]time.Duration{
				"p1/dev/flags.yaml": time.Minute, "p1/prod/flags.yaml": time.Minute, "p2/flags.yaml": time.Minute,
			},
			run: runDirectories,
		},
		{
			name:  "TTL expiry lists again and only the changed item is downloaded",
			items: map[string]time.Duration{"p1/dev/flags.yaml": 20 * time.Second, "p1/prod/flags.yaml": time.Minute},
			run:   runTtlExpiry,
		},
		{
			name:  "short polling interval uses the 1s minimum TTL",
			items: map[string]time.Duration{"p2/flags.yaml": time.Second},
			run:   runMinTTL,
		},
		{
			name:  "missing item",
			items: map[string]time.Duration{"p1/nope.yaml": 0},
			run:   runMissing,
		},
		{
			name:  "list error is returned once per window for all the callers",
			items: map[string]time.Duration{"p1/dev/flags.yaml": 20 * time.Second, "p2/flags.yaml": 20 * time.Second},
			setup: func(b *fakeBucket) { b.listErr = errors.New("AccessDenied") },
			run:   runListError,
		},
		{
			name:  "download error is returned",
			items: map[string]time.Duration{"p2/flags.yaml": 0},
			setup: func(b *fakeBucket) { b.getErr = errors.New("denied") },
			run:   runDownloadError,
		},
		{
			name:  "download without ETag is cached with the listed ETag",
			items: map[string]time.Duration{"p2/flags.yaml": time.Second},
			setup: func(b *fakeBucket) { b.noDownloadETag = true },
			run:   runNoDownloadETag,
		},
		{
			name:  "concurrent retrieves coalesce the listing",
			items: map[string]time.Duration{"p1/dev/flags.yaml": 0, "p1/prod/flags.yaml": 0},
			run:   runConcurrent,
		},
		{
			name:  "returned bytes can be modified by the caller",
			items: map[string]time.Duration{"p2/flags.yaml": 0},
			run:   runClone,
		},
		{
			name:  "item added after the first listing widens the listing",
			items: map[string]time.Duration{"p1/dev/flags.yaml": 0},
			run:   runWiden,
		},
		{
			name:  "empty file",
			items: map[string]time.Duration{"empty/flags.yaml": 0},
			run:   runEmptyFile,
		},
		{
			name:  "hung listing: a cancelled caller does not block the others and the error is cached",
			items: map[string]time.Duration{"p1/dev/flags.yaml": time.Minute, "p2/flags.yaml": time.Minute},
			setup: func(b *fakeBucket) { b.hang = true },
			run:   runHung,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) { runSharedCase(t, tt.items, tt.setup, tt.run) })
		})
	}

	t.Run("not initialized", func(t *testing.T) {
		t.Parallel()
		r := newGroup(t, newFakeBucket(baseObjects())).Retriever("p1/dev/flags.yaml", 0)
		_, err := r.Retrieve(t.Context())
		assert.ErrorContains(t, err, "not initialized")
		//nolint:staticcheck // a nil context is tolerated on purpose
		_, err = r.Retrieve(nil)
		assert.ErrorContains(t, err, "not initialized")
	})
}

// runSharedCase builds a bucket and a group with the initialized retrievers of items, then runs the case.
func runSharedCase(
	t *testing.T, items map[string]time.Duration, setup func(*fakeBucket),
	run func(*testing.T, *fakeBucket, map[string]*s3retrieverv2.SharedItemRetriever),
) {
	t.Helper()
	b := newFakeBucket(baseObjects())
	if setup != nil {
		setup(b)
	}
	g := newGroup(t, b)
	r := map[string]*s3retrieverv2.SharedItemRetriever{}
	for item, polling := range items {
		r[item] = initRetriever(t, g, item, polling)
	}
	run(t, b, r)
}

func baseObjects() map[string]fakeObject {
	return map[string]fakeObject{
		"p1/dev/flags.yaml":  {etag: `"e1"`, data: "dev"},
		"p1/prod/flags.yaml": {etag: `W/"e2"`, data: "prod"},
		"p2/flags.yaml":      {etag: `"e3"`, data: "p2"},
		"other/flags.yaml":   {etag: `"e4"`, data: "other"},
		"p1/dev/sub/x.yaml":  {etag: `"e5"`, data: "sub"},
		"empty/flags.yaml":   {etag: `"e6"`, data: ""},
	}
}

func runSameWindow(t *testing.T, b *fakeBucket, r map[string]*s3retrieverv2.SharedItemRetriever) {
	for key, want := range map[string]string{"p1/dev/flags.yaml": "dev", "p1/prod/flags.yaml": "prod"} {
		got, err := r[key].Retrieve(t.Context())
		require.NoError(t, err)
		assert.Equal(t, want, string(got))
	}
	lists := b.lists.Load()
	assert.Equal(t, 2, b.totalGets())
	for range 3 {
		for _, rr := range r {
			_, err := rr.Retrieve(t.Context())
			require.NoError(t, err)
		}
	}
	assert.Equal(t, lists, b.lists.Load())
	assert.Equal(t, 2, b.totalGets())
}

func runDirectories(t *testing.T, b *fakeBucket, r map[string]*s3retrieverv2.SharedItemRetriever) {
	for _, rr := range r {
		_, err := rr.Retrieve(t.Context())
		require.NoError(t, err)
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	assert.ElementsMatch(t, []string{"p1/dev/", "p1/prod/", "p2/"}, b.prefixes)
}

func runTtlExpiry(t *testing.T, b *fakeBucket, r map[string]*s3retrieverv2.SharedItemRetriever) {
	dev, prod := r["p1/dev/flags.yaml"], r["p1/prod/flags.yaml"]
	_, err := dev.Retrieve(t.Context())
	require.NoError(t, err)
	_, err = prod.Retrieve(t.Context())
	require.NoError(t, err)
	listsAfterFirst := b.lists.Load()

	b.put("p1/dev/flags.yaml", `"e1b"`, "dev2")
	// TTL = 20s/2 = 10s
	time.Sleep(5 * time.Second)
	got, err := dev.Retrieve(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "dev", string(got), "still inside the TTL window")
	assert.Equal(t, listsAfterFirst, b.lists.Load())

	time.Sleep(6 * time.Second)
	got, err = dev.Retrieve(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "dev2", string(got))
	got, err = prod.Retrieve(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "prod", string(got))
	assert.Greater(t, b.lists.Load(), listsAfterFirst)
	assert.Equal(t, 2, b.getCount("p1/dev/flags.yaml"))
	assert.Equal(t, 1, b.getCount("p1/prod/flags.yaml"))
}

func runMinTTL(t *testing.T, b *fakeBucket, r map[string]*s3retrieverv2.SharedItemRetriever) {
	_, err := r["p2/flags.yaml"].Retrieve(t.Context())
	require.NoError(t, err)
	time.Sleep(1100 * time.Millisecond)
	_, err = r["p2/flags.yaml"].Retrieve(t.Context())
	require.NoError(t, err)
	assert.EqualValues(t, 2, b.lists.Load())
	assert.Equal(t, 1, b.getCount("p2/flags.yaml"))
}

func runMissing(t *testing.T, _ *fakeBucket, r map[string]*s3retrieverv2.SharedItemRetriever) {
	_, err := r["p1/nope.yaml"].Retrieve(t.Context())
	assert.ErrorContains(t, err, `"p1/nope.yaml" not found in S3 bucket "flags"`)
}

func runListError(t *testing.T, b *fakeBucket, r map[string]*s3retrieverv2.SharedItemRetriever) {
	for range 3 {
		for _, rr := range r {
			_, err := rr.Retrieve(t.Context())
			assert.ErrorContains(t, err, "AccessDenied")
		}
	}
	// one failing list for the whole window, whatever the number of callers
	assert.EqualValues(t, 1, b.lists.Load())

	b.mu.Lock()
	b.listErr = nil
	b.mu.Unlock()
	time.Sleep(11 * time.Second)
	got, err := r["p2/flags.yaml"].Retrieve(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "p2", string(got))
}

func runDownloadError(t *testing.T, _ *fakeBucket, r map[string]*s3retrieverv2.SharedItemRetriever) {
	_, err := r["p2/flags.yaml"].Retrieve(t.Context())
	assert.ErrorContains(t, err, "unable to download item from S3 \"p2/flags.yaml\", denied")
}

func runNoDownloadETag(t *testing.T, b *fakeBucket, r map[string]*s3retrieverv2.SharedItemRetriever) {
	for range 3 {
		got, err := r["p2/flags.yaml"].Retrieve(t.Context())
		require.NoError(t, err)
		assert.Equal(t, "p2", string(got))
		time.Sleep(2 * time.Second)
	}
	assert.Equal(t, 1, b.getCount("p2/flags.yaml"))
}

func runConcurrent(t *testing.T, b *fakeBucket, r map[string]*s3retrieverv2.SharedItemRetriever) {
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := "p1/dev/flags.yaml"
			if i%2 == 0 {
				key = "p1/prod/flags.yaml"
			}
			_, err := r[key].Retrieve(t.Context())
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
	assert.EqualValues(t, 2, b.lists.Load(), "one listing per directory")
	assert.Equal(t, 1, b.getCount("p1/dev/flags.yaml"))
	assert.Equal(t, 1, b.getCount("p1/prod/flags.yaml"))
}

func runClone(t *testing.T, _ *fakeBucket, r map[string]*s3retrieverv2.SharedItemRetriever) {
	got, err := r["p2/flags.yaml"].Retrieve(t.Context())
	require.NoError(t, err)
	got[0] = 'X'
	again, err := r["p2/flags.yaml"].Retrieve(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "p2", string(again))
}

func runHung(t *testing.T, b *fakeBucket, r map[string]*s3retrieverv2.SharedItemRetriever) {
	var wg sync.WaitGroup
	var otherErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, otherErr = r["p2/flags.yaml"].Retrieve(t.Context())
	}()
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	start := time.Now()
	_, err := r["p1/dev/flags.yaml"].Retrieve(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, time.Second, time.Since(start))

	wg.Wait()
	assert.ErrorContains(t, otherErr, "unable to list S3 bucket")
	assert.GreaterOrEqual(t, time.Since(start), 30*time.Second)

	// the failure is shared: no new listing inside the window
	lists := b.lists.Load()
	_, err = r["p1/dev/flags.yaml"].Retrieve(t.Context())
	assert.Error(t, err)
	assert.Equal(t, lists, b.lists.Load())
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
			assert.ElementsMatch(t, []string{"a/", "b/"}, b.prefixes, "listing covers both items")

			require.NoError(t, r2.Shutdown(t.Context()))
			assert.Equal(t, retriever.RetrieverNotReady, r2.Status())
			_, err = r2.Retrieve(t.Context())
			assert.Error(t, err)

			time.Sleep(2 * time.Second)
			_, err = r1.Retrieve(t.Context())
			require.NoError(t, err)
			assert.ElementsMatch(t, []string{"a/", "b/", "a/"}, b.prefixes, "listing only covers the remaining item")
		})
	})

	t.Run("shutting down an item that is not registered is harmless", func(t *testing.T) {
		t.Parallel()
		r := newGroup(t, newFakeBucket(nil)).Retriever("", 0)
		assert.NoError(t, r.Shutdown(t.Context()))
		assert.NoError(t, r.Shutdown(t.Context()))
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

func runWiden(t *testing.T, b *fakeBucket, _ map[string]*s3retrieverv2.SharedItemRetriever) {
	g := newGroup(t, b)
	dev := initRetriever(t, g, "p1/dev/flags.yaml", 0)
	_, err := dev.Retrieve(t.Context())
	require.NoError(t, err)
	late := initRetriever(t, g, "p2/flags.yaml", 0)
	got, err := late.Retrieve(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "p2", string(got))
	assert.Equal(t, []string{"p1/dev/", "p1/dev/", "p2/"}, b.prefixes)
}

func runEmptyFile(t *testing.T, b *fakeBucket, r map[string]*s3retrieverv2.SharedItemRetriever) {
	for range 2 {
		got, err := r["empty/flags.yaml"].Retrieve(t.Context())
		require.NoError(t, err)
		assert.Empty(t, got)
	}
	assert.Equal(t, 1, b.getCount("empty/flags.yaml"))
}
