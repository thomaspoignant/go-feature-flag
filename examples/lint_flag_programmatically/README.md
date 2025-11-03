# Lint Feature Flags Programmatically

This example demonstrates how to lint/validate a GO Feature Flag configuration programmatically without using the CLI.

## Overview

This example shows how to:
1. Retrieve flag configurations from your storage
2. Unmarshal the configuration into DTOs
3. Validate each flag configuration programmatically

## Prerequisites

- Access to the GO Feature Flag core module

## Installation

### Step 1: Install Dependencies

Install the required dependencies:

```bash
go get github.com/thomaspoignant/go-feature-flag/modules/core
```

#### Step 2: Retrieve Flag Configuration from Storage

```go
myFlags := getMyFlagsFromStorage()
```

This simulates retrieving your flag configuration from a storage system (database, file system, API, etc.). In this example, it returns a YAML string, but you can adapt this to your specific storage mechanism.

**Replace `getMyFlagsFromStorage()`** with your actual retrieval logic.

#### Step 3: Unmarshal Flag Configuration

```go
var flags map[string]dto.DTO
err := yaml.Unmarshal([]byte(myFlags), &flags)
if err != nil {
    panic(err) // Don't forget to handle errors appropriately in production code.
}
```

The configuration is unmarshaled from YAML format into a map where:
- **Key**: The flag key (identifier)
- **Value**: A `dto.DTO` object representing the flag configuration

**Note**: You can also use JSON or TOML formats depending on your storage format. Simply use the appropriate unmarshaler (`json.Unmarshal` or `toml.Unmarshal`).

#### Step 4: Validate Each Flag Configuration

```go
for key, flagDto := range flags {
    convertedFlag := flagDto.Convert()
    if err := convertedFlag.IsValid(); err != nil {
        panic(fmt.Sprintf("Invalid flag %s: %s", key, err.Error()))
    }
}
```

For each flag:
1. Convert the DTO to a Flag object using `flagDto.Convert()`
2. Validate the flag using `IsValid()`
3. Handle any validation errors appropriately

**Handle validation errors** according to your application's error handling strategy.

## Additional Resources

- [GO Feature Flag Documentation](https://gofeatureflag.org/)
- [YAML Package Documentation](https://pkg.go.dev/gopkg.in/yaml.v3)
