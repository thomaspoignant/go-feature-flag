package utils_test

import (
	"fmt"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

var contextToMapSink map[string]any

func BenchmarkContextToMap(b *testing.B) {
	for _, n := range []int{0, 5, 15} {
		builder := ffcontext.NewEvaluationContextBuilder("3f1c1a52-8a5e-4c1e-9d0a-2b7f6c2e9a10")
		for i := range n {
			builder.AddCustom(fmt.Sprintf("attr%d", i), i)
		}
		ctx := builder.Build()
		b.Run(fmt.Sprintf("%d custom attributes", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				contextToMapSink = utils.ContextToMap(ctx)
			}
		})
	}
}
