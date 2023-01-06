package cache

import (
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/flagv1"
)

// Cache is the interface to represent a cache in the system.
type Cache interface {
	// addFlag add a flag in the cache
	addFlag(key string, value flagv1.FlagData)

	// getFlag get a specific flag by the flag key
	getFlag(key string) (flag.Flag, error)

	// Copy return a copy version of the cache.
	Copy() Cache

	// All return the complete list of the flags.
	All() map[string]flag.Flag

	// Init allow to initialize the cache with a collection of flags.
	Init(flags map[string]flagv1.FlagData)
}
