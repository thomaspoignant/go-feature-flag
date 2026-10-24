package notification

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

var _ Service = (*notificationService)(nil)

type Service interface {
	Close()
	Notify(oldCache, newCache map[string]flag.Flag, log *fflog.FFLogger)
}

// changeInfoTimeout is the maximum time we wait for the origin of a change.
const changeInfoTimeout = 5 * time.Second

// ChangeInfoResolver returns who made the last change on the flag configuration.
// It returns nil when the information is not available.
type ChangeInfoResolver func(ctx context.Context) (*notifier.ChangeInfo, error)

// Option configures the notification service.
type Option func(*notificationService)

// WithChangeInfoResolver sets the function used to find the origin of a change.
// The resolver is called only when a difference is detected.
func WithChangeInfoResolver(r ChangeInfoResolver) Option {
	return func(n *notificationService) { n.changeInfoResolver = r }
}

func NewService(notifiers []notifier.Notifier, opts ...Option) Service {
	svc := &notificationService{
		Notifiers: notifiers,
		waitGroup: &sync.WaitGroup{},
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

type notificationService struct {
	Notifiers          []notifier.Notifier
	waitGroup          *sync.WaitGroup
	changeInfoResolver ChangeInfoResolver
}

// Notify is sending the notification to the notifiers.
// oldCache is the old cache of the flags.
// newCache is the new cache of the flags.
// log is the logger to use.
func (c *notificationService) Notify(oldCache, newCache map[string]flag.Flag, log *fflog.FFLogger) {
	diff := c.getDifferences(oldCache, newCache)
	if diff.HasDiff() {
		diff.Change = c.resolveChangeInfo(log)
		for _, n := range c.Notifiers {
			c.waitGroup.Add(1)
			notif := n
			go func() {
				defer c.waitGroup.Done()
				err := notif.Notify(diff)
				if err != nil {
					log.Error("error while calling the notifier", slog.Any("error", err.Error()))
				}
			}()
		}
	}
}

// resolveChangeInfo finds the origin of the change, a failure never blocks the notifications.
func (c *notificationService) resolveChangeInfo(log *fflog.FFLogger) *notifier.ChangeInfo {
	if c.changeInfoResolver == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), changeInfoTimeout)
	defer cancel()
	info, err := c.changeInfoResolver(ctx)
	if err != nil {
		log.Warn("impossible to find the origin of the change", slog.Any("error", err.Error()))
		return nil
	}
	return info
}

func (c *notificationService) Close() {
	c.waitGroup.Wait()
}

// getDifferences is checking what are the difference in the updated cache.
func (c *notificationService) getDifferences(
	oldCache, newCache map[string]flag.Flag) notifier.DiffCache {
	diff := notifier.DiffCache{
		Deleted: map[string]flag.Flag{},
		Added:   map[string]flag.Flag{},
		Updated: map[string]notifier.DiffUpdated{},
	}
	for key := range oldCache {
		newFlag, inNewCache := newCache[key]
		oldFlag := oldCache[key]
		if !inNewCache {
			diff.Deleted[key] = oldFlag
			continue
		}

		if !cmp.Equal(oldCache[key], newCache[key]) {
			diff.Updated[key] = notifier.DiffUpdated{
				Before: oldFlag,
				After:  newFlag,
			}
		}
	}

	for key := range newCache {
		if _, inOldCache := oldCache[key]; !inOldCache {
			f := newCache[key]
			diff.Added[key] = f
		}
	}
	return diff
}
