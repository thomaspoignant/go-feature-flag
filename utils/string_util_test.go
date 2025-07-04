package utils

import (
	"reflect"
	"testing"
)

func TestStringToArray(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "single string with commas",
			input:    []string{"a,b,c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "slice with one empty string",
			input:    []string{""},
			expected: []string{""},
		},
		{
			name:     "slice with multiple elements (should split only first)",
			input:    []string{"a,b", "c"},
			expected: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringToArray(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("StringToArray(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
