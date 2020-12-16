package ffclient

import (
	"fmt"
	"github.com/go-co-op/gocron"
	"log"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/cache"
)

var flagUpdater gocron.Scheduler

// Init the feature flag component with the configuration of ffclient.Config
//  func main() {
//    err := ffclient.Init(ffclient.Config{
//             PollInterval: 3,
//             HTTPRetriever: &ffClient.HTTPRetriever{
//               URL:    "http://example.com/test.yaml",
//             },
//           })
//    defer ffclient.Close()
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
		log.Printf("error while getting the file retriever: %v", err)
		return err
	}

	loadedFlags, err := retriever.Retrieve()
	if err != nil {
		log.Printf("error: impossible to retrieve flags from the config file: %v", err)
		return err
	}

	err = cache.UpdateCache(loadedFlags)
	if err != nil {
		log.Printf("error: impossible to update the cache of the flags: %v", err)
		return err
	}

	return nil
}

// Close the component by stopping the background refresh and clean the cache.
func Close() {
	cache.Close()
	flagUpdater.Stop()
}
