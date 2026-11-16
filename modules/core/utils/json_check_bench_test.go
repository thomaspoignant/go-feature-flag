package utils_test

import (
	"testing"

	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

var isJSONObjectSink bool

func BenchmarkIsJSONObject(b *testing.B) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "nikunjy query",
			input: `language eq "ar" and isNewUser eq true and clubsTimeSpent gt 600`,
		},
		{
			name:  "jsonlogic query",
			input: `{"and":[{"==":[{"var":"language"},"ar"]},{">":[{"var":"clubsTimeSpent"},600]}]}`,
		},
		{name: "null", input: "null"},
		{
			name:  "whitespace-prefixed object",
			input: "  \n\t{\"==\":[{\"var\":\"language\"},\"ar\"]}",
		},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				isJSONObjectSink = utils.IsJSONObject(tt.input)
			}
		})
	}
}
