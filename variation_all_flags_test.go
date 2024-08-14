package ffclient

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"os"
	"testing"
	"time"
)

func TestAllFlagsState(t *testing.T) {
	tests := []struct {
		name       string
		config     Config
		valid      bool
		jsonOutput string
		initModule bool
	}{
		{
			name: "Valid multiple types",
			config: Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
			},
			valid:      true,
			jsonOutput: "./testdata/ffclient/all_flags/marshal_json/valid_multiple_types.json",
			initModule: true,
		},
		{
			name: "Error in flag-0",
			config: Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-with-error.yaml",
				},
			},
			valid:      false,
			jsonOutput: "./testdata/ffclient/all_flags/marshal_json/error_in_flag_0.json",
			initModule: true,
		},
		{
			name: "module not init",
			config: Config{
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
			config: Config{
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
			tt.config.DataExporter = DataExporter{
				FlushInterval:    1000,
				MaxEventInMemory: 1,
				Exporter:         &fileexporter.Exporter{OutputDir: exportDir},
			}

			var goff *GoFeatureFlag
			var err error
			if tt.initModule {
				goff, err = New(tt.config)
				assert.NoError(t, err)
				defer goff.Close()
			} else {
				// we close directly so we can test with module not init
				goff, _ = New(tt.config)
				goff.Close()
			}

			user := ffcontext.NewEvaluationContext("random-key")
			allFlagsState := goff.AllFlagsState(user)
			assert.Equal(t, tt.valid, allFlagsState.IsValid())

			// expected JSON output - we force the timestamp
			expected, _ := os.ReadFile(tt.jsonOutput)
			var f map[string]interface{}
			_ = json.Unmarshal(expected, &f)
			if expectedFlags, ok := f["flags"].(map[string]interface{}); ok {
				for _, value := range expectedFlags {
					if valueObj, ok := value.(map[string]interface{}); ok {
						assert.NotNil(t, valueObj["timestamp"])
						assert.NotEqual(t, 0, valueObj["timestamp"])
						valueObj["timestamp"] = time.Now().Unix()
					}
				}
			}
			expectedJSON, _ := json.Marshal(f)
			marshaled, err := allFlagsState.MarshalJSON()
			assert.NoError(t, err)
			assert.JSONEq(t, string(expectedJSON), string(marshaled))

			// no data exported
			files, _ := os.ReadDir(exportDir)
			assert.Equal(t, 0, len(files))
		})
	}
}

func TestGetFlagStates(t *testing.T) {
	tests := []struct {
		name              string
		config            Config
		valid             bool
		jsonOutput        string
		initModule        bool
		evaluationContext ffcontext.EvaluationContext
	}{
		{
			name: "Valid multiple flags",
			config: Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/get_flagstates/config_flag/flag-config-all-flags.yaml",
				},
			},
			valid:      true,
			jsonOutput: "./testdata/ffclient/get_flagstates/marshal_json/valid_flag1_flag4.json",
			initModule: true,
			evaluationContext: ffcontext.NewEvaluationContextBuilder("123").AddCustom("gofeatureflag", map[string]interface{}{
				"flagList": []string{"test-flag1", "test-flag4"},
			}).Build(),
		},
		{
			name: "empty list of flags in context",
			config: Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/get_flagstates/config_flag/flag-config-all-flags.yaml",
				},
			},
			valid:      true,
			jsonOutput: "./testdata/ffclient/get_flagstates/marshal_json/all_flags.json",
			initModule: true,
			evaluationContext: ffcontext.NewEvaluationContextBuilder("123").AddCustom("gofeatureflag", map[string]interface{}{
				"flagList": []string{},
			}).Build(),
		},
		{
			name: "no field in context context",
			config: Config{
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
			config: Config{
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
			tt.config.DataExporter = DataExporter{
				FlushInterval:    1000,
				MaxEventInMemory: 1,
				Exporter:         &fileexporter.Exporter{OutputDir: exportDir},
			}

			var goff *GoFeatureFlag
			var err error
			if tt.initModule {
				goff, err = New(tt.config)
				assert.NoError(t, err)
				defer goff.Close()
			} else {
				// we close directly so we can test with module not init
				goff, _ = New(tt.config)
				goff.Close()
			}

			allFlagsState := goff.GetFlagStates(tt.evaluationContext, tt.evaluationContext.ExtractGOFFProtectedFields().FlagList)
			assert.Equal(t, tt.valid, allFlagsState.IsValid())

			// expected JSON output - we force the timestamp
			expected, _ := os.ReadFile(tt.jsonOutput)
			var f map[string]interface{}
			_ = json.Unmarshal(expected, &f)
			if expectedFlags, ok := f["flags"].(map[string]interface{}); ok {
				for _, value := range expectedFlags {
					if valueObj, ok := value.(map[string]interface{}); ok {
						assert.NotNil(t, valueObj["timestamp"])
						assert.NotEqual(t, 0, valueObj["timestamp"])
						valueObj["timestamp"] = time.Now().Unix()
					}
				}
			}
			expectedJSON, _ := json.Marshal(f)
			marshaled, err := allFlagsState.MarshalJSON()
			assert.NoError(t, err)
			assert.JSONEq(t, string(expectedJSON), string(marshaled))

			// no data exported
			files, _ := os.ReadDir(exportDir)
			assert.Equal(t, 0, len(files))
		})
	}
}

func TestAllFlagsFromCache(t *testing.T) {
	tests := []struct {
		name       string
		config     Config
		initModule bool
	}{
		{
			name: "Valid multiple types",
			config: Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
			},
			initModule: true,
		},
		{
			name: "module not init",
			config: Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
			},
			initModule: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var goff *GoFeatureFlag
			var err error
			if tt.initModule {
				goff, err = New(tt.config)
				assert.NoError(t, err)
				defer goff.Close()

				flags, err := goff.GetFlagsFromCache()
				assert.NoError(t, err)

				cf, _ := goff.cache.AllFlags()
				assert.Equal(t, flags, cf)
			} else {
				// we close directly so we can test with module not init
				goff, _ = New(tt.config)
				goff.Close()

				_, err := goff.GetFlagsFromCache()
				assert.Error(t, err)
			}
		})
	}
}
