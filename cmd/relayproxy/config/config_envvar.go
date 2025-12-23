package config

import (
	"slices"
	"strconv"
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

func (c *ConfigLoader) mapEnvVariablesProvider(prefix string, log *zap.Logger) koanf.Provider {
	return env.ProviderWithValue(prefix, ".", func(key, v string) (string, any) {
		key = strings.TrimPrefix(key, prefix)
		switch {
		case strings.HasPrefix(key, "RETRIEVERS"),
			strings.HasPrefix(key, "NOTIFIER"),
			strings.HasPrefix(key, "NOTIFIERS"),
			strings.HasPrefix(key, "FLAGSETS"),
			strings.HasPrefix(key, "EXPORTERS"):
			configMap := c.k.Raw()
			modifiedConfigMap, err := loadArrayEnv(key, v, configMap)
			if err != nil {
				log.Error(
					"config: error loading array env",
					zap.String("key", key),
					zap.String("value", v),
					zap.Error(err),
				)
				return key, v
			}
			// Update the global config with the modified configMap
			for configKey, configValue := range modifiedConfigMap {
				_ = c.k.Set(configKey, configValue)
			}
			return key, v
		case strings.HasSuffix(key, "KAFKA_ADDRESSES"),
			strings.HasSuffix(key, "APIKEYS"),
			strings.HasPrefix(key, "AUTHORIZEDKEYS_EVALUATION"),
			strings.HasPrefix(key, "AUTHORIZEDKEYS_ADMIN"):
			transformedKey := strings.ReplaceAll(strings.ToLower(key), "_", ".")
			return transformedKey, strings.Split(v, ",")
		case key == "OTEL_RESOURCE_ATTRIBUTES":
			c.parseOtelResourceAttributes(v, log)
			return key, v
		default:
			return strings.ReplaceAll(strings.ToLower(key), "_", "."), v
		}
	})
}

// Load the ENV Like:RETRIEVERS_0_HEADERS_AUTHORIZATION
func loadArrayEnv(s string, v string, configMap map[string]any) (map[string]any, error) {
	paths := normalizePaths(s)
	prefixKey := paths[0]

	configArray, ok := configMap[prefixKey].([]any)
	if !ok {
		return configMap, nil
	}

	index, err := strconv.Atoi(paths[1])
	if err != nil {
		return configMap, err
	}

	configItem := fetchOrInitConfigItemAtIndex(configArray, index)
	keys := paths[2:]

	if shouldHandleRecursively(prefixKey, keys) {
		return handleRecursiveConfig(keys, v, configItem, configArray, index, configMap, prefixKey)
	}

	return handleDirectConfig(keys, v, configItem, configArray, index, configMap, prefixKey)
}

// normalizePaths splits the input string and converts all parts to lowercase
func normalizePaths(s string) []string {
	paths := strings.Split(s, "_")
	for i, str := range paths {
		paths[i] = strings.ToLower(str)
	}
	return paths
}

// fetchOrInitConfigItemAtIndex retrieves or creates a config item at the specified index
func fetchOrInitConfigItemAtIndex(configArray []any, index int) map[string]any {
	outRange := index > len(configArray)-1
	if outRange {
		return make(map[string]any)
	}
	return configArray[index].(map[string]any)
}

// shouldHandleRecursively determines if the configuration should be handled recursively
func shouldHandleRecursively(prefixKey string, keys []string) bool {
	if prefixKey != "flagsets" || len(keys) < 1 {
		return false
	}

	recursiveKeys := []string{"retrievers", "notifier", "notifiers", "exporters"}
	return slices.Contains(recursiveKeys, keys[0])
}

// handleRecursiveConfig processes recursive configuration for flagsets
func handleRecursiveConfig(
	keys []string,
	v string,
	configItem map[string]any,
	configArray []any,
	index int,
	configMap map[string]any,
	prefixKey string,
) (map[string]any, error) {
	recursiveKey := strings.Join(keys, "_")
	modifiedNestedConfig, err := loadArrayEnv(recursiveKey, v, configItem)
	if err != nil {
		return configMap, err
	}

	for k, val := range modifiedNestedConfig {
		configItem[k] = val
	}

	return updateConfigArray(configArray, index, configItem, configMap, prefixKey)
}

// handleDirectConfig processes direct configuration assignment
func handleDirectConfig(keys []string,
	v string,
	configItem map[string]any,
	configArray []any,
	index int,
	configMap map[string]any,
	prefixKey string,
) (map[string]any, error) {
	currentMap := configItem

	for i, key := range keys {
		currentMap = ensureMapExists(currentMap, key, i, len(keys)-1)
	}

	lastKey := keys[len(keys)-1]
	value := parseValue(lastKey, keys, v)
	currentMap[lastKey] = value

	return updateConfigArray(configArray, index, configItem, configMap, prefixKey)
}

// ensureMapExists ensures a map exists at the specified key path
func ensureMapExists(
	currentMap map[string]any,
	key string,
	currentIndex,
	lastIndex int,
) map[string]any {
	next, ok := currentMap[key].(map[string]any)
	if ok {
		return next
	}

	if currentIndex != lastIndex {
		newMap := make(map[string]any)
		currentMap[key] = newMap
		return newMap
	}

	return currentMap
}

// parseValue parses the value based on the key type
func parseValue(lastKey string, keys []string, v string) any {
	if isArrayValue(lastKey, keys) {
		return parseArrayValue(v)
	}
	return v
}

// isArrayValue determines if the value should be treated as an array
func isArrayValue(lastKey string, keys []string) bool {
	return (lastKey == "addresses" && len(keys) > 1 && keys[len(keys)-2] == "kafka") ||
		lastKey == "apikeys"
}

// parseArrayValue splits a comma-separated string and trims whitespace
func parseArrayValue(v string) []string {
	split := strings.Split(v, ",")
	for i, item := range split {
		split[i] = strings.TrimSpace(item)
	}
	return split
}

// updateConfigArray updates the configuration array and returns the modified config map
func updateConfigArray(configArray []any, index int, configItem map[string]any,
	configMap map[string]any, prefixKey string) (map[string]any, error) {
	outRange := index > len(configArray)-1

	if outRange {
		configArray = expandArray(configArray, index)
	}

	configArray[index] = configItem
	configMap[prefixKey] = configArray

	return configMap, nil
}

// expandArray expands the array to accommodate the new index
func expandArray(configArray []any, index int) []any {
	blank := index - len(configArray) + 1
	for i := 0; i < blank; i++ {
		configArray = append(configArray, make(map[string]any))
	}
	return configArray
}

// parseOtelResourceAttributes parses the OTEL_RESOURCE_ATTRIBUTES environment variable
// and sets the attributes in the koanf configuration.
// The expected format is "key1=value1,key2=value2,..."
func (c *ConfigLoader) parseOtelResourceAttributes(attributes string, log *zap.Logger) {
	configMap := c.k.Raw()
	otel, ok := configMap["otel"].(map[string]any)
	if !ok {
		configMap["otel"] = make(map[string]any)
		otel = configMap["otel"].(map[string]any)
	}

	resource, ok := otel["resource"].(map[string]any)
	if !ok {
		otel["resource"] = make(map[string]any)
		resource = otel["resource"].(map[string]any)
	}

	attrs, ok := resource["attributes"].(map[string]any)
	if !ok {
		resource["attributes"] = make(map[string]any)
		attrs = resource["attributes"].(map[string]any)
	}

	for _, attr := range strings.Split(attributes, ",") {
		k, v, found := strings.Cut(attr, "=")
		if !found {
			log.Error("config: error loading OTEL_RESOURCE_ATTRIBUTES - incorrect format",
				zap.String("key", k), zap.String("value", v))
			continue
		}

		attrs[k] = v
	}

	_ = c.k.Set("otel", otel)
}
