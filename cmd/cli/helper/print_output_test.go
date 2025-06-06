package helper_test

import (
	"bytes"
	"testing"

	"github.com/pterm/pterm" // Added pterm
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/helper"
)

func TestOutput_Add(t *testing.T) {
	tests := []struct {
		name     string
		initial  helper.Output
		line     string
		level    helper.Level
		expected helper.Output
	}{
		{
			name:    "add info level line",
			initial: helper.Output{},
			line:    "Info message",
			level:   helper.InfoLevel,
			expected: helper.Output{
				Lines: []helper.OutputLine{
					{Text: "Info message", Level: helper.InfoLevel},
				},
			},
		},
		{
			name:    "add warning level line",
			initial: helper.Output{},
			line:    "Warning message",
			level:   helper.WarnLevel,
			expected: helper.Output{
				Lines: []helper.OutputLine{
					{Text: "Warning message", Level: helper.WarnLevel},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initial.Add(tt.line, tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOutput_PrintLines(t *testing.T) {
	tests := []struct {
		name     string
		output   helper.Output
		expected string
		level    helper.Level // Added level for better assertions
	}{
		{
			name: "print info level line",
			output: helper.Output{
				Lines: []helper.OutputLine{
					{Text: "Info message", Level: helper.InfoLevel},
				},
			},
			expected: "Info message", // Will check for substring due to pterm formatting
			level:    helper.InfoLevel,
		},
		{
			name: "print warning level line",
			output: helper.Output{
				Lines: []helper.OutputLine{
					{Text: "Warning message", Level: helper.WarnLevel},
				},
			},
			expected: "Warning message", // Will check for substring due to pterm formatting
			level:    helper.WarnLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer // Scoped a new buffer for each test run
			cmd := &cobra.Command{}
			// cmd.SetOut(&buf) // Not needed as pterm writes globally or to its configured writers

			originalInfoWriter := pterm.Info.Writer
			originalWarningWriter := pterm.Warning.Writer

			pterm.Info.Writer = &buf
			pterm.Warning.Writer = &buf
			pterm.DisableStyling() // Disable colors/styles for simpler assertion

			tt.output.PrintLines(cmd)

			pterm.Info.Writer = originalInfoWriter // Restore original writer
			pterm.Warning.Writer = originalWarningWriter // Restore original writer
			pterm.EnableStyling() // Re-enable styling

			outputStr := buf.String()
			assert.Contains(t, outputStr, tt.expected, "The output string should contain the expected text")

			if tt.level == helper.WarnLevel {
				assert.Contains(t, outputStr, "WARNING", "Warning lines should contain WARNING prefix")
			} else {
				assert.Contains(t, outputStr, "INFO", "Info lines should contain INFO prefix")
			}
		})
	}
}
