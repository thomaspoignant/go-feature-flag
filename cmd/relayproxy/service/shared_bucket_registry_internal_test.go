package service

import (
	"context"
	"errors"
	"fmt"
	"os"
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
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	retrieverInit "github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf/init"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retrieverv2"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// fakeS3 serves the same flag file for every key and counts the requests.
type fakeS3 struct {
	content []byte
	listErr error
	keys    []string
	lists   atomic.Int32
	gets    atomic.Int32

	mu        sync.Mutex
	etags     map[string]string // key -> ETag, "x" by default
	getsByKey map[string]int
}

func (f *fakeS3) setETag(key, etag string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.etags == nil {
		f.etags = map[string]string{}
	}
	f.etags[key] = etag
}

func (f *fakeS3) etagOf(key string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if e, ok := f.etags[key]; ok {
		return e
	}
	return "x"
}

func (f *fakeS3) getCount(key string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.getsByKey[key]
}

func (f *fakeS3) ListObjectsV2(
	_ context.Context, in *s3.ListObjectsV2Input, _ ...func(*s3.Options),
) (*s3.ListObjectsV2Output, error) {
	f.lists.Add(1)
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := &s3.ListObjectsV2Output{}
	keys := f.keys
	if keys == nil {
		keys = []string{"p/a.yaml", "p/b.yaml", "p/c.yaml"}
	}
	for _, k := range keys {
		if strings.HasPrefix(k, aws.ToString(in.Prefix)) {
			out.Contents = append(out.Contents, s3types.Object{Key: aws.String(k), ETag: aws.String(f.etagOf(k))})
		}
	}
	return out, nil
}

func (f *fakeS3) DownloadObject(
	_ context.Context, in *transfermanager.DownloadObjectInput, _ ...func(*transfermanager.Options),
) (*transfermanager.DownloadObjectOutput, error) {
	f.gets.Add(1)
	f.mu.Lock()
	if f.getsByKey == nil {
		f.getsByKey = map[string]int{}
	}
	f.getsByKey[aws.ToString(in.Key)]++
	f.mu.Unlock()
	if _, err := in.WriterAt.WriteAt(f.content, 0); err != nil {
		return nil, err
	}
	return &transfermanager.DownloadObjectOutput{ETag: aws.String(f.etagOf(aws.ToString(in.Key)))}, nil
}

func newFakeRegistry(t *testing.T, f *fakeS3, created *atomic.Int32) *sharedBucketRegistry {
	t.Helper()
	return newFakeRegistryWithPolling(t, f, created, 0, false, nil)
}

func newFakeRegistryWithPolling(
	t *testing.T, f *fakeS3, created *atomic.Int32, interval time.Duration, jitter bool, logger *zap.Logger,
) *sharedBucketRegistry {
	t.Helper()
	return newSharedBucketRegistry(func(ctx context.Context, bucket string) (*s3retrieverv2.SharedItems, error) {
		created.Add(1)
		return s3retrieverv2.NewSharedItems(ctx, s3retrieverv2.SharedItemsOptions{
			Bucket: bucket, Lister: f, Downloader: f,
		})
	}, interval, jitter, logger)
}

func s3FlagSet(name, bucket, item string) config.FlagSet {
	return config.FlagSet{
		Name:    name,
		APIKeys: []string{name + "-key"},
		CommonFlagSet: config.CommonFlagSet{
			Retriever:                &retrieverconf.RetrieverConf{Kind: retrieverconf.S3Retriever, Bucket: bucket, Item: item},
			UseSharedClientPerBucket: true,
			PollingInterval:          60000,
		},
	}
}

func Test_sharedBucketRegistry_retrieverFor(t *testing.T) {
	t.Parallel()
	t.Run("same bucket shares a group, other bucket has its own", func(t *testing.T) {
		t.Parallel()
		var created atomic.Int32
		reg := newFakeRegistry(t, &fakeS3{}, &created)
		fs1, fs2, fs3 := s3FlagSet("a", "b1", "p/a.yaml"), s3FlagSet("b", "b1", "p/b.yaml"),
			s3FlagSet("c", "b2", "p/c.yaml")
		r1, err := reg.retrieverFor(&fs1)
		require.NoError(t, err)
		_, err = reg.retrieverFor(&fs2)
		require.NoError(t, err)
		_, err = reg.retrieverFor(&fs3)
		require.NoError(t, err)
		assert.EqualValues(t, 2, created.Load())
		assert.Equal(t, 2, reg.groupCount())

		// the group is dropped when its last retriever is shut down (Shutdown twice is harmless)
		require.NoError(t, r1.(*releasingRetriever).Shutdown(t.Context()))
		require.NoError(t, r1.(*releasingRetriever).Shutdown(t.Context()))
		assert.Equal(t, 2, reg.groupCount())
	})

	t.Run("group is dropped after the last shutdown", func(t *testing.T) {
		t.Parallel()
		var created atomic.Int32
		reg := newFakeRegistry(t, &fakeS3{}, &created)
		fs1, fs2 := s3FlagSet("a", "b1", "p/a.yaml"), s3FlagSet("b", "b1", "p/b.yaml")
		r1, err := reg.retrieverFor(&fs1)
		require.NoError(t, err)
		r2, err := reg.retrieverFor(&fs2)
		require.NoError(t, err)
		require.NoError(t, r1.(*releasingRetriever).Shutdown(t.Context()))
		assert.Equal(t, 1, reg.groupCount())
		require.NoError(t, r2.(*releasingRetriever).Shutdown(t.Context()))
		assert.Equal(t, 0, reg.groupCount())
	})

	t.Run("unsupported configurations", func(t *testing.T) {
		t.Parallel()
		reg := newFakeRegistry(t, &fakeS3{}, &atomic.Int32{})
		folder := s3FlagSet("a", "b1", "")
		other := s3FlagSet("a", "b1", "x")
		other.Retriever.Kind = retrieverconf.HTTPRetriever
		multi := s3FlagSet("a", "b1", "x")
		multi.Retrievers = &[]retrieverconf.RetrieverConf{*multi.Retriever}
		for _, fs := range []config.FlagSet{folder, other, multi} {
			_, err := reg.retrieverFor(&fs)
			assert.Error(t, err)
		}
		assert.Equal(t, 0, reg.groupCount())
	})
}

func TestFlagsetManager_SharedClientPerBucket(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		testSharedClientManager(t)
	})
}

func testSharedClientManager(t *testing.T) {
	t.Helper()
	content, err := os.ReadFile("../testdata/goff/configuration_flags.yaml")
	require.NoError(t, err)

	f := &fakeS3{content: content}
	var created atomic.Int32
	reg := newFakeRegistry(t, f, &created)
	nonShared := config.FlagSet{
		Name:    "isolated",
		APIKeys: []string{"isolated-key"},
		CommonFlagSet: config.CommonFlagSet{
			Retriever: &retrieverconf.RetrieverConf{Kind: retrieverconf.FileRetriever, Path: "../testdata/goff/configuration_flags.yaml"},
		},
	}
	cfg := &config.Config{FlagSets: []config.FlagSet{
		s3FlagSet("u1", "flags", "p/a.yaml"),
		s3FlagSet("u2", "flags", "p/b.yaml"),
		nonShared,
	}}
	mgr, err := newFlagsetManagerWithFlagsetsAndRegistry(cfg, zap.NewNop(), nil, nil, nil, reg)
	require.NoError(t, err)

	// two shared-client flagsets on the same bucket share one group, the non shared-client one does not use any
	assert.EqualValues(t, 1, created.Load())
	assert.Equal(t, 1, reg.groupCount())
	// all the shared-client items are pre-registered, so startup lists the directory once
	assert.EqualValues(t, 1, f.lists.Load())
	assert.EqualValues(t, 2, f.gets.Load())

	// a flagset added at runtime joins the same group
	impl := mgr.(*flagsetManagerImpl)
	impl.addFlagset(s3FlagSet("u3", "flags", "p/c.yaml"))
	assert.EqualValues(t, 1, created.Load())
	assert.Equal(t, 1, reg.groupCount())
	all, err := mgr.AllFlagSets()
	require.NoError(t, err)
	assert.Contains(t, all, "u3")

	// removing flagsets releases the group once the last one is gone
	impl.removeFlagset(s3FlagSet("u1", "flags", "p/a.yaml"))
	impl.removeFlagset(s3FlagSet("u2", "flags", "p/b.yaml"))
	assert.Equal(t, 1, reg.groupCount())
	mgr.Close()
	assert.Equal(t, 0, reg.groupCount())
}

func TestFlagsetManager_SharedClientPerBucket_PreRegister(t *testing.T) {
	t.Parallel()
	var created atomic.Int32
	reg := newFakeRegistry(t, &fakeS3{}, &created)
	fs1, fs2 := s3FlagSet("a", "b1", "p/a.yaml"), s3FlagSet("b", "b1", "q/b.yaml")
	require.NoError(t, reg.preRegister(&fs1))
	require.NoError(t, reg.preRegister(&fs2))
	assert.EqualValues(t, 1, created.Load())

	// claimed retrievers are the pre-registered ones, no additional reference is taken
	r1, err := reg.retrieverFor(&fs1)
	require.NoError(t, err)
	require.NoError(t, r1.(*releasingRetriever).Shutdown(t.Context()))
	assert.Equal(t, 1, reg.groupCount(), "fs2 is still pending")

	// the unclaimed ones are released
	reg.releasePending()
	assert.Equal(t, 0, reg.groupCount())

	bad := s3FlagSet("c", "b1", "")
	assert.Error(t, reg.preRegister(&bad))
}

func Test_newGoFeatureFlagClient_ReleasesSharedClientRetrieverOnFailure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s3   *fakeS3
		mod  func(fs *config.FlagSet)
	}{
		{
			name: "invalid exporter",
			s3:   &fakeS3{content: []byte("{}")},
			mod: func(fs *config.FlagSet) {
				fs.Exporter = &config.ExporterConf{Kind: "unknown"}
			},
		},
		{
			name: "first retrieve fails",
			s3:   &fakeS3{listErr: errors.New("AccessDenied")},
			mod:  func(*config.FlagSet) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				reg := newFakeRegistry(t, tt.s3, &atomic.Int32{})
				fs := s3FlagSet("a", "b1", "p/a.yaml")
				tt.mod(&fs)
				_, err := newGoFeatureFlagClient(&fs, zap.NewNop(), nil, withSharedBucketRegistry(reg))
				require.Error(t, err)
				assert.Equal(t, 0, reg.groupCount(), "the retriever must have left its group")
			})
		})
	}
}

func failingRegistry(logger *zap.Logger) *sharedBucketRegistry {
	return newSharedBucketRegistry(func(context.Context, string) (*s3retrieverv2.SharedItems, error) {
		return nil, errors.New("no AWS")
	}, 0, false, logger)
}

func Test_sharedBucketRegistry_factoryError(t *testing.T) {
	t.Parallel()
	fs := s3FlagSet("a", "b1", "p/a.yaml")
	reg := failingRegistry(nil)
	_, err := reg.retrieverFor(&fs)
	assert.ErrorContains(t, err, "no AWS")
	assert.ErrorContains(t, reg.preRegister(&fs), "no AWS")
	assert.Equal(t, 0, reg.groupCount())
}

func Test_sharedBucketRegistry_preRegisterAll(t *testing.T) {
	t.Parallel()
	valid := s3FlagSet("valid", "b1", "p/a.yaml")
	plain := s3FlagSet("plain", "b1", "p/b.yaml")
	plain.UseSharedClientPerBucket = false
	tests := []struct {
		name       string
		reg        func(logger *zap.Logger) *sharedBucketRegistry
		wantGroups int
		wantWarn   int
	}{
		{
			name: "valid flagsets are registered, others skipped",
			reg: func(l *zap.Logger) *sharedBucketRegistry {
				return newFakeRegistryWithPolling(t, &fakeS3{}, &atomic.Int32{}, 0, false, l)
			},
			wantGroups: 1,
		},
		{name: "factory failure is logged", reg: failingRegistry, wantWarn: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			core, logs := observer.New(zap.WarnLevel)
			reg := tt.reg(zap.New(core))
			reg.preRegisterAll([]config.FlagSet{valid, plain})
			assert.Equal(t, tt.wantGroups, reg.groupCount())
			assert.Equal(t, tt.wantWarn, logs.Len())
		})
	}
}

// Not parallel: the subtests use t.Setenv.
func Test_defaultSharedItemsFactory(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr assert.ErrorAssertionFunc
	}{
		{name: "default AWS config", wantErr: assert.NoError},
		{
			name: "invalid AWS profile",
			env: map[string]string{
				"AWS_PROFILE":                 "this-profile-does-not-exist",
				"AWS_CONFIG_FILE":             "/nonexistent/config",
				"AWS_SHARED_CREDENTIALS_FILE": "/nonexistent/credentials",
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			_, err := defaultSharedItemsFactory(t.Context(), "bucket")
			tt.wantErr(t, err)
		})
	}
}

func Test_initRetrieversWithRegistry_shared(t *testing.T) {
	t.Parallel()
	valid := s3FlagSet("a", "b1", "p/a.yaml")
	folder := s3FlagSet("a", "b1", "")
	tests := []struct {
		name    string
		fs      config.FlagSet
		reg     *sharedBucketRegistry
		wantLen int
		wantErr assert.ErrorAssertionFunc
	}{
		{name: "no registry gets a private one", fs: valid, reg: nil, wantLen: 1, wantErr: assert.NoError},
		{name: "invalid configuration", fs: folder, reg: nil, wantErr: assert.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := initRetrieversWithRegistry(&tt.fs, tt.reg)
			tt.wantErr(t, err)
			assert.Len(t, got, tt.wantLen)
		})
	}
}

func Test_flagsetManagerImpl_addFlagset_sharedDuplicate(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		content, err := os.ReadFile("../testdata/goff/configuration_flags.yaml")
		require.NoError(t, err)
		reg := newFakeRegistry(t, &fakeS3{content: content}, &atomic.Int32{})
		first := s3FlagSet("u1", "flags", "p/a.yaml")
		mgr, err := newFlagsetManagerWithFlagsetsAndRegistry(
			&config.Config{FlagSets: []config.FlagSet{first}}, zap.NewNop(), nil, nil, nil, reg)
		require.NoError(t, err)

		// the flagset name is already in the configuration: the new client is closed and released
		dup := s3FlagSet("u1", "flags", "p/b.yaml")
		mgr.(*flagsetManagerImpl).addFlagset(dup)
		assert.Equal(t, 1, reg.groupCount())
		mgr.Close()
		assert.Equal(t, 0, reg.groupCount())
	})
}

func Test_flagsetManagerImpl_reloadAddsSharedClientFlagsets(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		content, err := os.ReadFile("../testdata/goff/configuration_flags.yaml")
		require.NoError(t, err)
		f := &fakeS3{content: content}
		flagsets := make([]config.FlagSet, 11)
		for i := range flagsets {
			f.keys = append(f.keys, fmt.Sprintf("p/f%d.yaml", i))
			flagsets[i] = s3FlagSet(fmt.Sprintf("u%d", i), "flags", f.keys[i])
		}
		reg := newFakeRegistry(t, f, &atomic.Int32{})
		mgr, err := newFlagsetManagerWithFlagsetsAndRegistry(
			&config.Config{FlagSets: flagsets[:1]}, zap.NewNop(), nil, nil, nil, reg)
		require.NoError(t, err)
		require.EqualValues(t, 1, f.lists.Load())

		// 10 flagsets added in one reload share one listing
		mgr.OnConfigChange(&config.Config{Server: config.Server{Port: 8080}, FlagSets: flagsets})
		all, err := mgr.AllFlagSets()
		require.NoError(t, err)
		assert.Len(t, all, 11)
		assert.EqualValues(t, 2, f.lists.Load())
		mgr.Close()
		assert.Equal(t, 0, reg.groupCount())
	})
}

func Test_sharedBucketRegistry_concurrentGroupCreation(t *testing.T) {
	t.Parallel()
	const callers = 4
	var building sync.WaitGroup
	building.Add(callers)
	reg := newSharedBucketRegistry(func(ctx context.Context, bucket string) (*s3retrieverv2.SharedItems, error) {
		// every caller is building its own group at the same time: only one is kept
		building.Done()
		building.Wait()
		return s3retrieverv2.NewSharedItems(ctx, s3retrieverv2.SharedItemsOptions{
			Bucket: bucket, Lister: &fakeS3{}, Downloader: &fakeS3{},
		})
	}, 0, false, nil)
	var wg sync.WaitGroup
	retrievers := make(chan *releasingRetriever, callers)
	for i := range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fs := s3FlagSet(fmt.Sprintf("f%d", i), "b1", fmt.Sprintf("p/%d.yaml", i))
			r, err := reg.retrieverFor(&fs)
			assert.NoError(t, err)
			retrievers <- r.(*releasingRetriever)
		}()
	}
	wg.Wait()
	close(retrievers)
	assert.Equal(t, 1, reg.groupCount())
	for r := range retrievers {
		require.NoError(t, r.Shutdown(t.Context()))
	}
	assert.Equal(t, 0, reg.groupCount(), "all the references were taken on the same group")
}

func newPollingManager(
	t *testing.T, f *fakeS3, interval time.Duration, logger *zap.Logger,
) (*flagsetManagerImpl, *sharedBucketRegistry) {
	t.Helper()
	content, err := os.ReadFile("../testdata/goff/configuration_flags.yaml")
	require.NoError(t, err)
	f.content = content
	reg := newFakeRegistryWithPolling(t, f, &atomic.Int32{}, interval, false, logger)
	a, b := s3FlagSet("u1", "flags", "p/a.yaml"), s3FlagSet("u2", "flags", "p/b.yaml")
	a.PollingInterval, b.PollingInterval = 0, 0
	mgr, err := newFlagsetManagerWithFlagsetsAndRegistry(
		&config.Config{FlagSets: []config.FlagSet{a, b}}, logger, nil, nil, nil, reg)
	require.NoError(t, err)
	return mgr.(*flagsetManagerImpl), reg
}

func Test_sharedBucketRegistry_poller(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		f := &fakeS3{}
		mgr, reg := newPollingManager(t, f, 10*time.Second, zap.NewNop())
		require.EqualValues(t, 1, f.lists.Load())
		require.EqualValues(t, 2, f.gets.Load())

		// the clients do not poll, one listing per polling interval of the relay proxy
		time.Sleep(2 * time.Minute)
		synctest.Wait()
		assert.EqualValues(t, 1+12, f.lists.Load())
		assert.EqualValues(t, 2, f.gets.Load(), "nothing changed")

		// only the flagset of the changed object downloads it
		f.setETag("p/a.yaml", `"y"`)
		time.Sleep(10 * time.Second)
		synctest.Wait()
		assert.EqualValues(t, 1+13, f.lists.Load())
		assert.Equal(t, 2, f.getCount("p/a.yaml"))
		assert.Equal(t, 1, f.getCount("p/b.yaml"))

		mgr.Close()
		assert.Equal(t, 0, reg.groupCount())
	})
}

func Test_flagsetManagerImpl_OnConfigChange_sharedClientPolling(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		core, logs := observer.New(zap.InfoLevel)
		f := &fakeS3{}
		mgr, reg := newPollingManager(t, f, 10*time.Second, zap.New(core))
		lists := f.lists.Load()

		// a new top-level pollingInterval is applied to the running poller, the flagsets are not rebuilt
		a, b := s3FlagSet("u1", "flags", "p/a.yaml"), s3FlagSet("u2", "flags", "p/b.yaml")
		a.PollingInterval, b.PollingInterval = 0, 0
		b.PollingInterval = 20000
		before, err := mgr.AllFlagSets()
		require.NoError(t, err)
		mgr.config = &config.Config{FlagSets: []config.FlagSet{a, b}}
		mgr.OnConfigChange(&config.Config{
			Server:        config.Server{Port: 8080},
			CommonFlagSet: config.CommonFlagSet{PollingInterval: 30000},
			FlagSets:      []config.FlagSet{a, b},
		})
		after, err := mgr.AllFlagSets()
		require.NoError(t, err)
		assert.Same(t, before["u1"], after["u1"])

		time.Sleep(29 * time.Second)
		synctest.Wait()
		assert.Equal(t, lists, f.lists.Load())
		time.Sleep(time.Second)
		synctest.Wait()
		assert.Equal(t, lists+1, f.lists.Load())

		assert.Equal(t, 1, logs.FilterMessageSnippet("polling of the shared clients updated").Len())
		warns := logs.FilterMessageSnippet("the pollingInterval of the flagset is ignored")
		require.Equal(t, 1, warns.Len())
		assert.Equal(t, "u2", warns.All()[0].ContextMap()["flagset"])

		// the same polling again changes nothing
		mgr.OnConfigChange(&config.Config{
			Server:        config.Server{Port: 8080},
			CommonFlagSet: config.CommonFlagSet{PollingInterval: 30000},
			FlagSets:      []config.FlagSet{a, b},
		})
		assert.Equal(t, 1, logs.FilterMessageSnippet("polling of the shared clients updated").Len())
		mgr.Close()
		assert.Equal(t, 0, reg.groupCount())
	})
}

func Test_sharedBucketRegistry_warnPollingMismatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		parent      time.Duration
		parentJit   bool
		polling     int
		jitter      bool
		wantMessage []string
	}{
		{name: "unset values follow the parent", parent: 10 * time.Second},
		{name: "same values", parent: time.Minute, polling: 60000},
		{name: "same interval as the default parent", polling: 60000},
		{
			name: "different interval", parent: 10 * time.Second, polling: 5000,
			wantMessage: []string{"the pollingInterval of the flagset is ignored"},
		},
		{name: "jitter only on the flagset", jitter: true, wantMessage: []string{"the enablePollingJitter of the flagset is ignored"}},
		{name: "jitter on the parent", parentJit: true, jitter: true},
		{
			name: "both differ", parent: 10 * time.Second, polling: 1000, jitter: true,
			wantMessage: []string{"the pollingInterval of the flagset", "the enablePollingJitter of the flagset"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			core, logs := observer.New(zap.WarnLevel)
			reg := newFakeRegistryWithPolling(t, &fakeS3{}, &atomic.Int32{}, tt.parent, tt.parentJit, zap.New(core))
			fs := s3FlagSet("fs", "b1", "p/a.yaml")
			fs.PollingInterval, fs.EnablePollingJitter = tt.polling, tt.jitter
			r, err := reg.retrieverFor(&fs)
			require.NoError(t, err)
			require.NoError(t, r.(*releasingRetriever).Shutdown(t.Context()))

			require.Equal(t, len(tt.wantMessage), logs.Len())
			for i, snippet := range tt.wantMessage {
				entry := logs.All()[i]
				assert.Contains(t, entry.Message, snippet)
				assert.Equal(t, "fs", entry.ContextMap()["flagset"])
			}
		})
	}
}

func Test_firstSharedRetriever(t *testing.T) {
	t.Parallel()
	reg := newFakeRegistry(t, &fakeS3{}, &atomic.Int32{})
	fs := s3FlagSet("a", "b1", "p/a.yaml")
	r, err := reg.retrieverFor(&fs)
	require.NoError(t, err)
	file, err := retrieverInit.InitRetriever(&retrieverconf.RetrieverConf{
		Kind: retrieverconf.FileRetriever, Path: "../testdata/goff/configuration_flags.yaml",
	})
	require.NoError(t, err)
	tests := []struct {
		name       string
		retrievers []retriever.Retriever
		want       bool
	}{
		{name: "no retriever"},
		{name: "another kind of retriever", retrievers: []retriever.Retriever{file}},
		{name: "shared-client retriever", retrievers: []retriever.Retriever{r}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := firstSharedRetriever(tt.retrievers)
			assert.Equal(t, tt.want, ok)
			assert.Equal(t, tt.want, got != nil)
		})
	}
	require.NoError(t, r.(*releasingRetriever).Shutdown(t.Context()))
}
