package notifier

import (
	"log"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
)

type LogNotifier struct {
	Logger *log.Logger
}

func (c *LogNotifier) Notify(diff model.DiffCache, wg *sync.WaitGroup) {
	defer wg.Done()
	for key := range diff.Deleted {
		fflog.Printf(c.Logger, "flag %v removed\n", key)
	}

	for key := range diff.Added {
		fflog.Printf(c.Logger, "flag %v added\n", key)
	}

	for key, flagDiff := range diff.Updated {
		if flagDiff.After.GetDisable() != flagDiff.Before.GetDisable() {
			if flagDiff.After.GetDisable() {
				// Flag is disabled
				fflog.Printf(c.Logger, "flag %v is turned OFF\n", key)
				continue
			}
			fflog.Printf(c.Logger, "flag %v is turned ON (flag=[%v])  \n", key, flagDiff.After)
			continue
		}
		// key has changed in cache
		fflog.Printf(c.Logger, "flag %s updated, old=[%v], new=[%v]\n", key, flagDiff.Before, flagDiff.After)
	}
}
