package utils_test

import (
	"testing"

	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

var strTrimSink string

func BenchmarkStrTrim(b *testing.B) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "single line", input: `language eq "ar" and isNewUser eq true`},
		{
			name:  "multiline",
			input: "language eq \"ar\"\n   and isNewUser eq true\n   and clubsTimeSpent gt 600",
		},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				strTrimSink = utils.StrTrim(tt.input)
			}
		})
	}
}
