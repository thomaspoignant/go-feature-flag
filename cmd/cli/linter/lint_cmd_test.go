package linter_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/linter"
)

func TestCmdLint(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		lintFlagFormat string
		wantErr        assert.ErrorAssertionFunc
		expectedError  string
	}{
		{
			name:          "missing configuration file location",
			args:          []string{"--format=json"},
			wantErr:       assert.Error,
			expectedError: "impossible to find config file in the default locations [./,/goff/,/etc/opt/goff/]\nError: invalid GO Feature Flag configuration\n",
		},
		{
			name:          "invalid configuration",
			args:          []string{"testdata/invalid.json", "--format=json"},
			wantErr:       assert.Error,
			expectedError: "testdata/invalid.json: could not parse file (json): invalid character ':' after top-level value\nError: invalid GO Feature Flag configuration\n",
		},
		{
			name:    "valid configuration",
			args:    []string{"testdata/valid.yaml", "--format=yaml"},
			wantErr: assert.NoError,
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
			if tt.expectedError != "" {
				content, err := os.ReadFile(redirectionStderr.Name())
				require.NoError(t, err)
				assert.Equal(t, tt.expectedError, string(content))
			}
		})
	}
}
