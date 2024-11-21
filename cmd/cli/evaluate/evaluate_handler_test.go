package evaluate_test

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/evaluate"
)

func TestRunEvaluate(t *testing.T) {
	tests := []struct {
		name           string
		flagFormat     string
		configFile     string
		flag           string
		ctx            string
		wantErr        assert.ErrorAssertionFunc
		expectedResult string
	}{
		{
			name:           "should return an error if flag does not exists",
			flagFormat:     "yaml",
			configFile:     "testdata/flag.goff.yaml",
			flag:           "non-existent-flag",
			ctx:            `{"targetingKey": "user-123"}`,
			wantErr:        assert.NoError,
			expectedResult: "testdata/res/empty.json",
		},
		{
			name:           "should return all flags if no flag is provided",
			flagFormat:     "yaml",
			configFile:     "testdata/flag.goff.yaml",
			ctx:            `{"targetingKey": "user-123"}`,
			wantErr:        assert.NoError,
			expectedResult: "testdata/res/all-flag.json",
		},
		{
			name:           "should return a single flag if flag name is provided",
			flagFormat:     "yaml",
			configFile:     "testdata/flag.goff.yaml",
			flag:           "test-flag",
			ctx:            `{"targetingKey": "user-123"}`,
			wantErr:        assert.NoError,
			expectedResult: "testdata/res/single-flag.json",
		},
		{
			name:       "should return an error if configuration file does not exist",
			flagFormat: "yaml",
			configFile: "testdata/does-not-exist.yaml",
			flag:       "test-flag",
			ctx:        `{"targetingKey": "user-123"}`,
			wantErr:    assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentStdout := os.Stdout
			defer func() { os.Stdout = currentStdout }()

			os.Stdout, _ = os.CreateTemp("", "temp")
			defer func() { _ = os.Remove(os.Stdout.Name()) }()

			err := evaluate.RunEvaluate(&cobra.Command{}, []string{}, tt.flagFormat, tt.configFile, tt.flag, tt.ctx)
			tt.wantErr(t, err)

			if tt.expectedResult != "" {
				expectedContent, err := os.ReadFile(tt.expectedResult)
				require.NoError(t, err)
				gotContent, err := os.ReadFile(os.Stdout.Name())
				require.NoError(t, err)
				assert.JSONEq(t, string(expectedContent), string(gotContent))
			}
		})
	}
}
