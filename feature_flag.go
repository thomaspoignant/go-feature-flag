package ffclient

import (
	"fmt"
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
		ff, err = New(config)
	})
	return err
}

// Close the component by stopping the background refresh and clean the cache.
func Close() {
	ff.Close()
}

// GoFeatureFlag is the main object of the library
// it contains the cache, the config and the update.
type GoFeatureFlag struct {
	cache     cache.Cache
	config    Config
	bgUpdater backgroundUpdater
}

// ff is the default object for go-feature-flag
var ff *GoFeatureFlag
var onceFF sync.Once

// New creates a new go-feature-flag instance that retrieve the config from a YAML file
// and return everything you need to manage your flags.
func New(config Config) (*GoFeatureFlag, error) {
	// The default value for poll interval is 60 seconds
	if config.PollInterval == 0 {
		config.PollInterval = 60
	}

	// Check that value is not negative
	if config.PollInterval < 0 {
		return nil, fmt.Errorf("%d is not a valid PollInterval value, it need to be > 0", config.PollInterval)
	}

	goFF := &GoFeatureFlag{
		config:    config,
		bgUpdater: newBackgroundUpdater(config.PollInterval),
	}
	goFF.cache = cache.New(cache.NewService(goFF.getNotifiers()))

	// fail if we cannot retrieve the flags the 1st time
	err := retrieveFlagsAndUpdateCache(goFF.config, goFF.cache)
	if err != nil {
		return nil, fmt.Errorf("impossible to retrieve the flags, please check your configuration: %v", err)
	}

	// start the flag update in background
	go goFF.startFlagUpdaterDaemon()

	return goFF, nil
}

func (g *GoFeatureFlag) Close() {
	// clear the cache
	g.cache.Close()

	// stop the background updater
	g.bgUpdater.close()
}

// startFlagUpdaterDaemon is the daemon that refresh the cache every X seconds.
func (g *GoFeatureFlag) startFlagUpdaterDaemon() {
	for {
		select {
		case <-g.bgUpdater.ticker.C:
			err := retrieveFlagsAndUpdateCache(g.config, g.cache)
			if err != nil && g.config.Logger != nil {
				g.config.Logger.Printf("[%v] error while updating the cache: %v\n", time.Now().Format(time.RFC3339), err)
			}
		case <-g.bgUpdater.updaterChan:
			return
		}
	}
}

// getNotifiers is creating Notifier from the config
func (g *GoFeatureFlag) getNotifiers() []cache.Notifier {
	var notifiers []cache.Notifier
	if g.config.Logger != nil {
		notifiers = append(notifiers, &cache.LogNotifier{Logger: g.config.Logger})
	}
	return notifiers
}

// retrieveFlagsAndUpdateCache is called every X seconds to refresh the cache flag.
func retrieveFlagsAndUpdateCache(config Config, cache cache.Cache) error {
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

	err = cache.UpdateCache(loadedFlags)
	if err != nil {
		log.Printf("error: impossible to update the cache of the flags: %v", err)
		return err
	}
	return nil
}
