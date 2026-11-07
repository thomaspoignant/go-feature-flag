package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

func TestIsJSONObject(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid JSON object",
			input: `{"key": "value"}`,
			want:  true,
		},
		{
			name:  "invalid JSON",
			input: `{"key": "value"`,
			want:  false,
		},
		{
			name:  "empty string",
			input: ``,
			want:  false,
		},
		{
			name:  "non-JSON string",
			input: `not a json`,
			want:  false,
		},
		{
			name:  "null",
			input: `null`,
			want:  true,
		},
		{
			name:  "null with surrounding whitespace",
			input: " \t null\r\n ",
			want:  true,
		},
		{
			name:  "null followed by other characters",
			input: `nullx`,
			want:  false,
		},
		{
			name:  "null followed by a closing brace",
			input: `null}`,
			want:  false,
		},
		{
			name:  "two nulls",
			input: `null null`,
			want:  false,
		},
		{
			name:  "null followed by a non-breaking space",
			input: "null\u00a0",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.IsJSONObject(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
