package retriever_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/notification"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mockretriever"
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
				func() *mockretriever.InitializableRetrieverLegacy {
					r := mockretriever.NewInitializableRetrieverLegacy("legacy")
					r.InitShouldFail = true
					return r
				}(),
				mockretriever.NewInitializableRetriever("standard"),
				mockretriever.NewInitializableRetrieverWithFlagset("flagset"),
			},
			expectError: true,
			expectedInitCalls: map[string]bool{
				"legacy":   true,
				"standard": true,
				"flagset":  true,
			},
		},
		{
			name: "All retrievers fail initialization",
			retrievers: []retriever.Retriever{
				func() *mockretriever.InitializableRetrieverLegacy {
					r := mockretriever.NewInitializableRetrieverLegacy("legacy")
					r.InitShouldFail = true
					return r
				}(),
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
				"legacy":   true,
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
				Ctx:                     ctx,
				FileFormat:              "yaml",
				StartWithRetrieverError: tt.startWithRetrieverError,
				PollingInterval:         0, // Disable polling for tests
			}

			manager := retriever.NewManager(config, tt.retrievers, cacheManager, &logger)

			// Act
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

// TestManagerInit_InterfaceSpecificBehavior tests specific behaviors for each interface type
func TestManagerInit_InterfaceSpecificBehavior(t *testing.T) {
	ctx := context.Background()
	logger := fflog.FFLogger{LeveledLogger: slog.Default()}
	cacheManager := cache.New(notification.NewService([]notifier.Notifier{}), "", &logger)

	t.Run("InitializableRetrieverLegacy receives log.Logger", func(t *testing.T) {
		mockRetriever := mockretriever.NewInitializableRetrieverLegacy("legacy")
		config := retriever.ManagerConfig{
			Ctx:        ctx,
			FileFormat: "yaml",
		}
		manager := retriever.NewManager(config, []retriever.Retriever{mockRetriever}, cacheManager, &logger)

		err := manager.Init(ctx)
		require.NoError(t, err)

		assert.True(t, mockRetriever.InitCalled)
		assert.NotNil(t, mockRetriever.ReceivedLogger)

		_ = manager.Shutdown(ctx)
	})

	t.Run("InitializableRetriever receives fflog.FFLogger", func(t *testing.T) {
		mockRetriever := mockretriever.NewInitializableRetriever("standard")
		config := retriever.ManagerConfig{
			Ctx:        ctx,
			FileFormat: "yaml",
		}
		manager := retriever.NewManager(config, []retriever.Retriever{mockRetriever}, cacheManager, &logger)

		err := manager.Init(ctx)
		require.NoError(t, err)

		assert.True(t, mockRetriever.InitCalled)
		assert.NotNil(t, mockRetriever.ReceivedLogger)

		_ = manager.Shutdown(ctx)
	})

	t.Run("InitializableRetrieverWithFlagset receives flagset parameter", func(t *testing.T) {
		mockRetriever := mockretriever.NewInitializableRetrieverWithFlagset("flagset")
		flagsetName := "test-flagset"
		config := retriever.ManagerConfig{
			Ctx:        ctx,
			FileFormat: "yaml",
			Name:       &flagsetName,
		}
		manager := retriever.NewManager(config, []retriever.Retriever{mockRetriever}, cacheManager, &logger)

		err := manager.Init(ctx)
		require.NoError(t, err)

		assert.True(t, mockRetriever.InitCalled)
		assert.NotNil(t, mockRetriever.ReceivedLogger)
		assert.NotNil(t, mockRetriever.ReceivedFlagset)
		assert.Equal(t, flagsetName, *mockRetriever.ReceivedFlagset)

		_ = manager.Shutdown(ctx)
	})
}

// TestManagerInit_PollingBehavior tests that polling is started correctly
func TestManagerInit_PollingBehavior(t *testing.T) {
	ctx := context.Background()
	logger := fflog.FFLogger{LeveledLogger: slog.Default()}
	cacheManager := cache.New(notification.NewService([]notifier.Notifier{}), "", &logger)

	t.Run("Polling enabled with valid interval", func(t *testing.T) {
		mockRetriever := mockretriever.NewSimpleRetriever("simple")
		config := retriever.ManagerConfig{
			Ctx:             ctx,
			FileFormat:      "yaml",
			PollingInterval: 100 * time.Millisecond,
		}
		manager := retriever.NewManager(config, []retriever.Retriever{mockRetriever}, cacheManager, &logger)

		err := manager.Init(ctx)
		require.NoError(t, err)

		// Give some time for potential polling
		time.Sleep(50 * time.Millisecond)

		// Clean up
		manager.StopPolling()
		_ = manager.Shutdown(ctx)
	})

	t.Run("Polling disabled with zero interval", func(t *testing.T) {
		mockRetriever := mockretriever.NewSimpleRetriever("simple")
		config := retriever.ManagerConfig{
			Ctx:             ctx,
			FileFormat:      "yaml",
			PollingInterval: 0,
		}
		manager := retriever.NewManager(config, []retriever.Retriever{mockRetriever}, cacheManager, &logger)

		err := manager.Init(ctx)
		require.NoError(t, err)

		// No polling should be started
		_ = manager.Shutdown(ctx)
	})
}
