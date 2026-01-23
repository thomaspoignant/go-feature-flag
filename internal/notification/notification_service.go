package notification

import (
	"log/slog"
	"sync"

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

func NewService(notifiers []notifier.Notifier) Service {
	return &notificationService{
		Notifiers: notifiers,
		waitGroup: &sync.WaitGroup{},
	}
}

type notificationService struct {
	Notifiers []notifier.Notifier
	waitGroup *sync.WaitGroup
}

// Notify is sending the notification to the notifiers.
// oldCache is the old cache of the flags.
// newCache is the new cache of the flags.
// log is the logger to use.
func (c *notificationService) Notify(oldCache, newCache map[string]flag.Flag, log *fflog.FFLogger) {
	diff := c.getDifferences(oldCache, newCache)
	if diff.HasDiff() {
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
