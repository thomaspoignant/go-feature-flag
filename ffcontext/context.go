package ffcontext

// This file is a wrapper around the core context package to avoid any breaking changes
// when moving the logic to the new core package.

import coreCtx "github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"

type Context = coreCtx.Context
type EvaluationContext = coreCtx.EvaluationContext
type EvaluationContextBuilder = coreCtx.EvaluationContextBuilder
type GoffContextSpecifics = coreCtx.GoffContextSpecifics

// NewEvaluationContext creates a new evaluation context identified by the given targetingKey.
func NewEvaluationContext(key string) EvaluationContext {
	return coreCtx.NewEvaluationContext(key)
}

// Deprecated: NewAnonymousEvaluationContext is here for compatibility reason.
// Please use NewEvaluationContext instead and add a attributes attribute to know that it is an anonymous user.
//
// ctx := NewEvaluationContext("my-targetingKey")
// ctx.AddCustomAttribute("anonymous", true)
func NewAnonymousEvaluationContext(key string) EvaluationContext {
	return coreCtx.NewAnonymousEvaluationContext(key)
}

// NewEvaluationContextBuilder constructs a new EvaluationContextBuilder, specifying the user targetingKey.
//
// For authenticated users, the targetingKey may be a username or e-mail address. For anonymous users,
// this could be an IP address or session ID.
func NewEvaluationContextBuilder(key string) EvaluationContextBuilder {
	return coreCtx.NewEvaluationContextBuilder(key)
}
