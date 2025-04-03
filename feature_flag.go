package ffclient

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/notification"
	"github.com/thomaspoignant/go-feature-flag/model/dto"
	"github.com/thomaspoignant/go-feature-flag/notifier/logsnotifier"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
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
	cache                     cache.Manager
	config                    Config
	bgUpdater                 backgroundUpdater
	featureEventDataExporter  exporter.Manager[exporter.FeatureEvent]
	trackingEventDataExporter exporter.Manager[exporter.TrackingEvent]
	retrieverManager          *retriever.Manager
	// evalExporterWg is a wait group to wait for the evaluation exporter to finish the export before closing GOFF
	evalExporterWg sync.WaitGroup
}

// ff is the default object for go-feature-flag
var ff *GoFeatureFlag
var onceFF sync.Once

// New creates a new go-feature-flag instances that retrieve the config from a YAML file
// and return everything you need to manage your flags.
func New(config Config) (*GoFeatureFlag, error) {
	config.PollingInterval = adjustPollingInterval(config.PollingInterval)
	if config.offlineMutex == nil {
		config.offlineMutex = &sync.RWMutex{}
	}

	// initialize internal logger
	config.internalLogger = &fflog.FFLogger{
		LeveledLogger: config.LeveledLogger,
		LegacyLogger:  config.Logger,
	}

	goFF := &GoFeatureFlag{
		config:         config,
		evalExporterWg: sync.WaitGroup{},
	}

	if config.Offline {
		// in case we are in offline mode, we don't need to initialize the cache since we will not use it.
		goFF.config.internalLogger.Info("GO Feature Flag is in offline mode")
		return goFF, nil
	}

	notificationService := initializeNotificationService(config)

	// init internal cache
	goFF.cache = cache.New(
		notificationService,
		config.PersistentFlagConfigurationFile,
		config.internalLogger,
	)

	retrieverManager, err := initializeRetrieverManager(config)
	if err != nil && (retrieverManager == nil || !config.StartWithRetrieverError) {
		return nil, fmt.Errorf(
			"impossible to initialize the retrievers, please check your configuration: %v",
			err,
		)
	}
	goFF.retrieverManager = retrieverManager

	// first retrieval of the flags
	if err := retrieveFlagsAndUpdateCache(goFF.config, goFF.cache, goFF.retrieverManager, true); err != nil {
		if err := handleFirstRetrieverError(config, goFF.config.internalLogger, goFF.cache, err); err != nil {
			return nil, err
		}
	}

	// start the background task to update the flags periodically
	if config.PollingInterval > 0 {
		goFF.bgUpdater = newBackgroundUpdater(config.PollingInterval, config.EnablePollingJitter)
		go goFF.startFlagUpdaterDaemon()
	}

	goFF.featureEventDataExporter, goFF.trackingEventDataExporter =
		initializeDataExporters(config, goFF.config.internalLogger)
	config.internalLogger.Debug("GO Feature Flag is initialized")
	return goFF, nil
}

// adjustPollingInterval is a function that will check the polling interval and set it to the minimum value if it is
// lower than 1 second. It also set the default value to 60 seconds if the polling interval is 0.
func adjustPollingInterval(pollingInterval time.Duration) time.Duration {
	switch {
	case pollingInterval == 0:
		// The default value for the poll interval is 60 seconds
		return 60 * time.Second
	case pollingInterval > 0 && pollingInterval < time.Second:
		// the minimum value for the polling policy is 1 second
		return time.Second
	default:
		return pollingInterval
	}
}

// initializeNotificationService is a function that will initialize the notification service with the notifiers
func initializeNotificationService(config Config) notification.Service {
	notifiers := config.Notifiers
	notifiers = append(notifiers, &logsnotifier.Notifier{Logger: config.internalLogger})
	return notification.NewService(notifiers)
}

// initializeRetrieverManager is a function that will initialize the retriever manager with the retrievers
func initializeRetrieverManager(config Config) (*retriever.Manager, error) {
	retrievers, err := config.GetRetrievers()
	if err != nil {
		return nil, err
	}
	manager := retriever.NewManager(config.Context, retrievers, config.internalLogger)
	err = manager.Init(config.Context)
	return manager, err
}

func initializeDataExporters(config Config, logger *fflog.FFLogger) (
	exporter.Manager[exporter.FeatureEvent], exporter.Manager[exporter.TrackingEvent]) {
	exporters := config.GetDataExporters()
	featureEventExporterConfigs := make([]exporter.Config, 0)
	trackingEventExporterConfigs := make([]exporter.Config, 0)
	if len(exporters) > 0 {
		for _, exp := range exporters {
			c := exporter.Config{
				Exporter:         exp.Exporter,
				FlushInterval:    exp.FlushInterval,
				MaxEventInMemory: exp.MaxEventInMemory,
			}
			if exp.ExporterEventType == TrackingEventExporter {
				trackingEventExporterConfigs = append(trackingEventExporterConfigs, c)
				continue
			}
			featureEventExporterConfigs = append(featureEventExporterConfigs, c)
		}
	}

	var trackingEventManager exporter.Manager[exporter.TrackingEvent]
	if len(trackingEventExporterConfigs) > 0 {
		trackingEventManager = exporter.NewManager[exporter.TrackingEvent](
			config.Context, trackingEventExporterConfigs, config.ExporterCleanQueueInterval, logger)
		trackingEventManager.Start()
	}

	var featureEventManager exporter.Manager[exporter.FeatureEvent]
	if len(featureEventExporterConfigs) > 0 {
		featureEventManager = exporter.NewManager[exporter.FeatureEvent](
			config.Context, featureEventExporterConfigs, config.ExporterCleanQueueInterval, logger)
		featureEventManager.Start()
	}
	return featureEventManager, trackingEventManager
}

// handleFirstRetrieverError is a function that will handle the first error when trying to retrieve
// the flags the first time when starting GO Feature Flag.
func handleFirstRetrieverError(
	config Config,
	logger *fflog.FFLogger,
	cache cache.Manager,
	err error,
) error {
	switch {
	case config.PersistentFlagConfigurationFile != "":
		errPersist := retrievePersistentLocalDisk(config.Context, config, cache)
		if errPersist != nil && !config.StartWithRetrieverError {
			return fmt.Errorf("impossible to use the persistent flag configuration file: %v "+
				"[original error: %v]", errPersist, err)
		}
	case !config.StartWithRetrieverError:
		return fmt.Errorf(
			"impossible to retrieve the flags, please check your configuration: %v",
			err,
		)
	default:
		// We accept to start with a retriever error, we will serve only default value
		logger.Error("Impossible to retrieve the flags, starting with the "+
			"retriever error", slog.Any("error", err))
	}
	return nil
}

// retrievePersistentLocalDisk is a function used in case we are not able to retrieve any flag when starting
// GO Feature Flag.
// This function will look at any pre-existent persistent configuration and start with it.
func retrievePersistentLocalDisk(ctx context.Context, config Config, cache cache.Manager) error {
	if config.PersistentFlagConfigurationFile != "" {
		config.internalLogger.Error(
			"Impossible to retrieve your flag configuration, trying to use the persistent"+
				" flag configuration file.",
			slog.String("path", config.PersistentFlagConfigurationFile),
		)
		if _, err := os.Stat(config.PersistentFlagConfigurationFile); err == nil {
			// we found the configuration file on the disk
			r := &fileretriever.Retriever{Path: config.PersistentFlagConfigurationFile}
			fallBackRetrieverManager := retriever.NewManager(
				config.Context,
				[]retriever.Retriever{r},
				config.internalLogger,
			)
			err := fallBackRetrieverManager.Init(ctx)
			if err != nil {
				return err
			}
			defer func() { _ = fallBackRetrieverManager.Shutdown(ctx) }()
			err = retrieveFlagsAndUpdateCache(config, cache, fallBackRetrieverManager, true)
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
			g.cache.Close()
		}
		if g.bgUpdater.updaterChan != nil && g.bgUpdater.ticker != nil {
			g.bgUpdater.close()
		}
		// we have to wait for the GO routine before stopping the exporter
		g.evalExporterWg.Wait()
		if g.featureEventDataExporter != nil {
			g.featureEventDataExporter.Stop()
		}

		if g.trackingEventDataExporter != nil {
			g.trackingEventDataExporter.Stop()
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
				err := retrieveFlagsAndUpdateCache(g.config, g.cache, g.retrieverManager, false)
				if err != nil {
					g.config.internalLogger.Error(
						"Error while updating the cache.",
						slog.Any("error", err),
					)
				}
			}
		case <-g.bgUpdater.updaterChan:
			return
		}
	}
}

// retreiveFlags is a function that will retrieve the flags from the retrievers,
// merge them and convert them to the flag struct.
func retreiveFlags(
	config Config,
	cache cache.Manager,
	retrieverManager *retriever.Manager,
) (map[string]dto.DTO, error) {
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
			if rr, ok := r.(retriever.CommonInitializableRetriever); ok &&
				rr.Status() != retriever.RetrieverReady {
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
			return nil, v.Error
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
	return newFlags, nil
}

// retrieveFlagsAndUpdateCache is a function that retrieves the flags from the retrievers,
// and update the cache with the new flags.
func retrieveFlagsAndUpdateCache(config Config, cache cache.Manager,
	retrieverManager *retriever.Manager, isInit bool) error {
	newFlags, err := retreiveFlags(config, cache, retrieverManager)
	if err != nil {
		return err
	}

	err = cache.UpdateCache(
		newFlags,
		config.internalLogger,
		!isInit || !config.DisableNotifierOnInit,
	)
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
	err := retrieveFlagsAndUpdateCache(g.config, g.cache, g.retrieverManager, false)
	if err != nil {
		g.config.internalLogger.Error(
			"Error while force updating the cache.",
			slog.Any("error", err),
		)
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
