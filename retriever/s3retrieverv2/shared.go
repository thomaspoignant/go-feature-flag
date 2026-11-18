package s3retrieverv2

import (
	"context"
	"errors"
	"fmt"
	"path"
	"slices"
	"sort"
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
	// sharedListTimeout bounds one listing, which is not tied to the context of a caller.
	sharedListTimeout = 30 * time.Second
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
// requests: the directories of the registered items are listed once per listing TTL window for
// all of them, and an object is only downloaded when its ETag differs from the cached one.
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
	dirs     []string      // sorted distinct prefixes ("" or "dir/") of the registered items
	ttl      time.Duration // listing TTL of the registered items
	contents map[string]*sharedContent

	listing    map[string]string // key -> normalized ETag
	listedDirs map[string]struct{}
	listErr    error
	listedAt   time.Time
	listed     bool
	inflight   *sharedListing
}

// sharedListing is a listing in progress, shared by all the callers waiting for it.
type sharedListing struct {
	done chan struct{}
}

// sharedContent is the cached content of one object.
type sharedContent struct {
	mu   sync.Mutex
	etag string
	data []byte
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
		contents:   map[string]*sharedContent{},
		ttl:        minSharedListingTTL,
	}, nil
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
	g.keyRefs[r.item]++
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
	}
	g.recompute()
}

// recompute refreshes the listed directories and the TTL. The caller must hold g.mu.
func (g *SharedItems) recompute() {
	minPolling := time.Duration(0)
	dirs := map[string]struct{}{}
	for it := range g.items {
		dirs[dirPrefix(it.item)] = struct{}{}
		if minPolling == 0 || it.pollingInterval < minPolling {
			minPolling = it.pollingInterval
		}
	}
	g.dirs = g.dirs[:0]
	for d := range dirs {
		g.dirs = append(g.dirs, d)
	}
	sort.Strings(g.dirs)
	g.ttl = max(minSharedListingTTL, minPolling/2)
}

// dirPrefix is the listing prefix of the directory of the key: "" for the root, "dir/" otherwise.
func dirPrefix(key string) string {
	dir := path.Dir(key)
	if dir == "." || dir == "/" {
		return ""
	}
	return dir + "/"
}

// normalizeETag ignores the weak prefix and the surrounding double quotes.
func normalizeETag(v string) string {
	return strings.Trim(strings.TrimPrefix(v, "W/"), `"`)
}

// freshLocked tells if the current listing can serve the key. The caller must hold g.mu.
func (g *SharedItems) freshLocked(key string) bool {
	if !g.listed {
		return false
	}
	if _, ok := g.listedDirs[dirPrefix(key)]; !ok {
		return false
	}
	ttl := g.ttl
	if g.listErr != nil {
		ttl = min(ttl, maxSharedErrorTTL)
	}
	return time.Since(g.listedAt) < ttl
}

// etagOf returns the ETag of the key in a listing no older than the TTL, listing if needed.
// Concurrent callers share one listing, and a caller stops waiting when its ctx is done.
func (g *SharedItems) etagOf(ctx context.Context, key string) (string, bool, error) {
	for range 3 {
		g.mu.Lock()
		if g.freshLocked(key) {
			etag, ok := g.listing[key]
			err := g.listErr
			g.mu.Unlock()
			return etag, ok, err
		}
		fl := g.inflight
		if fl == nil {
			fl = &sharedListing{done: make(chan struct{})}
			g.inflight = fl
			go g.runListing(context.WithoutCancel(ctx), fl, slices.Clone(g.dirs))
		}
		g.mu.Unlock()

		select {
		case <-fl.done:
		case <-ctx.Done():
			return "", false, ctx.Err()
		}
	}
	return "", false, errors.New("item is not registered in the shared listing")
}

// runListing lists the directories and publishes the result. It does not depend on the context of
// a caller so that a cancelled caller does not fail the others.
func (g *SharedItems) runListing(base context.Context, fl *sharedListing, dirs []string) {
	ctx, cancel := context.WithTimeout(base, sharedListTimeout)
	defer cancel()
	listing, err := g.list(ctx, dirs)

	g.mu.Lock()
	defer g.mu.Unlock()
	g.listing, g.listErr, g.listedAt, g.listed = listing, err, time.Now(), true
	g.listedDirs = make(map[string]struct{}, len(dirs))
	for _, d := range dirs {
		g.listedDirs[d] = struct{}{}
	}
	g.inflight = nil
	close(fl.done)
}

// list walks all the pages of each directory (not recursively).
func (g *SharedItems) list(ctx context.Context, dirs []string) (map[string]string, error) {
	res := map[string]string{}
	for _, prefix := range dirs {
		input := &s3.ListObjectsV2Input{Bucket: aws.String(g.bucket), Delimiter: aws.String("/")}
		if prefix != "" {
			input.Prefix = aws.String(prefix)
		}
		paginator := s3.NewListObjectsV2Paginator(g.lister, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			for _, obj := range page.Contents {
				res[aws.ToString(obj.Key)] = normalizeETag(aws.ToString(obj.ETag))
			}
		}
	}
	return res, nil
}

// get returns the content of the key, downloading it only if the listed ETag changed.
func (g *SharedItems) get(ctx context.Context, key string) ([]byte, error) {
	etag, found, err := g.etagOf(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to list S3 bucket %q, %w", g.bucket, err)
	}
	if !found {
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

	c.mu.Lock()
	defer c.mu.Unlock()
	if etag != "" && c.etag == etag && c.data != nil {
		return slices.Clone(c.data), nil
	}
	data, dlETag, err := downloadObject(ctx, g.downloader, g.bucket, key)
	if err != nil {
		return nil, fmt.Errorf("unable to download item from S3 %q, %v", key, err)
	}
	dlETag = normalizeETag(dlETag)
	if dlETag == "" {
		// no ETag in the download, the listed one identifies the content
		dlETag = etag
	}
	if data == nil {
		data = []byte{}
	}
	c.etag, c.data = dlETag, data
	return slices.Clone(c.data), nil
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
