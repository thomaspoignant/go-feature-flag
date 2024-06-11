package cache

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/internal/dto"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"log/slog"

	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

type InMemoryCache struct {
	Flags  map[string]flag.InternalFlag
	Logger *fflog.FFLogger
}

func NewInMemoryCache(logger *fflog.FFLogger) *InMemoryCache {
	return &InMemoryCache{
		Flags:  map[string]flag.InternalFlag{},
		Logger: logger,
	}
}

func (fc *InMemoryCache) addFlag(key string, value flag.InternalFlag) {
	if err := value.IsValid(); err == nil {
		fc.Flags[key] = value
	} else {
		fc.Logger.Error("[cache] invalid configuration for flag",
			slog.String("key", key), slog.Any("error", err))
	}
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
	inMemoryCache := NewInMemoryCache(fc.Logger)
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
	for key, flagDto := range flags {
		flagToAdd := flagDto.Convert()
		if err := flagToAdd.IsValid(); err == nil {
			cache[key] = flagDto.Convert()
		} else {
			fc.Logger.Error("[cache] invalid configuration for flag",
				slog.String("key", key), slog.Any("error", err))
		}
	}
	fc.Flags = cache
}
