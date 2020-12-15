package ffclient

import (
	"fmt"
	"github.com/go-co-op/gocron"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/flags"
)

type Flags map[string]flags.Flag

var flagUpdater gocron.Scheduler

// Init the feature flag component.
func Init(config Config) error {
	cache.Init()
	err := retrieveFlagsAndUpdateCache(config)
	if err != nil {
		return fmt.Errorf("impossible to retrieve the flags, please check your configuration: %v", err)
	}
	flagUpdater = *gocron.NewScheduler(time.UTC)

	// The default value for poll interval is 1
	if config.PollInterval == 0 {
		config.PollInterval = 1
	}
	_, err = flagUpdater.Every(uint64(config.PollInterval)).Seconds().Do(retrieveFlagsAndUpdateCache, config)
	if err != nil {
		return fmt.Errorf("impossible to launch background updater: %v", err)
	}

	flagUpdater.StartAsync()
	return nil
}

func retrieveFlagsAndUpdateCache(config Config) error {
	retriever, err := config.GetRetriever()
	if err != nil {
		return err
	}

	loadedFlags, err := retriever.Retrieve()
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
