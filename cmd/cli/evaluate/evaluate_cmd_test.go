package evaluate_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/evaluate"
)

func TestCmdEvaluate(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantErr        assert.ErrorAssertionFunc
		expectedResult string
	}{
		{
			name:           "should return an error if flag does not exists",
			args:           []string{"--config", "testdata/flag.goff.yaml", "--flag", "non-existent-flag", "--ctx", `{"targetingKey": "user-123"}`, "--format", "yaml"},
			wantErr:        assert.NoError,
			expectedResult: "testdata/res/empty.json",
		},
		{
			name:           "should return all flags if no flag is provided",
			args:           []string{"--config", "testdata/flag.goff.yaml", "--ctx", `{"targetingKey": "user-123"}`, "--format", "yaml"},
			wantErr:        assert.NoError,
			expectedResult: "testdata/res/all-flag.json",
		},
		{
			name:           "should return a single flag if flag name is provided",
			args:           []string{"--config", "testdata/flag.goff.yaml", "--flag", "test-flag", "--ctx", `{"targetingKey": "user-123"}`, "--format", "yaml"},
			wantErr:        assert.NoError,
			expectedResult: "testdata/res/single-flag.json",
		},
		{
			name:    "should return an error if configuration file does not exist",
			args:    []string{"--config", "testdata/does-not-exist.yaml", "--flag", "test-flag", "--ctx", `{"targetingKey": "user-123"}`, "--format", "yaml"},
			wantErr: assert.Error,
		},
		{
			name:    "should return an error if context has no targeting key",
			args:    []string{"--config", "testdata/flag.goff.yaml", "--ctx", `{"id": "user-123"}`, "--format", "yaml"},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdOut, err := os.CreateTemp("", "temp")
			require.NoError(t, err)
			stdErr, err := os.CreateTemp("", "temp")
			require.NoError(t, err)
			defer func() {
				_ = os.Remove(stdOut.Name())
				_ = os.Remove(stdErr.Name())
			}()

			cmd := evaluate.NewEvaluateCmd()
			cmd.SetErr(stdErr)
			cmd.SetOut(stdOut)
			cmd.SetArgs(tt.args)
			err = cmd.Execute()
			tt.wantErr(t, err)
			if tt.expectedResult != "" {
				expectedContent, err := os.ReadFile(tt.expectedResult)
				require.NoError(t, err)
				gotContent, err := os.ReadFile(stdOut.Name())
				require.NoError(t, err)
				assert.JSONEq(t, string(expectedContent), string(gotContent))
			}
		})
	}
}
