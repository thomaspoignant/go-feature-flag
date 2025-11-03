package main

import (
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"go.yaml.in/yaml/v3"
)

// This example is demonstrating how to lint/validate a GO Feature Flag configuration programmatically
// without using the CLI.
// The example is retrieving a flag configuration from a storage (could be a database, a file, etc...),
// unmarshaling it into a DTO structure, and validating each flag configuration.
// If you're storage contains only 1 flag this will still work as the configuration is a map of flag key to flag DTO.
func main() {
	// Step 1: Retrieve your flag configuration from your storage
	myFlags := getMyFlagsFromStorage()

	// Step 2: Unmarshal the flag configuration into a flag DTO
	// In the example we are using YAML but you can use JSON or TOML as well depending on your storage format.
	var flags map[string]dto.DTO
	err := yaml.Unmarshal([]byte(myFlags), &flags)
	if err != nil {
		panic(err) // Don't forget to handle errors appropriately in production code.
	}

	// Step 3: Validate the flags configuration
	// We are converting each DTO into a Flag and calling the IsValid method exactly like the CLI linter would do.
	for key, flagDto := range flags {
		convertedFlag := flagDto.Convert()
		if err := convertedFlag.IsValid(); err != nil {
			panic(fmt.Sprintf("Invalid flag %s: %s", key, err.Error())) // Don't forget to handle errors appropriately in production code.
		}
	}
}

// getMyFlagsFromStorage simulates retrieving a flag configuration from some storage
// and returns it as a string.
func getMyFlagsFromStorage() string {
	return `
my-test-flag:
  variations:
    enabled: true
    disabled: false
  targeting:
    - query: key eq "785a14bf-d2c5-4caa-9c70-2bbc4e3732a5"
      percentage:
        enabled: 0
        disabled: 100
  defaultRule:
    variation: disabled
`
}
