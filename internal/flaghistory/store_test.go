//go:build docker

package flaghistory_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	testcontainerPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/thomaspoignant/go-feature-flag/internal/flaghistory"
)

func newStore(t *testing.T) *flaghistory.PostgresStore {
	t.Helper()
	store, _ := newStoreWithURI(t)
	return store
}

func newStoreWithURI(t *testing.T) (*flaghistory.PostgresStore, string) {
	t.Helper()
	ctx := context.Background()
	container, err := testcontainerPostgres.Run(ctx, "postgres:16-alpine",
		testcontainerPostgres.WithDatabase("goff"),
		testcontainerPostgres.WithUsername("goff"),
		testcontainerPostgres.WithPassword("goff"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second)),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	uri, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	store, err := flaghistory.NewPostgresStore(ctx, uri, "")
	require.NoError(t, err)
	t.Cleanup(func() { store.Close(ctx) })
	require.NoError(t, store.Migrate(ctx))
	require.NoError(t, store.Migrate(ctx), "migration must be idempotent")
	return store, uri
}

func entry(key string, action flaghistory.Action, hash string) flaghistory.Entry {
	return flaghistory.Entry{
		Flagset: "default", FlagKey: key, Action: action, AfterHash: hash,
		Author: "jane", Source: "test",
		Diff: flaghistory.Diff{Changes: []flaghistory.FieldChange{{Path: "version", Old: "1", New: "2"}}},
	}
}

func TestPostgresStore_RecordDeduplicates(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	tests := []struct {
		name     string
		entry    flaghistory.Entry
		inserted bool
	}{
		{"first change", entry("f", flaghistory.ActionAdded, "h1"), true},
		{"same change seen by another instance", entry("f", flaghistory.ActionAdded, "h1"), false},
		{"update", entry("f", flaghistory.ActionUpdated, "h2"), true},
		{"toggle back to a previous value", entry("f", flaghistory.ActionUpdated, "h1"), true},
		{"toggle again", entry("f", flaghistory.ActionUpdated, "h2"), true},
		{"first load after restart (same hash)", entry("f", flaghistory.ActionAdded, "h2"), false},
		{"first load after restart (changed hash)", entry("f", flaghistory.ActionAdded, "h3"), true},
		{"deletion", entry("f", flaghistory.ActionDeleted, ""), true},
		{"deletion seen twice", entry("f", flaghistory.ActionDeleted, ""), false},
		{"re-created", entry("f", flaghistory.ActionAdded, "h3"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inserted, err := store.Record(ctx, tt.entry)
			require.NoError(t, err)
			assert.Equal(t, tt.inserted, inserted)
		})
	}

	// the "first load after restart (changed hash)" must be stored as an update, not as an addition.
	page, err := store.List(ctx, flaghistory.Filter{FlagKey: "f", Limit: 100})
	require.NoError(t, err)
	require.Len(t, page.Entries, 7)
	assert.Equal(t, flaghistory.ActionAdded, page.Entries[0].Action) // re-created
	assert.Equal(t, flaghistory.ActionDeleted, page.Entries[1].Action)
	assert.Equal(t, flaghistory.ActionUpdated, page.Entries[2].Action)
	assert.True(t, page.Entries[2].Diff.Reconciled)
}

func TestPostgresStore_RecordConcurrentInstances(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	var wg sync.WaitGroup
	var mu sync.Mutex
	inserted := 0
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, err := store.Record(ctx, entry("f", flaghistory.ActionAdded, "h1"))
			assert.NoError(t, err)
			if ok {
				mu.Lock()
				inserted++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, 1, inserted)
}

func TestPostgresStore_ListFiltersAndPagination(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	for _, k := range []string{"a", "b", "c", "d", "e"} {
		_, err := store.Record(ctx, entry(k, flaghistory.ActionAdded, "h-"+k))
		require.NoError(t, err)
	}
	other := entry("a", flaghistory.ActionUpdated, "h-a2")
	other.Author = "john"
	other.Flagset = "other"
	_, err := store.Record(ctx, other)
	require.NoError(t, err)

	first, err := store.List(ctx, flaghistory.Filter{Flagset: "default", Limit: 2})
	require.NoError(t, err)
	require.Len(t, first.Entries, 2)
	assert.Equal(t, "e", first.Entries[0].FlagKey)
	assert.NotZero(t, first.NextCursor)

	second, err := store.List(ctx, flaghistory.Filter{Flagset: "default", Limit: 2, Cursor: first.NextCursor})
	require.NoError(t, err)
	assert.Equal(t, []string{"c", "b"}, []string{second.Entries[0].FlagKey, second.Entries[1].FlagKey})

	last, err := store.List(ctx, flaghistory.Filter{Flagset: "default", Limit: 2, Cursor: second.NextCursor})
	require.NoError(t, err)
	require.Len(t, last.Entries, 1)
	assert.Zero(t, last.NextCursor)

	byAuthor, err := store.List(ctx, flaghistory.Filter{Author: "john"})
	require.NoError(t, err)
	assert.Len(t, byAuthor.Entries, 1)

	byAction, err := store.List(ctx, flaghistory.Filter{Action: flaghistory.ActionUpdated})
	require.NoError(t, err)
	assert.Len(t, byAction.Entries, 1)

	future, err := store.List(ctx, flaghistory.Filter{From: time.Now().Add(time.Hour)})
	require.NoError(t, err)
	assert.Empty(t, future.Entries)

	latest, err := store.LatestPerFlag(ctx, "default")
	require.NoError(t, err)
	assert.Len(t, latest, 5)
	one, err := store.Latest(ctx, "default", "zzz")
	require.NoError(t, err)
	assert.Nil(t, one)
}

func TestPostgresStore_PurgeKeepsTheBaselineOfLiveFlags(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	for _, e := range []flaghistory.Entry{
		entry("live", flaghistory.ActionAdded, "h1"),
		entry("live", flaghistory.ActionUpdated, "h2"),
		entry("gone", flaghistory.ActionAdded, "g1"),
		entry("gone", flaghistory.ActionDeleted, ""),
	} {
		_, err := store.Record(ctx, e)
		require.NoError(t, err)
	}

	deleted, err := store.Purge(ctx, time.Now().Add(time.Hour))
	require.NoError(t, err)
	assert.Equal(t, int64(3), deleted)

	page, err := store.List(ctx, flaghistory.Filter{})
	require.NoError(t, err)
	require.Len(t, page.Entries, 1)
	assert.Equal(t, "live", page.Entries[0].FlagKey)
	assert.Equal(t, "h2", page.Entries[0].AfterHash)
}
