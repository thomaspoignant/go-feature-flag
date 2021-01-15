package ffclient

import (
	"fmt"
	"github.com/go-co-op/gocron"
	"log"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/cache"
)

// Init the feature flag component with the configuration of ffclient.Config
//  func main() {
//    err := ffclient.Init(ffclient.Config{
//             PollInterval: 3,
//             Retriever: &ffClient.HTTPRetriever{
//               URL:    "http://example.com/test.yaml",
//             },
//           })
//    defer ffclient.Close()
func Init(config Config) error {
	var err error = nil
	onceFF.Do(func() {
		ff = newGoFeatureFlag(config)
		err = ff.startUpdater()
	})
	return err
}

// Close the component by stopping the background refresh and clean the cache.
func Close() {
	ff.cache.Close()
	ff.flagUpdater.Stop()
}

// goFeatureFlag is the main object of the library
// it contains the cache, the config and the update.
type goFeatureFlag struct {
	flagUpdater gocron.Scheduler
	cache       cache.Cache
	config      Config
}

// ff is the default object for go-feature-flag
var ff *goFeatureFlag
var onceFF sync.Once

func newGoFeatureFlag(config Config) *goFeatureFlag {
	flagUpdater := *gocron.NewScheduler(time.UTC)

	// The default value for poll interval is 60 seconds
	if config.PollInterval == 0 {
		config.PollInterval = 60
	}

	return &goFeatureFlag{
		cache:       cache.New(config.Logger),
		flagUpdater: flagUpdater,
		config:      config,
	}
}

func (g *goFeatureFlag) startUpdater() error {
	// fail if we cannot retrieve the flags the 1st time
	err := retrieveFlagsAndUpdateCache(g.config)
	if err != nil {
		return fmt.Errorf("impossible to retrieve the flags, please check your configuration: %v", err)
	}

	if g.config.PollInterval < 0 {
		return fmt.Errorf("%d is not a valid PollInterval value, it need to be > 0", g.config.PollInterval)
	}

	// start flag updater
	_, err = g.flagUpdater.Every(uint64(g.config.PollInterval)).Seconds().Do(retrieveFlagsAndUpdateCache, g.config)
	if err != nil {
		return fmt.Errorf("impossible to launch background updater: %v", err)
	}
	g.flagUpdater.StartAsync()
	return nil
}

// retrieveFlagsAndUpdateCache is called every X seconds to refresh the cache flag.
func retrieveFlagsAndUpdateCache(config Config) error {
	retriever, err := config.GetRetriever()
	if err != nil {
		log.Printf("error while getting the file retriever: %v", err)
		return err
	}

	loadedFlags, err := retriever.Retrieve(config.Context)
	if err != nil {
		log.Printf("error: impossible to retrieve flags from the config file: %v", err)
		return err
	}

	err = ff.cache.UpdateCache(loadedFlags)
	if err != nil {
		log.Printf("error: impossible to update the cache of the flags: %v", err)
		return err
	}
	return nil
}
