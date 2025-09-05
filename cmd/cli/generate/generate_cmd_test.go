package generate_test

import (
	"bytes"
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/generate"
)

func TestNewGenerateCmd(t *testing.T) {
	pterm.DisableStyling()
	pterm.DisableColor()

	tests := []struct {
		name          string
		args          []string
		wantErr       bool
		expectedErr   string
		expectedOut   string
		checkManifest bool
	}{
		{
			name:        "no subcommand provided",
			args:        []string{},
			wantErr:     true,
			expectedErr: "no subcommand provided",
			expectedOut: "ERROR: you must specify a subcommand (e.g., manifest)\n",
		},
		{
			name:          "manifest subcommand is available",
			args:          []string{"manifest", "--help"},
			wantErr:       false,
			expectedOut:   "", // help goes to stdout, we don’t check in detail here
			checkManifest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := generate.NewGenerateCmd()
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
			}

			if tt.expectedOut != "" {
				assert.Equal(t, tt.expectedOut, buf.String())
			}

			if tt.checkManifest {
				found := false
				for _, c := range cmd.Commands() {
					if c.Name() == "manifest" {
						found = true
						break
					}
				}
				assert.True(t, found, "manifest subcommand should be wired into generate")
			}
		})
	}
}
