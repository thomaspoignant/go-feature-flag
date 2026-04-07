package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_localEvaluation(t *testing.T) {
	tests := []struct {
		name          string
		inputLocation string
		wantLocation  string
		err           assert.ErrorAssertionFunc
	}{
		{
			name:          "Test with a valid input",
			inputLocation: "testdata/local_evaluation_inputs/valid.json",
			wantLocation:  "testdata/local_evaluation_outputs/valid.json",
		},
		{
			name:          "Test with invalid json input",
			inputLocation: "testdata/local_evaluation_inputs/invalid.json",
			wantLocation:  "testdata/local_evaluation_outputs/invalid.json",
		},
		{
			name:          "Test with invalid json input",
			inputLocation: "testdata/local_evaluation_inputs/missing-targeting-key.json",
			wantLocation:  "testdata/local_evaluation_outputs/missing-targeting-key.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := os.ReadFile(tt.inputLocation)
			assert.NoError(t, err)
			want, err := os.ReadFile(tt.wantLocation)
			assert.NoError(t, err)
			got := localEvaluation(content)
			assert.JSONEq(t, string(want), string(got))
		})
	}
}

func Test_main(_ *testing.T) {
	// just to make sure that the main function exists for tinygo
	main()
}

func Benchmark_localEvaluation(b *testing.B) {
	tests := []struct {
		name          string
		inputLocation string
	}{
		{
			name:          "simple flag (no rules)",
			inputLocation: "testdata/local_evaluation_inputs/bench_simple.json",
		},
		{
			name:          "flag with targeting rule and percentage rollout",
			inputLocation: "testdata/local_evaluation_inputs/valid.json",
		},
		{
			name:          "flag with scheduled rollout",
			inputLocation: "testdata/local_evaluation_inputs/bench_scheduled.json",
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			content, err := os.ReadFile(tt.inputLocation)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				localEvaluation(content)
			}
		})
	}
}
