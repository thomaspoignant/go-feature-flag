package cache

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/thomaspoignant/go-feature-flag/internal/dto"

	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"gopkg.in/yaml.v3"
)

type Manager interface {
	ConvertToFlagStruct(loadedFlags []byte, fileFormat string) (map[string]dto.DTO, error)
	UpdateCache(newFlags map[string]dto.DTO, log *log.Logger) error
	Close()
	GetFlag(key string) (flag.Flag, error)
	AllFlags() (map[string]flag.Flag, error)
	GetLatestUpdateDate() time.Time
}

type cacheManagerImpl struct {
	inMemoryCache       Cache
	mutex               sync.RWMutex
	notificationService Service
	latestUpdate        time.Time
	logger              *log.Logger
}

func New(notificationService Service, logger *log.Logger) Manager {
	return &cacheManagerImpl{
		logger:              logger,
		inMemoryCache:       NewInMemoryCache(logger),
		mutex:               sync.RWMutex{},
		notificationService: notificationService,
	}
}

func (c *cacheManagerImpl) ConvertToFlagStruct(loadedFlags []byte, fileFormat string) (map[string]dto.DTO, error) {
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

func (c *cacheManagerImpl) UpdateCache(newFlags map[string]dto.DTO, log *log.Logger) error {
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

	// notify the changes
	c.notificationService.Notify(oldCacheFlags, newCacheFlags, log)
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
