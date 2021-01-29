package cache

import (
	"github.com/google/go-cmp/cmp"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/internal/flags"
)

type Service interface {
	Close()
	Notify(oldCache FlagsCache, newCache FlagsCache)
}

func NewService(notifiers []Notifier) Service {
	return &notificationService{
		Notifiers: notifiers,
		waitGroup: &sync.WaitGroup{},
	}
}

type notificationService struct {
	Notifiers []Notifier
	waitGroup *sync.WaitGroup
}

func (c *notificationService) Notify(oldCache FlagsCache, newCache FlagsCache) {
	diff := c.getDifferences(oldCache, newCache)
	if diff.hasDiff() {
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
	oldCache FlagsCache, newCache FlagsCache) diffCache {
	diff := diffCache{
		Deleted: map[string]flags.Flag{},
		Added:   map[string]flags.Flag{},
		Updated: map[string]diffUpdated{},
	}
	for key := range oldCache {
		_, inNewCache := newCache[key]
		if !inNewCache {
			diff.Deleted[key] = oldCache[key]
			continue
		}

		if !cmp.Equal(oldCache[key], newCache[key]) {
			diff.Updated[key] = diffUpdated{
				Before: oldCache[key],
				After:  newCache[key],
			}
		}
	}

	for key := range newCache {
		if _, inOldCache := oldCache[key]; !inOldCache {
			diff.Added[key] = newCache[key]
		}
	}
	return diff
}
