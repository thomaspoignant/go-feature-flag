package cache

import (
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"log/slog"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type Service interface {
	Close()
	Notify(oldCache map[string]flag.Flag, newCache map[string]flag.Flag, log *fflog.FFLogger)
}

func NewNotificationService(notifiers []notifier.Notifier) Service {
	return &notificationService{
		Notifiers: notifiers,
		waitGroup: &sync.WaitGroup{},
	}
}

type notificationService struct {
	Notifiers []notifier.Notifier
	waitGroup *sync.WaitGroup
}

func (c *notificationService) Notify(
	oldCache map[string]flag.Flag,
	newCache map[string]flag.Flag,
	log *fflog.FFLogger) {
	diff := c.getDifferences(oldCache, newCache)
	if diff.HasDiff() {
		for _, n := range c.Notifiers {
			c.waitGroup.Add(1)
			notif := n
			go func() {
				defer c.waitGroup.Done()
				err := notif.Notify(diff)
				if err != nil {
					log.Error("error while calling the notifier", slog.Any("err", err))
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
	oldCache map[string]flag.Flag, newCache map[string]flag.Flag,
) notifier.DiffCache {
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
