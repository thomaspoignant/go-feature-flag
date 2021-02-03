package cache

import "github.com/thomaspoignant/go-feature-flag/internal/model"

type FlagsCache map[string]model.Flag

func (fc FlagsCache) Copy() FlagsCache {
	copyCache := make(FlagsCache)
	for k, v := range fc {
		copyCache[k] = v
	}
	return copyCache
}
