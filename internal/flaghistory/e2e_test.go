//go:build docker

package flaghistory_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/internal/flaghistory"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/retriever"
)

const flagV1 = `
my-flag:
  variations: {on: true, off: false}
  defaultRule: {variation: on}
`

const flagV2 = `
my-flag:
  variations: {on: true, off: false}
  defaultRule: {variation: off}
other-flag:
  variations: {on: true, off: false}
  defaultRule: {variation: on}
`

// authoredRetriever serves a flag file and knows who modified it.
type authoredRetriever struct {
	content atomic.Value
	author  atomic.Value
}

func (r *authoredRetriever) set(content, author string) {
	r.author.Store(author)
	r.content.Store(content)
}

func (r *authoredRetriever) Retrieve(context.Context) ([]byte, error) {
	return []byte(r.content.Load().(string)), nil
}

func (r *authoredRetriever) LastChange(context.Context) (*retriever.ChangeInfo, error) {
	return &retriever.ChangeInfo{
		Author: r.author.Load().(string), Reference: "sha-" + r.author.Load().(string), ChangedAt: time.Now(),
	}, nil
}

func startHistory(t *testing.T, store flaghistory.Store) *flaghistory.History {
	t.Helper()
	h := flaghistory.New(flaghistory.Config{
		Connect: func(context.Context) (flaghistory.Store, error) { return store, nil },
		Source:  "e2e",
	})
	h.Start()
	t.Cleanup(h.Close)
	return h
}

func startClient(t *testing.T, r retriever.Retriever, h *flaghistory.History) *ffclient.GoFeatureFlag {
	t.Helper()
	client, err := ffclient.New(ffclient.Config{
		Retriever:  r,
		FileFormat: "yaml",
		Notifiers:  []notifier.Notifier{h.NotifierFor("default")},
	})
	require.NoError(t, err)
	t.Cleanup(client.Close)
	return client
}

func TestHistory_endToEnd(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	r := &authoredRetriever{}
	r.set(flagV1, "jane")

	h := startHistory(t, store)
	client := startClient(t, r, h)

	// first load: the flag is added by jane
	require.Eventually(t, func() bool {
		p, _ := store.List(ctx, flaghistory.Filter{})
		return len(p.Entries) == 1
	}, 5*time.Second, 50*time.Millisecond)

	// john modifies one flag and adds another
	r.set(flagV2, "john")
	require.True(t, client.ForceRefresh())
	require.Eventually(t, func() bool {
		p, _ := store.List(ctx, flaghistory.Filter{})
		return len(p.Entries) == 3
	}, 5*time.Second, 50*time.Millisecond)

	byJohn, err := store.List(ctx, flaghistory.Filter{Author: "john"})
	require.NoError(t, err)
	require.Len(t, byJohn.Entries, 2)
	actions := map[string]flaghistory.Action{}
	for _, e := range byJohn.Entries {
		actions[e.FlagKey] = e.Action
		assert.Equal(t, "sha-john", e.ChangeRef)
		assert.Equal(t, "e2e", e.Source)
	}
	assert.Equal(t, map[string]flaghistory.Action{
		"my-flag": flaghistory.ActionUpdated, "other-flag": flaghistory.ActionAdded}, actions)

	updated, err := store.Latest(ctx, "default", "my-flag")
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Len(t, updated.Diff.Changes, 1)
	assert.Equal(t, "defaultRule.variation", updated.Diff.Changes[0].Path)
	assert.Equal(t, "on", updated.Diff.Changes[0].Old)
	assert.Equal(t, "off", updated.Diff.Changes[0].New)

	// a second relay proxy instance starts on the same database with the same flags:
	// nothing new must be recorded.
	h2 := startHistory(t, store)
	client2 := startClient(t, r, h2)
	flags, err := client2.GetFlagsFromCache()
	require.NoError(t, err)
	h2.Reconcile("default", flags)
	time.Sleep(500 * time.Millisecond)
	all, err := store.List(ctx, flaghistory.Filter{})
	require.NoError(t, err)
	assert.Len(t, all.Entries, 3, "a restart must not duplicate the history")
}

func TestHistory_reconcileAfterDowntime(t *testing.T) {
	store, uri := newStoreWithURI(t)
	ctx := context.Background()
	r := &authoredRetriever{}
	r.set(flagV1, "jane")

	h := startHistory(t, store)
	client := startClient(t, r, h)
	require.Eventually(t, func() bool {
		p, _ := store.List(ctx, flaghistory.Filter{})
		return len(p.Entries) == 1
	}, 5*time.Second, 50*time.Millisecond)
	client.Close()
	h.Close() // closes the connection, like a stopped relay proxy

	// the file is modified while the relay proxy is stopped, my-flag is removed and another added.
	r.set("new-flag:\n  variations: {on: true, off: false}\n  defaultRule: {variation: on}\n", "john")
	restartedStore, err := flaghistory.NewPostgresStore(ctx, uri, "")
	require.NoError(t, err)
	store = restartedStore
	h2 := startHistory(t, store)
	client2 := startClient(t, r, h2)
	flags, err := client2.GetFlagsFromCache()
	require.NoError(t, err)
	h2.Reconcile("default", flags)

	require.Eventually(t, func() bool {
		p, _ := store.List(ctx, flaghistory.Filter{})
		return len(p.Entries) == 3
	}, 5*time.Second, 50*time.Millisecond)

	deleted, err := store.Latest(ctx, "default", "my-flag")
	require.NoError(t, err)
	assert.Equal(t, flaghistory.ActionDeleted, deleted.Action)
	assert.True(t, deleted.Diff.Reconciled)
	added, err := store.Latest(ctx, "default", "new-flag")
	require.NoError(t, err)
	assert.Equal(t, flaghistory.ActionAdded, added.Action)
}
