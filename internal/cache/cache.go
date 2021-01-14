package cache

import (
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
	"log"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/flags"
)

var once sync.Once
var mutex = &sync.Mutex{}

// FlagsCache is the cache of your flags.
var FlagsCache map[string]flags.Flag

// Init init the cache of all flags.
// We are using a singleton to avoid multiple init and to be sure
// we have only one cache for all the flags in our system.
func Init() map[string]flags.Flag {
	once.Do(func() {
		FlagsCache = make(map[string]flags.Flag)
	})
	return FlagsCache
}

// UpdateCache retrieve the flags from the backend file and update the cache,
// we are using a mutex during the update to be sure to stay consistent.
func UpdateCache(logger *log.Logger, loadedFlags []byte) error {
	var flags map[string]flags.Flag
	err := yaml.Unmarshal(loadedFlags, &flags)
	if err != nil {
		return err
	}

	// launching a go routine to log the differences
	if logger != nil {
		go logFlagChanges(logger, FlagsCache, flags)
	}

	mutex.Lock()
	FlagsCache = flags
	mutex.Unlock()
	return nil
}

// Close is removing everything from the cache.
func Close() {
	FlagsCache = nil
}

// logFlagChanges is logging if something has changed in your flag config file
func logFlagChanges(logger *log.Logger, oldCache map[string]flags.Flag, newCache map[string]flags.Flag) {
	date := time.Now().Format(time.RFC3339)
	for key := range oldCache {
		_, inNewCache := newCache[key]
		if !inNewCache {
			logger.Printf("[%v] flag %v removed\n", date, key)
			continue
		}

		if oldCache[key].Disable != newCache[key].Disable {
			if newCache[key].Disable {
				// Flag is disabled
				logger.Printf("[%v] flag %v is turned OFF\n", date, key)
			} else {
				logger.Printf("[%v] flag %v is turned ON (flag=[%v])  \n", date, key, newCache[key])
			}
		} else if !cmp.Equal(oldCache[key], newCache[key]) {
			// key has changed in cache
			logger.Printf("[%v] flag %s updated, old=[%v], new=[%v]\n", date, key, oldCache[key], newCache[key])
		}
	}

	for key := range newCache {
		_, inOldCache := oldCache[key]
		if !inOldCache && !newCache[key].Disable {
			logger.Printf("[%v] flag %v added\n", date, key)
		}
	}
}
