package flaghistory

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type fakeStore struct {
	mu        sync.Mutex
	entries   []Entry
	latest    map[string]Entry
	failUntil int32 // number of Record calls that fail
	calls     atomic.Int32
	purged    atomic.Int32
	closed    atomic.Bool
}

func (f *fakeStore) Migrate(context.Context) error { return nil }
func (f *fakeStore) Record(_ context.Context, e Entry) (bool, error) {
	if f.calls.Add(1) <= atomic.LoadInt32(&f.failUntil) {
		return false, errors.New("db down")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries = append(f.entries, e)
	return true, nil
}
func (f *fakeStore) Latest(context.Context, string, string) (*Entry, error) { return nil, nil }
func (f *fakeStore) LatestPerFlag(context.Context, string) (map[string]Entry, error) {
	return f.latest, nil
}
func (f *fakeStore) List(context.Context, Filter) (Page, error) { return Page{}, nil }
func (f *fakeStore) Purge(context.Context, time.Time) (int64, error) {
	f.purged.Add(1)
	return 0, nil
}
func (f *fakeStore) Ping(context.Context) error { return nil }
func (f *fakeStore) Close(context.Context)      { f.closed.Store(true) }
func (f *fakeStore) recorded() []Entry {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]Entry(nil), f.entries...)
}

func fastConfig(connect func(ctx context.Context) (Store, error)) Config {
	return Config{
		Connect:      connect,
		Source:       "test",
		backoffMin:   time.Millisecond,
		backoffMax:   5 * time.Millisecond,
		flushTimeout: time.Second,
		purgeEvery:   time.Hour,
	}
}

func addedDiff(keys ...string) notifier.DiffCache {
	added := map[string]flag.Flag{}
	for _, k := range keys {
		added[k] = newFlag("1", false)
	}
	return notifier.DiffCache{Added: added}
}

func TestHistory_recordsChanges(t *testing.T) {
	store := &fakeStore{}
	h := New(fastConfig(func(context.Context) (Store, error) { return store, nil }))
	h.Start()
	defer h.Close()

	require.NoError(t, h.NotifierFor("default").Notify(addedDiff("a", "b")))
	assert.Eventually(t, func() bool { return len(store.recorded()) == 2 }, time.Second, 5*time.Millisecond)
	assert.False(t, store.recorded()[0].DetectedAt.IsZero())
}

func TestHistory_databaseDownAtStartup(t *testing.T) {
	store := &fakeStore{}
	var attempts atomic.Int32
	h := New(fastConfig(func(context.Context) (Store, error) {
		if attempts.Add(1) < 4 {
			return nil, errors.New("connection refused")
		}
		return store, nil
	}))
	h.Start()
	defer h.Close()

	_, err := h.Store()
	assert.ErrorIs(t, err, ErrNotReady)

	// the change happens while the database is still down: it must be kept in memory.
	require.NoError(t, h.NotifierFor("default").Notify(addedDiff("a")))
	assert.Eventually(t, func() bool { return len(store.recorded()) == 1 }, time.Second, 5*time.Millisecond)
	_, err = h.Store()
	assert.NoError(t, err)
}

func TestHistory_retriesWhenWriteFails(t *testing.T) {
	store := &fakeStore{failUntil: 3}
	h := New(fastConfig(func(context.Context) (Store, error) { return store, nil }))
	h.Start()
	defer h.Close()

	require.NoError(t, h.NotifierFor("default").Notify(addedDiff("a")))
	assert.Eventually(t, func() bool { return len(store.recorded()) == 1 }, time.Second, 5*time.Millisecond)
	assert.GreaterOrEqual(t, int(store.calls.Load()), 4)
}

func TestHistory_queueFullDropsAndCounts(t *testing.T) {
	cfg := fastConfig(func(ctx context.Context) (Store, error) {
		<-ctx.Done() // never connects
		return nil, ctx.Err()
	})
	cfg.QueueSize = 2
	h := New(cfg)
	defer h.Close()

	require.NoError(t, h.NotifierFor("default").Notify(addedDiff("a", "b", "c", "d")))
	assert.Equal(t, int64(2), h.Dropped())
}

func TestHistory_closeFlushesQueue(t *testing.T) {
	store := &fakeStore{failUntil: 2}
	h := New(fastConfig(func(context.Context) (Store, error) { return store, nil }))
	h.Start()
	require.NoError(t, h.NotifierFor("default").Notify(addedDiff("a", "b", "c")))
	h.Close()

	assert.Len(t, store.recorded(), 3)
	assert.True(t, store.closed.Load())
}

func TestHistory_closeDoesNotHangWhenDatabaseIsDown(t *testing.T) {
	h := New(fastConfig(func(ctx context.Context) (Store, error) {
		return nil, errors.New("down")
	}))
	h.Start()
	h.Reconcile("default", map[string]flag.Flag{"a": newFlag("1", false)})

	done := make(chan struct{})
	go func() { h.Close(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Close is blocked while the database is down")
	}
}

func TestHistory_reconcileEntries(t *testing.T) {
	unchanged := newFlag("1", false)
	unchangedHash := mustHash(t, unchanged)
	tests := []struct {
		name   string
		latest map[string]Entry
		flags  map[string]flag.Flag
		want   map[string]Action
	}{
		{
			name:   "first start records the flags as added",
			latest: map[string]Entry{},
			flags:  map[string]flag.Flag{"a": newFlag("1", false)},
			want:   map[string]Action{"a": ActionAdded},
		},
		{
			name: "same hash is ignored",
			latest: map[string]Entry{
				"a": {FlagKey: "a", Action: ActionAdded, AfterHash: unchangedHash},
			},
			flags: map[string]flag.Flag{"a": unchanged},
			want:  map[string]Action{},
		},
		{
			name: "modified while stopped",
			latest: map[string]Entry{
				"a": {FlagKey: "a", Action: ActionUpdated, AfterHash: "old"},
			},
			flags: map[string]flag.Flag{"a": unchanged},
			want:  map[string]Action{"a": ActionUpdated},
		},
		{
			name: "deleted while stopped",
			latest: map[string]Entry{
				"gone": {FlagKey: "gone", Action: ActionAdded, AfterHash: "x"},
			},
			flags: map[string]flag.Flag{},
			want:  map[string]Action{"gone": ActionDeleted},
		},
		{
			name: "already deleted is not recorded again",
			latest: map[string]Entry{
				"gone": {FlagKey: "gone", Action: ActionDeleted},
			},
			flags: map[string]flag.Flag{},
			want:  map[string]Action{},
		},
		{
			name: "re-created after a deletion",
			latest: map[string]Entry{
				"a": {FlagKey: "a", Action: ActionDeleted},
			},
			flags: map[string]flag.Flag{"a": unchanged},
			want:  map[string]Action{"a": ActionAdded},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{latest: tt.latest}
			h := New(fastConfig(func(context.Context) (Store, error) { return store, nil }))
			h.Start()
			defer h.Close()
			require.True(t, h.waitReady())

			entries, err := h.reconcileEntries("default", tt.flags)
			require.NoError(t, err)
			got := map[string]Action{}
			for _, e := range entries {
				got[e.FlagKey] = e.Action
				assert.True(t, e.Diff.Reconciled)
				assert.Equal(t, UnknownAuthor, e.Author)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHistory_reconcileWritesEntries(t *testing.T) {
	store := &fakeStore{latest: map[string]Entry{}}
	h := New(fastConfig(func(context.Context) (Store, error) { return store, nil }))
	h.Start()
	defer h.Close()

	h.Reconcile("default", map[string]flag.Flag{"a": newFlag("1", false)})
	assert.Eventually(t, func() bool { return len(store.recorded()) == 1 }, time.Second, 5*time.Millisecond)
}

func TestHistory_purgeRunsWhenRetentionIsSet(t *testing.T) {
	store := &fakeStore{}
	cfg := fastConfig(func(context.Context) (Store, error) { return store, nil })
	cfg.Retention = time.Hour
	h := New(cfg)
	h.Start()
	defer h.Close()
	assert.Eventually(t, func() bool { return store.purged.Load() >= 1 }, time.Second, 5*time.Millisecond)
}

func mustHash(t *testing.T, f flag.Flag) string {
	t.Helper()
	m, err := toMap(f)
	require.NoError(t, err)
	h, err := hashFlag(m)
	require.NoError(t, err)
	return h
}
