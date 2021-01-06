package cache

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
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

	go FlagChanges(FlagsCache, flags)

	mutex.Lock()
	FlagsCache = flags
	mutex.Unlock()
	return nil
}

// Close is removing everything from the cache.
func Close() {
	FlagsCache = nil
}

func FlagChanges(oldCache map[string]flags.Flag, newCache map[string]flags.Flag) map[string]flags.Flag {
	for key := range oldCache {
		_, inNewCache := newCache[key]

		if !inNewCache {
			fmt.Printf("flag %v removed\n", key)
			continue
		}

		if oldCache[key].Disable != newCache[key].Disable {
			if newCache[key].Disable {
				// Flag is disabled
				fmt.Printf("The flag %v is turned off\n", key)
			} else {
				flag := newCache[key]
				fmt.Printf(
					"The flag %s is turned ON \n\t"+
						"- percentage is %d%% \n\t"+
						"- rule is [%s] \n", key, flag.Percentage, flag.Rule)
			}
		} else if !cmp.Equal(oldCache[key], newCache[key]) {
			// key has changed in cache
			fmt.Printf("Flag %v hasChanged: %v\n", key, cmp.Diff(oldCache[key], newCache[key]))
		}
	}

	for key := range newCache {
		_, inOldCache := oldCache[key]
		if !inOldCache && !newCache[key].Disable {
			fmt.Printf("flag %v addedd\n", key)
		}
	}

	// TODO: valid return
	return make(map[string]flags.Flag)
}
