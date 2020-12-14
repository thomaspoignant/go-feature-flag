package ffclient

import (
	"github.com/go-co-op/gocron"
	"log"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/flags"
)

type Flags map[string]flags.Flag

var flagUpdater gocron.Scheduler

// Init the feature flag component.
func Init(config Config) {
	cache.Init()
	err := retrieveFlagsAndUpdateCache(config)
	if err != nil {
		log.Fatalf("impossible to retrieve the flags, please check your configuration")
	}
	flagUpdater = *gocron.NewScheduler(time.UTC)
	_, err = flagUpdater.Every(uint64(config.PollInterval)).Seconds().Do(retrieveFlagsAndUpdateCache, config)
	if err != nil {
		log.Fatalf("impossible to launch background updater")
	}

	flagUpdater.StartAsync()
}

func retrieveFlagsAndUpdateCache(config Config) error {
	loadedFlags, err := config.GetRetriever().Retrieve()
	if err != nil {
		return err
	}

	err = cache.UpdateCache(loadedFlags)
	if err != nil {
		return err
	}

	return nil
}

// Close the component by stopping the background refresh.
func Close() {
	flagUpdater.Stop()
}
