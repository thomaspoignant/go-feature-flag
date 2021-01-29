package cache

import (
	"log"
	"sync"
	"time"
)

type LogNotifier struct {
	Logger *log.Logger
}

func (c *LogNotifier) Notify(diff diffCache, wg *sync.WaitGroup) {
	defer wg.Done()
	date := time.Now().Format(time.RFC3339)
	for key := range diff.Deleted {
		c.Logger.Printf("[%v] flag %v removed\n", date, key)
	}

	for key := range diff.Added {
		c.Logger.Printf("[%v] flag %v added\n", date, key)
	}

	for key, flagDiff := range diff.Updated {
		if flagDiff.After.Disable != flagDiff.Before.Disable {
			if flagDiff.After.Disable {
				// Flag is disabled
				c.Logger.Printf("[%v] flag %v is turned OFF\n", date, key)
				continue
			}
			c.Logger.Printf("[%v] flag %v is turned ON (flag=[%v])  \n", date, key, flagDiff.After)
			continue
		}
		// key has changed in cache
		c.Logger.Printf("[%v] flag %s updated, old=[%v], new=[%v]\n", date, key, flagDiff.Before, flagDiff.After)
	}
}
