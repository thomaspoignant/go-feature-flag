package cache

import (
	"gopkg.in/yaml.v2"
	"sync"

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
func UpdateCache(loadedFlags []byte) error {
	var flags map[string]flags.Flag
	err := yaml.Unmarshal(loadedFlags, &flags)
	if err != nil {
		return err
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
