package cache

import (
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
	"log"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/flags"
)

type Cache interface {
	UpdateCache(loadedFlags []byte) error
	Close()
	GetFlag(key string) (flags.Flag, error)
}
type cacheImpl struct {
	Logger     *log.Logger
	flagsCache map[string]flags.Flag
	mutex      sync.Mutex
	waitGroup  sync.WaitGroup
}

func New(logger *log.Logger) Cache {
	return &cacheImpl{
		flagsCache: make(map[string]flags.Flag),
		mutex:      sync.Mutex{},
		Logger:     logger,
		waitGroup:  sync.WaitGroup{},
	}
}

func (c *cacheImpl) UpdateCache(loadedFlags []byte) error {
	var newCache map[string]flags.Flag
	err := yaml.Unmarshal(loadedFlags, &newCache)
	if err != nil {
		return err
	}

	// launching a go routine to log the differences
	if c.Logger != nil {
		// copy cache for difference checks async
		cacheCopy := c.getCacheCopy()
		c.waitGroup.Add(1)
		go c.logFlagChangesRoutine(cacheCopy, newCache)
	}

	c.mutex.Lock()
	c.flagsCache = newCache
	c.mutex.Unlock()
	return nil
}

func (c *cacheImpl) Close() {
	c.waitGroup.Wait()
	c.flagsCache = nil
}

func (c *cacheImpl) getCacheCopy() map[string]flags.Flag {
	copyCache := make(map[string]flags.Flag)
	for k, v := range c.flagsCache {
		copyCache[k] = v
	}
	return copyCache
}

func (c *cacheImpl) GetFlag(key string) (flags.Flag, error) {
	if c.flagsCache == nil {
		return flags.Flag{}, errors.New("impossible to read the toggle before the initialisation")
	}

	flag, ok := c.flagsCache[key]
	if !ok {
		return flags.Flag{}, fmt.Errorf("flag [%v] does not exists", key)
	}
	return flag, nil
}

func (c *cacheImpl) logFlagChangesRoutine(oldCache map[string]flags.Flag, newCache map[string]flags.Flag) {
	c.logFlagChanges(oldCache, newCache)
	c.waitGroup.Done()
}

// logFlagChanges is logging if something has changed in your flag config file
func (c *cacheImpl) logFlagChanges(oldCache map[string]flags.Flag, newCache map[string]flags.Flag) {
	date := time.Now().Format(time.RFC3339)
	for key := range oldCache {
		_, inNewCache := newCache[key]
		if !inNewCache {
			c.Logger.Printf("[%v] flag %v removed\n", date, key)
			continue
		}

		if oldCache[key].Disable != newCache[key].Disable {
			if newCache[key].Disable {
				// Flag is disabled
				c.Logger.Printf("[%v] flag %v is turned OFF\n", date, key)
			} else {
				c.Logger.Printf("[%v] flag %v is turned ON (flag=[%v])  \n", date, key, newCache[key])
			}
		} else if !cmp.Equal(oldCache[key], newCache[key]) {
			// key has changed in cache
			c.Logger.Printf("[%v] flag %s updated, old=[%v], new=[%v]\n", date, key, c.flagsCache[key], newCache[key])
		}
	}

	for key := range newCache {
		_, inOldCache := oldCache[key]
		if !inOldCache && !newCache[key].Disable {
			c.Logger.Printf("[%v] flag %v added\n", date, key)
		}
	}
}
