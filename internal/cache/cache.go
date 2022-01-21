package cache

import (
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

// Cache is the interface to represent a cache in the system.
type Cache interface {
	// getFlag get a specific flag by the flag key
	getFlag(key string) (flag.Flag, error)

	// Copy return a copy version of the cache.
	Copy() Cache

	// All return the complete list of the flags.
	All() map[string]flag.Flag

	// Init allow to initialize the cache with a collection of flags.
	Init(flags map[string]flag.DtoFlag)
}
