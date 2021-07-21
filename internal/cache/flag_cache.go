package cache

import flagv1 "github.com/thomaspoignant/go-feature-flag/internal/flagv1"

type FlagsCache map[string]flagv1.FlagData

func (fc FlagsCache) Copy() FlagsCache {
	copyCache := make(FlagsCache)
	for k, v := range fc {
		copyCache[k] = v
	}
	return copyCache
}
