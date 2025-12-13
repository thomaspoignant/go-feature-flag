package cache

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/go-cmp/cmp"
	"github.com/thomaspoignant/go-feature-flag/internal/notification"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"gopkg.in/yaml.v3"
)

type Manager interface {
	UpdateCache(newFlags map[string]dto.DTO, log *fflog.FFLogger, notifyChanges bool) error
	Close()
	GetFlag(key string) (flag.Flag, error)
	AllFlags() (map[string]flag.Flag, error)
	GetLatestUpdateDate() time.Time
}

type cacheManagerImpl struct {
	inMemoryCache                   Cache
	mutex                           sync.RWMutex
	notificationService             notification.Service
	latestUpdate                    time.Time
	logger                          *fflog.FFLogger
	persistentFlagConfigurationFile string
}

func New(
	notificationService notification.Service,
	persistentFlagConfigurationFile string,
	logger *fflog.FFLogger,
) Manager {
	return &cacheManagerImpl{
		logger:                          logger,
		inMemoryCache:                   NewInMemoryCache(logger),
		mutex:                           sync.RWMutex{},
		notificationService:             notificationService,
		persistentFlagConfigurationFile: persistentFlagConfigurationFile,
	}
}

func ConvertToFlagStruct(
	loadedFlags []byte,
	fileFormat string,
) (map[string]dto.DTO, error) {
	var newFlags map[string]dto.DTO
	var err error
	switch strings.ToLower(fileFormat) {
	case "toml":
		err = toml.Unmarshal(loadedFlags, &newFlags)
	case "json":
		err = json.Unmarshal(loadedFlags, &newFlags)
	default:
		// default unmarshaller is YAML
		err = yaml.Unmarshal(loadedFlags, &newFlags)
	}
	return newFlags, err
}

func (c *cacheManagerImpl) UpdateCache(
	newFlags map[string]dto.DTO,
	log *fflog.FFLogger,
	notifyChanges bool,
) error {
	newCache := NewInMemoryCache(c.logger)
	newCache.Init(newFlags)
	newCacheFlags := newCache.All()
	oldCacheFlags := map[string]flag.Flag{}

	c.mutex.Lock()
	// collect flags for compare.
	if c.inMemoryCache != nil {
		oldCacheFlags = c.inMemoryCache.All()
	}
	c.inMemoryCache = newCache
	c.latestUpdate = time.Now()
	c.mutex.Unlock()

	if notifyChanges {
		// notify the changes
		c.notificationService.Notify(oldCacheFlags, newCacheFlags, log)
	}
	// persist the cache on disk
	if c.persistentFlagConfigurationFile != "" {
		c.PersistCache(oldCacheFlags, newCacheFlags)
	}
	return nil
}

func (c *cacheManagerImpl) Close() {
	// Clear the cache
	c.mutex.Lock()
	c.inMemoryCache = nil
	c.mutex.Unlock()
	if c.notificationService != nil {
		c.notificationService.Close()
	}
}

func (c *cacheManagerImpl) GetFlag(key string) (flag.Flag, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.inMemoryCache == nil {
		return nil, errors.New("impossible to read the flag before the initialisation")
	}
	return c.inMemoryCache.getFlag(key)
}

func (c *cacheManagerImpl) AllFlags() (map[string]flag.Flag, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.inMemoryCache == nil {
		return nil, errors.New("impossible to read the flag before the initialisation")
	}
	return c.inMemoryCache.All(), nil
}

func (c *cacheManagerImpl) GetLatestUpdateDate() time.Time {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.latestUpdate
}

// PersistCache is writing the flags to a file to be able to restart without being able to access the retrievers.
// It is useful to have a fallback in case of a problem with the retrievers, such as a network issue.
//
// The persistence is done in a goroutine to not block the main thread.
func (c *cacheManagerImpl) PersistCache(oldCache, newCache map[string]flag.Flag) {
	go func() {
		if _, err := os.Stat(c.persistentFlagConfigurationFile); !os.IsNotExist(err) &&
			cmp.Equal(oldCache, newCache) {
			c.logger.Debug("No change in the cache, skipping the persist")
			return
		}
		data, err := yaml.Marshal(newCache)
		if err != nil {
			c.logger.Error("Error while marshalling flags to persist", slog.Any("error", err.Error()))
			return
		}

		err = os.WriteFile(c.persistentFlagConfigurationFile, data, 0600)
		if err != nil {
			c.logger.Error("Error while writing flags to file", slog.Any("error", err.Error()))
			return
		}
		c.logger.Info(
			"Flags cache persisted to file",
			slog.String("file", c.persistentFlagConfigurationFile),
		)
	}()
}
