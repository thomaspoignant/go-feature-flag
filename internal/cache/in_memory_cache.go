package cache

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/flagv1"
	"github.com/thomaspoignant/go-feature-flag/internal/flagv2"
)

type InMemoryCache struct {
	flagsV1 map[string]flagv1.FlagData
	flagsV2 map[string]flagv2.FlagData
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		flagsV1: map[string]flagv1.FlagData{},
		flagsV2: map[string]flagv2.FlagData{},
	}
}

func (fc *InMemoryCache) addDtoFlag(key string, dto flag.DtoFlag) {
	if dto.IsFlagV2() {
		fc.addFlag(key, dto.ToFlagV2())
	} else {
		fc.addFlag(key, dto.ToFlagV1())
	}
}

func (fc *InMemoryCache) addFlag(key string, fd flagv2.FlagData) {
	fc.flagsV2[key] = fd
}

func (fc *InMemoryCache) getFlag(key string) (flag.Flag, error) {
	fV1, okV1 := fc.flagsV1[key]
	fV2, okV2 := fc.flagsV2[key]
	if !okV1 && !okV2 {
		return nil, fmt.Errorf("flag [%v] does not exists", key)
	}
	if okV2 {
		return &fV2, nil
	}
	return &fV1, nil
}

func (fc *InMemoryCache) keys() []string {
	var keys = make([]string, 0, len(fc.flagsV1)+len(fc.flagsV2))
	for k := range fc.flagsV1 {
		keys = append(keys, k)
	}
	for k := range fc.flagsV2 {
		keys = append(keys, k)
	}
	return keys
}

func (fc *InMemoryCache) Copy() Cache {
	inMemoryCache := NewInMemoryCache()
	for k, v := range fc.flagsV2 {
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

func (fc *InMemoryCache) Init(flags map[string]flag.DtoFlag) {
	for key, dto := range flags {
		fc.addDtoFlag(key, dto)
	}
}
