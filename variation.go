package ffclient

import (
	"fmt"
	"maps"

	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/model"
	"github.com/thomaspoignant/go-feature-flag/modules/evaluation"
)

const (
	errorFlagNotAvailable = "flag %v is not present or disabled"
)

// BoolVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func BoolVariation(flagKey string, ctx ffcontext.Context, defaultValue bool) (bool, error) {
	return ff.BoolVariation(flagKey, ctx, defaultValue)
}

// BoolVariation return the value of the flag in boolean.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) BoolVariation(
	flagKey string,
	ctx ffcontext.Context,
	defaultValue bool,
) (bool, error) {
	res, err := g.BoolVariationDetails(flagKey, ctx, defaultValue)
	return res.Value, err
}

// BoolVariationDetails return the details of the evaluation for boolean flag.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func BoolVariationDetails(flagKey string, ctx ffcontext.Context, defaultValue bool) (
	model.VariationResult[bool], error) {
	return ff.BoolVariationDetails(flagKey, ctx, defaultValue)
}

// BoolVariationDetails return the details of the evaluation for boolean flag.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) BoolVariationDetails(
	flagKey string,
	ctx ffcontext.Context,
	defaultValue bool,
) (model.VariationResult[bool], error) {
	res, err := getVariation(g, flagKey, ctx, defaultValue, "bool")
	notifyVariation(g, flagKey, ctx, res)
	return res, err
}

// IntVariation return the value of the flag in int.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func IntVariation(flagKey string, ctx ffcontext.Context, defaultValue int) (int, error) {
	return ff.IntVariation(flagKey, ctx, defaultValue)
}

// IntVariation return the value of the flag in int.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) IntVariation(
	flagKey string,
	ctx ffcontext.Context,
	defaultValue int,
) (int, error) {
	res, err := g.IntVariationDetails(flagKey, ctx, defaultValue)
	return res.Value, err
}

// IntVariationDetails return the details of the evaluation for int flag.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func IntVariationDetails(
	flagKey string,
	ctx ffcontext.Context,
	defaultValue int,
) (model.VariationResult[int], error) {
	return ff.IntVariationDetails(flagKey, ctx, defaultValue)
}

// IntVariationDetails return the details of the evaluation for int flag.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) IntVariationDetails(flagKey string, ctx ffcontext.Context, defaultValue int,
) (model.VariationResult[int], error) {
	res, err := getVariation(g, flagKey, ctx, defaultValue, "int")
	notifyVariation(g, flagKey, ctx, res)
	return res, err
}

// Float64Variation return the value of the flag in float64.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func Float64Variation(
	flagKey string,
	ctx ffcontext.Context,
	defaultValue float64,
) (float64, error) {
	return ff.Float64Variation(flagKey, ctx, defaultValue)
}

// Float64Variation return the value of the flag in float64.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) Float64Variation(
	flagKey string,
	ctx ffcontext.Context,
	defaultValue float64,
) (float64, error) {
	res, err := g.Float64VariationDetails(flagKey, ctx, defaultValue)
	return res.Value, err
}

// Float64VariationDetails return the details of the evaluation for float64 flag.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func Float64VariationDetails(flagKey string, ctx ffcontext.Context, defaultValue float64,
) (model.VariationResult[float64], error) {
	return ff.Float64VariationDetails(flagKey, ctx, defaultValue)
}

// Float64VariationDetails return the details of the evaluation for float64 flag.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) Float64VariationDetails(
	flagKey string,
	ctx ffcontext.Context,
	defaultValue float64,
) (model.VariationResult[float64], error) {
	res, err := getVariation(g, flagKey, ctx, defaultValue, "float64")
	notifyVariation(g, flagKey, ctx, res)
	return res, err
}

// StringVariation return the value of the flag in string.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func StringVariation(flagKey string, ctx ffcontext.Context, defaultValue string) (string, error) {
	return ff.StringVariation(flagKey, ctx, defaultValue)
}

// StringVariation return the value of the flag in string.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) StringVariation(
	flagKey string,
	ctx ffcontext.Context,
	defaultValue string,
) (string, error) {
	res, err := g.StringVariationDetails(flagKey, ctx, defaultValue)
	return res.Value, err
}

// StringVariationDetails return the details of the evaluation for string flag.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func StringVariationDetails(flagKey string, ctx ffcontext.Context, defaultValue string,
) (model.VariationResult[string], error) {
	return ff.StringVariationDetails(flagKey, ctx, defaultValue)
}

// StringVariationDetails return the details of the evaluation for string flag.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) StringVariationDetails(
	flagKey string,
	ctx ffcontext.Context,
	defaultValue string,
) (model.VariationResult[string], error) {
	res, err := getVariation(g, flagKey, ctx, defaultValue, "string")
	notifyVariation(g, flagKey, ctx, res)
	return res, err
}

// JSONArrayVariation return the value of the flag in []interface{}.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func JSONArrayVariation(
	flagKey string,
	ctx ffcontext.Context,
	defaultValue []interface{},
) ([]interface{}, error) {
	return ff.JSONArrayVariation(flagKey, ctx, defaultValue)
}

// JSONArrayVariation return the value of the flag in []interface{}.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) JSONArrayVariation(
	flagKey string, ctx ffcontext.Context, defaultValue []interface{},
) ([]interface{}, error) {
	res, err := g.JSONArrayVariationDetails(flagKey, ctx, defaultValue)
	return res.Value, err
}

// JSONArrayVariationDetails return the details of the evaluation for []interface{} flag.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func JSONArrayVariationDetails(flagKey string, ctx ffcontext.Context, defaultValue []interface{},
) (model.VariationResult[[]interface{}], error) {
	return ff.JSONArrayVariationDetails(flagKey, ctx, defaultValue)
}

// JSONArrayVariationDetails return the details of the evaluation for []interface{} flag.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) JSONArrayVariationDetails(
	flagKey string,
	ctx ffcontext.Context,
	defaultValue []interface{},
) (model.VariationResult[[]interface{}], error) {
	res, err := getVariation(g, flagKey, ctx, defaultValue, "[]interface{}")
	notifyVariation(g, flagKey, ctx, res)
	return res, err
}

// JSONVariation return the value of the flag in map[string]interface{}.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func JSONVariation(
	flagKey string, ctx ffcontext.Context, defaultValue map[string]interface{},
) (map[string]interface{}, error) {
	return ff.JSONVariation(flagKey, ctx, defaultValue)
}

// JSONVariation return the value of the flag in map[string]interface{}.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) JSONVariation(
	flagKey string, ctx ffcontext.Context, defaultValue map[string]interface{},
) (map[string]interface{}, error) {
	res, err := g.JSONVariationDetails(flagKey, ctx, defaultValue)
	return res.Value, err
}

// JSONVariationDetails return the details of the evaluation for map[string]interface{} flag.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
func JSONVariationDetails(
	flagKey string,
	ctx ffcontext.Context,
	defaultValue map[string]interface{},
) (model.VariationResult[map[string]interface{}], error) {
	return ff.JSONVariationDetails(flagKey, ctx, defaultValue)
}

// JSONVariationDetails return the details of the evaluation for []interface{} flag.
// An error is return if you don't have init the library before calling the function.
// If the key does not exist we return the default value.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) JSONVariationDetails(
	flagKey string,
	ctx ffcontext.Context,
	defaultValue map[string]interface{},
) (model.VariationResult[map[string]interface{}], error) {
	res, err := getVariation(g, flagKey, ctx, defaultValue, "bool")
	notifyVariation(g, flagKey, ctx, res)
	return res, err
}

// RawVariation return the raw value of the flag (without any types).
// This raw result is mostly used by software built on top of go-feature-flag such as
// go-feature-flag relay proxy.
// If you are using directly the library you should avoid calling this function.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) RawVariation(
	flagKey string,
	ctx ffcontext.Context,
	sdkDefaultValue interface{},
) (model.RawVarResult, error) {
	res, err := getVariation(g, flagKey, ctx, sdkDefaultValue, "interface{}")
	notifyVariation(g, flagKey, ctx, res)
	return model.RawVarResult(res), err
}

// getFlagFromCache try to get the flag from the cache.
// It returns an error if the cache is not init or if the flag is not present or disabled.
func (g *GoFeatureFlag) getFlagFromCache(flagKey string) (flag.Flag, error) {
	f, err := g.retrieverManager.GetFlag(flagKey)
	if err != nil {
		return f, fmt.Errorf(errorFlagNotAvailable, flagKey)
	}
	return f, nil
}

// CollectEventData is collecting events and sending them to the data exporter to be stored.
func (g *GoFeatureFlag) CollectEventData(event exporter.FeatureEvent) {
	if g != nil && g.featureEventDataExporter != nil {
		// Add event in the exporter
		g.featureEventDataExporter.AddEvent(event)
	}
}

// CollectTrackingEventData is collecting tracking events and sending them to the data exporter to be stored.
func (g *GoFeatureFlag) CollectTrackingEventData(event exporter.TrackingEvent) {
	if g != nil && g.trackingEventDataExporter != nil {
		// Add event in the exporter
		g.trackingEventDataExporter.AddEvent(event)
	}
}

// notifyVariation is logging the evaluation result for a flag
// if no logger is provided in the configuration we are not logging anything.
func notifyVariation[T model.JSONType](
	g *GoFeatureFlag,
	flagKey string,
	ctx ffcontext.Context,
	result model.VariationResult[T],
) {
	if result.TrackEvents {
		event := exporter.NewFeatureEvent(
			ctx,
			flagKey,
			result.Value,
			result.VariationType,
			result.Failed,
			result.Version,
			"SERVER",
			ctx.ExtractGOFFProtectedFields().ExporterMetadata,
		)
		g.evalExporterWg.Add(1)
		go func() {
			defer g.evalExporterWg.Done()
			g.CollectEventData(event)
		}()
	}
}

// getVariation is the internal generic func that handle the logic of a variation the result will always
// contain a valid model.VariationResult
// nolint:funlen
func getVariation[T model.JSONType](
	g *GoFeatureFlag,
	flagKey string,
	evaluationCtx ffcontext.Context,
	sdkDefaultValue T,
	expectedType string,
) (model.VariationResult[T], error) {
	if g == nil {
		return model.VariationResult[T]{
			Value:         sdkDefaultValue,
			VariationType: flag.VariationSDKDefault,
			Failed:        true,
			Reason:        flag.ReasonError,
			ErrorCode:     flag.ErrorCodeProviderNotReady,
			Cacheable:     false,
		}, fmt.Errorf("go-feature-flag is not initialised, default value is used")
	}
	if g.config.Offline {
		return model.VariationResult[T]{
			Value:         sdkDefaultValue,
			VariationType: flag.VariationSDKDefault,
			Failed:        false,
			Reason:        flag.ReasonOffline,
			Cacheable:     false,
		}, nil
	}

	f, err := g.getFlagFromCache(flagKey)
	if err != nil {
		varResult := model.VariationResult[T]{
			Value:         sdkDefaultValue,
			VariationType: flag.VariationSDKDefault,
			ErrorCode:     flag.ErrorCodeFlagNotFound,
			Failed:        true,
			Reason:        flag.ReasonError,
			Cacheable:     false,
		}
		if f != nil {
			varResult.TrackEvents = f.IsTrackEvents()
			varResult.Version = f.GetVersion()
			varResult.Metadata = f.GetMetadata()
		}
		return varResult, err
	}

	flagCtx := flag.Context{
		DefaultSdkValue:             sdkDefaultValue,
		EvaluationContextEnrichment: maps.Clone(g.config.EvaluationContextEnrichment),
	}
	if g.config.Environment != "" {
		flagCtx.AddIntoEvaluationContextEnrichment("env", g.config.Environment)
	}
	return evaluation.Evaluate[T](f, flagKey, evaluationCtx, flagCtx, expectedType, sdkDefaultValue)
}
