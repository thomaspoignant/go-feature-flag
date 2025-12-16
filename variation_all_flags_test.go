package ffclient_test

import (
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

// compareJSONWithTimestampHandling compares expected and actual JSON while handling timestamp fields properly
// to avoid flaky tests due to timing differences
func compareJSONWithTimestampHandling(t *testing.T, expectedFilePath string, actualJSON []byte) {
	t.Helper()

	expected, err := os.ReadFile(expectedFilePath)
	assert.NoError(t, err, "Failed to read expected JSON file")

	var expectedFlags map[string]any
	err = json.Unmarshal(expected, &expectedFlags)
	assert.NoError(t, err, "Failed to unmarshal expected JSON")

	var actualFlags map[string]any
	err = json.Unmarshal(actualJSON, &actualFlags)
	assert.NoError(t, err, "Failed to unmarshal actual JSON")

	// Handle cases where there might not be a "flags" field (e.g., offline, module_not_init)
	expectedFlagData, hasExpectedFlags := expectedFlags["flags"].(map[string]any)
	actualFlagData, hasActualFlags := actualFlags["flags"].(map[string]any)

	// Only proceed with timestamp validation if both have flags
	if hasExpectedFlags && hasActualFlags {
		// Compare structure without timestamps first
		for flagName, expectedFlag := range expectedFlagData {
			actualFlag, exists := actualFlagData[flagName]
			require.True(t, exists, "Flag %s should exist in actual results", flagName)

			expectedFlagObj, ok := expectedFlag.(map[string]any)
			require.True(t, ok, "expected flag %s should be a map", flagName)

			actualFlagObj, ok := actualFlag.(map[string]any)
			require.True(t, ok, "actual flag %s should be a map", flagName)

			// Verify timestamp exists and is reasonable
			require.Contains(t, actualFlagObj, "timestamp", "timestamp should exist in flag %s", flagName)
			actualTimestamp, ok := actualFlagObj["timestamp"].(float64)
			require.True(t, ok, "timestamp should be a number for flag %s", flagName)
			require.NotEqual(t, 0, actualTimestamp, "timestamp should not be zero for flag %s", flagName)

			// Timestamp should be recent (within a small delta of the current time)
			assert.InDelta(t, float64(time.Now().Unix()), actualTimestamp, 5.0, "timestamp should be recent for flag %s", flagName)

			// Normalize timestamps to compare structure
			expectedFlagObj["timestamp"] = actualFlagObj["timestamp"]
		}
	}

	// Now compare the full JSON structures
	expectedJSON, err := json.Marshal(expectedFlags)
	assert.NoError(t, err, "Failed to marshal normalized expected JSON")
	assert.JSONEq(t, string(expectedJSON), string(actualJSON), "JSON structures should match after timestamp normalization")
}

func TestAllFlagsState(t *testing.T) {
	tests := []struct {
		name       string
		config     ffclient.Config
		valid      bool
		jsonOutput string
		initModule bool
	}{
		{
			name: "Valid multiple types",
			config: ffclient.Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
				LeveledLogger: slog.Default(),
			},
			valid:      true,
			jsonOutput: "./testdata/ffclient/all_flags/marshal_json/valid_multiple_types.json",
			initModule: true,
		},
		{
			name: "module not init",
			config: ffclient.Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
			},
			valid:      false,
			jsonOutput: "./testdata/ffclient/all_flags/marshal_json/module_not_init.json",
			initModule: false,
		},
		{
			name: "offline",
			config: ffclient.Config{
				Offline: true,
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
			},
			valid:      true,
			jsonOutput: "./testdata/ffclient/all_flags/marshal_json/offline.json",
			initModule: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exportDir, _ := os.MkdirTemp("", "export")
			tt.config.DataExporter = ffclient.DataExporter{
				FlushInterval:    1000,
				MaxEventInMemory: 1,
				Exporter:         &fileexporter.Exporter{OutputDir: exportDir},
			}

			var goff *ffclient.GoFeatureFlag
			var err error
			if tt.initModule {
				goff, err = ffclient.New(tt.config)
				assert.NoError(t, err)
				defer goff.Close()
			} else {
				// we close directly so we can test with module not init
				goff, _ = ffclient.New(tt.config)
				goff.Close()
			}

			user := ffcontext.NewEvaluationContext("random-key")
			allFlagsState := goff.AllFlagsState(user)
			assert.Equal(t, tt.valid, allFlagsState.IsValid())

			// Compare JSON output with proper timestamp handling
			marshaled, err := allFlagsState.MarshalJSON()
			assert.NoError(t, err)
			compareJSONWithTimestampHandling(t, tt.jsonOutput, marshaled)

			// no data exported
			files, _ := os.ReadDir(exportDir)
			assert.Equal(t, 0, len(files))
		})
	}
}

func TestGetFlagStates(t *testing.T) {
	tests := []struct {
		name              string
		config            ffclient.Config
		valid             bool
		jsonOutput        string
		initModule        bool
		evaluationContext ffcontext.EvaluationContext
	}{
		{
			name: "Valid multiple flags",
			config: ffclient.Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/get_flagstates/config_flag/flag-config-all-flags.yaml",
				},
			},
			valid:      true,
			jsonOutput: "./testdata/ffclient/get_flagstates/marshal_json/valid_flag1_flag4.json",
			initModule: true,
			evaluationContext: ffcontext.NewEvaluationContextBuilder("123").
				AddCustom("gofeatureflag", map[string]any{
					"flagList": []string{"test-flag1", "test-flag4"},
				}).
				Build(),
		},
		{
			name: "empty list of flags in context",
			config: ffclient.Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/get_flagstates/config_flag/flag-config-all-flags.yaml",
				},
			},
			valid:      true,
			jsonOutput: "./testdata/ffclient/get_flagstates/marshal_json/all_flags.json",
			initModule: true,
			evaluationContext: ffcontext.NewEvaluationContextBuilder("123").
				AddCustom("gofeatureflag", map[string]any{
					"flagList": []string{},
				}).
				Build(),
		},
		{
			name: "no field in context context",
			config: ffclient.Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/get_flagstates/config_flag/flag-config-all-flags.yaml",
				},
			},
			valid:             true,
			jsonOutput:        "./testdata/ffclient/get_flagstates/marshal_json/all_flags.json",
			initModule:        true,
			evaluationContext: ffcontext.NewEvaluationContextBuilder("123").Build(),
		},
		{
			name: "offline",
			config: ffclient.Config{
				Offline: true,
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
			},
			valid:      true,
			jsonOutput: "./testdata/ffclient/all_flags/marshal_json/offline.json",
			initModule: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init logger
			exportDir, _ := os.MkdirTemp("", "export")
			tt.config.DataExporter = ffclient.DataExporter{
				FlushInterval:    1000,
				MaxEventInMemory: 1,
				Exporter:         &fileexporter.Exporter{OutputDir: exportDir},
			}

			var goff *ffclient.GoFeatureFlag
			var err error
			if tt.initModule {
				goff, err = ffclient.New(tt.config)
				assert.NoError(t, err)
				defer goff.Close()
			} else {
				// we close directly so we can test with module not init
				goff, _ = ffclient.New(tt.config)
				goff.Close()
			}

			allFlagsState := goff.GetFlagStates(
				tt.evaluationContext,
				tt.evaluationContext.ExtractGOFFProtectedFields().FlagList,
			)
			assert.Equal(t, tt.valid, allFlagsState.IsValid())

			// Compare JSON output with proper timestamp handling
			marshaled, err := allFlagsState.MarshalJSON()
			assert.NoError(t, err)
			compareJSONWithTimestampHandling(t, tt.jsonOutput, marshaled)

			// no data exported
			files, _ := os.ReadDir(exportDir)
			assert.Equal(t, 0, len(files))
		})
	}
}

func TestAllFlagsFromCache(t *testing.T) {
	tests := []struct {
		name       string
		config     ffclient.Config
		initModule bool
		numberFlag int
		err        assert.ErrorAssertionFunc
	}{
		{
			name: "Valid multiple types",
			config: ffclient.Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
			},
			initModule: true,
			numberFlag: 7,
			err:        assert.NoError,
		},
		{
			name: "module not init",
			config: ffclient.Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
			},
			initModule: false,
			err:        assert.NoError,
		},
		{
			name: "offline",
			config: ffclient.Config{
				Offline: true,
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
			},
			initModule: true,
			err:        assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var goff *ffclient.GoFeatureFlag
			var err error
			if tt.initModule {
				goff, err = ffclient.New(tt.config)
				assert.NoError(t, err)
				defer goff.Close()

				flags, err := goff.GetFlagsFromCache()
				tt.err(t, err)

				if err != nil {
					assert.Equal(t, tt.numberFlag, len(flags))
				}
			} else {
				// we close directly so we can test with module not init
				goff, _ = ffclient.New(tt.config)
				goff.Close()

				_, err := goff.GetFlagsFromCache()
				assert.Error(t, err)
			}
		})
	}
}
