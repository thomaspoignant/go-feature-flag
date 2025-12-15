package retriever_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/notification"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock/mockretriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// TestManagerInit_AllRetrieverTypes tests the Init function with all possible retriever types
// and interface combinations to ensure comprehensive coverage.
func TestManagerInit_AllRetrieverTypes(t *testing.T) {
	tests := []struct {
		name                    string
		retrievers              []retriever.Retriever
		expectError             bool
		expectedInitCalls       map[string]bool
		expectedStatusCalls     map[string]bool
		expectedShutdownCalls   map[string]bool
		startWithRetrieverError bool
	}{
		{
			name: "All retriever interface types - successful initialization",
			retrievers: []retriever.Retriever{
				mockretriever.NewSimpleRetriever("simple"),
				mockretriever.NewInitializableRetrieverLegacy("legacy"),
				mockretriever.NewInitializableRetriever("standard"),
				mockretriever.NewInitializableRetrieverWithFlagset("flagset"),
			},
			expectError: false,
			expectedInitCalls: map[string]bool{
				"legacy":   true,
				"standard": true,
				"flagset":  true,
			},
			expectedStatusCalls: map[string]bool{
				"legacy":   true,
				"standard": true,
				"flagset":  true,
			},
		},
		{
			name: "Mixed retriever types with one initialization failure",
			retrievers: []retriever.Retriever{
				mockretriever.NewSimpleRetriever("simple"),
				mockretriever.NewInitializableRetrieverLegacy("legacy"),
				mockretriever.NewInitializableRetriever("standard"),
				mockretriever.NewInitializableRetrieverWithFlagset("flagset"),
			},
			expectError: false,
			expectedInitCalls: map[string]bool{
				"legacy":   true,
				"standard": true,
				"flagset":  true,
			},
		},
		{
			name: "All retrievers fail initialization",
			retrievers: []retriever.Retriever{
				func() *mockretriever.InitializableRetriever {
					r := mockretriever.NewInitializableRetriever("standard")
					r.InitShouldFail = true
					return r
				}(),
				func() *mockretriever.InitializableRetrieverWithFlagset {
					r := mockretriever.NewInitializableRetrieverWithFlagset("flagset")
					r.InitShouldFail = true
					return r
				}(),
			},
			expectError: true,
			expectedInitCalls: map[string]bool{
				"standard": true,
				"flagset":  true,
			},
		},
		{
			name: "Only simple retrievers (no initialization needed)",
			retrievers: []retriever.Retriever{
				mockretriever.NewSimpleRetriever("simple1"),
				mockretriever.NewSimpleRetriever("simple2"),
				mockretriever.NewSimpleRetriever("simple3"),
			},
			expectError: false,
		},
		{
			name:        "Empty retrievers slice",
			retrievers:  []retriever.Retriever{},
			expectError: false,
		},
		{
			name: "Mixed success and failure with StartWithRetrieverError enabled",
			retrievers: []retriever.Retriever{
				func() *mockretriever.InitializableRetriever {
					r := mockretriever.NewInitializableRetriever("standard")
					r.InitShouldFail = true
					return r
				}(),
				mockretriever.NewInitializableRetrieverWithFlagset("flagset"),
			},
			startWithRetrieverError: true,
			expectError:             true, // Still errors during initialization phase, StartWithRetrieverError only applies to retrieval phase
			expectedInitCalls: map[string]bool{
				"standard": true,
				"flagset":  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := fflog.FFLogger{}
			cacheManager := cache.New(notification.NewService([]notifier.Notifier{}), "", &logger)

			config := retriever.ManagerConfig{
				FileFormat:              "yaml",
				StartWithRetrieverError: tt.startWithRetrieverError,
				PollingInterval:         0, // Disable polling for tests
			}

			manager := retriever.NewManager(config, tt.retrievers, cacheManager, &logger)
			err := manager.Init(ctx)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify that Init was called on the expected retrievers
			for retrieverName, shouldBeCalled := range tt.expectedInitCalls {
				retrieverFound := false
				for _, r := range tt.retrievers {
					if mockRetriever, ok := r.(interface{ GetName() string }); ok {
						if mockRetriever.GetName() == retrieverName {
							retrieverFound = true
							if legacyRetriever, ok := r.(*mockretriever.InitializableRetrieverLegacy); ok {
								if shouldBeCalled {
									assert.True(t, legacyRetriever.InitCalled, "Init should have been called on %s", retrieverName)
								}
							}
							if standardRetriever, ok := r.(*mockretriever.InitializableRetriever); ok {
								if shouldBeCalled {
									assert.True(t, standardRetriever.InitCalled, "Init should have been called on %s", retrieverName)
								}
							}
							if flagsetRetriever, ok := r.(*mockretriever.InitializableRetrieverWithFlagset); ok {
								if shouldBeCalled {
									assert.True(t, flagsetRetriever.InitCalled, "Init should have been called on %s", retrieverName)
								}
							}
							break
						}
					}
				}
				if shouldBeCalled {
					assert.True(t, retrieverFound, "Retriever %s should have been found", retrieverName)
				}
			}

			// Clean up
			_ = manager.Shutdown(ctx)
		})
	}
}
