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
			name:  "JSON object after leading whitespace",
			input: " \t\r\n{\"key\": \"value\"}",
			want:  true,
		},
		{
			name:  "JSON null with surrounding whitespace",
			input: ` null `,
			want:  true,
		},
		{
			name:  "JSON array",
			input: `[{"key": "value"}]`,
			want:  false,
		},
		{
			name:  "JSON string",
			input: `"value"`,
			want:  false,
		},
		{
			name:  "nikunjy query",
			input: `key in ["a", "b"]`,
			want:  false,
		},
		{
			name:  "nikunjy query starting with null",
			input: `nullable eq true`,
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
