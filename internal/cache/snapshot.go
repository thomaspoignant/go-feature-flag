package cache

import "github.com/thomaspoignant/go-feature-flag/modules/core/flag"

// Snapshot gives access to the flags of the cache.
// The one of the cache manager is the cache as it was when it was taken: a refresh of the flags is
// not visible through it (the cache is replaced, never modified, when the flags are refreshed).
type Snapshot struct {
	getFlag func(key string) (flag.Flag, error)
	all     func() (map[string]flag.Flag, error)
}

// NewSnapshot creates a snapshot from the functions reading the flags.
func NewSnapshot(
	getFlag func(key string) (flag.Flag, error),
	all func() (map[string]flag.Flag, error),
) Snapshot {
	return Snapshot{getFlag: getFlag, all: all}
}

// GetFlag returns the flag from the snapshot.
func (s Snapshot) GetFlag(key string) (flag.Flag, error) {
	return s.getFlag(key)
}

// All returns all the flags of the snapshot.
func (s Snapshot) All() (map[string]flag.Flag, error) {
	return s.all()
}
