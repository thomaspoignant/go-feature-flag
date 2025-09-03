package manifest_test

import (
	"os"
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/generate/manifest"
)

func TestManifestCmd(t *testing.T) {
	pterm.DisableStyling()
	pterm.DisableColor()
	tests := []struct {
		name             string
		args             []string
		expectedManifest string
		expectedOutput   string
		expectedError    string
		assertError      assert.ErrorAssertionFunc
	}{
		{
			name:             "should return success if everything is ok",
			args:             []string{"--config=testdata/input/flag.goff.yaml"},
			expectedManifest: "testdata/output/flag.goff.json",
			expectedOutput:   "INFO: 🎉 Manifest has been created\n",
			assertError:      assert.NoError,
		},
		{
			name:             "should ignore flag without value",
			args:             []string{"--config=testdata/input/flag-no-default.yaml"},
			expectedManifest: "testdata/output/flag-no-default.json",
			expectedOutput:   "WARNING: flag test-flag ignored: no default value provided\nINFO: 🎉 Manifest has been created\n",
			assertError:      assert.NoError,
		},
		{
			name:          "should error if flag type is invalid",
			args:          []string{"--config=testdata/input/flag-invalid-flag-type.yaml"},
			assertError:   assert.Error,
			expectedError: "invalid configuration for flag test-flag: impossible to find type",
		},
		{
			name:             "should have int as type if float with .0 and int are mixed",
			args:             []string{"--config=testdata/input/flag-int-as-float.yaml"},
			assertError:      assert.NoError,
			expectedManifest: "testdata/output/flag-int-as-float.json",
			expectedOutput:   "INFO: 🎉 Manifest has been created\n",
		},
		{
			name:             "should have float as type if 1 float and int are mixed",
			args:             []string{"--config=testdata/input/flag-float-and-int.yaml"},
			assertError:      assert.NoError,
			expectedManifest: "testdata/output/flag-float-and-int.json",
			expectedOutput:   "INFO: 🎉 Manifest has been created\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpManifest, err := os.CreateTemp("", "temp")
			require.NoError(t, err)
			os.Remove(tmpManifest.Name()) // delete, we only want the path

			redirectionStd, err := os.CreateTemp("", "temp")
			require.NoError(t, err)
			defer func() { _ = os.Remove(redirectionStd.Name()) }()

			tt.args = append(tt.args, "--flag_manifest_destination", tmpManifest.Name())

			cmd := manifest.NewManifestCmd()
			cmd.SetErr(redirectionStd)
			cmd.SetOut(redirectionStd)
			cmd.SetArgs(tt.args)

			err = cmd.Execute()
			tt.assertError(t, err)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			}

			output, err := os.ReadFile(redirectionStd.Name())
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, string(output), "output is not expected")

			if tt.expectedManifest != "" {
				wantManifest, err := os.ReadFile(tt.expectedManifest)
				assert.NoError(t, err)
				gotManifest, err := os.ReadFile(tmpManifest.Name())
				assert.NoError(t, err)
				assert.Equal(t, string(wantManifest), string(gotManifest), "manifest is not expected")
			}
		})
	}
}
