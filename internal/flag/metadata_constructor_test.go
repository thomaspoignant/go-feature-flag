// Since the function we are about to test is internal,
// I've added this test package in the main pack instead of proper one.
package flag

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func Test_constructMetadataParallel(t *testing.T) {
	sharedMetadata := map[string]any{
		"key1": "value1",
	}

	var tests []struct {
		name                  string
		wantEvaluatedRuleName string
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	// generate test cases
	for i := 0; i < 10_000; i++ {
		ruleName := fmt.Sprintf("rule-%d", i)
		tests = append(tests, struct {
			name                  string
			wantEvaluatedRuleName string
		}{
			name:                  fmt.Sprintf("Rule %d", i),
			wantEvaluatedRuleName: ruleName,
		})
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := constructMetadata(sharedMetadata, &tt.wantEvaluatedRuleName)
			assert.Equal(t, tt.wantEvaluatedRuleName, got["evaluatedRuleName"])
		})
	}
}
