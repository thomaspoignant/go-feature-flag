package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

func mapEnvVariablesProvider(prefix string, log *zap.Logger) koanf.Provider {
	return env.ProviderWithValue(prefix, ".", func(key string, v string) (string, interface{}) {
		key = strings.TrimPrefix(key, prefix)
		switch {
		case strings.HasPrefix(key, "RETRIEVERS"),
			strings.HasPrefix(key, "NOTIFIER"),
			strings.HasPrefix(key, "NOTIFIERS"),
			strings.HasPrefix(key, "FLAGSETS"),
			strings.HasPrefix(key, "EXPORTERS"):
			configMap := k.Raw()
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
				_ = k.Set(configKey, configValue)
			}
			return key, v
			// case strings.HasPrefix(key, "FLAGSETS"):
			// 	parseFlagSetsEnv(key, v, log)
			// return key, v
		case strings.HasSuffix(key, "KAFKA_ADDRESSES"):
			return "exporter.kafka.addresses", strings.Split(v, ",")
		case strings.HasPrefix(key, "AUTHORIZEDKEYS_EVALUATION"):
			return "authorizedKeys.evaluation", strings.Split(v, ",")
		case strings.HasPrefix(key, "AUTHORIZEDKEYS_ADMIN"):
			return "authorizedKeys.admin", strings.Split(v, ",")
		case key == "OTEL_RESOURCE_ATTRIBUTES":
			parseOtelResourceAttributes(v, log)
			return key, v
		default:
			return strings.ReplaceAll(strings.ToLower(key), "_", "."), v
		}
	})
}

// Load the ENV Like:RETRIEVERS_0_HEADERS_AUTHORIZATION
func loadArrayEnv(s string, v string, configMap map[string]interface{}) (map[string]interface{}, error) {
	paths := strings.Split(s, "_")
	for i, str := range paths {
		paths[i] = strings.ToLower(str)
	}
	prefixKey := paths[0]
	if configArray, ok := configMap[prefixKey].([]interface{}); ok {
		index, err := strconv.Atoi(paths[1])
		if err != nil {
			return configMap, err
		}
		var configItem map[string]interface{}
		outRange := index > len(configArray)-1
		if outRange {
			configItem = make(map[string]interface{})
		} else {
			configItem = configArray[index].(map[string]interface{})
		}

		keys := paths[2:]
		currentMap := configItem
		for i, key := range keys {
			hasKey := false
			lowerKey := key
			for y := range currentMap {
				if y != lowerKey {
					continue
				}
				if nextMap, ok := currentMap[y].(map[string]interface{}); ok {
					currentMap = nextMap
					hasKey = true
					break
				}
			}
			if !hasKey && i != len(keys)-1 {
				newMap := make(map[string]interface{})
				currentMap[lowerKey] = newMap
				currentMap = newMap
			}
		}
		lastKey := keys[len(keys)-1]
		currentMap[lastKey] = v
		if outRange {
			blank := index - len(configArray) + 1
			for i := 0; i < blank; i++ {
				configArray = append(configArray, make(map[string]interface{}))
			}
			configArray[index] = configItem
		} else {
			configArray[index] = configItem
		}
		configMap[prefixKey] = configArray
	}
	return configMap, nil
}

func parseFlagSetsEnv(key, v string, log *zap.Logger) {
	configMap := k.Raw()
	flagSets, ok := configMap["flagsets"].([]interface{})
	if !ok {
		flagSets = make([]interface{}, 0)
		// TODO - should ignore if index to be created is not the last one
	}
	flagSets = configMap["flagsets"].([]interface{})
	keyIndex := strings.Split(key, "_")[1]
	index, err := strconv.Atoi(keyIndex)
	if err != nil {
		log.Error("config: error loading FLAGSETS - incorrect format",
			zap.String("key", key), zap.String("value", v))
		return
	}
	flagSet := flagSets[index]
	flagSetMap, ok := flagSet.(map[string]interface{})
	if !ok {
		log.Error("config: error loading FLAGSETS - incorrect format",
			zap.String("key", key), zap.String("value", v))
		return
	}
	fmt.Println(flagSetMap)

	panic("unimplemented")
}

// parseOtelResourceAttributes parses the OTEL_RESOURCE_ATTRIBUTES environment variable
// and sets the attributes in the koanf configuration.
// The expected format is "key1=value1,key2=value2,..."
func parseOtelResourceAttributes(attributes string, log *zap.Logger) {
	configMap := k.Raw()
	otel, ok := configMap["otel"].(map[string]interface{})
	if !ok {
		configMap["otel"] = make(map[string]interface{})
		otel = configMap["otel"].(map[string]interface{})
	}

	resource, ok := otel["resource"].(map[string]interface{})
	if !ok {
		otel["resource"] = make(map[string]interface{})
		resource = otel["resource"].(map[string]interface{})
	}

	attrs, ok := resource["attributes"].(map[string]interface{})
	if !ok {
		resource["attributes"] = make(map[string]interface{})
		attrs = resource["attributes"].(map[string]interface{})
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

	_ = k.Set("otel", otel)
}
