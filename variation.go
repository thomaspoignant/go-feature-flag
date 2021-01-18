package ffclient

import (
	"fmt"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/flags"
)

const errorFlagNotAvailable = "flag %v is not present or disabled"
const errorWrongVariation = "wrong variation used for flag %v"

// BoolVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func BoolVariation(flagKey string, user ffuser.User, defaultValue bool) (bool, error) {
	flag, err := getFlagFromCache(flagKey)
	if err != nil {
		notifyVariation(flagKey, user.GetKey(), defaultValue)
		return defaultValue, err
	}

	res, ok := flag.Value(flagKey, user).(bool)
	if !ok {
		notifyVariation(flagKey, user.GetKey(), defaultValue)
		return defaultValue, fmt.Errorf(errorWrongVariation, flagKey)
	}
	notifyVariation(flagKey, user.GetKey(), res)
	return res, nil
}

// IntVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func IntVariation(flagKey string, user ffuser.User, defaultValue int) (int, error) {
	flag, err := getFlagFromCache(flagKey)
	if err != nil {
		notifyVariation(flagKey, user.GetKey(), defaultValue)
		return defaultValue, err
	}

	res, ok := flag.Value(flagKey, user).(int)
	if !ok {
		notifyVariation(flagKey, user.GetKey(), defaultValue)
		return defaultValue, fmt.Errorf(errorWrongVariation, flagKey)
	}
	notifyVariation(flagKey, user.GetKey(), res)
	return res, nil
}

// Float64Variation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func Float64Variation(flagKey string, user ffuser.User, defaultValue float64) (float64, error) {
	flag, err := getFlagFromCache(flagKey)
	if err != nil {
		notifyVariation(flagKey, user.GetKey(), defaultValue)
		return defaultValue, err
	}

	res, ok := flag.Value(flagKey, user).(float64)
	if !ok {
		notifyVariation(flagKey, user.GetKey(), defaultValue)
		return defaultValue, fmt.Errorf(errorWrongVariation, flagKey)
	}
	notifyVariation(flagKey, user.GetKey(), res)
	return res, nil
}

// StringVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func StringVariation(flagKey string, user ffuser.User, defaultValue string) (string, error) {
	flag, err := getFlagFromCache(flagKey)
	if err != nil {
		notifyVariation(flagKey, user.GetKey(), defaultValue)
		return defaultValue, err
	}

	res, ok := flag.Value(flagKey, user).(string)
	if !ok {
		notifyVariation(flagKey, user.GetKey(), defaultValue)
		return defaultValue, fmt.Errorf(errorWrongVariation, flagKey)
	}
	notifyVariation(flagKey, user.GetKey(), res)
	return res, nil
}

// JSONArrayVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func JSONArrayVariation(flagKey string, user ffuser.User, defaultValue []interface{}) ([]interface{}, error) {
	flag, err := getFlagFromCache(flagKey)
	if err != nil {
		notifyVariation(flagKey, user.GetKey(), defaultValue)
		return defaultValue, err
	}

	res, ok := flag.Value(flagKey, user).([]interface{})
	if !ok {
		notifyVariation(flagKey, user.GetKey(), defaultValue)
		return defaultValue, fmt.Errorf(errorWrongVariation, flagKey)
	}
	notifyVariation(flagKey, user.GetKey(), res)
	return res, nil
}

// JSONVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func JSONVariation(
	flagKey string, user ffuser.User, defaultValue map[string]interface{}) (map[string]interface{}, error) {
	flag, err := getFlagFromCache(flagKey)
	if err != nil {
		notifyVariation(flagKey, user.GetKey(), defaultValue)
		return defaultValue, err
	}

	res, ok := flag.Value(flagKey, user).(map[string]interface{})
	if !ok {
		notifyVariation(flagKey, user.GetKey(), defaultValue)
		return defaultValue, fmt.Errorf(errorWrongVariation, flagKey)
	}
	notifyVariation(flagKey, user.GetKey(), res)
	return res, nil
}

// notifyVariation is logging the evaluation result for a flag
// if no logger is provided in the configuration we are not logging anything.
func notifyVariation(flagKey string, userKey string, value interface{}) {
	if ff.config.Logger != nil {
		ff.config.Logger.Printf(
			"[%v] user=\"%s\", flag=\"%s\", value=\"%v\"",
			time.Now().Format(time.RFC3339), userKey, flagKey, value)
	}
}

// getFlagFromCache try to get the flag from the cache.
// It returns an error if the cache is not init or if the flag is not present or disabled.
func getFlagFromCache(flagKey string) (flags.Flag, error) {
	flag, err := ff.cache.GetFlag(flagKey)
	if err != nil || flag.Disable {
		return flag, fmt.Errorf(errorFlagNotAvailable, flagKey)
	}
	return flag, nil
}
