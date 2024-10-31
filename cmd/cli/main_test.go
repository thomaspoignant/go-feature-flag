package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/variation"
)

func TestEvaluateFlag(t *testing.T) {
	// Create temporary config file
	configContent := `
flags:
  test-flag:
    variations:
      - value: true
      - value: false
    defaultRule:
      variation: 0
`
	tmpfile, err := ioutil.TempFile("", "test-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(configContent))
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name            string
		configFile      string
		flagName        string
		context         string
		expectedResult  bool
		expectedError   bool
	}{
		{
			name:           "Valid evaluation",
			configFile:     tmpfile.Name(),
			flagName:       "test-flag",
			context:        `{"targetingKey": "user-123"}`,
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "Invalid flag name",
			configFile:     tmpfile.Name(),
			flagName:       "non-existent-flag",
			context:        `{"targetingKey": "user-123"}`,
			expectedError:  true,
		},
		{
			name:           "Invalid context",
			configFile:     tmpfile.Name(),
			flagName:       "test-flag",
			context:        `invalid-json`,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up command line args
			configFile = tt.configFile
			flagName = tt.flagName
			evaluationContext = tt.context

			// Capture output
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)

			// Run command
			err := runEvaluate(nil, nil)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Parse output
			var result EvaluationResult
			err = json.Unmarshal(buf.Bytes(), &result)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedResult, result.Value)
		})
	}
}

func TestLint(t *testing.T) {
	// Create temporary config file
	configContent := `
flags:
  valid-flag:
    variations:
      - value: true
      - value: false
    defaultRule:
      variation: 0
  INVALID-FLAG:
    variations:
      - value: true
`
	tmpfile, err := ioutil.TempFile("", "test-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(configContent))
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

	// Test linting
	configFile = tmpfile.Name()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	err = runLint(nil, nil)
	assert.NoError(t, err)

	var result LintResult
	err = json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)

	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors, "Invalid flag name 'INVALID-FLAG': should be lowercase with hyphens")
}