package cache

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/internal/flags"
)

type Cache interface {
	UpdateCache(loadedFlags []byte) error
	Close()
	GetFlag(key string) (flags.Flag, error)
}

type cacheImpl struct {
	flagsCache          FlagsCache
	mutex               sync.RWMutex
	notificationService Service
}

func New(notificationService Service) Cache {
	return &cacheImpl{
		flagsCache:          make(map[string]flags.Flag),
		mutex:               sync.RWMutex{},
		notificationService: notificationService,
	}
}

func (c *cacheImpl) UpdateCache(loadedFlags []byte) error {
	var newCache FlagsCache
	err := yaml.Unmarshal(loadedFlags, &newCache)
	if err != nil {
		return err
	}

	// copy cache for difference checks async
	cacheCopy := c.flagsCache.Copy()

	c.mutex.Lock()
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
	c.notificationService.Close()
}

func (c *cacheImpl) GetFlag(key string) (flags.Flag, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.flagsCache == nil {
		return flags.Flag{}, errors.New("impossible to read the toggle before the initialisation")
	}

	flag, ok := c.flagsCache[key]
	if !ok {
		return flags.Flag{}, fmt.Errorf("flag [%v] does not exists", key)
	}
	return flag, nil
}
