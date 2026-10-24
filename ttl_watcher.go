package ffclient

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

const ttlCheckInterval = 5 * time.Second

// expirable is implemented by the flags that have a TTL.
type expirable interface {
	IsExpired(now time.Time, defaultTTL time.Duration) bool
}

// ttlWatcher detects the flags for which the TTL is over and keeps track of them.
// The evaluation does not depend on it (the flag checks its own expiration), the watcher
// is what makes the transition to the expired state visible.
type ttlWatcher struct {
	defaultTTL time.Duration
	getFlags   func(ctx context.Context) (map[string]flag.Flag, error)
	logger     *fflog.FFLogger

	mutex   sync.RWMutex
	expired map[string]struct{}
	stop    chan struct{}
	done    chan struct{}
}

func newTTLWatcher(
	defaultTTL time.Duration,
	getFlags func(ctx context.Context) (map[string]flag.Flag, error),
	logger *fflog.FFLogger,
) *ttlWatcher {
	return &ttlWatcher{
		defaultTTL: defaultTTL,
		getFlags:   getFlags,
		logger:     logger,
		expired:    map[string]struct{}{},
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
	}
}

// start runs the check every ttlCheckInterval until stopAndWait is called.
func (w *ttlWatcher) start(ctx context.Context) {
	go func() {
		defer close(w.done)
		ticker := time.NewTicker(ttlCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-w.stop:
				return
			case <-ticker.C:
				w.check(ctx, time.Now())
			}
		}
	}()
}

func (w *ttlWatcher) stopAndWait() {
	close(w.stop)
	<-w.done
}

// check refreshes the set of expired flags.
func (w *ttlWatcher) check(ctx context.Context, now time.Time) {
	flags, err := w.getFlags(ctx)
	if err != nil {
		return
	}
	expired := make(map[string]struct{})
	for key, f := range flags {
		if e, ok := f.(expirable); ok && e.IsExpired(now, w.defaultTTL) {
			expired[key] = struct{}{}
		}
	}

	w.mutex.Lock()
	previous := w.expired
	w.expired = expired
	w.mutex.Unlock()

	for key := range expired {
		if _, known := previous[key]; !known {
			w.logger.Warn("flag TTL expired, evaluations now return an error", slog.String("flag", key))
		}
	}
}

// expiredFlags returns the sorted names of the expired flags.
func (w *ttlWatcher) expiredFlags() []string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return slices.Sorted(maps.Keys(w.expired))
}
