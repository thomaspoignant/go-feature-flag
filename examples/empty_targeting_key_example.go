// Package main demonstrates how to use GO Feature Flag with empty evaluation contexts
// for flags that don't require bucketing (no percentage-based rules).
//
// This example shows the new functionality introduced in issue #2533:
// https://github.com/thomaspoignant/go-feature-flag/issues/2533
package main

import (
	"fmt"
	"log"
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func main() {
	// Create a flag configuration that demonstrates both cases
	// This would normally be in a separate YAML file
	flagConfig := `my-static-feature:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: disabled

my-percentage-feature:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    percentage:
      enabled: 20
      disabled: 80

my-targeted-feature:
  variations:
    enabled: true
    disabled: false
  targeting:
    - query: role eq "admin"
      variation: enabled
  defaultRule:
    variation: disabled`

	// Write the config to a temporary file
	configFile := "/tmp/goff-empty-context-example.yaml"
	if err := writeStringToFile(configFile, flagConfig); err != nil {
		log.Fatalf("Failed to write config file: %v", err)
	}

	// Initialize GO Feature Flag
	err := ffclient.Init(ffclient.Config{
		PollingInterval: 3 * time.Second,
		Retriever: &fileretriever.Retriever{
			Path: configFile,
		},
	})
	if err != nil {
		log.Fatalf("Failed to initialize GO Feature Flag: %v", err)
	}
	defer ffclient.Close()

	fmt.Println("=== Empty Targeting Key Examples ===\n")

	// Example 1: Static flag (should work with empty context)
	fmt.Println("1. Static flag with empty evaluation context:")
	emptyContext := ffcontext.NewEvaluationContextWithoutTargetingKey()

	staticResult, err := ffclient.BoolVariationDetails("my-static-feature", emptyContext, false)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Value: %v, Reason: %s, ErrorCode: %s\n",
			staticResult.Value, staticResult.Reason, staticResult.ErrorCode)
	}

	// Example 2: Percentage-based flag (should fail with empty context)
	fmt.Println("\n2. Percentage-based flag with empty evaluation context:")
	percentageResult, err := ffclient.BoolVariationDetails("my-percentage-feature", emptyContext, false)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	}
	fmt.Printf("   Value: %v, Reason: %s, ErrorCode: %s\n",
		percentageResult.Value, percentageResult.Reason, percentageResult.ErrorCode)

	// Example 3: Same percentage-based flag with proper targeting key (should work)
	fmt.Println("\n3. Percentage-based flag with targeting key:")
	contextWithKey := ffcontext.NewEvaluationContext("user-123")

	percentageWithKeyResult, err := ffclient.BoolVariationDetails("my-percentage-feature", contextWithKey, false)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Value: %v, Reason: %s, ErrorCode: %s\n",
			percentageWithKeyResult.Value, percentageWithKeyResult.Reason, percentageWithKeyResult.ErrorCode)
	}

	// Example 4: Targeted flag with empty context and attributes (should work)
	fmt.Println("\n4. Targeted flag (non-percentage) with empty targeting key but custom attributes:")
	emptyContextWithRole := ffcontext.NewEvaluationContextWithoutTargetingKey()
	emptyContextWithRole.AddCustomAttribute("role", "admin")

	targetedResult, err := ffclient.BoolVariationDetails("my-targeted-feature", emptyContextWithRole, false)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Value: %v, Reason: %s, ErrorCode: %s\n",
			targetedResult.Value, targetedResult.Reason, targetedResult.ErrorCode)
	}

	// Example 5: Builder pattern for empty context
	fmt.Println("\n5. Using builder pattern for empty context:")
	builderContext := ffcontext.NewEvaluationContextBuilderWithoutTargetingKey().
		AddCustom("environment", "development").
		AddCustom("version", "v1.2.3").
		Build()

	builderResult, err := ffclient.BoolVariationDetails("my-static-feature", builderContext, false)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Value: %v, Reason: %s, ErrorCode: %s\n",
			builderResult.Value, builderResult.Reason, builderResult.ErrorCode)
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("✅ Flags without bucketing requirements (static variations, targeting without percentages) work with empty targeting keys")
	fmt.Println("❌ Flags with bucketing requirements (percentage-based rules, progressive rollouts) require targeting keys")
	fmt.Println("✅ Custom attributes still work for targeting rules even without targeting keys")
}

func writeStringToFile(filename, content string) error {
	file, err := createFile(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// Mock function to simulate file creation (would use os.Create in real code)
func createFile(filename string) (fileInterface, error) {
	// In a real implementation, this would use os.Create
	// For this example, we'll simulate it
	return &mockFile{content: ""}, nil
}

type fileInterface interface {
	WriteString(s string) (int, error)
	Close() error
}

type mockFile struct {
	content string
}

func (m *mockFile) WriteString(s string) (int, error) {
	m.content = s
	return len(s), nil
}

func (m *mockFile) Close() error {
	return nil
}