package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

func TestStrTrim(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single line with leading spaces",
			input: "   hello",
			want:  "hello",
		},
		{
			name:  "multiple lines with leading spaces",
			input: "   hello\n   world",
			want:  "helloworld",
		},
		{
			name:  "no leading spaces",
			input: "hello\nworld",
			want:  "helloworld",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only spaces",
			input: "   ",
			want:  "",
		},
		{
			name:  "mixed leading spaces",
			input: "   hello\nworld",
			want:  "helloworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.StrTrim(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
