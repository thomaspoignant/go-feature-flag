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
			got := localEvaluation(string(content))
			assert.JSONEq(t, string(want), got)
		})
	}
}

func Test_main(_ *testing.T) {
	// just to make sure that the main function exists for tinygo
	main()
}
