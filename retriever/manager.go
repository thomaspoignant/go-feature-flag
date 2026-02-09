package retriever

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// ManagerConfig is the configuration of the retriever manager.
type ManagerConfig struct {
	FileFormat                      string
	DisableNotifierOnInit           bool
	PersistentFlagConfigurationFile string
	StartWithRetrieverError         bool
	EnablePollingJitter             bool
	PollingInterval                 time.Duration
	Name                            *string
}

// Manager is a struct that managed the retrievers.
type Manager struct {
	retrievers       []Retriever
	onErrorRetriever []Retriever
	logger           *fflog.FFLogger
	cacheManager     cache.Manager
	config           ManagerConfig
	bgUpdater        backgroundUpdater
}

// NewManager create a new Manager.
func NewManager(
	config ManagerConfig, retrievers []Retriever, cacheManager cache.Manager, logger *fflog.FFLogger) *Manager {
	return &Manager{
		retrievers:       retrievers,
		onErrorRetriever: make([]Retriever, 0),
		logger:           logger,
		cacheManager:     cacheManager,
		config:           config,
	}
}

// Init the retrievers.
// This function will call the Init function of the retrievers that implements the InitializableRetriever interface.
func (m *Manager) Init(ctx context.Context) error {
	if err := m.initRetrievers(ctx, m.retrievers); err != nil {
		return err
	}
	if err := m.retrieveFlagsAndUpdateCache(ctx, true); err != nil {
		if err := m.handleFirstRetrieverError(ctx, err); err != nil {
			return err
		}
	}

	if m.config.PollingInterval > 0 {
		m.bgUpdater = newBackgroundUpdater(m.config.PollingInterval, m.config.EnablePollingJitter)
		go m.StartPolling(ctx)
	}
	return nil
}

// StartPolling is the daemon that refreshes the cache every X seconds.
func (m *Manager) StartPolling(ctx context.Context) {
	for {
		select {
		case <-m.bgUpdater.ticker.C:
			err := m.retrieveFlagsAndUpdateCache(ctx, false)
			if err != nil {
				m.logger.Error(
					"Error while updating the cache.",
					slog.Any("error", err.Error()),
				)
			}
		case <-m.bgUpdater.updaterChan:
			return
		}
	}
}

// StopPolling is the function to stop the background updater.
func (m *Manager) StopPolling() {
	m.bgUpdater.close()
}

// initRetrievers is a helper function to initialize the retrievers.
func (m *Manager) initRetrievers(ctx context.Context, retrieversToInit []Retriever) error {
	m.onErrorRetriever = make([]Retriever, 0)
	for _, retriever := range retrieversToInit {
		m.tryInitializeLegacy(ctx, retriever)
		m.tryInitializeStandard(ctx, retriever)
		m.tryInitializeWithFlagset(ctx, retriever)
	}
	return m.checkInitializationErrors()
}

// tryInitializeLegacy attempts to initialize a retriever using the legacy interface.
// This function will append the retriever to the onErrorRetriever slice if the initialization fails.
// If retriever implements the InitializableRetrieverLegacy interface, it will be initialized using
// the legacy interface.
func (m *Manager) tryInitializeLegacy(ctx context.Context, retriever Retriever) {
	if r, ok := retriever.(InitializableRetrieverLegacy); ok {
		if r.Init(ctx, m.logger.GetLogLogger(slog.LevelError)) != nil {
			m.onErrorRetriever = append(m.onErrorRetriever, retriever)
		}
	}
}

// tryInitializeStandard attempts to initialize a retriever using the standard interface.
// This function will append the retriever to the onErrorRetriever slice if the initialization fails.
// If retriever implements the InitializableRetriever interface, it will be initialized using
// the standard interface.
func (m *Manager) tryInitializeStandard(ctx context.Context, retriever Retriever) {
	if r, ok := retriever.(InitializableRetriever); ok {
		if r.Init(ctx, m.logger) != nil {
			m.onErrorRetriever = append(m.onErrorRetriever, retriever)
		}
	}
}

// tryInitializeWithFlagset attempts to initialize a retriever using the flagset interface.
// This function will append the retriever to the onErrorRetriever slice if the initialization fails.
// If retriever implements the InitializableRetrieverWithFlagset interface, it will be initialized using
// the flagset interface.
func (m *Manager) tryInitializeWithFlagset(ctx context.Context, retriever Retriever) {
	if r, ok := retriever.(InitializableRetrieverWithFlagset); ok {
		if r.Init(ctx, m.logger, m.config.Name) != nil {
			m.onErrorRetriever = append(m.onErrorRetriever, retriever)
		}
	}
}

// checkInitializationErrors returns an error if any retrievers failed to initialize.
func (m *Manager) checkInitializationErrors() error {
	if len(m.onErrorRetriever) > 0 {
		return fmt.Errorf("error while initializing the retrievers: %v", m.onErrorRetriever)
	}
	return nil
}

// Shutdown the retrievers.
// This function will call the Shutdown function of the retrievers that implements the InitializableRetriever interface.
func (m *Manager) Shutdown(ctx context.Context) error {
	onErrorRetriever := make([]Retriever, 0)
	for _, retriever := range m.retrievers {
		if r, ok := retriever.(CommonInitializableRetriever); ok {
			err := r.Shutdown(ctx)
			if err != nil {
				onErrorRetriever = append(onErrorRetriever, retriever)
			}
		}
	}
	if len(onErrorRetriever) > 0 {
		return fmt.Errorf("error while shutting down the retrievers: %v", onErrorRetriever)
	}

	m.cacheManager.Close()
	return nil
}

// retrieveFlagsAndUpdateCache is a function that will retrieve the flags from the retrievers and update the cache.
func (m *Manager) retrieveFlagsAndUpdateCache(ctx context.Context, isInit bool) error {
	if len(m.onErrorRetriever) > 0 {
		_ = m.initRetrievers(ctx, m.onErrorRetriever)
	}
	newFlags, err := retrieve(ctx, m.retrievers, m.config.FileFormat)
	if err != nil {
		return err
	}
	return m.updateCacheWithRetriever(newFlags, isInit)
}

// updateCacheWithRetriever is a function that will update the cache with the new flags received from the retriever.
func (m *Manager) updateCacheWithRetriever(newFlags map[string]dto.DTO, isInit bool) error {
	err := m.cacheManager.UpdateCache(
		newFlags,
		m.logger,
		!isInit || !m.config.DisableNotifierOnInit,
	)
	if err != nil {
		m.logger.Error("error: impossible to update the cache of the flags: %v", err)
		return err
	}
	return nil
}

// handleFirstRetrieverError is a function that will handle the first error when trying to retrieve
// the flags the first time when starting GO Feature Flag.
func (m *Manager) handleFirstRetrieverError(ctx context.Context, err error) error {
	switch {
	case m.config.PersistentFlagConfigurationFile != "":
		errPersist := m.retrievePersistentLocalDisk(ctx)
		if errPersist != nil && !m.config.StartWithRetrieverError {
			return fmt.Errorf("impossible to use the persistent flag configuration file: %v "+
				"[original error: %v]", errPersist, err)
		}
	case !m.config.StartWithRetrieverError:
		return fmt.Errorf(
			"impossible to retrieve the flags, please check your configuration: %v",
			err,
		)
	default:
		// We accept to start with a retriever error, we will serve only default value
		m.logger.Error("Impossible to retrieve the flags, starting with the "+
			"retriever error", slog.Any("error", err.Error()))
	}
	return nil
}

// retrievePersistentLocalDisk is a function used in case we are not able to retrieve any flag when starting
// GO Feature Flag.
// This function will look at any pre-existent persistent configuration and start with it.
func (m *Manager) retrievePersistentLocalDisk(ctx context.Context) error {
	if m.config.PersistentFlagConfigurationFile != "" {
		m.logger.Error(
			"Impossible to retrieve your flag configuration, trying to use the persistent"+
				" flag configuration file.",
			slog.String("path", m.config.PersistentFlagConfigurationFile),
		)
		if _, err := os.Stat(m.config.PersistentFlagConfigurationFile); err == nil {
			// we found the configuration file on the disk
			r := &fileretriever.Retriever{Path: m.config.PersistentFlagConfigurationFile}
			newFlags, err := retrieve(ctx, []Retriever{r}, m.config.FileFormat)
			if err != nil {
				return err
			}
			return m.updateCacheWithRetriever(newFlags, true)
		}
		m.logger.Warn("No persistent flag configuration found",
			slog.String("path", m.config.PersistentFlagConfigurationFile))
	}
	return fmt.Errorf("no persistent flag available")
}

// GetCacheRefreshDate gives the last refresh date of the cache
func (m *Manager) GetCacheRefreshDate() time.Time {
	return m.cacheManager.GetLatestUpdateDate()
}

// GetFlag returns the flag from the cache with the current state when calling this method.
func (m *Manager) GetFlag(flagKey string) (flag.Flag, error) {
	return m.cacheManager.GetFlag(flagKey)
}

// GetFlagsFromCache returns all the flags present in the cache with their
// current state when calling this method. If cache hasn't been initialized, an
// error reporting this is returned.
func (m *Manager) GetFlagsFromCache(_ context.Context) (map[string]flag.Flag, error) {
	if m == nil || m.cacheManager == nil {
		return nil, fmt.Errorf("cache is not initialized")
	}
	return m.cacheManager.AllFlags()
}

func (m *Manager) ForceRefresh(ctx context.Context) bool {
	err := m.retrieveFlagsAndUpdateCache(ctx, false)
	if err != nil {
		m.logger.Error(
			"Error while force updating the cache.",
			slog.Any("error", err.Error()),
		)
		return false
	}
	return true
}

// retrieve is a function that will retrieve the flags from all the retrievers in parallel.
func retrieve(ctx context.Context, retrievers []Retriever, fileFormat string) (map[string]dto.DTO, error) {
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
		go func(r Retriever, format string, index int, ctx context.Context) {
			defer wg.Done()

			// If the retriever is not ready, we ignore it
			if rr, ok := r.(CommonInitializableRetriever); ok &&
				rr.Status() != RetrieverReady {
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
		}(r, fileFormat, index, ctx)
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
