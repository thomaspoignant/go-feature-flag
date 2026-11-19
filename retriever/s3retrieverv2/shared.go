package s3retrieverv2

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

const (
	// DefaultSharedPollingInterval is the polling interval assumed for an item registered without one.
	DefaultSharedPollingInterval = 60 * time.Second
	// minSharedListingTTL is the lowest listing TTL.
	minSharedListingTTL = time.Second
	// maxSharedErrorTTL is the longest time a listing error is served without listing again.
	maxSharedErrorTTL = 10 * time.Second
	// sharedRequestTimeout bounds one S3 request (a listing page or a download), which is not
	// tied to the context of a caller.
	sharedRequestTimeout = 30 * time.Second
	// MaxConcurrentDownloads bounds the GetObject requests a group sends at the same time.
	MaxConcurrentDownloads = 16
	// maxSharedListPages bounds the number of pages of one listing.
	maxSharedListPages = 10000
)

// SharedItemsOptions configures a SharedItems group.
type SharedItemsOptions struct {
	// Bucket is the name of the S3 bucket shared by all the items.
	Bucket string
	// AwsConfig is the AWS SDK configuration, the default one is loaded if nil.
	AwsConfig *aws.Config
	// S3ClientOptions are functional options of the S3 client.
	S3ClientOptions []func(*s3.Options)
	// Lister overrides the S3 lister (for tests), built from AwsConfig if nil.
	Lister ListerAPI
	// Downloader overrides the S3 downloader (for tests), built from AwsConfig if nil.
	Downloader DownloaderAPI
}

// SharedItems lets several single-file S3 retrievers of the same bucket share their storage
// requests: the objects under the longest common key prefix of the registered items (cut at its
// last "/") are listed
// once per listing TTL window (recursively, all the pages) for all of them, and an object is
// only downloaded when its ETag differs from the cached one. The listing covers every object
// under that prefix, including the unrelated ones, so the flag files should live under a
// dedicated prefix.
//
// The listing TTL is max(1s, minPollingInterval/2), where minPollingInterval is the lowest
// polling interval of the registered items. A change in the bucket can therefore be detected up
// to one TTL later than with an independent retriever. Listing errors are cached for
// min(TTL, 10s) so that an outage causes one failing listing per window, not one per item.
type SharedItems struct {
	bucket     string
	lister     ListerAPI
	downloader DownloaderAPI

	// mu protects everything below.
	mu       sync.Mutex
	items    map[*SharedItemRetriever]struct{}
	keyRefs  map[string]int
	seq      uint64            // incremented each time a new key is registered
	keySeq   map[string]uint64 // key -> seq of its registration
	listSeq  uint64            // seq at the start of the current listing
	prefix   string            // listing prefix of the registered items
	ttl      time.Duration     // listing TTL of the registered items
	contents map[string]*sharedContent
	// base is the context of the requests of the group, it is cancelled when the last item leaves.
	base   context.Context
	cancel context.CancelFunc
	// sem bounds the concurrent downloads.
	sem chan struct{}

	listing      map[string]string // registered key -> normalized ETag
	listedPrefix string
	listErr      error
	listedAt     time.Time // end of the listing
	listStart    time.Time // start of the listing
	listed       bool
	inflight     *sharedListing
}

// sharedListing is a listing in progress, shared by all the callers waiting for it.
type sharedListing struct {
	done    chan struct{}
	start   time.Time
	prefix  string
	seq     uint64
	listing map[string]string
	err     error
}

// sharedContent is the cached content of one object.
type sharedContent struct {
	mu   sync.Mutex
	etag string
	data []byte
	// fetchedAt is the start of the download of data.
	fetchedAt time.Time
	dl        *sharedDownload // download in progress
}

// sharedDownload is a download in progress, shared by all the callers waiting for it.
type sharedDownload struct {
	done chan struct{}
	data []byte
	err  error
}

// NewSharedItems builds the group of a bucket. The S3 clients are built like Retriever.Init does.
func NewSharedItems(ctx context.Context, opts SharedItemsOptions) (*SharedItems, error) {
	if opts.Bucket == "" {
		return nil, errors.New("impossible to init shared S3 items: bucket is empty")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	lister, downloader := opts.Lister, opts.Downloader
	if lister == nil || downloader == nil {
		cfg := opts.AwsConfig
		if cfg == nil {
			loaded, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				return nil, fmt.Errorf("impossible to init shared S3 items: %w", err)
			}
			cfg = &loaded
		}
		client := s3.NewFromConfig(*cfg, opts.S3ClientOptions...)
		if lister == nil {
			lister = client
		}
		if downloader == nil {
			downloader = transfermanager.New(client)
		}
	}
	return &SharedItems{
		bucket:     opts.Bucket,
		lister:     lister,
		downloader: downloader,
		items:      map[*SharedItemRetriever]struct{}{},
		keyRefs:    map[string]int{},
		keySeq:     map[string]uint64{},
		contents:   map[string]*sharedContent{},
		ttl:        minSharedListingTTL,
		sem:        make(chan struct{}, MaxConcurrentDownloads),
	}, nil
}

// baseContext returns the context of the requests of the group, renewed if it was cancelled
// because every item left.
func (g *SharedItems) baseContext() context.Context {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.baseLocked()
}

func (g *SharedItems) baseLocked() context.Context {
	if g.base == nil || g.base.Err() != nil {
		g.base, g.cancel = context.WithCancel(context.Background())
	}
	return g.base
}

// Bucket returns the bucket of the group.
func (g *SharedItems) Bucket() string { return g.bucket }

// Retriever returns the retriever of one item of the bucket. The item joins the group
// immediately (so that the listing covers it from the first poll) and leaves it on Shutdown.
// pollingInterval is used to compute the listing TTL, 0 means the default one (60s).
func (g *SharedItems) Retriever(item string, pollingInterval time.Duration) *SharedItemRetriever {
	if pollingInterval <= 0 {
		pollingInterval = DefaultSharedPollingInterval
	}
	r := &SharedItemRetriever{group: g, item: item, pollingInterval: pollingInterval}
	g.register(r)
	return r
}

func (g *SharedItems) register(r *SharedItemRetriever) {
	if r.item == "" {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.items[r]; ok {
		return
	}
	g.items[r] = struct{}{}
	if g.keyRefs[r.item]++; g.keyRefs[r.item] == 1 {
		// a listing started before this registration may not know the key: force a refresh
		g.seq++
		g.keySeq[r.item] = g.seq
	}
	g.recompute()
}

func (g *SharedItems) deregister(r *SharedItemRetriever) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.items[r]; !ok {
		return
	}
	delete(g.items, r)
	if g.keyRefs[r.item]--; g.keyRefs[r.item] <= 0 {
		delete(g.keyRefs, r.item)
		delete(g.contents, r.item)
		delete(g.keySeq, r.item)
	}
	if len(g.items) == 0 && g.cancel != nil {
		// nobody needs the requests in flight anymore
		g.cancel()
	}
	g.recompute()
}

// recompute refreshes the listing prefix and the TTL. The caller must hold g.mu.
func (g *SharedItems) recompute() {
	minPolling := time.Duration(0)
	keys := make([]string, 0, len(g.items))
	for it := range g.items {
		keys = append(keys, it.item)
		if minPolling == 0 || it.pollingInterval < minPolling {
			minPolling = it.pollingInterval
		}
	}
	g.prefix = listingPrefix(keys)
	g.ttl = max(minSharedListingTTL, minPolling/2)
}

// listingPrefix is the longest common prefix of the keys, cut after its last "/" ("" if there is
// none). S3 keys are opaque so they are not normalized in any way, and cutting at a "/" never
// splits a multi-byte character.
func listingPrefix(keys []string) string {
	if len(keys) == 0 {
		return ""
	}
	prefix := keys[0]
	for _, k := range keys[1:] {
		for !strings.HasPrefix(k, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix[:strings.LastIndex(prefix, "/")+1]
}

// normalizeETag ignores the weak prefix and the surrounding double quotes.
func normalizeETag(v string) string {
	return strings.Trim(strings.TrimPrefix(v, "W/"), `"`)
}

// coversLocked tells if a listing of prefix started at seq knows the key: the key is under the
// prefix and was registered before the listing started. The caller must hold g.mu.
func (g *SharedItems) coversLocked(key, prefix string, seq uint64) bool {
	return strings.HasPrefix(key, prefix) && seq >= g.keySeq[key]
}

// freshLocked tells if the current listing can serve the key. The caller must hold g.mu.
func (g *SharedItems) freshLocked(key string) bool {
	if !g.listed {
		return false
	}
	if !g.coversLocked(key, g.listedPrefix, g.listSeq) {
		return false
	}
	ttl := g.ttl
	if g.listErr != nil {
		ttl = min(ttl, maxSharedErrorTTL)
	}
	return time.Since(g.listedAt) < ttl
}

// listedItem is the result of a lookup in the listing.
type listedItem struct {
	etag  string
	found bool
	// start is the start of the listing that returned it.
	start time.Time
}

// etagOf returns the ETag of the key in a listing no older than the TTL, listing if needed.
// Concurrent callers share one listing, and a caller stops waiting when its ctx is done.
func (g *SharedItems) etagOf(ctx context.Context, key string) (listedItem, error) {
	for range 3 {
		g.mu.Lock()
		if g.freshLocked(key) {
			etag, ok := g.listing[key]
			item, err := listedItem{etag: etag, found: ok, start: g.listStart}, g.listErr
			g.mu.Unlock()
			return item, err
		}
		fl := g.inflight
		if fl == nil {
			fl = &sharedListing{done: make(chan struct{})}
			g.inflight = fl
			wanted := make(map[string]struct{}, len(g.keyRefs))
			for k := range g.keyRefs {
				wanted[k] = struct{}{}
			}
			go g.runListing(g.baseLocked(), fl, g.prefix, wanted, g.seq)
		}
		g.mu.Unlock()

		select {
		case <-fl.done:
		case <-ctx.Done():
			return listedItem{}, ctx.Err()
		}
		// the waiters use the result of the listing they waited for, even if it is already older
		// than the TTL (a slow listing), as long as it covers the key.
		g.mu.Lock()
		covered := g.coversLocked(key, fl.prefix, fl.seq)
		g.mu.Unlock()
		if covered {
			etag, ok := fl.listing[key]
			return listedItem{etag: etag, found: ok, start: fl.start}, fl.err
		}
	}
	return listedItem{}, errors.New("item is not registered in the shared listing")
}

// runListing lists the objects under prefix and publishes the result. It does not depend on the
// context of a caller so that a cancelled caller does not fail the others.
func (g *SharedItems) runListing(
	base context.Context, fl *sharedListing, prefix string, wanted map[string]struct{}, seq uint64,
) {
	start := time.Now()
	listing, err := g.list(base, prefix, wanted)

	g.mu.Lock()
	defer g.mu.Unlock()
	fl.start, fl.prefix, fl.seq, fl.listing, fl.err = start, prefix, seq, listing, err
	// a listing cancelled because every item left is not kept
	if base.Err() == nil {
		g.listing, g.listErr, g.listed, g.listSeq = listing, err, true, seq
		g.listedAt, g.listStart, g.listedPrefix = time.Now(), start, prefix
	}
	g.inflight = nil
	close(fl.done)
}

// list walks all the pages of the objects under prefix (recursively), each page request has its
// own timeout. Only the wanted keys are kept. A listing that can't be proven complete is an error.
func (g *SharedItems) list(
	base context.Context, prefix string, wanted map[string]struct{},
) (map[string]string, error) {
	res := map[string]string{}
	input := &s3.ListObjectsV2Input{Bucket: aws.String(g.bucket)}
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}
	paginator := s3.NewListObjectsV2Paginator(g.lister, input, func(o *s3.ListObjectsV2PaginatorOptions) {
		o.StopOnDuplicateToken = true
	})
	truncated := false
	for pages := 0; paginator.HasMorePages(); pages++ {
		if pages >= maxSharedListPages {
			return nil, fmt.Errorf("the listing of prefix %q has more than %d pages", prefix, maxSharedListPages)
		}
		page, err := g.listPage(base, paginator)
		if err != nil {
			return nil, err
		}
		truncated = aws.ToBool(page.IsTruncated)
		for _, obj := range page.Contents {
			if _, ok := wanted[aws.ToString(obj.Key)]; ok {
				res[aws.ToString(obj.Key)] = normalizeETag(aws.ToString(obj.ETag))
			}
		}
	}
	if truncated {
		// no (new) continuation token: the end of the listing was not reached
		return nil, fmt.Errorf("the listing of prefix %q is incomplete", prefix)
	}
	return res, nil
}

func (g *SharedItems) listPage(base context.Context, p *s3.ListObjectsV2Paginator) (*s3.ListObjectsV2Output, error) {
	ctx, cancel := context.WithTimeout(base, sharedRequestTimeout)
	defer cancel()
	return p.NextPage(ctx)
}

// get returns the content of the key, downloading it only if the listed ETag changed.
func (g *SharedItems) get(ctx context.Context, key string) ([]byte, error) {
	listed, err := g.etagOf(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to list S3 bucket %q, %w", g.bucket, err)
	}
	if !listed.found {
		return nil, fmt.Errorf("item %q not found in S3 bucket %q", key, g.bucket)
	}

	g.mu.Lock()
	c, ok := g.contents[key]
	if !ok {
		c = &sharedContent{}
		// an item that left the group in the meantime must not be cached again
		if g.keyRefs[key] > 0 {
			g.contents[key] = c
		}
	}
	g.mu.Unlock()

	data, err := g.content(ctx, c, key, listed)
	if err != nil {
		return nil, fmt.Errorf("unable to download item from S3 %q, %w", key, err)
	}
	return slices.Clone(data), nil
}

// content returns the cached content if it is still valid for the listing, otherwise it is
// downloaded. The content is valid if its ETag is the listed one, or if it was downloaded after
// the start of the listing (it is then at least as recent as the listing, even if the object
// changed between the listing and the download). The callers of the same object share one
// download, and a caller stops waiting when its ctx is done.
func (g *SharedItems) content(ctx context.Context, c *sharedContent, key string, listed listedItem) ([]byte, error) {
	c.mu.Lock()
	if c.data != nil && ((listed.etag != "" && c.etag == listed.etag) || !c.fetchedAt.Before(listed.start)) {
		defer c.mu.Unlock()
		return c.data, nil
	}
	dl := c.dl
	if dl == nil {
		dl = &sharedDownload{done: make(chan struct{})}
		c.dl = dl
		go g.runDownload(g.baseContext(), c, dl, key)
	}
	c.mu.Unlock()

	select {
	case <-dl.done:
		return dl.data, dl.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// runDownload downloads the object and publishes the result, it does not depend on the context
// of a caller so that a cancelled caller does not fail the others.
func (g *SharedItems) runDownload(base context.Context, c *sharedContent, dl *sharedDownload, key string) {
	data, dlETag, startedAt, err := g.download(base, key)

	c.mu.Lock()
	defer c.mu.Unlock()
	if err == nil {
		// an empty ETag is never a hit
		c.etag, c.data, c.fetchedAt = normalizeETag(dlETag), data, startedAt
	}
	dl.data, dl.err = data, err
	c.dl = nil
	close(dl.done)
}

// download waits for a free download slot, then downloads the object.
func (g *SharedItems) download(base context.Context, key string) ([]byte, string, time.Time, error) {
	select {
	case g.sem <- struct{}{}:
		defer func() { <-g.sem }()
	case <-base.Done():
		return nil, "", time.Time{}, base.Err()
	}
	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(base, sharedRequestTimeout)
	defer cancel()
	data, etag, err := downloadObject(ctx, g.downloader, g.bucket, key)
	return data, etag, startedAt, err
}

var (
	_ retriever.Retriever              = (*SharedItemRetriever)(nil)
	_ retriever.InitializableRetriever = (*SharedItemRetriever)(nil)
	_ retriever.FormatHinter           = (*SharedItemRetriever)(nil)
)

// SharedItemRetriever is the retriever of one item of a SharedItems group.
type SharedItemRetriever struct {
	group           *SharedItems
	item            string
	pollingInterval time.Duration
	status          atomic.Value // retriever.Status
}

// Init registers the item in the group.
func (r *SharedItemRetriever) Init(_ context.Context, _ *fflog.FFLogger) error {
	if r.item == "" {
		r.status.Store(retriever.RetrieverError)
		return errors.New("impossible to init shared S3 retriever: item is empty")
	}
	r.group.register(r)
	r.status.Store(retriever.RetrieverReady)
	return nil
}

// Shutdown removes the item from the group.
func (r *SharedItemRetriever) Shutdown(_ context.Context) error {
	r.group.deregister(r)
	r.status.Store(retriever.RetrieverNotReady)
	return nil
}

// Status is the internal state of the retriever.
func (r *SharedItemRetriever) Status() retriever.Status {
	if s, ok := r.status.Load().(retriever.Status); ok {
		return s
	}
	return retriever.RetrieverNotReady
}

// Retrieve returns the content of the item.
func (r *SharedItemRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if r.Status() != retriever.RetrieverReady {
		return nil, errors.New("shared S3 retriever is not initialized")
	}
	return r.group.get(ctx, r.item)
}

// OutputFormat returns an empty string like the single file Retriever: the format of the file is
// unknown, the global file format applies.
func (r *SharedItemRetriever) OutputFormat() string { return "" }
