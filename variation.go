package ffclient

import (
	"errors"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
)

const errorCacheNotInit = "impossible to read the toggle before the initialisation"

// BoolVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func BoolVariation(flagKey string, user ffuser.User, defaultValue bool) (bool, error) {
	if !cacheIsInitialized() {
		return false, errors.New(errorCacheNotInit)
	}

	flag, ok := cache.FlagsCache[flagKey]
	if !ok || flag.Disable {
		return defaultValue, nil
	}

	res, ok := flag.Value(flagKey, user).(bool)
	if !ok {
		return defaultValue, nil
	}
	return res, nil
}

// IntVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func IntVariation(flagKey string, user ffuser.User, defaultValue int) (int, error) {
	if !cacheIsInitialized() {
		return 0, errors.New(errorCacheNotInit)
	}

	flag, ok := cache.FlagsCache[flagKey]
	if !ok || flag.Disable {
		return defaultValue, nil
	}

	res, ok := flag.Value(flagKey, user).(int)
	if !ok {
		return defaultValue, nil
	}
	return res, nil
}

// Float64Variation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func Float64Variation(flagKey string, user ffuser.User, defaultValue float64) (float64, error) {
	if !cacheIsInitialized() {
		return 0, errors.New(errorCacheNotInit)
	}

	flag, ok := cache.FlagsCache[flagKey]
	if !ok || flag.Disable {
		return defaultValue, nil
	}

	res, ok := flag.Value(flagKey, user).(float64)
	if !ok {
		return defaultValue, nil
	}
	return res, nil
}

// StringVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func StringVariation(flagKey string, user ffuser.User, defaultValue string) (string, error) {
	if !cacheIsInitialized() {
		return "", errors.New(errorCacheNotInit)
	}

	flag, ok := cache.FlagsCache[flagKey]
	if !ok || flag.Disable {
		return defaultValue, nil
	}

	res, ok := flag.Value(flagKey, user).(string)
	if !ok || flag.Disable {
		return defaultValue, nil
	}
	return res, nil
}

// JSONArrayVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func JSONArrayVariation(flagKey string, user ffuser.User, defaultValue []interface{}) ([]interface{}, error) {
	if !cacheIsInitialized() {
		return nil, errors.New(errorCacheNotInit)
	}

	flag, ok := cache.FlagsCache[flagKey]
	if !ok || flag.Disable {
		return defaultValue, nil
	}

	res, ok := flag.Value(flagKey, user).([]interface{})
	if !ok {
		return defaultValue, nil
	}
	return res, nil
}

// JSONVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func JSONVariation(
	flagKey string, user ffuser.User, defaultValue map[string]interface{}) (map[string]interface{}, error) {
	if !cacheIsInitialized() {
		return nil, errors.New(errorCacheNotInit)
	}

	flag, ok := cache.FlagsCache[flagKey]
	if !ok || flag.Disable {
		return defaultValue, nil
	}

	res, ok := flag.Value(flagKey, user).(map[string]interface{})
	if !ok {
		return defaultValue, nil
	}
	return res, nil
}

func cacheIsInitialized() bool {
	return cache.FlagsCache != nil
}
