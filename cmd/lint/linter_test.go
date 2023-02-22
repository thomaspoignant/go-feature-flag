package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
)
import "testing"

func TestLinter_Lint(t *testing.T) {
	tests := []struct {
		name    string
		linter  Linter
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "flag-config.yaml",
			linter: Linter{
				InputFile:   "../../testdata/flag-config.yaml",
				InputFormat: "YAML",
			},
			wantErr: assert.NoError,
		},
		{
			name: "flag-config.json",
			linter: Linter{
				InputFile:   "../../testdata/flag-config.json",
				InputFormat: "JSON",
			},
			wantErr: assert.NoError,
		},
		{
			name: "flag-config.toml",
			linter: Linter{
				InputFile:   "../../testdata/flag-config.toml",
				InputFormat: "TOML",
			},
			wantErr: assert.NoError,
		},
		{
			name: "invalid yaml",
			linter: Linter{
				InputFile:   "testdata/invalid.yaml",
				InputFormat: "yaml",
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid json",
			linter: Linter{
				InputFile:   "testdata/invalid.json",
				InputFormat: "json",
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid input format",
			linter: Linter{
				InputFile:   "testdata/invalid.json",
				InputFormat: "swift",
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid toml",
			linter: Linter{
				InputFile:   "testdata/invalid.toml",
				InputFormat: "toml",
			},
			wantErr: assert.Error,
		},
		{
			name: "no variation",
			linter: Linter{
				InputFile:   "testdata/no-variation.yaml",
				InputFormat: "yaml",
			},
			wantErr: assert.Error,
		},
		{
			name: "no default rule",
			linter: Linter{
				InputFile:   "testdata/no-default-rule.yaml",
				InputFormat: "yaml",
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid rule",
			linter: Linter{
				InputFile:   "testdata/invalid-rule.yaml",
				InputFormat: "yaml",
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.linter.Lint()
			for _, err := range errs {
				tt.wantErr(t, err, fmt.Sprintf("Lint(): %s", err))
			}
		})
	}
}
