package service

// Whitebox: flagsetManagerImpl, newFlagsetManagerWithFlagsetsAndRegistry and addFlagset are
// unexported, and these tests inject a fake shared bucket registry. The external tests of the
// manager are in flagset_manager_test.go.

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func Test_flagsetManagerImpl_sharedClientPerBucket(t *testing.T) {
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

func Test_flagsetManagerImpl_sharedClientPoller(t *testing.T) {
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
		// a new top-level pollingInterval is applied to the running poller, the flagsets are not rebuilt
		a, b := s3FlagSet("u1", "flags", "p/a.yaml"), s3FlagSet("u2", "flags", "p/b.yaml")
		a.PollingInterval, b.PollingInterval = 0, 0
		b.PollingInterval = 20000
		added := s3FlagSet("u3", "flags", "p/c.yaml")
		added.PollingInterval = 20000
		before, err := mgr.AllFlagSets()
		require.NoError(t, err)
		mgr.config = &config.Config{FlagSets: []config.FlagSet{a, b}}
		mgr.OnConfigChange(&config.Config{
			Server:        config.Server{Port: 8080},
			CommonFlagSet: config.CommonFlagSet{PollingInterval: 30000},
			FlagSets:      []config.FlagSet{a, b, added},
		})
		after, err := mgr.AllFlagSets()
		require.NoError(t, err)
		assert.Same(t, before["u1"], after["u1"])
		lists := f.lists.Load() // the added flagset did its own initial listing

		time.Sleep(29 * time.Second)
		synctest.Wait()
		assert.Equal(t, lists, f.lists.Load())
		time.Sleep(time.Second)
		synctest.Wait()
		assert.Equal(t, lists+1, f.lists.Load())

		assert.Equal(t, 1, logs.FilterMessageSnippet("polling of the shared clients updated").Len())
		// every flagset (the existing one and the one added by the reload) is warned once
		warns := logs.FilterMessageSnippet("the pollingInterval of the flagset is ignored")
		require.Equal(t, 2, warns.Len())
		warned := []any{warns.All()[0].ContextMap()["flagset"], warns.All()[1].ContextMap()["flagset"]}
		assert.ElementsMatch(t, []any{"u2", "u3"}, warned)

		// the same polling again changes nothing
		mgr.OnConfigChange(&config.Config{
			Server:        config.Server{Port: 8080},
			CommonFlagSet: config.CommonFlagSet{PollingInterval: 30000},
			FlagSets:      []config.FlagSet{a, b, added},
		})
		assert.Equal(t, 1, logs.FilterMessageSnippet("polling of the shared clients updated").Len())
		mgr.Close()
		assert.Equal(t, 0, reg.groupCount())
	})
}
