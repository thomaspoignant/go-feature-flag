package helper_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/helper"
)

func TestOutput_PrintLines(t *testing.T) {
	pterm.DisableStyling()
	pterm.DisableColor()

	tests := []struct {
		name     string
		output   helper.Output
		expected string
	}{
		{
			name: "print info level line",
			output: helper.Output{
				Lines: []helper.OutputLine{
					{Text: "Info message", Level: helper.InfoLevel},
				},
			},
			expected: "INFO: Info message\n",
		},
		{
			name: "print warning level line",
			output: helper.Output{
				Lines: []helper.OutputLine{
					{Text: "Warning message", Level: helper.WarnLevel},
				},
			},
			expected: "WARNING: Warning message\n",
		},
		{
			name: "print error level line",
			output: helper.Output{
				Lines: []helper.OutputLine{
					{Text: "Error message", Level: helper.ErrorLevel},
				},
			},
			expected: "ERROR: Error message\n",
		},
		{
			name: "print default level line",
			output: helper.Output{
				Lines: []helper.OutputLine{
					{Text: "Default message", Level: helper.DefaultLevel},
				},
			},
			expected: "Default message\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			tt.output.PrintLines(cmd)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestPrintFatalAndExit(t *testing.T) {
	pterm.DisableStyling()
	pterm.DisableColor()

	var buf bytes.Buffer
	pterm.Error.Writer = &buf // redirect Error output to buf

	called := false
	restore := helper.SetExitFuncForTest(func(code int) {
		assert.Equal(t, 1, code)
		called = true
	})
	defer restore()

	helper.PrintFatalAndExit(fmt.Errorf("boom"))

	assert.True(t, called, "exitFunc should have been called")
	assert.Equal(t, "ERROR: error executing command: boom\n", buf.String())
}
