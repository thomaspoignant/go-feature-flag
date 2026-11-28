package ffclient

import (
	"errors"
	"slices"
)

// GetFlagMetadata returns the configured metadata for flagKey from the default client's cache.
func GetFlagMetadata(flagKey string) (map[string]any, error) {
	return ff.GetFlagMetadata(flagKey)
}

// GetFlagMetadata reads a flag's configured metadata without evaluating the flag.
// The caller may modify the returned map and its contents. A flag without metadata
// returns nil, nil. A missing flag or an unavailable cache returns an error.
func (g *GoFeatureFlag) GetFlagMetadata(flagKey string) (map[string]any, error) {
	if g == nil || g.retrieverManager == nil {
		return nil, errors.New("cache is not initialized")
	}
	f, err := g.retrieverManager.GetFlag(flagKey)
	if err != nil {
		return nil, err
	}
	return cloneMetadataMap(f.GetMetadata()), nil
}

func cloneMetadataMap[K comparable](src map[K]any) map[K]any {
	if src == nil {
		return nil
	}
	dst := make(map[K]any, len(src))
	for key, value := range src {
		dst[key] = cloneMetadataValue(value)
	}
	return dst
}

// Nested YAML mappings can also have non-string keys.
func cloneMetadataValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return cloneMetadataMap(v)
	case map[any]any:
		return cloneMetadataMap(v)
	case []any:
		out := slices.Clone(v)
		for i, element := range v {
			out[i] = cloneMetadataValue(element)
		}
		return out
	default:
		return value
	}
}
