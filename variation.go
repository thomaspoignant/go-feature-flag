package ffclient

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/exporter"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/flagstate"
	flagv1 "github.com/thomaspoignant/go-feature-flag/internal/flagv1"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
)

const errorFlagNotAvailable = "flag %v is not present or disabled"
const errorWrongVariation = "wrong variation used for flag %v"

var offlineVariationResult = model.VariationResult{VariationType: flagv1.VariationSDKDefault, Failed: true}

// BoolVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func BoolVariation(flagKey string, user ffuser.User, defaultValue bool) (bool, error) {
	return ff.BoolVariation(flagKey, user, defaultValue)
}

// IntVariation return the value of the flag in int.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func IntVariation(flagKey string, user ffuser.User, defaultValue int) (int, error) {
	return ff.IntVariation(flagKey, user, defaultValue)
}

// Float64Variation return the value of the flag in float64.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func Float64Variation(flagKey string, user ffuser.User, defaultValue float64) (float64, error) {
	return ff.Float64Variation(flagKey, user, defaultValue)
}

// StringVariation return the value of the flag in string.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func StringVariation(flagKey string, user ffuser.User, defaultValue string) (string, error) {
	return ff.StringVariation(flagKey, user, defaultValue)
}

// JSONArrayVariation return the value of the flag in []interface{}.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func JSONArrayVariation(flagKey string, user ffuser.User, defaultValue []interface{}) ([]interface{}, error) {
	return ff.JSONArrayVariation(flagKey, user, defaultValue)
}

// JSONVariation return the value of the flag in map[string]interface{}.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func JSONVariation(
	flagKey string, user ffuser.User, defaultValue map[string]interface{}) (map[string]interface{}, error) {
	return ff.JSONVariation(flagKey, user, defaultValue)
}

// AllFlagsState return the values of all the flags for a specific user.
// If valid field is false it means that we had an error when checking the flags.
func AllFlagsState(user ffuser.User) flagstate.AllFlags {
	return ff.AllFlagsState(user)
}

// BoolVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) BoolVariation(flagKey string, user ffuser.User, defaultValue bool) (bool, error) {
	res, err := g.boolVariation(flagKey, user, defaultValue)
	g.notifyVariation(flagKey, user, res.VariationResult, res.Value)
	return res.Value, err
}

// IntVariation return the value of the flag in int.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) IntVariation(flagKey string, user ffuser.User, defaultValue int) (int, error) {
	res, err := g.intVariation(flagKey, user, defaultValue)
	g.notifyVariation(flagKey, user, res.VariationResult, res.Value)
	return res.Value, err
}

// Float64Variation return the value of the flag in float64.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) Float64Variation(flagKey string, user ffuser.User, defaultValue float64) (float64, error) {
	res, err := g.float64Variation(flagKey, user, defaultValue)
	g.notifyVariation(flagKey, user, res.VariationResult, res.Value)
	return res.Value, err
}

// StringVariation return the value of the flag in string.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) StringVariation(flagKey string, user ffuser.User, defaultValue string) (string, error) {
	res, err := g.stringVariation(flagKey, user, defaultValue)
	g.notifyVariation(flagKey, user, res.VariationResult, res.Value)
	return res.Value, err
}

// JSONArrayVariation return the value of the flag in []interface{}.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) JSONArrayVariation(
	flagKey string, user ffuser.User, defaultValue []interface{}) ([]interface{}, error) {
	res, err := g.jsonArrayVariation(flagKey, user, defaultValue)
	g.notifyVariation(flagKey, user, res.VariationResult, res.Value)
	return res.Value, err
}

// JSONVariation return the value of the flag in map[string]interface{}.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) JSONVariation(
	flagKey string, user ffuser.User, defaultValue map[string]interface{}) (map[string]interface{}, error) {
	res, err := g.jsonVariation(flagKey, user, defaultValue)
	g.notifyVariation(flagKey, user, res.VariationResult, res.Value)
	return res.Value, err
}

// AllFlagsState return a flagstate.AllFlags that contains all the flags for a specific user.
func (g *GoFeatureFlag) AllFlagsState(user ffuser.User) flagstate.AllFlags {
	flags := cache.FlagsCache{}

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
		flagValue, varType := currentFlag.Value(key, user)
		switch v := flagValue; v.(type) {
		case int, float64, bool, string, []interface{}, map[string]interface{}:
			allFlags.AddFlag(key, flagstate.NewFlagState(currentFlag.GetTrackEvents(), v, varType, false))

		default:
			defaultVariationName := currentFlag.GetDefaultVariation()
			defaultVariationValue := currentFlag.GetVariationValue(defaultVariationName)
			allFlags.AddFlag(
				key, flagstate.NewFlagState(currentFlag.GetTrackEvents(), defaultVariationValue, defaultVariationName, true))
		}
	}
	return allFlags
}

// boolVariation is the internal func that handle the logic of a variation with a bool value
// the result will always contains a valid model.BoolVarResult
func (g *GoFeatureFlag) boolVariation(flagKey string, user ffuser.User, sdkDefaultValue bool,
) (model.BoolVarResult, error) {
	if g.config.Offline {
		return model.BoolVarResult{Value: sdkDefaultValue, VariationResult: offlineVariationResult}, nil
	}

	flag, err := g.getFlagFromCache(flagKey)
	if err != nil {
		return model.BoolVarResult{
			Value:           sdkDefaultValue,
			VariationResult: computeVariationResult(flag, flagv1.VariationSDKDefault, true),
		}, err
	}

	flagValue, variationType := flag.Value(flagKey, user)
	res, ok := flagValue.(bool)
	if !ok {
		return model.BoolVarResult{
			Value:           sdkDefaultValue,
			VariationResult: computeVariationResult(flag, flagv1.VariationSDKDefault, true),
		}, fmt.Errorf(errorWrongVariation, flagKey)
	}
	return model.BoolVarResult{Value: res,
		VariationResult: computeVariationResult(flag, variationType, false),
	}, nil
}

// intVariation is the internal func that handle the logic of a variation with an int value
// the result will always contains a valid model.IntVarResult
func (g *GoFeatureFlag) intVariation(flagKey string, user ffuser.User, sdkDefaultValue int,
) (model.IntVarResult, error) {
	if g.config.Offline {
		return model.IntVarResult{Value: sdkDefaultValue, VariationResult: offlineVariationResult}, nil
	}

	flag, err := g.getFlagFromCache(flagKey)
	if err != nil {
		return model.IntVarResult{Value: sdkDefaultValue,
			VariationResult: computeVariationResult(flag, flagv1.VariationSDKDefault, true),
		}, err
	}

	flagValue, variationType := flag.Value(flagKey, user)
	res, ok := flagValue.(int)
	if !ok {
		// if this is a float64 we convert it to int
		if resFloat, okFloat := flagValue.(float64); okFloat {
			return model.IntVarResult{
				Value:           int(resFloat),
				VariationResult: computeVariationResult(flag, variationType, false),
			}, nil
		}

		return model.IntVarResult{
			Value:           sdkDefaultValue,
			VariationResult: computeVariationResult(flag, flagv1.VariationSDKDefault, true),
		}, fmt.Errorf(errorWrongVariation, flagKey)
	}
	return model.IntVarResult{Value: res,
		VariationResult: computeVariationResult(flag, variationType, false),
	}, nil
}

// float64Variation is the internal func that handle the logic of a variation with a float64 value
// the result will always contains a valid model.Float64VarResult
func (g *GoFeatureFlag) float64Variation(flagKey string, user ffuser.User, sdkDefaultValue float64,
) (model.Float64VarResult, error) {
	if g.config.Offline {
		return model.Float64VarResult{Value: sdkDefaultValue, VariationResult: offlineVariationResult}, nil
	}

	flag, err := g.getFlagFromCache(flagKey)
	if err != nil {
		return model.Float64VarResult{
			Value:           sdkDefaultValue,
			VariationResult: computeVariationResult(flag, flagv1.VariationSDKDefault, true),
		}, err
	}

	flagValue, variationType := flag.Value(flagKey, user)
	res, ok := flagValue.(float64)
	if !ok {
		return model.Float64VarResult{
			Value:           sdkDefaultValue,
			VariationResult: computeVariationResult(flag, flagv1.VariationSDKDefault, true),
		}, fmt.Errorf(errorWrongVariation, flagKey)
	}
	return model.Float64VarResult{
		Value:           res,
		VariationResult: computeVariationResult(flag, variationType, false),
	}, nil
}

// stringVariation is the internal func that handle the logic of a variation with a string value
// the result will always contains a valid model.StringVarResult
func (g *GoFeatureFlag) stringVariation(flagKey string, user ffuser.User, sdkDefaultValue string,
) (model.StringVarResult, error) {
	if g.config.Offline {
		return model.StringVarResult{Value: sdkDefaultValue, VariationResult: offlineVariationResult}, nil
	}

	flag, err := g.getFlagFromCache(flagKey)
	if err != nil {
		return model.StringVarResult{
			Value:           sdkDefaultValue,
			VariationResult: computeVariationResult(flag, flagv1.VariationSDKDefault, true),
		}, err
	}

	flagValue, variationType := flag.Value(flagKey, user)
	res, ok := flagValue.(string)
	if !ok {
		return model.StringVarResult{
			Value:           sdkDefaultValue,
			VariationResult: computeVariationResult(flag, flagv1.VariationSDKDefault, true),
		}, fmt.Errorf(errorWrongVariation, flagKey)
	}
	return model.StringVarResult{
		Value:           res,
		VariationResult: computeVariationResult(flag, variationType, false),
	}, nil
}

// jsonArrayVariation is the internal func that handle the logic of a variation with a json value
// the result will always contains a valid model.JSONArrayVarResult
func (g *GoFeatureFlag) jsonArrayVariation(flagKey string, user ffuser.User, sdkDefaultValue []interface{},
) (model.JSONArrayVarResult, error) {
	if g.config.Offline {
		return model.JSONArrayVarResult{Value: sdkDefaultValue, VariationResult: offlineVariationResult}, nil
	}

	flag, err := g.getFlagFromCache(flagKey)
	if err != nil {
		return model.JSONArrayVarResult{
			Value:           sdkDefaultValue,
			VariationResult: computeVariationResult(flag, flagv1.VariationSDKDefault, true),
		}, err
	}

	flagValue, variationType := flag.Value(flagKey, user)
	res, ok := flagValue.([]interface{})
	if !ok {
		return model.JSONArrayVarResult{
			Value:           sdkDefaultValue,
			VariationResult: computeVariationResult(flag, flagv1.VariationSDKDefault, true),
		}, fmt.Errorf(errorWrongVariation, flagKey)
	}
	return model.JSONArrayVarResult{
		Value:           res,
		VariationResult: computeVariationResult(flag, variationType, false),
	}, nil
}

// jsonVariation is the internal func that handle the logic of a variation with a json value
// the result will always contains a valid model.JSONVarResult
func (g *GoFeatureFlag) jsonVariation(flagKey string, user ffuser.User, sdkDefaultValue map[string]interface{},
) (model.JSONVarResult, error) {
	if g.config.Offline {
		return model.JSONVarResult{Value: sdkDefaultValue, VariationResult: offlineVariationResult}, nil
	}

	flag, err := g.getFlagFromCache(flagKey)
	if err != nil {
		return model.JSONVarResult{
			Value:           sdkDefaultValue,
			VariationResult: computeVariationResult(flag, flagv1.VariationSDKDefault, true),
		}, err
	}

	flagValue, variationType := flag.Value(flagKey, user)
	res, ok := flagValue.(map[string]interface{})
	if !ok {
		return model.JSONVarResult{
			Value:           sdkDefaultValue,
			VariationResult: computeVariationResult(flag, flagv1.VariationSDKDefault, true),
		}, fmt.Errorf(errorWrongVariation, flagKey)
	}
	return model.JSONVarResult{
		Value:           res,
		VariationResult: computeVariationResult(flag, variationType, false),
	}, nil
}

// computeVariationResult is creating a model.VariationResult
func computeVariationResult(flag flag.Flag, variationType string, failed bool) model.VariationResult {
	varResult := model.VariationResult{
		VariationType: variationType,
		Failed:        failed,
	}

	if flag != nil {
		varResult.TrackEvents = flag.GetTrackEvents()
		varResult.Version = flag.GetVersion()
	}

	return varResult
}

// notifyVariation is logging the evaluation result for a flag
// if no logger is provided in the configuration we are not logging anything.
func (g *GoFeatureFlag) notifyVariation(
	flagKey string,
	user ffuser.User,
	result model.VariationResult,
	value interface{}) {
	if result.TrackEvents {
		event := exporter.NewFeatureEvent(user, flagKey, value, result.VariationType, result.Failed, result.Version)

		// Add event in the exporter
		if g.dataExporter != nil {
			g.dataExporter.AddEvent(event)
		}
	}
}

// getFlagFromCache try to get the flag from the cache.
// It returns an error if the cache is not init or if the flag is not present or disabled.
func (g *GoFeatureFlag) getFlagFromCache(flagKey string) (flag.Flag, error) {
	flag, err := g.cache.GetFlag(flagKey)
	if err != nil || flag.GetDisable() {
		return flag, fmt.Errorf(errorFlagNotAvailable, flagKey)
	}
	return flag, nil
}
