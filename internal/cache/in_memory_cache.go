package cache

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/flagv1"
)

type InMemoryCache struct {
	Flags map[string]flagv1.FlagData
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		Flags: map[string]flagv1.FlagData{},
	}
}

func (fc *InMemoryCache) addFlag(key string, value flagv1.FlagData) {
	fc.Flags[key] = value
}

func (fc *InMemoryCache) getFlag(key string) (flag.Flag, error) {
	f, ok := fc.Flags[key]
	if !ok {
		return &f, fmt.Errorf("flag [%v] does not exists", key)
	}
	return &f, nil
}

func (fc *InMemoryCache) keys() []string {
	var keys = make([]string, 0, len(fc.Flags))
	for k := range fc.Flags {
		keys = append(keys, k)
	}
	return keys
}

func (fc *InMemoryCache) Copy() Cache {
	inMemoryCache := NewInMemoryCache()
	for k, v := range fc.Flags {
		inMemoryCache.addFlag(k, v)
	}
	return inMemoryCache
}

func (fc *InMemoryCache) All() map[string]flag.Flag {
	c := map[string]flag.Flag{}
	for _, key := range fc.keys() {
		val, _ := fc.getFlag(key)
		c[key] = val
	}
	return c
}

func (fc *InMemoryCache) Init(flags map[string]flagv1.FlagData) {
	fc.Flags = flags
}
