package ffclient

import (
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/flagstate"
)

// AllFlagsState return the values of all the flags for a specific user.
// If a valid field is false, it means that we had an error when checking the flags.
func AllFlagsState(ctx ffcontext.Context) flagstate.AllFlags {
	return ff.AllFlagsState(ctx)
}

// GetFlagsFromCache returns all the flags present in the cache with their
// current state when calling this method. If cache hasn't been initialized, an
// error reporting this is returned.
func GetFlagsFromCache() (map[string]flag.Flag, error) {
	return ff.GetFlagsFromCache()
}

// GetFlagStates is evaluating all the flags in flagsToEvaluate based on the context provided.
// If flagsToEvaluate is nil or empty, it will evaluate all the flags available in GO Feature Flag.
func (g *GoFeatureFlag) GetFlagStates(evaluationCtx ffcontext.Context, flagsToEvaluate []string) flagstate.AllFlags {
	if g == nil {
		return flagstate.AllFlags{}
	}
	if g.config.Offline {
		return flagstate.NewAllFlags()
	}

	// prepare evaluation context enrichment
	flagCtx := flag.Context{
		EvaluationContextEnrichment: g.config.EvaluationContextEnrichment,
		DefaultSdkValue:             nil,
	}
	flagCtx.AddIntoEvaluationContextEnrichment("env", g.config.Environment)

	// Evaluate only the flags in flagsToEvaluate
	if len(flagsToEvaluate) != 0 {
		flagStates := flagstate.NewAllFlags()
		for _, key := range flagsToEvaluate {
			currentFlag, err := g.cache.GetFlag(key)
			if err != nil {
				// We ignore flags in error
				continue
			}
			flagStates.AddFlag(key, flagstate.FromFlagEvaluation(key, evaluationCtx, flagCtx, currentFlag))
		}
		return flagStates
	}

	// Evaluate all the flags
	flags, err := g.GetFlagsFromCache()
	if err != nil {
		return flagstate.AllFlags{}
	}
	allFlags := flagstate.NewAllFlags()
	for key, currentFlag := range flags {
		allFlags.AddFlag(key, flagstate.FromFlagEvaluation(key, evaluationCtx, flagCtx, currentFlag))
	}
	return allFlags
}

// AllFlagsState return a flagstate.AllFlags that contains all the flags for a specific user.
func (g *GoFeatureFlag) AllFlagsState(evaluationCtx ffcontext.Context) flagstate.AllFlags {
	if g == nil {
		return flagstate.AllFlags{}
	}
	return g.GetFlagStates(evaluationCtx, []string{})
}

// GetFlagsFromCache returns all the flags present in the cache with their
// current state when calling this method. If cache hasn't been initialized, an
// error reporting this is returned.
func (g *GoFeatureFlag) GetFlagsFromCache() (map[string]flag.Flag, error) {
	return g.cache.AllFlags()
}
