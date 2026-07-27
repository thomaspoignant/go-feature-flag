package configfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
	"github.com/thomaspoignant/go-feature-flag/model/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestLoadConfigFile(t *testing.T) {
	tests := []struct {
		name               string
		inputFilePath      string
		configFormat       string
		defaultLocations   []string
		expected           map[string]dto.DTO
		expectErr          bool
		useDefaultLocation bool
	}{
		{
			name:          "should load the flag file as yaml",
			inputFilePath: "testdata/flag.goff.yaml",
			configFormat:  "yaml",
			expected: map[string]dto.DTO{
				"test-flag": {
					Variations: &map[string]*any{
						"var_a": testconvert.Interface(true),
						"var_b": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_a"),
					},
					Metadata: &map[string]any{
						"description":  "this is a simple feature flag",
						"defaultValue": false,
					},
				},
				"test-flag2": {
					Variations: &map[string]*any{
						"var_a": testconvert.Interface(1),
						"var_b": testconvert.Interface(2),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_b"),
					},
					Metadata: &map[string]any{
						"defaultValue": 123,
					},
				},
			},
			expectErr: false,
		},
		{
			name:          "should load the flag file and use yaml as default parser",
			inputFilePath: "testdata/flag.goff.yaml",
			expected: map[string]dto.DTO{
				"test-flag": {
					Variations: &map[string]*any{
						"var_a": testconvert.Interface(true),
						"var_b": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_a"),
					},
					Metadata: &map[string]any{
						"description":  "this is a simple feature flag",
						"defaultValue": false,
					},
				},
				"test-flag2": {
					Variations: &map[string]*any{
						"var_a": testconvert.Interface(1),
						"var_b": testconvert.Interface(2),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_b"),
					},
					Metadata: &map[string]any{
						"defaultValue": 123,
					},
				},
			},
			expectErr: false,
		},
		{
			name:          "should load the flag file as json",
			inputFilePath: "testdata/flag.goff.json",
			configFormat:  "json",
			expected: map[string]dto.DTO{
				"test-flag": {
					Variations: &map[string]*any{
						"var_a": testconvert.Interface(true),
						"var_b": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_a"),
					},
					Metadata: &map[string]any{
						"description":  "this is a simple feature flag",
						"defaultValue": false,
					},
				},
				"test-flag2": {
					Variations: &map[string]*any{
						"var_a": testconvert.Interface(1.0),
						"var_b": testconvert.Interface(2.0),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_b"),
					},
					Metadata: &map[string]any{
						"defaultValue": 123.0,
					},
				},
			},
			expectErr: false,
		},
		{
			name:          "should load the flag file as toml",
			inputFilePath: "testdata/flag.goff.toml",
			configFormat:  "toml",
			expected: map[string]dto.DTO{
				"test-flag": {
					Variations: &map[string]*any{
						"var_a": testconvert.Interface(true),
						"var_b": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_a"),
					},
					Metadata: &map[string]any{
						"description":  "this is a simple feature flag",
						"defaultValue": false,
					},
				},
				"test-flag2": {
					Variations: &map[string]*any{
						"var_a": testconvert.Interface(int64(1)),
						"var_b": testconvert.Interface(int64(2)),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_b"),
					},
					Metadata: &map[string]any{
						"defaultValue": int64(123),
					},
				},
			},
			expectErr: false,
		},
		{
			name:               "should load the flag file as yaml from a default location",
			useDefaultLocation: true,
			inputFilePath:      "testdata/flag.goff.yaml",
			configFormat:       "yaml",
			expected: map[string]dto.DTO{
				"test-flag": {
					Variations: &map[string]*any{
						"var_a": testconvert.Interface(true),
						"var_b": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_a"),
					},
					Metadata: &map[string]any{
						"description":  "this is a simple feature flag",
						"defaultValue": false,
					},
				},
				"test-flag2": {
					Variations: &map[string]*any{
						"var_a": testconvert.Interface(1),
						"var_b": testconvert.Interface(2),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_b"),
					},
					Metadata: &map[string]any{
						"defaultValue": 123,
					},
				},
			},
			expectErr: false,
		},
		{
			name:          "should error if file does not exist",
			inputFilePath: "testdata/does-not-exist.yaml",
			configFormat:  "yaml",
			expectErr:     true,
		},
		{
			name:      "should error if file does not exist in default locations",
			expectErr: true,
		},
		{
			name:          "should error if json is invalid",
			inputFilePath: "testdata/invalid.json",
			configFormat:  "json",
			expectErr:     true,
		},
		{
			name:          "should error if yaml invalid",
			inputFilePath: "testdata/invalid.yaml",
			configFormat:  "yaml",
			expectErr:     true,
		},
		{
			name:          "should error if toml is invalid",
			inputFilePath: "testdata/invalid.toml",
			configFormat:  "toml",
			expectErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.useDefaultLocation {
				// write the config into an isolated per-test dir (auto-removed by the
				// test framework). Using t.TempDir() instead of a fixed name in the
				// shared system temp dir avoids leftovers/collisions between runs and
				// the Windows "file in use" errors from renaming an open handle.
				dir := t.TempDir()
				content, err := os.ReadFile(tt.inputFilePath)
				assert.NoError(t, err)
				err = os.WriteFile(filepath.Join(dir, "flags.goff.yaml"), content, 0600)
				assert.NoError(t, err)
				tt.inputFilePath = ""
				tt.defaultLocations = []string{dir + "/"}
			}

			result, err := configfile.LoadConfigFile(
				tt.inputFilePath,
				tt.configFormat,
				tt.defaultLocations,
			)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
