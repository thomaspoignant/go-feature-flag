package cache

import "github.com/thomaspoignant/go-feature-flag/internal/flags"

type FlagsCache map[string]flags.Flag

func (fc FlagsCache) Copy() FlagsCache {
	copyCache := make(FlagsCache)
	for k, v := range fc {
		copyCache[k] = v
	}
	return copyCache
}
