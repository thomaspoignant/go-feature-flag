package utils_test

import (
	"testing"

	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

var buildHashSink uint32

func BenchmarkBuildHash(b *testing.B) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "short key", key: "user-1"},
		{name: "uuid key", key: "3f1c1a52-8a5e-4c1e-9d0a-2b7f6c2e9a10"},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				buildHashSink = utils.BuildHash("my-flag", tt.key, 100000)
			}
		})
	}
}
