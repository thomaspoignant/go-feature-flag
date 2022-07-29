package cache

import (
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/internal/dto"

	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

type InMemoryCache struct {
	Flags map[string]flag.InternalFlag
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		Flags: map[string]flag.InternalFlag{},
	}
}

func (fc *InMemoryCache) addFlag(key string, value flag.InternalFlag) {
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
	keys := make([]string, 0, len(fc.Flags))
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

func (fc *InMemoryCache) Init(flags map[string]dto.DTO) {
	cache := make(map[string]flag.InternalFlag, 0)
	for k, v := range flags {
		cache[k] = v.Convert()
	}
	fc.Flags = cache
}
