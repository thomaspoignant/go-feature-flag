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
	"go.uber.org/zap"
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
	// pending are the retrievers created by preRegister and not claimed yet.
	pending map[string][]*releasingRetriever
}

type registryGroup struct {
	items *s3retrieverv2.SharedItems
	refs  int
}

func newUnifiedRetrieverRegistry(factory sharedItemsFactory) *unifiedRetrieverRegistry {
	if factory == nil {
		factory = defaultSharedItemsFactory
	}
	return &unifiedRetrieverRegistry{
		factory: factory,
		groups:  map[string]*registryGroup{},
		pending: map[string][]*releasingRetriever{},
	}
}

// pendingKey identifies the retrievers created by preRegister for a flagset configuration.
func pendingKey(fs *config.FlagSet) string {
	rc := fs.Retriever
	return fmt.Sprintf("%s\x00%s\x00%d", rc.Bucket, rc.Item, fs.PollingInterval)
}

// retrieverFor returns the retriever of the flagset (its configuration is expected to be valid).
// The reference on the bucket group is released when the retriever is shut down.
func (r *unifiedRetrieverRegistry) retrieverFor(fs *config.FlagSet) (retriever.Retriever, error) {
	if err := checkUnifiedFlagSet(fs); err != nil {
		return nil, err
	}
	r.mu.Lock()
	key := pendingKey(fs)
	if p := r.pending[key]; len(p) > 0 {
		r.pending[key] = p[1:]
		r.mu.Unlock()
		return p[0], nil
	}
	r.mu.Unlock()
	return r.newRetriever(fs)
}

// preRegister registers the item of the flagset in its bucket group so that the first listing
// already covers it. The next retrieverFor with the same configuration returns this retriever,
// the ones that are not claimed must be released with releasePending.
func (r *unifiedRetrieverRegistry) preRegister(fs *config.FlagSet) error {
	if err := checkUnifiedFlagSet(fs); err != nil {
		return err
	}
	rr, err := r.newRetriever(fs)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := pendingKey(fs)
	r.pending[key] = append(r.pending[key], rr)
	return nil
}

// preRegisterAll pre-registers the flagsets using useUnifiedRetriever, an invalid one is only
// logged (its client creation reports the error).
func (r *unifiedRetrieverRegistry) preRegisterAll(flagsets []config.FlagSet, logger *zap.Logger) {
	for i := range flagsets {
		if !flagsets[i].UseUnifiedRetriever {
			continue
		}
		if err := r.preRegister(&flagsets[i]); err != nil {
			logger.Warn("unable to pre-register the unified retriever of the flagset",
				zap.String("flagset", flagsets[i].Name), zap.Error(err))
		}
	}
}

// releasePending shuts down the pre-registered retrievers nobody claimed.
func (r *unifiedRetrieverRegistry) releasePending() {
	r.mu.Lock()
	var unclaimed []*releasingRetriever
	for k, p := range r.pending {
		unclaimed = append(unclaimed, p...)
		delete(r.pending, k)
	}
	r.mu.Unlock()
	for _, rr := range unclaimed {
		_ = rr.Shutdown(context.Background())
	}
}

func checkUnifiedFlagSet(fs *config.FlagSet) error {
	rc := fs.Retriever
	if fs.Retrievers != nil || rc == nil || rc.Kind != retrieverconf.S3Retriever || rc.Item == "" {
		return fmt.Errorf("useUnifiedRetriever requires a single s3 retriever with item set")
	}
	return nil
}

// acquire returns the group of the bucket, with one more reference. The group is built outside of
// the lock (it loads the AWS configuration), if another caller built it meanwhile theirs is used.
func (r *unifiedRetrieverRegistry) acquire(bucket string) (*registryGroup, error) {
	r.mu.Lock()
	if g, ok := r.groups[bucket]; ok {
		g.refs++
		r.mu.Unlock()
		return g, nil
	}
	r.mu.Unlock()

	items, err := r.factory(context.Background(), bucket)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.groups[bucket]
	if !ok {
		g = &registryGroup{items: items}
		r.groups[bucket] = g
	}
	g.refs++
	return g, nil
}

// newRetriever creates the retriever of the flagset in the group of its bucket.
func (r *unifiedRetrieverRegistry) newRetriever(fs *config.FlagSet) (*releasingRetriever, error) {
	rc := fs.Retriever
	g, err := r.acquire(rc.Bucket)
	if err != nil {
		return nil, err
	}
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
