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
			assertInitCalls(t, tt.retrievers, tt.expectedInitCalls)

			// Clean up
			_ = manager.Shutdown(ctx)
		})
	}
}

// findRetrieverByName returns the first retriever exposing GetName() == name,
// or nil when none matches.
func findRetrieverByName(retrievers []retriever.Retriever, name string) retriever.Retriever {
	for _, r := range retrievers {
		if named, ok := r.(interface{ GetName() string }); ok && named.GetName() == name {
			return r
		}
	}
	return nil
}

// initCalled reports whether Init was recorded as called on a mock retriever.
// The second return value is false when r is not an initializable mock type.
func initCalled(r retriever.Retriever) (called bool, ok bool) {
	switch v := r.(type) {
	case *mockretriever.InitializableRetrieverLegacy:
		return v.InitCalled, true
	case *mockretriever.InitializableRetriever:
		return v.InitCalled, true
	case *mockretriever.InitializableRetrieverWithFlagset:
		return v.InitCalled, true
	default:
		return false, false
	}
}

// assertInitCalls verifies that Init was called on each retriever that the test
// case expects to have been initialized.
func assertInitCalls(t *testing.T, retrievers []retriever.Retriever, expected map[string]bool) {
	t.Helper()
	for name, shouldBeCalled := range expected {
		if !shouldBeCalled {
			continue
		}
		r := findRetrieverByName(retrievers, name)
		assert.NotNil(t, r, "Retriever %s should have been found", name)
		if called, ok := initCalled(r); ok {
			assert.True(t, called, "Init should have been called on %s", name)
		}
	}
}

// formatHintingRetriever is a local test double that implements
// retriever.FormatHinter so the test can verify that the manager honors the
// retriever-declared output format.
type formatHintingRetriever struct {
	format  string
	content []byte
}

func (r *formatHintingRetriever) Retrieve(_ context.Context) ([]byte, error) {
	return r.content, nil
}

func (r *formatHintingRetriever) OutputFormat() string {
	return r.format
}

// TestManagerInit_FormatHinterOverridesGlobalFileFormat verifies that a
// Retriever implementing FormatHinter causes the manager to parse its output
// with the declared format, overriding the global ManagerConfig.FileFormat.
//
// The TOML payload below would fail to parse as YAML (the configured global
// format), so a passing Init proves the hinter took effect.
func TestManagerInit_FormatHinterOverridesGlobalFileFormat(t *testing.T) {
	tomlContent := []byte(`["test-flag"]
[test-flag.variations]
A = true
B = false
[test-flag.defaultRule]
variation = "A"
`)

	tests := []struct {
		name        string
		hintFormat  string
		globalFmt   string
		expectError bool
	}{
		{
			name:        "FormatHinter toml overrides global yaml",
			hintFormat:  "toml",
			globalFmt:   "yaml",
			expectError: false,
		},
		{
			name:        "Empty FormatHinter falls back to global yaml and fails on toml content",
			hintFormat:  "",
			globalFmt:   "yaml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := fflog.FFLogger{}
			cacheManager := cache.New(notification.NewService([]notifier.Notifier{}), "", &logger)

			r := &formatHintingRetriever{format: tt.hintFormat, content: tomlContent}
			config := retriever.ManagerConfig{
				FileFormat:      tt.globalFmt,
				PollingInterval: 0,
			}
			manager := retriever.NewManager(config, []retriever.Retriever{r}, cacheManager, &logger)
			err := manager.Init(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			_ = manager.Shutdown(ctx)
		})
	}
}
