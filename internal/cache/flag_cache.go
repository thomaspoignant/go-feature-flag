package cache

import (
	"github.com/thomaspoignant/go-feature-flag/internal/flag_v1"
)

type FlagsCache map[string]flag_v1.FlagData

func (fc FlagsCache) Copy() FlagsCache {
	copyCache := make(FlagsCache)
	for k, v := range fc {
		copyCache[k] = v
	}
	return copyCache
}
