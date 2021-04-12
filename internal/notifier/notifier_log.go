package notifier

import (
	"log"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
)

type LogNotifier struct {
	Logger *log.Logger
}

func (c *LogNotifier) Notify(diff model.DiffCache, wg *sync.WaitGroup) {
	defer wg.Done()
	date := time.Now().Format(time.RFC3339)
	for key := range diff.Deleted {
		fflog.Printf(c.Logger, "[%v] flag %v removed\n", date, key)
	}

	for key := range diff.Added {
		fflog.Printf(c.Logger, "[%v] flag %v added\n", date, key)
	}

	for key, flagDiff := range diff.Updated {
		if flagDiff.After.Disable != flagDiff.Before.Disable {
			if flagDiff.After.Disable {
				// Flag is disabled
				fflog.Printf(c.Logger, "[%v] flag %v is turned OFF\n", date, key)
				continue
			}
			fflog.Printf(c.Logger, "[%v] flag %v is turned ON (flag=[%v])  \n", date, key, flagDiff.After)
			continue
		}
		// key has changed in cache
		fflog.Printf(c.Logger, "[%v] flag %s updated, old=[%v], new=[%v]\n", date, key, flagDiff.Before, flagDiff.After)
	}
}
