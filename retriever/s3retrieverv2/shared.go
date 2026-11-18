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
// requests: the bucket is listed once per listing TTL window for all of them, and an object is
// only downloaded when its ETag differs from the one of the cached content.
//
// The listing TTL is max(1s, minPollingInterval/10), where minPollingInterval is the lowest
// polling interval of the registered items. A change in the bucket can therefore be detected up
// to one TTL later than with an independent retriever.
type SharedItems struct {
	bucket     string
	lister     ListerAPI
	downloader DownloaderAPI

	// mu protects the registered items, the listing and the content cache map.
	mu    sync.Mutex
	items map[*SharedItemRetriever]struct{}
	// listMu serializes the refresh of the listing so that concurrent callers reuse its result.
	listMu       sync.Mutex
	listing      map[string]string // key -> normalized ETag
	listedPrefix string
	listedAt     time.Time
	listed       bool
	contents     map[string]*sharedContent
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
		contents:   map[string]*sharedContent{},
	}, nil
}

// Bucket returns the bucket of the group.
func (g *SharedItems) Bucket() string { return g.bucket }

// Retriever returns the retriever of one item of the bucket. It joins the group when Init is
// called and leaves it on Shutdown. pollingInterval is used to compute the listing TTL, 0 means
// the default one (60s).
func (g *SharedItems) Retriever(item string, pollingInterval time.Duration) *SharedItemRetriever {
	if pollingInterval <= 0 {
		pollingInterval = DefaultSharedPollingInterval
	}
	return &SharedItemRetriever{group: g, item: item, pollingInterval: pollingInterval}
}

func (g *SharedItems) register(r *SharedItemRetriever) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.items[r] = struct{}{}
}

func (g *SharedItems) deregister(r *SharedItemRetriever) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.items, r)
	// Drop the contents that are not used anymore.
	used := map[string]struct{}{}
	for it := range g.items {
		used[it.item] = struct{}{}
	}
	for key := range g.contents {
		if _, ok := used[key]; !ok {
			delete(g.contents, key)
		}
	}
}

// ttlAndPrefix computes the listing TTL and prefix from the registered items.
func (g *SharedItems) ttlAndPrefix() (time.Duration, string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	minPolling := time.Duration(0)
	keys := make([]string, 0, len(g.items))
	for it := range g.items {
		keys = append(keys, it.item)
		if minPolling == 0 || it.pollingInterval < minPolling {
			minPolling = it.pollingInterval
		}
	}
	ttl := max(minSharedListingTTL, minPolling/10)
	return ttl, commonPrefix(keys)
}

// commonPrefix is the longest common prefix of the keys.
func commonPrefix(keys []string) string {
	if len(keys) == 0 {
		return ""
	}
	prefix := keys[0]
	for _, k := range keys[1:] {
		for !strings.HasPrefix(k, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}

// normalizeETag ignores the weak prefix and the surrounding double quotes.
func normalizeETag(v string) string {
	return strings.Trim(strings.TrimPrefix(v, "W/"), `"`)
}

// etagOf returns the ETag of the key in a listing no older than the TTL (listing it if needed).
func (g *SharedItems) etagOf(ctx context.Context, key string) (string, bool, error) {
	g.listMu.Lock()
	defer g.listMu.Unlock()
	ttl, prefix := g.ttlAndPrefix()
	fresh := g.listed && time.Since(g.listedAt) < ttl && strings.HasPrefix(key, g.listedPrefix)
	if !fresh {
		listing, err := g.list(ctx, prefix)
		if err != nil {
			g.listed = false
			return "", false, err
		}
		g.listing, g.listedPrefix, g.listedAt, g.listed = listing, prefix, time.Now(), true
	}
	etag, ok := g.listing[key]
	return etag, ok, nil
}

// list walks all the pages of the objects under prefix (recursively).
func (g *SharedItems) list(ctx context.Context, prefix string) (map[string]string, error) {
	input := &s3.ListObjectsV2Input{Bucket: aws.String(g.bucket)}
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}
	res := map[string]string{}
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
	return res, nil
}

// get returns the content of the key, downloading it only if the listed ETag changed.
func (g *SharedItems) get(ctx context.Context, key string) ([]byte, error) {
	etag, found, err := g.etagOf(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to list bucket %q: %w", g.bucket, err)
	}
	if !found {
		return nil, fmt.Errorf("item %q not found in S3 bucket %q", key, g.bucket)
	}

	g.mu.Lock()
	c, ok := g.contents[key]
	if !ok {
		c = &sharedContent{}
		g.contents[key] = c
	}
	g.mu.Unlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if etag != "" && c.etag == etag && c.data != nil {
		return slices.Clone(c.data), nil
	}
	data, dlETag, err := downloadObject(ctx, g.downloader, g.bucket, key)
	if err != nil {
		return nil, fmt.Errorf("unable to download item %q from S3 bucket %q: %w", key, g.bucket, err)
	}
	c.etag, c.data = normalizeETag(dlETag), data
	if c.data == nil {
		c.data = []byte{}
	}
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
