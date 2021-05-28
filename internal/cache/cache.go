package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
	"sync"

	"github.com/pelletier/go-toml"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
)

type Cache interface {
	UpdateCache(loadedFlags []byte, fileFormat string) error
	Close()
	GetFlag(key string) (model.Flag, error)
	AllFlags() (FlagsCache, error)
}

type cacheImpl struct {
	flagsCache          FlagsCache
	mutex               sync.RWMutex
	notificationService Service
}

func New(notificationService Service) Cache {
	return &cacheImpl{
		flagsCache:          make(map[string]model.FlagData),
		mutex:               sync.RWMutex{},
		notificationService: notificationService,
	}
}

func (c *cacheImpl) UpdateCache(loadedFlags []byte, fileFormat string) error {
	var newCache FlagsCache
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
	cacheCopy := c.flagsCache.Copy()
	c.flagsCache = newCache
	c.mutex.Unlock()

	// notify the changes
	c.notificationService.Notify(cacheCopy, newCache)
	return nil
}

func (c *cacheImpl) Close() {
	// Clear the cache
	c.mutex.Lock()
	c.flagsCache = nil
	c.mutex.Unlock()
	if c.notificationService != nil {
		c.notificationService.Close()
	}
}

func (c *cacheImpl) GetFlag(key string) (model.Flag, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.flagsCache == nil {
		return &model.FlagData{}, errors.New("impossible to read the flag before the initialisation")
	}

	flag, ok := c.flagsCache[key]
	if !ok {
		return &model.FlagData{}, fmt.Errorf("flag [%v] does not exists", key)
	}

	return &flag, nil
}

func (c *cacheImpl) AllFlags() (FlagsCache, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.flagsCache == nil {
		return nil, errors.New("impossible to read the flags before the initialisation")
	}

	return c.flagsCache.Copy(), nil
}
