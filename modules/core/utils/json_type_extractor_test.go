package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

func Test_JSONTypeExtractor(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"string", "hello", "(string)"},
		{"integer", 42, "(number)"},
		{"float", 3.14, "(number)"},
		{"bool", true, "(bool)"},
		{"[]interface", []any{1, "two", 3.0}, "([]interface{})"},
		{"map", map[string]any{"key1": "value1", "key2": 2}, "(map[string]interface{})"},
		{"null", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.JSONTypeExtractor(tt.input)
			assert.NoError(t, err, "unexpected error")
			assert.Equal(t, tt.expected, got)
		})
	}
}
