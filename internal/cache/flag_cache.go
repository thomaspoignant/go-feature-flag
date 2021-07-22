package cache

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/flagv1"
)

type FlagsCache map[string]flagv1.FlagData

func (fc FlagsCache) Copy() FlagsCache {
	copyCache := make(FlagsCache)
	for k, v := range fc {
		copyCache[k] = v
	}
	return copyCache
}

func (fc *FlagsCache) GetFlag(key string) (flag.Flag, error) {
	cache := *fc
	f, ok := cache[key]
	if !ok {
		return &f, fmt.Errorf("flag [%v] does not exists", key)
	}
	return &f, nil
}
