package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

// TestEmptyEvaluationContext demonstrates the new functionality for issue #2533
// https://github.com/thomaspoignant/go-feature-flag/issues/2533
func TestEmptyEvaluationContext(t *testing.T) {
	// Create a configuration file with different types of flags
	configContent := `static-flag:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: disabled

percentage-flag:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    percentage:
      enabled: 30
      disabled: 70

targeted-flag:
  variations:
    enabled: true
    disabled: false
  targeting:
    - query: role eq "admin"
      variation: enabled
  defaultRule:
    variation: disabled
`

	// Create a temporary config file
	configFile, err := ioutil.TempFile("", "goff-test-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(configFile.Name())

	_, err = configFile.WriteString(configContent)
	assert.NoError(t, err)
	configFile.Close()

	// Initialize the client with test configuration
	err = ffclient.Init(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Retriever: &fileretriever.Retriever{
			Path: configFile.Name(),
		},
	})
	assert.NoError(t, err)
	defer ffclient.Close()

	t.Run("Static flag should work with empty context", func(t *testing.T) {
		ctx := ffcontext.NewEvaluationContextWithoutTargetingKey()
		
		result, err := ffclient.BoolVariationDetails("static-flag", ctx, true)
		
		assert.NoError(t, err)
		assert.False(t, result.Value, "Should return the static variation")
		assert.Equal(t, "STATIC", result.Reason)
		assert.Empty(t, result.ErrorCode)
	})

	t.Run("Percentage flag should fail with empty context", func(t *testing.T) {
		ctx := ffcontext.NewEvaluationContextWithoutTargetingKey()
		
		result, err := ffclient.BoolVariationDetails("percentage-flag", ctx, true)
		
		// Should return default value due to error
		assert.Error(t, err)
		assert.True(t, result.Value, "Should return SDK default value")
		assert.Equal(t, "ERROR", result.Reason)
		assert.Equal(t, "TARGETING_KEY_MISSING", result.ErrorCode)
	})

	t.Run("Percentage flag should work with targeting key", func(t *testing.T) {
		ctx := ffcontext.NewEvaluationContext("user-123")
		
		result, err := ffclient.BoolVariationDetails("percentage-flag", ctx, true)
		
		assert.NoError(t, err)
		assert.Equal(t, "SPLIT", result.Reason)
		assert.Empty(t, result.ErrorCode)
		// Value will depend on hash, but should be either true or false
		assert.IsType(t, false, result.Value)
	})

	t.Run("Targeted flag should work with empty context and custom attributes", func(t *testing.T) {
		ctx := ffcontext.NewEvaluationContextWithoutTargetingKey()
		ctx.AddCustomAttribute("role", "admin")
		
		result, err := ffclient.BoolVariationDetails("targeted-flag", ctx, false)
		
		assert.NoError(t, err)
		assert.True(t, result.Value, "Should match targeting rule")
		assert.Equal(t, "TARGETING_MATCH", result.Reason)
		assert.Empty(t, result.ErrorCode)
	})

	t.Run("Builder pattern should work for empty context", func(t *testing.T) {
		ctx := ffcontext.NewEvaluationContextBuilderWithoutTargetingKey().
			AddCustom("environment", "test").
			AddCustom("version", "1.0.0").
			Build()
		
		result, err := ffclient.BoolVariationDetails("static-flag", ctx, true)
		
		assert.NoError(t, err)
		assert.False(t, result.Value)
		assert.Equal(t, "STATIC", result.Reason)
		assert.Empty(t, result.ErrorCode)
		
		// Verify custom attributes are preserved
		assert.Equal(t, "test", ctx.GetCustom()["environment"])
		assert.Equal(t, "1.0.0", ctx.GetCustom()["version"])
	})
}

// TestBackwardsCompatibility ensures existing functionality still works
func TestBackwardsCompatibility(t *testing.T) {
	configContent := `test-flag:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    percentage:
      enabled: 50
      disabled: 50
`

	// Create a temporary config file
	configFile, err := ioutil.TempFile("", "goff-test-compat-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(configFile.Name())

	_, err = configFile.WriteString(configContent)
	assert.NoError(t, err)
	configFile.Close()

	err = ffclient.Init(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Retriever: &fileretriever.Retriever{
			Path: configFile.Name(),
		},
	})
	assert.NoError(t, err)
	defer ffclient.Close()

	t.Run("Existing usage with targeting key should still work", func(t *testing.T) {
		ctx := ffcontext.NewEvaluationContext("user-456")
		
		result, err := ffclient.BoolVariationDetails("test-flag", ctx, false)
		
		assert.NoError(t, err)
		assert.Equal(t, "SPLIT", result.Reason)
		assert.Empty(t, result.ErrorCode)
		assert.IsType(t, false, result.Value)
	})
}