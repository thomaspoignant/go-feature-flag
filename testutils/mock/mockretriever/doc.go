// Package mockretriever provides mock implementations of various retriever interfaces
// for testing purposes in the go-feature-flag project.
//
// This package contains mock retrievers that implement different retriever interfaces:
//   - SimpleRetriever: Basic Retriever interface only (only retrieve)
//   - InitializableRetriever: Standard interface with *fflog.FFLogger (init and shutdown)
//   - InitializableRetrieverWithFlagset: Flagset interface with flagset parameter (init and shutdown)
//   - FileInitializableRetriever: File-based mock for compatibility (init and shutdown)
//   - ContextAwareRetriever: Respects context cancellation
//   - StatusChangingRetriever: Allows testing status changes
//   - RecoverableRetriever: Can fail initially but succeed on retry
//
// Each mock provides configurable behavior for testing different scenarios
// such as initialization failures, retrieval errors, and status changes.
package mockretriever

// Default flag configuration used by all mock retrievers
const defaultFlagConfig = `{
    "test-flag": {
        "variations": {
            "true": true,
            "false": false
        },
        "defaultRule": {
            "variation": "false"
        }
    }
}`
