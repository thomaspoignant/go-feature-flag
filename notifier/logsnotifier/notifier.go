package logsnotifier

import (
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"log"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type Notifier struct {
	Logger *log.Logger
}

func (c *Notifier) Notify(diff notifier.DiffCache, wg *sync.WaitGroup) error {
	defer wg.Done()
	for key := range diff.Deleted {
		fflog.Printf(c.Logger, "flag %v removed\n", key)
	}

	for key := range diff.Added {
		fflog.Printf(c.Logger, "flag %v added\n", key)
	}

	for key, flagDiff := range diff.Updated {
		if flagDiff.After.IsDisable() != flagDiff.Before.IsDisable() {
			if flagDiff.After.IsDisable() {
				// Flag is disabled
				fflog.Printf(c.Logger, "flag %v is turned OFF\n", key)
				continue
			}
			fflog.Printf(c.Logger, "flag %v is turned ON\n", key)
			continue
		}
		// key has changed in cache
		fflog.Printf(c.Logger, "flag %s updated\n", key)
	}

	return nil
}
