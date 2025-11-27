package flag

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
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

func Test_ConstructMetadata(t *testing.T) {
	rule := "test-rule"

	tests := []struct {
		name         string
		flagMetadata map[string]any
		ruleName     *string
		want         map[string]any
	}{
		{
			name:         "nil metadata, nil rule",
			flagMetadata: nil,
			ruleName:     nil,
			want:         nil,
		},
		{
			name:         "nil metadata, empty rule",
			flagMetadata: nil,
			ruleName:     stringToPointer(""),
			want:         nil,
		},
		{
			name:         "nil metadata, non-empty rule",
			flagMetadata: nil,
			ruleName:     &rule,
			want:         map[string]any{"evaluatedRuleName": "test-rule"},
		},
		{
			name:         "non-nil metadata, nil rule",
			flagMetadata: map[string]any{"foo": "bar"},
			ruleName:     nil,
			want:         map[string]any{"foo": "bar"},
		},
		{
			name:         "non-nil metadata, empty rule",
			flagMetadata: map[string]any{"foo": "bar"},
			ruleName:     stringToPointer(""),
			want:         map[string]any{"foo": "bar"},
		},
		{
			name:         "non-nil metadata, non-empty rule",
			flagMetadata: map[string]any{"foo": "bar"},
			ruleName:     &rule,
			want:         map[string]any{"foo": "bar", "evaluatedRuleName": "test-rule"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := constructMetadata(tt.flagMetadata, tt.ruleName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func stringToPointer(s string) *string {
	return &s
}
