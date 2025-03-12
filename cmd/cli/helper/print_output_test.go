package helper_test

import (
	"bytes"
	"testing"

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
	}{
		{
			name: "print info level line",
			output: helper.Output{
				Lines: []helper.OutputLine{
					{Text: "Info message", Level: helper.InfoLevel},
				},
			},
			expected: "Info message\n",
		},
		{
			name: "print warning level line",
			output: helper.Output{
				Lines: []helper.OutputLine{
					{Text: "Warning message", Level: helper.WarnLevel},
				},
			},
			expected: "⚠️ Warning message\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&buf)

			tt.output.PrintLines(cmd)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}
