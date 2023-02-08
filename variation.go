// nolint: dupl
package ffclient

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/flagstate"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
)

const (
	errorFlagNotAvailable = "flag %v is not present or disabled"
	errorWrongVariation   = "wrong variation used for flag %v"
)

// BoolVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func BoolVariation(flagKey string, user ffuser.User, defaultValue bool) (bool, error) {
	return ff.BoolVariation(flagKey, user, defaultValue)
}

// BoolVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) BoolVariation(flagKey string, user ffuser.User, defaultValue bool) (bool, error) {
	res, err := getVariation[bool](g, flagKey, user, defaultValue, "bool")
	notifyVariation(g, flagKey, user, res)
	return res.Value, err
}

// IntVariation return the value of the flag in int.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func IntVariation(flagKey string, user ffuser.User, defaultValue int) (int, error) {
	return ff.IntVariation(flagKey, user, defaultValue)
}

// IntVariation return the value of the flag in int.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) IntVariation(flagKey string, user ffuser.User, defaultValue int) (int, error) {
	res, err := getVariation[int](g, flagKey, user, defaultValue, "int")
	notifyVariation(g, flagKey, user, res)
	return res.Value, err
}

// Float64Variation return the value of the flag in float64.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func Float64Variation(flagKey string, user ffuser.User, defaultValue float64) (float64, error) {
	return ff.Float64Variation(flagKey, user, defaultValue)
}

// Float64Variation return the value of the flag in float64.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) Float64Variation(flagKey string, user ffuser.User, defaultValue float64) (float64, error) {
	res, err := getVariation[float64](g, flagKey, user, defaultValue, "float64")
	notifyVariation(g, flagKey, user, res)
	return res.Value, err
}

// StringVariation return the value of the flag in string.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func StringVariation(flagKey string, user ffuser.User, defaultValue string) (string, error) {
	return ff.StringVariation(flagKey, user, defaultValue)
}

// StringVariation return the value of the flag in string.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) StringVariation(flagKey string, user ffuser.User, defaultValue string) (string, error) {
	res, err := getVariation[string](g, flagKey, user, defaultValue, "string")
	notifyVariation(g, flagKey, user, res)
	return res.Value, err
}

// JSONArrayVariation return the value of the flag in []interface{}.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func JSONArrayVariation(flagKey string, user ffuser.User, defaultValue []interface{}) ([]interface{}, error) {
	return ff.JSONArrayVariation(flagKey, user, defaultValue)
}

// JSONArrayVariation return the value of the flag in []interface{}.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) JSONArrayVariation(
	flagKey string, user ffuser.User, defaultValue []interface{},
) ([]interface{}, error) {
	res, err := getVariation[[]interface{}](g, flagKey, user, defaultValue, "[]interface{}")
	notifyVariation(g, flagKey, user, res)
	return res.Value, err
}

// JSONVariation return the value of the flag in map[string]interface{}.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func JSONVariation(
	flagKey string, user ffuser.User, defaultValue map[string]interface{},
) (map[string]interface{}, error) {
	return ff.JSONVariation(flagKey, user, defaultValue)
}

// JSONVariation return the value of the flag in map[string]interface{}.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) JSONVariation(
	flagKey string, user ffuser.User, defaultValue map[string]interface{},
) (map[string]interface{}, error) {
	res, err := getVariation[map[string]interface{}](g, flagKey, user, defaultValue, "map[string]interface{}")
	notifyVariation(g, flagKey, user, res)
	return res.Value, err
}

// AllFlagsState return the values of all the flags for a specific user.
// If valid field is false it means that we had an error when checking the flags.
func AllFlagsState(user ffuser.User) flagstate.AllFlags {
	return ff.AllFlagsState(user)
}

// GetFlagsFromCache returns all the flags present in the cache with their
// current state when calling this method. If cache hasn't been initialized, an
// error reporting this is returned.
func GetFlagsFromCache() (map[string]flag.Flag, error) {
	return ff.GetFlagsFromCache()
}

// AllFlagsState return a flagstate.AllFlags that contains all the flags for a specific user.
func (g *GoFeatureFlag) AllFlagsState(user ffuser.User) flagstate.AllFlags {
	flags := map[string]flag.Flag{}
	if g == nil {
		// empty AllFlags will set valid to false
		return flagstate.AllFlags{}
	}

	if !g.config.Offline {
		var err error
		flags, err = g.cache.AllFlags()
		if err != nil {
			// empty AllFlags will set valid to false
			return flagstate.AllFlags{}
		}
	}

	allFlags := flagstate.NewAllFlags()
	for key, currentFlag := range flags {
		flagValue, resolutionDetails := currentFlag.Value(key, user, flag.EvaluationContext{
			Environment:     g.config.Environment,
			DefaultSdkValue: nil,
		})

		// if the flag is disabled we are ignoring it.
		if resolutionDetails.Reason == flag.ReasonDisabled {
			allFlags.AddFlag(key, flagstate.FlagState{
				Timestamp:   time.Now().Unix(),
				TrackEvents: currentFlag.IsTrackEvents(),
				Failed:      resolutionDetails.ErrorCode != "",
				ErrorCode:   resolutionDetails.ErrorCode,
				Reason:      resolutionDetails.Reason,
			})
			continue
		}

		switch v := flagValue; v.(type) {
		case int, float64, bool, string, []interface{}, map[string]interface{}:
			allFlags.AddFlag(key, flagstate.FlagState{
				Value:         v,
				Timestamp:     time.Now().Unix(),
				VariationType: resolutionDetails.Variant,
				TrackEvents:   currentFlag.IsTrackEvents(),
				Failed:        resolutionDetails.ErrorCode != "",
				ErrorCode:     resolutionDetails.ErrorCode,
				Reason:        resolutionDetails.Reason,
			})

		default:
			defaultVariationName := flag.VariationSDKDefault
			defaultVariationValue := currentFlag.GetVariationValue(defaultVariationName)
			allFlags.AddFlag(
				key,
				flagstate.FlagState{
					Value:         defaultVariationValue,
					Timestamp:     time.Now().Unix(),
					VariationType: defaultVariationName,
					TrackEvents:   currentFlag.IsTrackEvents(),
					Failed:        true,
					ErrorCode:     flag.ErrorCodeTypeMismatch,
					Reason:        flag.ReasonError,
				})
		}
	}
	return allFlags
}

// GetFlagsFromCache returns all the flags present in the cache with their
// current state when calling this method. If cache hasn't been initialized, an
// error reporting this is returned.
func (g *GoFeatureFlag) GetFlagsFromCache() (map[string]flag.Flag, error) {
	return g.cache.AllFlags()
}

// RawVariation return the raw value of the flag (without any types).
// This raw result is mostly used by software built on top of go-feature-flag such as
// go-feature-flag relay proxy.
// If you are using directly the library you should avoid calling this function.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) RawVariation(flagKey string, user ffuser.User, sdkDefaultValue interface{},
) (model.RawVarResult, error) {
	res, err := getVariation[interface{}](g, flagKey, user, sdkDefaultValue, "interface{}")
	notifyVariation(g, flagKey, user, res)
	return model.RawVarResult{
		VariationResult: model.VariationResult{
			TrackEvents:   res.TrackEvents,
			VariationType: res.VariationType,
			Failed:        res.Failed,
			Version:       res.Version,
			Reason:        res.Reason,
			ErrorCode:     res.ErrorCode,
		},
		Value: res.Value,
	}, err
}

// getFlagFromCache try to get the flag from the cache.
// It returns an error if the cache is not init or if the flag is not present or disabled.
func (g *GoFeatureFlag) getFlagFromCache(flagKey string) (flag.Flag, error) {
	f, err := g.cache.GetFlag(flagKey)
	if err != nil {
		return f, fmt.Errorf(errorFlagNotAvailable, flagKey)
	}
	return f, nil
}

// notifyVariation is logging the evaluation result for a flag
// if no logger is provided in the configuration we are not logging anything.
func notifyVariation[T model.JSONType](
	g *GoFeatureFlag,
	flagKey string,
	user ffuser.User,
	result model.GenericVariationResult[T],
) {
	if result.TrackEvents {
		event := exporter.NewFeatureEvent(user, flagKey, result.Value, result.VariationType, result.Failed, result.Version)

		// Add event in the exporter
		if g != nil && g.dataExporter != nil {
			g.dataExporter.AddEvent(event)
		}
	}
}

// getVariation is the internal generic func that handle the logic of a variation the result will always
// contain a valid model.GenericVariationResult
func getVariation[T model.JSONType](
	g *GoFeatureFlag, flagKey string, user ffuser.User, sdkDefaultValue T, expectedType string,
) (model.GenericVariationResult[T], error) {
	if g == nil {
		return model.GenericVariationResult[T]{
			Value:         sdkDefaultValue,
			VariationType: flag.VariationSDKDefault,
			Failed:        true,
			Reason:        flag.ReasonError,
			ErrorCode:     flag.ErrorCodeProviderNotReady,
		}, fmt.Errorf("go-feature-flag is not initialised, default value is used")
	}

	if g.config.Offline {
		return model.GenericVariationResult[T]{
			Value:         sdkDefaultValue,
			VariationType: flag.VariationSDKDefault,
			Failed:        false,
		}, nil
	}

	f, err := g.getFlagFromCache(flagKey)
	if err != nil {
		varResult := model.GenericVariationResult[T]{
			Value:         sdkDefaultValue,
			VariationType: flag.VariationSDKDefault,
			ErrorCode:     flag.ErrorCodeFlagNotFound,
			Failed:        true,
			Reason:        flag.ReasonError,
		}
		if f != nil {
			varResult.TrackEvents = f.IsTrackEvents()
			varResult.Version = f.GetVersion()
		}
		return varResult, err
	}

	flagValue, resolutionDetails := f.Value(flagKey, user,
		flag.EvaluationContext{Environment: g.config.Environment, DefaultSdkValue: sdkDefaultValue})

	var convertedValue interface{}
	switch value := flagValue.(type) {
	case float64:
		// this part ensure that we convert float64 value into int if we call IntVariation on a float64 value.
		if expectedType == "int" {
			convertedValue = int(value)
		} else {
			convertedValue = value
		}
	default:
		convertedValue = value
	}

	var v T
	switch val := convertedValue.(type) {
	case T:
		v = val
	default:
		return model.GenericVariationResult[T]{
			Value:         sdkDefaultValue,
			VariationType: flag.VariationSDKDefault,
			Reason:        flag.ReasonError,
			ErrorCode:     flag.ErrorCodeTypeMismatch,
			Failed:        true,
			TrackEvents:   f.IsTrackEvents(),
			Version:       f.GetVersion(),
		}, fmt.Errorf(errorWrongVariation, flagKey)
	}

	return model.GenericVariationResult[T]{
		Value:         v,
		VariationType: resolutionDetails.Variant,
		Reason:        resolutionDetails.Reason,
		ErrorCode:     resolutionDetails.ErrorCode,
		Failed:        resolutionDetails.ErrorCode != "",
		TrackEvents:   f.IsTrackEvents(),
		Version:       f.GetVersion(),
	}, nil
}
