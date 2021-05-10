package cache

import model "github.com/thomaspoignant/go-feature-flag/internal/model"

type FlagsCache map[string]model.FlagData

func (fc FlagsCache) Copy() FlagsCache {
	copyCache := make(FlagsCache)
	for k, v := range fc {
		copyCache[k] = v
	}
	return copyCache
}
