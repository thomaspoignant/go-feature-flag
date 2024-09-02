package ffclient

import (
	"context"
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/internal/dto"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"

	"github.com/thomaspoignant/go-feature-flag/notifier/logsnotifier"

	"github.com/thomaspoignant/go-feature-flag/internal/cache"
)

// Init the feature flag component with the configuration of ffclient.Config
//
//	func main() {
//	  err := ffclient.Init(ffclient.Config{
//	           PollingInterval: 3 * time.Second,
//	           Retriever: &httpretriever.Retriever{
//	             URL:    "http://example.com/flag-config.yaml",
//	           },
//	         })
//	  defer ffclient.Close()
func Init(config Config) error {
	var err error
	onceFF.Do(func() {
		var tmpFF *GoFeatureFlag
		tmpFF, err = New(config)
		if err == nil {
			ff = tmpFF
		}
	})
	return err
}

// GoFeatureFlag is the main object of the library
// it contains the cache, the config, the updater and the exporter.
type GoFeatureFlag struct {
	cache            cache.Manager
	config           Config
	bgUpdater        backgroundUpdater
	dataExporter     *exporter.Scheduler
	retrieverManager *retriever.Manager
}

// ff is the default object for go-feature-flag
var ff *GoFeatureFlag
var onceFF sync.Once

// New creates a new go-feature-flag instances that retrieve the config from a YAML file
// and return everything you need to manage your flags.
func New(config Config) (*GoFeatureFlag, error) {
	switch {
	case config.PollingInterval == 0:
		// The default value for the poll interval is 60 seconds
		config.PollingInterval = 60 * time.Second
	case config.PollingInterval < 0:
		// Check that value is not negative
		return nil, fmt.Errorf("%d is not a valid PollingInterval value, it need to be > 0", config.PollingInterval)
	case config.PollingInterval < time.Second:
		// the minimum value for the polling policy is 1 second
		config.PollingInterval = time.Second
	default:
		// do nothing
	}

	if config.offlineMutex == nil {
		config.offlineMutex = &sync.RWMutex{}
	}

	config.internalLogger = &fflog.FFLogger{
		LeveledLogger: config.LeveledLogger,
		LegacyLogger:  config.Logger,
	}

	goFF := &GoFeatureFlag{
		config: config,
	}

	if !config.Offline {
		notifiers := config.Notifiers
		notifiers = append(notifiers, &logsnotifier.Notifier{Logger: config.internalLogger})

		notificationService := cache.NewNotificationService(notifiers)
		goFF.bgUpdater = newBackgroundUpdater(config.PollingInterval, config.EnablePollingJitter)
		goFF.cache = cache.New(notificationService, config.PersistentFlagConfigurationFile, config.internalLogger)

		retrievers, err := config.GetRetrievers()
		if err != nil {
			return nil, err
		}
		goFF.retrieverManager = retriever.NewManager(config.Context, retrievers, config.internalLogger)
		err = goFF.retrieverManager.Init(config.Context)
		if err != nil && !config.StartWithRetrieverError {
			return nil, fmt.Errorf("impossible to initialize the retrievers, please check your configuration: %v", err)
		}

		err = retrieveFlagsAndUpdateCache(goFF.config, goFF.cache, goFF.retrieverManager)
		if err != nil {
			switch {
			case config.PersistentFlagConfigurationFile != "":
				errPersist := retrievePersistentLocalDisk(config.Context, config, goFF)
				if errPersist != nil && !config.StartWithRetrieverError {
					return nil, fmt.Errorf("impossible to use the persistent flag configuration file: %v "+
						"[original error: %v]", errPersist, err)
				}
			case !config.StartWithRetrieverError:
				return nil, fmt.Errorf("impossible to retrieve the flags, please check your configuration: %v", err)
			default:
				// We accept to start with a retriever error, we will serve only default value
				goFF.config.internalLogger.Error("Impossible to retrieve the flags, starting with the "+
					"retriever error", slog.Any("error", err))
			}
		}

		go goFF.startFlagUpdaterDaemon()

		if goFF.config.DataExporter.Exporter != nil {
			// init the data exporter
			goFF.dataExporter = exporter.NewScheduler(goFF.config.Context, goFF.config.DataExporter.FlushInterval,
				goFF.config.DataExporter.MaxEventInMemory, goFF.config.DataExporter.Exporter, goFF.config.internalLogger)

			// we start the daemon only if we have a bulk exporter
			if goFF.config.DataExporter.Exporter.IsBulk() {
				go goFF.dataExporter.StartDaemon()
			}
		}
	}
	config.internalLogger.Debug("GO Feature Flag is initialized")
	return goFF, nil
}

// retrievePersistentLocalDisk is a function used in case we are not able to retrieve any flag when starting
// GO Feature Flag.
// This function will look at any pre-existent persistent configuration and start with it.
func retrievePersistentLocalDisk(ctx context.Context, config Config, goFF *GoFeatureFlag) error {
	if config.PersistentFlagConfigurationFile != "" {
		config.internalLogger.Error("Impossible to retrieve your flag configuration, trying to use the persistent"+
			" flag configuration file.", slog.String("path", config.PersistentFlagConfigurationFile))
		if _, err := os.Stat(config.PersistentFlagConfigurationFile); err == nil {
			// we found the configuration file on the disk
			r := &fileretriever.Retriever{Path: config.PersistentFlagConfigurationFile}

			fallBackRetrieverManager := retriever.NewManager(config.Context, []retriever.Retriever{r}, config.internalLogger)
			err := fallBackRetrieverManager.Init(ctx)
			if err != nil {
				return err
			}
			defer func() { _ = fallBackRetrieverManager.Shutdown(ctx) }()
			err = retrieveFlagsAndUpdateCache(goFF.config, goFF.cache, fallBackRetrieverManager)
			if err != nil {
				return err
			}
			return nil
		}
		config.internalLogger.Warn("No persistent flag configuration found",
			slog.String("path", config.PersistentFlagConfigurationFile))
	}
	return fmt.Errorf("no persistent flag available")
}

// Close wait until thread are done
func (g *GoFeatureFlag) Close() {
	if g != nil {
		if g.cache != nil {
			// clear the cache
			g.cache.Close()
		}
		if g.bgUpdater.updaterChan != nil && g.bgUpdater.ticker != nil {
			g.bgUpdater.close()
		}

		if g.dataExporter != nil {
			g.dataExporter.Close()
		}
		if g.retrieverManager != nil {
			_ = g.retrieverManager.Shutdown(g.config.Context)
		}
	}
}

// startFlagUpdaterDaemon is the daemon that refreshes the cache every X seconds.
func (g *GoFeatureFlag) startFlagUpdaterDaemon() {
	for {
		select {
		case <-g.bgUpdater.ticker.C:
			if !g.IsOffline() {
				err := retrieveFlagsAndUpdateCache(g.config, g.cache, g.retrieverManager)
				if err != nil {
					g.config.internalLogger.Error("Error while updating the cache.", slog.Any("error", err))
				}
			}
		case <-g.bgUpdater.updaterChan:
			return
		}
	}
}

// retrieveFlagsAndUpdateCache is called every X seconds to refresh the cache flag.
func retrieveFlagsAndUpdateCache(config Config, cache cache.Manager, retrieverManager *retriever.Manager) error {
	retrievers := retrieverManager.GetRetrievers()
	// Results is the type that will receive the results when calling
	// all the retrievers.
	type Results struct {
		Error error
		Value map[string]dto.DTO
		Index int
	}

	// resultsChan is the channel that will receive all the results.
	resultsChan := make(chan Results)
	var wg sync.WaitGroup
	wg.Add(len(retrievers))

	// Launching a goroutine that will wait until the waiting group is complete.
	// It closes the channel when ready
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for index, r := range retrievers {
		// Launching GO routines to retrieve all files in parallel.
		go func(r retriever.Retriever, format string, index int, ctx context.Context) {
			defer wg.Done()

			// If the retriever is not ready, we ignore it
			if rr, ok := r.(retriever.CommonInitializableRetriever); ok && rr.Status() != retriever.RetrieverReady {
				resultsChan <- Results{Error: nil, Value: map[string]dto.DTO{}, Index: index}
				return
			}

			rawValue, err := r.Retrieve(ctx)
			if err != nil {
				resultsChan <- Results{Error: err, Value: nil, Index: index}
				return
			}
			convertedFlag, err := cache.ConvertToFlagStruct(rawValue, format)
			resultsChan <- Results{Error: err, Value: convertedFlag, Index: index}
		}(r, config.FileFormat, index, config.Context)
	}

	retrieversResults := make([]map[string]dto.DTO, len(retrievers))
	for v := range resultsChan {
		if v.Error != nil {
			return v.Error
		}
		retrieversResults[v.Index] = v.Value
	}

	// merge all the flags
	newFlags := map[string]dto.DTO{}
	for _, flags := range retrieversResults {
		for flagName, value := range flags {
			newFlags[flagName] = value
		}
	}

	err := cache.UpdateCache(newFlags, config.internalLogger)
	if err != nil {
		log.Printf("error: impossible to update the cache of the flags: %v", err)
		return err
	}
	return nil
}

// GetCacheRefreshDate gives the last refresh date of the cache
func (g *GoFeatureFlag) GetCacheRefreshDate() time.Time {
	if g.config.Offline {
		return time.Time{}
	}
	return g.cache.GetLatestUpdateDate()
}

// ForceRefresh is a function that forces to call the retrievers and refresh the configuration of flags.
// This function can be called explicitly to refresh the flags if you know that a change has been made in
// the configuration.
func (g *GoFeatureFlag) ForceRefresh() bool {
	if g.IsOffline() {
		return false
	}
	err := retrieveFlagsAndUpdateCache(g.config, g.cache, g.retrieverManager)
	if err != nil {
		g.config.internalLogger.Error("Error while force updating the cache.", slog.Any("error", err))
		return false
	}
	return true
}

// SetOffline updates the config Offline parameter
func (g *GoFeatureFlag) SetOffline(control bool) {
	g.config.SetOffline(control)
}

// IsOffline allows knowing if the feature flag is in offline mode
func (g *GoFeatureFlag) IsOffline() bool {
	return g.config.IsOffline()
}

// GetPollingInterval is the polling interval between two refreshes of the cache
func (g *GoFeatureFlag) GetPollingInterval() int64 {
	return g.config.PollingInterval.Milliseconds()
}

// SetOffline updates the config Offline parameter
func SetOffline(control bool) {
	ff.SetOffline(control)
}

// IsOffline allows knowing if the feature flag is in offline mode
func IsOffline() bool {
	return ff.IsOffline()
}

// GetCacheRefreshDate gives the last refresh date of the cache
func GetCacheRefreshDate() time.Time {
	return ff.GetCacheRefreshDate()
}

// ForceRefresh is a function that forces to call the retrievers and refresh the configuration of flags.
// This function can be called explicitly to refresh the flags if you know that a change has been made in
// the configuration.
func ForceRefresh() bool {
	return ff.ForceRefresh()
}

// Close the component by stopping the background refresh and clean the cache.
func Close() {
	onceFF = sync.Once{}
	ff.Close()
}
