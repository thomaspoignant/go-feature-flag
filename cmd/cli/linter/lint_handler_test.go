package linter_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/linter"
)

func TestRunLint(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		lintFlagFormat string
		wantErr        assert.ErrorAssertionFunc
		expectedError  string
	}{
		{
			name:           "missing configuration file location",
			args:           []string{""},
			lintFlagFormat: "json",
			wantErr:        assert.Error,
			expectedError:  "missing configuration file location argument, please provide the location of the configuration file",
		},
		{
			name:           "invalid configuration",
			args:           []string{"testdata/invalid.json"},
			lintFlagFormat: "json",
			wantErr:        assert.Error,
			expectedError:  "testdata/invalid.json: could not parse file: invalid character ':' after top-level value\n",
		},
		{
			name:           "valid configuration",
			args:           []string{"testdata/valid.yaml"},
			lintFlagFormat: "yaml",
			wantErr:        assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentStderr := os.Stderr
			redirectionStderr, err := os.CreateTemp("", "temp")
			require.NoError(t, err)
			os.Stderr = redirectionStderr
			defer func() {
				if r := recover(); r != nil {
					f, _ := os.ReadFile(os.Stderr.Name())
					assert.Equal(t, tt.expectedError, string(f))
				}
				_ = os.Remove(redirectionStderr.Name())
				os.Stderr = currentStderr
			}()

			err = linter.RunLint(nil, tt.args, tt.lintFlagFormat)
			tt.wantErr(t, err, fmt.Sprintf("RunLint(%v, %v)", tt.args, tt.lintFlagFormat))
			if tt.expectedError != "" {
				assert.Equal(t, err.Error(), tt.expectedError)
			}
		})
	}
}
