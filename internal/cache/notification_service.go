package cache

import (
	"github.com/google/go-cmp/cmp"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/internal/notifier"
)

type Service interface {
	Close()
	Notify(oldCache FlagsCache, newCache FlagsCache)
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

func (c *notificationService) Notify(oldCache FlagsCache, newCache FlagsCache) {
	diff := c.getDifferences(oldCache, newCache)
	if diff.HasDiff() {
		for _, notifier := range c.Notifiers {
			c.waitGroup.Add(1)
			go notifier.Notify(diff, c.waitGroup)
		}
	}
}

func (c *notificationService) Close() {
	c.waitGroup.Wait()
}

// getDifferences is checking what are the difference in the updated cache.
func (c *notificationService) getDifferences(
	oldCache FlagsCache, newCache FlagsCache) model.DiffCache {
	diff := model.DiffCache{
		Deleted: map[string]flag.Flag{},
		Added:   map[string]flag.Flag{},
		Updated: map[string]model.DiffUpdated{},
	}
	for key := range oldCache {
		newFlag, inNewCache := newCache[key]
		oldFlag := oldCache[key]
		if !inNewCache {
			diff.Deleted[key] = &oldFlag
			continue
		}

		if !cmp.Equal(oldCache[key], newCache[key]) {
			diff.Updated[key] = model.DiffUpdated{
				Before: &oldFlag,
				After:  &newFlag,
			}
		}
	}

	for key := range newCache {
		if _, inOldCache := oldCache[key]; !inOldCache {
			f := newCache[key]
			diff.Added[key] = &f
		}
	}
	return diff
}
