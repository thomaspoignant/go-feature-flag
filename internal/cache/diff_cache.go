package cache

import "github.com/thomaspoignant/go-feature-flag/internal/flags"

// diffCache contains the changes made in the cache, to be able
// to notify the user that something has changed (logs, webhook ...)
type diffCache struct {
	Deleted map[string]flags.Flag
	Added   map[string]flags.Flag
	Updated map[string]diffUpdated
}

// hasDiff check if we have differences
func (d diffCache) hasDiff() bool {
	return len(d.Deleted) > 0 || len(d.Added) > 0 || len(d.Updated) > 0
}

type diffUpdated struct {
	Before flags.Flag
	After  flags.Flag
}
