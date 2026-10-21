package ffclient

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

func TestTTLWatcher_check(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	short := "1d"
	flags := map[string]flag.Flag{
		"no-created-at": &flag.InternalFlag{},
		"short-ttl":     &flag.InternalFlag{CreatedAt: &created, TTL: &short},
		"default-ttl":   &flag.InternalFlag{CreatedAt: &created},
	}
	w := newTTLWatcher(
		30*24*time.Hour,
		func(context.Context) (map[string]flag.Flag, error) { return flags, nil },
		&fflog.FFLogger{},
	)

	tests := []struct {
		name string
		now  time.Time
		want []string
	}{
		{name: "nothing expired", now: created.Add(time.Hour), want: []string{}},
		{name: "short ttl expired", now: created.AddDate(0, 0, 2), want: []string{"short-ttl"}},
		{name: "default ttl expired", now: created.AddDate(0, 0, 31), want: []string{"default-ttl", "short-ttl"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w.check(context.Background(), tt.now)
			assert.ElementsMatch(t, tt.want, w.expiredFlags())
		})
	}
}

func TestTTLWatcher_startStop(t *testing.T) {
	w := newTTLWatcher(0,
		func(context.Context) (map[string]flag.Flag, error) { return nil, nil },
		&fflog.FFLogger{})
	w.start(context.Background())
	w.stopAndWait()
	assert.NotPanics(t, w.stopAndWait)
}

func TestGoFeatureFlag_ExpiredFlag(t *testing.T) {
	flagFile := filepath.Join(t.TempDir(), "flags.yaml")
	content := `
expired-flag:
  createdAt: 2020-01-01T00:00:00Z
  ttl: 1d
  variations: {on: true, off: false}
  defaultRule: {variation: on}
fresh-flag:
  variations: {on: true, off: false}
  defaultRule: {variation: on}
`
	require.NoError(t, os.WriteFile(flagFile, []byte(content), 0o600))

	g, err := New(Config{
		PollingInterval: time.Minute,
		Retriever:       &fileretriever.Retriever{Path: flagFile},
	})
	require.NoError(t, err)
	defer g.Close()
	ctx := ffcontext.NewEvaluationContext("user")

	res, _ := g.BoolVariationDetails("expired-flag", ctx, false)
	assert.False(t, res.Value)
	assert.Equal(t, flag.ErrorCodeFlagExpired, res.ErrorCode)
	assert.Contains(t, res.ErrorDetails, `flag "expired-flag" expired on 2020-01-02T00:00:00Z`)

	res, err = g.BoolVariationDetails("fresh-flag", ctx, false)
	assert.NoError(t, err)
	assert.True(t, res.Value)

	g.ttlWatcher.check(context.Background(), time.Now())
	assert.Equal(t, []string{"expired-flag"}, g.GetExpiredFlags())
}

func TestTTLWatcher_check_dependencies(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	short := "1d"
	follows := func(follow bool) *[]flag.FlagDependency {
		on := "on"
		return &[]flag.FlagDependency{{Flag: "parent", Variation: &on, ExpiresWithParent: follow}}
	}
	flags := map[string]flag.Flag{
		"parent":        &flag.InternalFlag{CreatedAt: &created, TTL: &short},
		"follower":      &flag.InternalFlag{CreatedAt: &created, DependsOn: follows(true)},
		"not-follower":  &flag.InternalFlag{CreatedAt: &created, DependsOn: follows(false)},
		"no-created-at": &flag.InternalFlag{DependsOn: follows(true)},
	}
	w := newTTLWatcher(
		30*24*time.Hour,
		func(context.Context) (map[string]flag.Flag, error) { return flags, nil },
		&fflog.FFLogger{},
	)

	w.check(context.Background(), created.AddDate(0, 0, 2))
	assert.Equal(t, []string{"follower", "no-created-at", "parent"}, w.expiredFlags())
}

func TestGoFeatureFlag_FlagDependencies(t *testing.T) {
	flagFile := filepath.Join(t.TempDir(), "flags.yaml")
	content := `
parent:
  variations: {on: true, off: false}
  defaultRule: {variation: off}
child:
  variations: {on: true, off: false}
  defaultRule: {variation: on}
  dependsOn:
    - flag: parent
      variation: "on"
      appliesTo: ["on"]
orphan:
  variations: {on: true, off: false}
  defaultRule: {variation: on}
  dependsOn:
    - flag: unknown
      notVariation: "off"
expiring-child:
  createdAt: 2020-01-01T00:00:00Z
  ttl: 36500d
  variations: {on: true, off: false}
  defaultRule: {variation: on}
  dependsOn:
    - flag: expired-parent
      notVariation: "off"
      expiresWithParent: true
expired-parent:
  expiresAt: 2020-01-02T00:00:00Z
  variations: {on: true, off: false}
  defaultRule: {variation: on}
`
	require.NoError(t, os.WriteFile(flagFile, []byte(content), 0o600))
	g, err := New(Config{PollingInterval: time.Minute, Retriever: &fileretriever.Retriever{Path: flagFile}})
	require.NoError(t, err)
	defer g.Close()
	ctx := ffcontext.NewEvaluationContext("user")

	res, _ := g.BoolVariationDetails("child", ctx, false)
	assert.False(t, res.Value)
	assert.Equal(t, flag.ErrorCodeFlagDependency, res.ErrorCode)
	assert.False(t, res.Cacheable)

	res, _ = g.BoolVariationDetails("orphan", ctx, false)
	assert.Equal(t, flag.ErrorCodeFlagDependency, res.ErrorCode)

	res, _ = g.BoolVariationDetails("expiring-child", ctx, false)
	assert.Equal(t, flag.ErrorCodeFlagExpired, res.ErrorCode)

	states := g.AllFlagsState(ctx)
	assert.Equal(t, flag.ErrorCodeFlagDependency, states.GetFlags()["child"].ErrorCode)
	assert.Equal(t, flag.ErrorCodeFlagExpired, states.GetFlags()["expiring-child"].ErrorCode)

	g.ttlWatcher.check(context.Background(), time.Now())
	assert.Equal(t, []string{"expired-parent", "expiring-child"}, g.GetExpiredFlags())
}
