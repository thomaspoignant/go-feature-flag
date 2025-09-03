package linter_test

import (
	"os"
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/linter"
)

func TestCmdLint(t *testing.T) {
	pterm.DisableStyling()
	pterm.DisableColor()
	tests := []struct {
		name              string
		args              []string
		wantErr           assert.ErrorAssertionFunc
		expectedStderr    string
		expectedErrString string
	}{
		{
			name:              "missing configuration file location",
			args:              []string{"--format=json"},
			wantErr:           assert.Error,
			expectedStderr:    "ERROR: impossible to find config file in the default locations [./,/goff/,/etc/opt/goff/]\n",
			expectedErrString: "invalid GO Feature Flag configuration",
		},
		{
			name:              "invalid configuration",
			args:              []string{"testdata/invalid.json", "--format=json"},
			wantErr:           assert.Error,
			expectedStderr:    "ERROR: testdata/invalid.json: could not parse file (json): invalid character ':' after top-level value\n",
			expectedErrString: "invalid GO Feature Flag configuration",
		},
		{
			name:           "valid configuration",
			args:           []string{"testdata/valid.yaml", "--format=yaml"},
			wantErr:        assert.NoError,
			expectedStderr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redirectionStderr, err := os.CreateTemp("", "temp")
			require.NoError(t, err)
			redirectionStdout, err := os.CreateTemp("", "out")
			require.NoError(t, err)
			defer func() {
				_ = os.Remove(redirectionStderr.Name())
				_ = os.Remove(redirectionStdout.Name())
			}()

			cmd := linter.NewLintCmd()
			cmd.SetErr(redirectionStderr)
			cmd.SetOut(redirectionStdout)
			cmd.SetArgs(tt.args)
			err = cmd.Execute()
			tt.wantErr(t, err)

			// Check stderr content
			content, rerr := os.ReadFile(redirectionStderr.Name())
			require.NoError(t, rerr)
			assert.Equal(t, tt.expectedStderr, string(content))

			// If we expect an error string, check it separately
			if tt.expectedErrString != "" {
				assert.EqualError(t, err, tt.expectedErrString)
			}
		})
	}
}
