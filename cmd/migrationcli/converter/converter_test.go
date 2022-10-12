package converter_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/stretchr/testify/assert"

	"github.com/thomaspoignant/go-feature-flag/cmd/migrationcli/converter"
)

func TestFlagConverter_Migrate(t *testing.T) {
	tests := []struct {
		name             string
		converter        converter.FlagConverter
		wantFileLocation string
		wantErr          assert.ErrorAssertionFunc
	}{
		{
			name: "Simple flag YAML to YAML",
			converter: converter.FlagConverter{
				InputFile:    "../testdata/input-simple.yaml",
				InputFormat:  "YAML",
				OutputFormat: "YAML",
			},
			wantFileLocation: "../testdata/output-simple.yaml",
			wantErr:          assert.NoError,
		},
		{
			name: "Simple flag YAML to JSON",
			converter: converter.FlagConverter{
				InputFile:    "../testdata/input-simple.yaml",
				InputFormat:  "YAML",
				OutputFormat: "JSON",
			},
			wantFileLocation: "../testdata/output-simple.json",
			wantErr:          assert.NoError,
		},
		{
			name: "Simple flag YAML to TOML",
			converter: converter.FlagConverter{
				InputFile:    "../testdata/input-simple.yaml",
				InputFormat:  "YAML",
				OutputFormat: "TOML",
			},
			wantFileLocation: "../testdata/output-simple.toml",
			wantErr:          assert.NoError,
		},
		{
			name: "Simple flag TOML to TOML",
			converter: converter.FlagConverter{
				InputFile:    "../testdata/input-simple.toml",
				InputFormat:  "TOML",
				OutputFormat: "TOML",
			},
			wantFileLocation: "../testdata/output-simple.toml",
			wantErr:          assert.NoError,
		},
		{
			name: "Simple flag TOML to JSON",
			converter: converter.FlagConverter{
				InputFile:    "../testdata/input-simple.toml",
				InputFormat:  "TOML",
				OutputFormat: "JSON",
			},
			wantFileLocation: "../testdata/output-simple.json",
			wantErr:          assert.NoError,
		},
		{
			name: "Simple flag TOML to YAML",
			converter: converter.FlagConverter{
				InputFile:    "../testdata/input-simple.toml",
				InputFormat:  "TOML",
				OutputFormat: "YAML",
			},
			wantFileLocation: "../testdata/output-simple.yaml",
			wantErr:          assert.NoError,
		},
		{
			name: "Simple flag JSON to JSON",
			converter: converter.FlagConverter{
				InputFile:    "../testdata/input-simple.json",
				InputFormat:  "JSON",
				OutputFormat: "JSON",
			},
			wantFileLocation: "../testdata/output-simple.json",
			wantErr:          assert.NoError,
		},
		{
			name: "Simple flag JSON to TOML",
			converter: converter.FlagConverter{
				InputFile:    "../testdata/input-simple.json",
				InputFormat:  "JSON",
				OutputFormat: "TOML",
			},
			wantFileLocation: "../testdata/output-simple.toml",
			wantErr:          assert.NoError,
		},
		{
			name: "Simple flag JSON to YAML",
			converter: converter.FlagConverter{
				InputFile:    "../testdata/input-simple.json",
				InputFormat:  "JSON",
				OutputFormat: "YAML",
			},
			wantFileLocation: "../testdata/output-simple.yaml",
			wantErr:          assert.NoError,
		},
		{
			name: "Progressive rollout flag YAML to YAML",
			converter: converter.FlagConverter{
				InputFile:    "../testdata/input-progressive-rollout.yaml",
				InputFormat:  "YAML",
				OutputFormat: "YAML",
			},
			wantFileLocation: "../testdata/output-progressive-rollout.yaml",
			wantErr:          assert.NoError,
		},
		{
			name: "Experimentation rollout flag YAML to YAML",
			converter: converter.FlagConverter{
				InputFile:    "../testdata/input-experimentation.yaml",
				InputFormat:  "YAML",
				OutputFormat: "YAML",
			},
			wantFileLocation: "../testdata/output-experimentation.yaml",
			wantErr:          assert.NoError,
		},
		{
			name: "Scheduled rollout flag YAML to YAML",
			converter: converter.FlagConverter{
				InputFile:    "../testdata/input-scheduled.yaml",
				InputFormat:  "YAML",
				OutputFormat: "YAML",
			},
			wantFileLocation: "../testdata/output-scheduled.yaml",
			wantErr:          assert.NoError,
		},
		{
			name: "Not valid input format",
			converter: converter.FlagConverter{
				InputFile:    "../testdata/input-scheduled.yaml",
				InputFormat:  "YSON",
				OutputFormat: "YAML",
			},
			wantErr: assert.Error,
		},
		{
			name: "Invalid input file",
			converter: converter.FlagConverter{
				InputFile:    "../testdata/input-invalid.yaml",
				InputFormat:  "YAML",
				OutputFormat: "YAML",
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.converter.Migrate()
			tt.wantErr(t, err, fmt.Sprintf("Migrate(): %s", err))
			if tt.wantFileLocation != "" {
				want, err := os.ReadFile(tt.wantFileLocation)
				assert.NoError(t, err)
				assert.Equal(t, want, got, cmp.Diff(string(want), string(got)))
			}
		})
	}
}
