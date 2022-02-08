package cache

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

// InMemoryCache is a memory cache to store the status of each flag.
type InMemoryCache struct {
	Flags  map[string]flag.FlagData
	logger fflog.Logger
}

// NewInMemoryCache is the constructor to create a new in memory cache.
func NewInMemoryCache(logger fflog.Logger) *InMemoryCache {
	return &InMemoryCache{
		Flags:  map[string]flag.FlagData{},
		logger: logger,
	}
}

// addDtoFlag is converting the dto into a flag.FlagData and adds it into the memory cache.
func (fc *InMemoryCache) addDtoFlag(key string, dto flag.DtoFlag) {
	convertedDto := dto.ConvertToFlagData(false)
	fc.addFlag(key, convertedDto)
}

// addFlag is adding a new flag into the cache
func (fc *InMemoryCache) addFlag(key string, value flag.FlagData) {
	fc.Flags[key] = value
}

// getFlag is an accessor to a flag by the name
func (fc *InMemoryCache) getFlag(key string) (flag.Flag, error) {
	f, ok := fc.Flags[key]
	if !ok {
		return &f, fmt.Errorf("flag [%v] does not exists", key)
	}
	return &f, nil
}

// keys returns the list of the flag name in the cache
func (fc *InMemoryCache) keys() []string {
	var keys = make([]string, 0, len(fc.Flags))
	for k := range fc.Flags {
		keys = append(keys, k)
	}
	return keys
}

// Copy is duplicating the cache, it is used when updating the cache.
func (fc *InMemoryCache) Copy() Cache {
	inMemoryCache := NewInMemoryCache(fc.logger)
	for k, v := range fc.Flags {
		inMemoryCache.addFlag(k, v)
	}
	return inMemoryCache
}

// All return all the flags available in the cache.
func (fc *InMemoryCache) All() map[string]flag.Flag {
	c := map[string]flag.Flag{}
	for _, key := range fc.keys() {
		val, _ := fc.getFlag(key)
		c[key] = val
	}
	return c
}

// Init is initializing the cache with all the flags.
func (fc *InMemoryCache) Init(flags map[string]flag.DtoFlag) {
	for key, dto := range flags {
		fc.addDtoFlag(key, dto)
	}
}
