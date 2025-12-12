package ffclient

import (
	"fmt"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/notification"
	"github.com/thomaspoignant/go-feature-flag/notifier/logsnotifier"
	"github.com/thomaspoignant/go-feature-flag/retriever"
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
	config                    Config
	featureEventDataExporter  exporter.Manager[exporter.FeatureEvent]
	trackingEventDataExporter exporter.Manager[exporter.TrackingEvent]
	retrieverManager          *retriever.Manager
	// evalExporterWg is a wait group to wait for the evaluation exporter to finish the export before closing GOFF
	evalExporterWg sync.WaitGroup
}

// ff is the default object for go-feature-flag
var (
	ff     *GoFeatureFlag
	onceFF sync.Once
)

// New creates a new go-feature-flag instances that retrieve the config from a YAML file
// and return everything you need to manage your flags.
func New(config Config) (*GoFeatureFlag, error) {
	config.Initialize()

	goFF := &GoFeatureFlag{
		config:         config,
		evalExporterWg: sync.WaitGroup{},
	}

	if config.Offline {
		// in case we are in offline mode, we don't need to initialize the cache since we will not use it.
		goFF.config.internalLogger.Info("GO Feature Flag is in offline mode")
		return goFF, nil
	}

	retrieverManager, err := initializeRetrieverManager(config)
	if err != nil && (goFF.retrieverManager == nil || !config.StartWithRetrieverError) {
		return nil, fmt.Errorf(
			"impossible to initialize the retrievers, please check your configuration: %v",
			err,
		)
	}
	goFF.retrieverManager = retrieverManager
	goFF.featureEventDataExporter, goFF.trackingEventDataExporter = initializeDataExporters(
		config, goFF.config.internalLogger)
	config.internalLogger.Debug("GO Feature Flag is initialized")
	return goFF, nil
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
	mngrConfig := retriever.ManagerConfig{
		Ctx:                             config.Context,
		FileFormat:                      config.FileFormat,
		DisableNotifierOnInit:           config.DisableNotifierOnInit,
		PersistentFlagConfigurationFile: config.PersistentFlagConfigurationFile,
		StartWithRetrieverError:         config.StartWithRetrieverError,
		EnablePollingJitter:             config.EnablePollingJitter,
		PollingInterval:                 config.PollingInterval,
		Name:                            config.Name,
	}

	notificationService := initializeNotificationService(config)
	// init internal cache
	cacheMngr := cache.New(
		notificationService,
		config.PersistentFlagConfigurationFile,
		config.internalLogger,
	)

	manager := retriever.NewManager(mngrConfig, retrievers, cacheMngr, config.internalLogger)
	err = manager.Init(config.Context)
	return manager, err
}

func initializeDataExporters(config Config, logger *fflog.FFLogger) (
	exporter.Manager[exporter.FeatureEvent], exporter.Manager[exporter.TrackingEvent],
) {
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
			trackingEventExporterConfigs, config.ExporterCleanQueueInterval, logger)
		trackingEventManager.Start()
	}

	var featureEventManager exporter.Manager[exporter.FeatureEvent]
	if len(featureEventExporterConfigs) > 0 {
		featureEventManager = exporter.NewManager[exporter.FeatureEvent](
			featureEventExporterConfigs, config.ExporterCleanQueueInterval, logger)
		featureEventManager.Start()
	}
	return featureEventManager, trackingEventManager
}

// Close wait until thread are done
func (g *GoFeatureFlag) Close() {
	if g != nil {
		if g.retrieverManager != nil {
			_ = g.retrieverManager.Shutdown(g.config.Context)
		}
		// we have to wait for the GO routine before stopping the exporter
		g.evalExporterWg.Wait()
		if g.featureEventDataExporter != nil {
			g.featureEventDataExporter.Stop()
		}

		if g.trackingEventDataExporter != nil {
			g.trackingEventDataExporter.Stop()
		}
	}
}

// GetCacheRefreshDate gives the last refresh date of the cache
func (g *GoFeatureFlag) GetCacheRefreshDate() time.Time {
	if g.IsOffline() {
		return time.Time{}
	}
	return g.retrieverManager.GetCacheRefreshDate()
}

// GetEvaluationContextEnrichment returns the evaluation context enrichment
func (g *GoFeatureFlag) GetEvaluationContextEnrichment() map[string]any {
	return g.config.EvaluationContextEnrichment
}

// ForceRefresh is a function that forces to call the retrievers and refresh the configuration of flags.
// This function can be called explicitly to refresh the flags if you know that a change has been made in
// the configuration.
func (g *GoFeatureFlag) ForceRefresh() bool {
	return g.retrieverManager.ForceRefresh()
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
	if !ff.IsOffline() && control {
		ff.retrieverManager.StopPolling()
	}
	if ff.IsOffline() && !control {
		ff.retrieverManager.StartPolling()
	}
	ff.SetOffline(control)
}

// IsOffline allows knowing if the feature flag is in offline mode
func IsOffline() bool {
	return ff.IsOffline()
}

// GetCacheRefreshDate gives the last refresh date of the cache
func GetCacheRefreshDate() time.Time {
	if ff.IsOffline() {
		return time.Time{}
	}
	return ff.GetCacheRefreshDate()
}

// ForceRefresh is a function that forces to call the retrievers and refresh the configuration of flags.
// This function can be called explicitly to refresh the flags if you know that a change has been made in
// the configuration.
func ForceRefresh() bool {
	if ff.IsOffline() {
		return false
	}
	return ff.ForceRefresh()
}

// Close the component by stopping the background refresh and clean the cache.
func Close() {
	onceFF = sync.Once{}
	ff.Close()
}
