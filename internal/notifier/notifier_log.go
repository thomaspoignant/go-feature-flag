package notifier

import (
	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"sync"
	"time"
)

type LogNotifier struct {
	Logger fflog.Logger
}

func (c *LogNotifier) Notify(diff model.DiffCache, wg *sync.WaitGroup) {
	defer wg.Done()
	for key := range diff.Deleted {
		c.Logger.Printf("[%v] flag %v removed\n", time.Now().Format(fflog.LogDateFormat), key)
	}

	for key := range diff.Added {
		c.Logger.Printf("[%v] flag %v added\n", time.Now().Format(fflog.LogDateFormat), key)
	}

	for key, flagDiff := range diff.Updated {
		if flagDiff.After.IsDisable() != flagDiff.Before.IsDisable() {
			if flagDiff.After.IsDisable() {
				// Flag is disabled
				c.Logger.Printf("[%v] flag %v is turned OFF\n", time.Now().Format(fflog.LogDateFormat), key)
				continue
			}
			c.Logger.Printf("[%v] flag %v is turned ON (flag=[%v])  \n",
				time.Now().Format(fflog.LogDateFormat), key, flagDiff.After)
			continue
		}
		// key has changed in cache
		c.Logger.Printf("[%v] flag %s updated, old=[%v], new=[%v]\n",
			time.Now().Format(fflog.LogDateFormat), key, flagDiff.Before, flagDiff.After)
	}
}
