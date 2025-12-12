package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidInputFile(t *testing.T) {
	inputFile := "testdata/invalid-rule.yaml"
	inputFormat := "yaml"
	testName := "Test_Invalid_input_file"

	if os.Getenv("SHOULD_CRASH") == "true" {
		runAppMain(inputFile, inputFormat)
		return
	}
	// nolint: gosec
	cmd := exec.Command(os.Args[0], "-test.run="+testName)
	cmd.Env = append(os.Environ(), "SHOULD_CRASH=true")
	err := cmd.Run()

	e, ok := err.(*exec.ExitError)
	assert.True(t, ok && !e.Success())
}

func runAppMain(fileName, format string) {
	args := strings.Split(os.Getenv("SHOULD_CRASH"), " ")
	os.Args = append([]string{os.Args[0]}, args...)
	os.Args = append(os.Args, "--input-file="+fileName)
	os.Args = append(os.Args, "--input-format="+format)
	main()
}
