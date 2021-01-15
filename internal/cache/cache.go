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
	flagsCache map[string]flags.Flag
	mutex      sync.Mutex
	Logger     *log.Logger
}

func New(logger *log.Logger) Cache {
	return &cacheImpl{
		flagsCache: make(map[string]flags.Flag),
		mutex:      sync.Mutex{},
		Logger:     logger,
	}
}

func (c *cacheImpl) UpdateCache(loadedFlags []byte) error {
	var flags map[string]flags.Flag
	err := yaml.Unmarshal(loadedFlags, &flags)
	if err != nil {
		return err
	}

	// launching a go routine to log the differences
	if c.Logger != nil {
		go c.logFlagChanges(flags)
	}

	c.mutex.Lock()
	c.flagsCache = flags
	c.mutex.Unlock()
	return nil
}

func (c *cacheImpl) Close() {
	c.flagsCache = nil
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

// logFlagChanges is logging if something has changed in your flag config file
func (c *cacheImpl) logFlagChanges(newCache map[string]flags.Flag) {
	date := time.Now().Format(time.RFC3339)
	for key := range c.flagsCache {
		_, inNewCache := newCache[key]
		if !inNewCache {
			c.Logger.Printf("[%v] flag %v removed\n", date, key)
			continue
		}

		if c.flagsCache[key].Disable != newCache[key].Disable {
			if newCache[key].Disable {
				// Flag is disabled
				c.Logger.Printf("[%v] flag %v is turned OFF\n", date, key)
			} else {
				c.Logger.Printf("[%v] flag %v is turned ON (flag=[%v])  \n", date, key, newCache[key])
			}
		} else if !cmp.Equal(c.flagsCache[key], newCache[key]) {
			// key has changed in cache
			c.Logger.Printf("[%v] flag %s updated, old=[%v], new=[%v]\n", date, key, c.flagsCache[key], newCache[key])
		}
	}

	for key := range newCache {
		_, inOldCache := c.flagsCache[key]
		if !inOldCache && !newCache[key].Disable {
			c.Logger.Printf("[%v] flag %v added\n", date, key)
		}
	}
}
