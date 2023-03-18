package main

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func Test_OutputResult_stdout(t *testing.T) {
	rescueStdout := os.Stdout
	defer func() { os.Stdout = rescueStdout }()
	r, w, _ := os.Pipe()
	os.Stdout = w

	content := []byte("test content")
	// Test when outputFile is empty
	err := outputResult(content, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	w.Close()
	// Verify that the content was printed to the console
	consoleOutput, err := io.ReadAll(r)
	if err != nil {
		assert.NoError(t, err, "Unexpected error reading console output: %v", err)
	}
	assert.Equal(t, string(content)+"\n", string(consoleOutput))
}
func Test_OutputResult_file(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "output-test-*.txt")
	if err != nil {
		assert.NoError(t, err, "Failed to create temporary file")
	}
	defer os.Remove(tempFile.Name())
	content := []byte("test content")

	// Test when outputFile is empty
	err = outputResult(content, tempFile.Name())
	if err != nil {
		assert.NoError(t, err, "Unexpected error")
	}

	// Test when outputFile is not empty
	err = outputResult(content, tempFile.Name())
	if err != nil {
		assert.NoError(t, err, "Unexpected error")
	}
	// Verify that the content was written to the file
	fileContent, err := os.ReadFile(tempFile.Name())
	if err != nil {
		assert.NoError(t, err, "Unexpected error reading file")
	}
	assert.Equal(t, string(content), string(fileContent))
}
