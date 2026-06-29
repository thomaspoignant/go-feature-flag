package generate_test

import (
	"bytes"
	"testing"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/generate"
)

type generateCmdTestCase struct {
	name          string
	args          []string
	wantErr       bool
	expectedErr   string
	expectedOut   string
	checkManifest bool
}

func TestNewGenerateCmd(t *testing.T) {
	pterm.DisableStyling()
	pterm.DisableColor()

	tests := []generateCmdTestCase{
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

			verifyGenerateCmdResult(t, tt, cmd, &buf, err)
		})
	}
}

// verifyGenerateCmdResult asserts the outcome of executing the generate command
// for a single test case. Keeping it flat (no deep nesting) avoids inflating the
// cognitive complexity of the parent test.
func verifyGenerateCmdResult(
	t *testing.T,
	tt generateCmdTestCase,
	cmd *cobra.Command,
	buf *bytes.Buffer,
	err error,
) {
	t.Helper()

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
		assert.True(t, hasManifestSubcommand(cmd),
			"manifest subcommand should be wired into generate")
	}
}

// hasManifestSubcommand reports whether the manifest subcommand is wired into cmd.
func hasManifestSubcommand(cmd *cobra.Command) bool {
	for _, c := range cmd.Commands() {
		if c.Name() == "manifest" {
			return true
		}
	}
	return false
}
