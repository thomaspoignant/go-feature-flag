package ffcontext_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

func TestNewEvaluationContextWithoutTargetingKey(t *testing.T) {
	ctx := ffcontext.NewEvaluationContextWithoutTargetingKey()

	assert.Equal(t, "", ctx.GetKey(), "Should have empty targeting key")
	assert.Equal(t, map[string]interface{}{}, ctx.GetCustom(), "Should have empty custom attributes")
	assert.False(t, ctx.IsAnonymous(), "Should not be anonymous by default")
}

func TestNewEvaluationContextBuilderWithoutTargetingKey(t *testing.T) {
	ctx := ffcontext.NewEvaluationContextBuilderWithoutTargetingKey().
		AddCustom("role", "admin").
		AddCustom("anonymous", true).
		Build()

	assert.Equal(t, "", ctx.GetKey(), "Should have empty targeting key")
	assert.Equal(t, "admin", ctx.GetCustom()["role"], "Should have custom attributes")
	assert.True(t, ctx.IsAnonymous(), "Should be anonymous when set")
}

func TestEvaluationContextWithoutTargetingKey_Functionality(t *testing.T) {
	// Test that context without targeting key works like a normal context
	ctx := ffcontext.NewEvaluationContextWithoutTargetingKey()

	// Add custom attributes
	ctx.AddCustomAttribute("feature", "enabled")
	ctx.AddCustomAttribute("team", "engineering")

	assert.Equal(t, "enabled", ctx.GetCustom()["feature"])
	assert.Equal(t, "engineering", ctx.GetCustom()["team"])

	// Test ToMap functionality
	contextMap := ctx.ToMap()
	assert.Equal(t, "", contextMap["targetingKey"])
	assert.Equal(t, "enabled", contextMap["feature"])
	assert.Equal(t, "engineering", contextMap["team"])
}

func TestEvaluationContextComparison(t *testing.T) {
	// Test that contexts created with and without targeting keys work similarly
	ctxWithKey := ffcontext.NewEvaluationContext("user-123")
	ctxWithKey.AddCustomAttribute("role", "admin")

	ctxWithoutKey := ffcontext.NewEvaluationContextWithoutTargetingKey()
	ctxWithoutKey.AddCustomAttribute("role", "admin")

	// Both should have the same custom attributes
	assert.Equal(t, ctxWithKey.GetCustom()["role"], ctxWithoutKey.GetCustom()["role"])

	// But different targeting keys
	assert.Equal(t, "user-123", ctxWithKey.GetKey())
	assert.Equal(t, "", ctxWithoutKey.GetKey())
}
