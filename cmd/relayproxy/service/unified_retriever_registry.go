package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	awsConf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retrieverv2"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// sharedItemsFactory builds the shared group of a bucket.
type sharedItemsFactory func(ctx context.Context, bucket string) (*s3retrieverv2.SharedItems, error)

func defaultSharedItemsFactory(ctx context.Context, bucket string) (*s3retrieverv2.SharedItems, error) {
	awsConfig, err := awsConf.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return s3retrieverv2.NewSharedItems(ctx, s3retrieverv2.SharedItemsOptions{
		Bucket:    bucket,
		AwsConfig: &awsConfig,
	})
}

// unifiedRetrieverRegistry lazily creates one shared S3 group per bucket and hands out the
// retrievers of the flagsets using useUnifiedRetriever. A group is dropped once all its
// retrievers have been shut down.
type unifiedRetrieverRegistry struct {
	factory sharedItemsFactory

	mu     sync.Mutex
	groups map[string]*registryGroup
}

type registryGroup struct {
	items *s3retrieverv2.SharedItems
	refs  int
}

func newUnifiedRetrieverRegistry(factory sharedItemsFactory) *unifiedRetrieverRegistry {
	if factory == nil {
		factory = defaultSharedItemsFactory
	}
	return &unifiedRetrieverRegistry{factory: factory, groups: map[string]*registryGroup{}}
}

// retrieverFor returns the retriever of the flagset (its configuration is expected to be valid).
// The reference on the bucket group is released when the retriever is shut down.
func (r *unifiedRetrieverRegistry) retrieverFor(fs *config.FlagSet) (retriever.Retriever, error) {
	rc := fs.Retriever
	if fs.Retrievers != nil || rc == nil || rc.Kind != retrieverconf.S3Retriever || rc.Item == "" {
		return nil, fmt.Errorf("useUnifiedRetriever requires a single s3 retriever with item set")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.groups[rc.Bucket]
	if !ok {
		items, err := r.factory(context.Background(), rc.Bucket)
		if err != nil {
			return nil, err
		}
		g = &registryGroup{items: items}
		r.groups[rc.Bucket] = g
	}
	g.refs++
	polling := time.Duration(fs.PollingInterval) * time.Millisecond
	return &releasingRetriever{
		SharedItemRetriever: g.items.Retriever(rc.Item, polling),
		release:             func() { r.release(rc.Bucket, g) },
	}, nil
}

func (r *unifiedRetrieverRegistry) release(bucket string, g *registryGroup) {
	r.mu.Lock()
	defer r.mu.Unlock()
	g.refs--
	if g.refs <= 0 && r.groups[bucket] == g {
		delete(r.groups, bucket)
	}
}

// groupCount is the number of live groups.
func (r *unifiedRetrieverRegistry) groupCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.groups)
}

// releasingRetriever releases its reference on the group the first time it is shut down.
type releasingRetriever struct {
	*s3retrieverv2.SharedItemRetriever
	release     func()
	releaseOnce sync.Once
}

var _ retriever.InitializableRetriever = (*releasingRetriever)(nil)

func (r *releasingRetriever) Init(ctx context.Context, logger *fflog.FFLogger) error {
	return r.SharedItemRetriever.Init(ctx, logger)
}

func (r *releasingRetriever) Shutdown(ctx context.Context) error {
	err := r.SharedItemRetriever.Shutdown(ctx)
	r.releaseOnce.Do(r.release)
	return err
}
