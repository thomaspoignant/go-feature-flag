package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			name:          "Test with missing targeting key",
			inputLocation: "testdata/local_evaluation_inputs/missing-targeting-key.json",
			wantLocation:  "testdata/local_evaluation_outputs/missing-targeting-key.json",
		},
		{
			name:          "Dependency met",
			inputLocation: "testdata/local_evaluation_inputs/dependency-met.json",
			wantLocation:  "testdata/local_evaluation_outputs/dependency-met.json",
		},
		{
			name:          "Dependency not met",
			inputLocation: "testdata/local_evaluation_inputs/dependency-not-met.json",
			wantLocation:  "testdata/local_evaluation_outputs/dependency-not-met.json",
		},
		{
			name:          "Missing parent flag",
			inputLocation: "testdata/local_evaluation_inputs/dependency-missing-parent.json",
			wantLocation:  "testdata/local_evaluation_outputs/dependency-missing-parent.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := os.ReadFile(tt.inputLocation)
			assert.NoError(t, err)
			want, err := os.ReadFile(tt.wantLocation)
			assert.NoError(t, err)
			got := localEvaluation(string(content))
			assert.JSONEq(t, string(want), got)
		})
	}
}

func Test_main(_ *testing.T) {
	// just to make sure that the main function exists for tinygo
	main()
}

// FuzzLocalEvaluation checks that any input yields a JSON result the host can
// parse, and never a panic (which would trap the WASM instance in production).
func FuzzLocalEvaluation(f *testing.F) {
	inputs, err := filepath.Glob("testdata/local_evaluation_inputs/*.json")
	require.NoError(f, err)
	for _, input := range inputs {
		content, err := os.ReadFile(input)
		require.NoError(f, err)
		f.Add(string(content))
	}
	f.Fuzz(func(t *testing.T, input string) {
		got := localEvaluation(input)
		assert.True(t, json.Valid([]byte(got)), "result is not valid JSON: %q", got)
	})
}
