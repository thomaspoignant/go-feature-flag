package cache

import (
	"encoding/json"
	"errors"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	flagv1 "github.com/thomaspoignant/go-feature-flag/internal/flagv1"
	"gopkg.in/yaml.v3"
	"strings"
	"sync"

	"github.com/pelletier/go-toml"
)

type Manager interface {
	UpdateCache(loadedFlags []byte, fileFormat string) error
	Close()
	GetFlag(key string) (flag.Flag, error)
	AllFlags() (map[string]flag.Flag, error)
}

type cacheManagerImpl struct {
	inMemoryCache       Cache
	mutex               sync.RWMutex
	notificationService Service
}

func New(notificationService Service) Manager {
	return &cacheManagerImpl{
		inMemoryCache:       NewInMemoryCache(),
		mutex:               sync.RWMutex{},
		notificationService: notificationService,
	}
}

func (c *cacheManagerImpl) UpdateCache(loadedFlags []byte, fileFormat string) error {
	var newCache map[string]flagv1.FlagData
	var err error
	switch strings.ToLower(fileFormat) {
	case "toml":
		err = toml.Unmarshal(loadedFlags, &newCache)
	case "json":
		err = json.Unmarshal(loadedFlags, &newCache)
	default:
		// default unmarshaller is YAML
		err = yaml.Unmarshal(loadedFlags, &newCache)
	}

	if err != nil {
		return err
	}
	c.mutex.Lock()
	// copy cache for difference checks async
	cacheCopy := c.inMemoryCache.Copy()
	c.inMemoryCache.Init(newCache)
	c.mutex.Unlock()

	// notify the changes
	c.notificationService.Notify(cacheCopy.All(), c.inMemoryCache.All())
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
