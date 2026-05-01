package manifest_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/manifest"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

// captureHandler stores every slog record so tests can assert message content
// and severity emitted by GenerateDefinition. The handler is wired in as the
// default slog logger via slog.SetDefault so it intercepts the package-level
// slog calls performed by the code under test.
type captureHandler struct {
	msgs   []string
	levels []slog.Level
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.msgs = append(h.msgs, r.Message)
	h.levels = append(h.levels, r.Level)
	return nil
}
func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }

// installCaptureLogger swaps the default slog logger for one backed by a
// captureHandler and registers a cleanup that restores the previous default,
// so each subtest gets an isolated capture.
func installCaptureLogger(t *testing.T) *captureHandler {
	t.Helper()
	h := &captureHandler{}
	previous := slog.Default()
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return h
}

func ptrString(v string) *string { return &v }

// variationsPtr builds the *map[string]*any value expected by InternalFlag.Variations
// from a plain map[string]any literal.
func variationsPtr(values map[string]any) *map[string]*any {
	out := make(map[string]*any, len(values))
	for k, v := range values {
		v := v
		out[k] = &v
	}
	return &out
}

func metadataPtr(m map[string]any) *map[string]any {
	if m == nil {
		return nil
	}
	return &m
}

func defaultRule(variation string) *flag.Rule {
	return &flag.Rule{
		Name:            ptrString("defaultRule"),
		VariationResult: ptrString(variation),
	}
}

// boolFlag returns a minimal boolean InternalFlag whose variations are True/False
// and whose default rule selects the variation passed in.
func boolFlag(defaultVariation string, metadata map[string]any) flag.InternalFlag {
	return flag.InternalFlag{
		Variations:  variationsPtr(map[string]any{"True": true, "False": false}),
		DefaultRule: defaultRule(defaultVariation),
		Metadata:    metadataPtr(metadata),
	}
}

// boolFlagWithExperimentation is like boolFlag but attaches an experimentation window.
func boolFlagWithExperimentation(
	defaultVariation string,
	metadata map[string]any,
	experimentationStart, experimentationEnd time.Time,
) flag.InternalFlag {
	f := boolFlag(defaultVariation, metadata)
	s, e := experimentationStart, experimentationEnd
	f.Experimentation = &flag.ExperimentationRollout{
		Start: &s,
		End:   &e,
	}
	return f
}

func readGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "generate_definition", name))
	require.NoError(t, err)
	return string(data)
}

func TestGenerateDefinition(t *testing.T) {
	tests := []struct {
		name       string
		flags      map[string]flag.InternalFlag
		goldenFile string
		// wantLogs is the set of expected slog messages. Map iteration is
		// non-deterministic, so order is not asserted.
		wantLogs []string
		// wantErr, when non-empty, asserts the returned error contains it.
		// In that case the manifest output (and golden file) are not checked.
		wantErr string
	}{
		{
			name:       "empty input produces empty manifest",
			flags:      map[string]flag.InternalFlag{},
			goldenFile: "empty.json",
		},
		{
			name: "boolean flag",
			flags: map[string]flag.InternalFlag{
				"boolean-flag": boolFlag("False", map[string]any{
					"defaultValue": false,
					"description":  "a boolean flag",
				}),
			},
			goldenFile: "boolean.json",
		},
		{
			name: "boolean flag with experimentation",
			flags: map[string]flag.InternalFlag{
				"exp-flag": boolFlagWithExperimentation(
					"False",
					map[string]any{
						"defaultValue": false,
						"description":  "A/B experiment",
					},
					time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
					time.Date(2026, 6, 30, 23, 59, 59, 0, time.UTC),
				),
			},
			goldenFile: "experimentation.json",
		},
		{
			name: "string flag",
			flags: map[string]flag.InternalFlag{
				"string-flag": {
					Variations:  variationsPtr(map[string]any{"A": "a", "B": "b"}),
					DefaultRule: defaultRule("A"),
					Metadata: metadataPtr(map[string]any{
						"defaultValue": "a",
						"description":  "strings",
					}),
				},
			},
			goldenFile: "string.json",
		},
		{
			name: "integer flag",
			flags: map[string]flag.InternalFlag{
				"int-flag": {
					Variations:  variationsPtr(map[string]any{"One": 1, "Two": 2}),
					DefaultRule: defaultRule("One"),
					Metadata: metadataPtr(map[string]any{
						"defaultValue": 1,
						"description":  "ints",
					}),
				},
			},
			goldenFile: "integer.json",
		},
		{
			name: "float flag without description",
			flags: map[string]flag.InternalFlag{
				"float-flag": {
					Variations:  variationsPtr(map[string]any{"Half": 0.5, "OneAndHalf": 1.5}),
					DefaultRule: defaultRule("Half"),
					Metadata: metadataPtr(map[string]any{
						"defaultValue": 0.5,
					}),
				},
			},
			goldenFile: "float.json",
		},
		{
			name: "object flag",
			flags: map[string]flag.InternalFlag{
				"object-flag": {
					Variations: variationsPtr(map[string]any{
						"A": map[string]any{"key": "value-a"},
						"B": map[string]any{"key": "value-b"},
					}),
					DefaultRule: defaultRule("A"),
					Metadata: metadataPtr(map[string]any{
						"defaultValue": map[string]any{"key": "value-a"},
						"description":  "objects",
					}),
				},
			},
			goldenFile: "object.json",
		},
		{
			name: "int and float .0 mix resolves to integer",
			flags: map[string]flag.InternalFlag{
				"mixed-flag": {
					Variations:  variationsPtr(map[string]any{"Whole": 1, "Half": 2.0}),
					DefaultRule: defaultRule("Whole"),
					Metadata: metadataPtr(map[string]any{
						"defaultValue": 1,
						"description":  "int as float",
					}),
				},
			},
			goldenFile: "int_as_float.json",
		},
		{
			name: "int and non-whole float mix resolves to float",
			flags: map[string]flag.InternalFlag{
				"mixed-flag": {
					Variations:  variationsPtr(map[string]any{"One": 1, "OneAndHalf": 1.5}),
					DefaultRule: defaultRule("One"),
					Metadata: metadataPtr(map[string]any{
						"defaultValue": 1,
						"description":  "float and int",
					}),
				},
			},
			goldenFile: "float_and_int.json",
		},
		{
			name: "missing description metadata defaults to empty string",
			flags: map[string]flag.InternalFlag{
				"minimal-flag": boolFlag("False", map[string]any{
					"defaultValue": false,
				}),
			},
			goldenFile: "no_description.json",
		},
		{
			name: "non-string description is ignored",
			flags: map[string]flag.InternalFlag{
				"weird-flag": boolFlag("False", map[string]any{
					"defaultValue": false,
					"description":  42,
				}),
			},
			goldenFile: "description_non_string.json",
		},
		{
			name: "flag without metadata is skipped and logged",
			flags: map[string]flag.InternalFlag{
				"no-meta": {
					Variations:  variationsPtr(map[string]any{"True": true, "False": false}),
					DefaultRule: defaultRule("False"),
				},
			},
			goldenFile: "empty.json",
			wantLogs:   []string{"flag no-meta ignored: no default value provided"},
		},
		{
			name: "metadata without defaultValue is skipped and logged",
			flags: map[string]flag.InternalFlag{
				"no-default": boolFlag("False", map[string]any{
					"description": "no default",
				}),
			},
			goldenFile: "empty.json",
			wantLogs:   []string{"flag no-default ignored: no default value provided"},
		},
		{
			name: "valid flag is kept while invalid sibling is skipped",
			flags: map[string]flag.InternalFlag{
				"valid-flag": boolFlag("True", map[string]any{
					"defaultValue": true,
				}),
				"skipped-flag": boolFlag("True", map[string]any{
					"description": "no default",
				}),
			},
			goldenFile: "partial_skip.json",
			wantLogs:   []string{"flag skipped-flag ignored: no default value provided"},
		},
		{
			name: "multiple valid flags are all serialized",
			flags: map[string]flag.InternalFlag{
				"flag-a": boolFlag("True", map[string]any{
					"defaultValue": true,
				}),
				"flag-b": {
					Variations:  variationsPtr(map[string]any{"One": 1, "Two": 2}),
					DefaultRule: defaultRule("One"),
					Metadata: metadataPtr(map[string]any{
						"defaultValue": 1,
						"description":  "second",
					}),
				},
			},
			goldenFile: "multi_flags.json",
		},
		{
			name: "unsupported variation type returns wrapped error",
			flags: map[string]flag.InternalFlag{
				"bad-flag": {
					Variations: variationsPtr(map[string]any{
						"Slice": []int{1, 2, 3},
					}),
					DefaultRule: defaultRule("Slice"),
					Metadata: metadataPtr(map[string]any{
						"defaultValue": []int{1, 2, 3},
					}),
				},
			},
			wantErr: "invalid configuration for flag bad-flag: impossible to find type",
		},
		{
			name: "incompatible variation types return wrapped error",
			flags: map[string]flag.InternalFlag{
				"bad-flag": {
					Variations:  variationsPtr(map[string]any{"A": true, "B": "string"}),
					DefaultRule: defaultRule("A"),
					Metadata: metadataPtr(map[string]any{
						"defaultValue": true,
					}),
				},
			},
			wantErr: "invalid configuration for flag bad-flag: impossible to find type",
		},
		{
			name: "empty variations map returns wrapped error",
			flags: map[string]flag.InternalFlag{
				"bad-flag": {
					Variations:  variationsPtr(map[string]any{}),
					DefaultRule: defaultRule(""),
					Metadata: metadataPtr(map[string]any{
						"defaultValue": "ignored",
					}),
				},
			},
			wantErr: "invalid configuration for flag bad-flag: impossible to find type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := installCaptureLogger(t)

			got, err := manifest.GenerateDefinitionFromInternalFlags(tt.flags)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			actual, mErr := json.MarshalIndent(got, "", "  ")
			require.NoError(t, mErr)
			expected := readGolden(t, tt.goldenFile)
			assert.Equal(t, expected, string(actual))

			assert.ElementsMatch(t, tt.wantLogs, handler.msgs)
			for _, lvl := range handler.levels {
				assert.Equal(t, slog.LevelWarn, lvl)
			}
		})
	}
}
