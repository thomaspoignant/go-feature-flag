---
sidebar_position: 30
title: 🏗️ Generate OpenFeature flag manifest
description: Evaluate feature flags directly in your terminal
---

# 🏗️ Generate OpenFeature flag manifest

:::warning
This feature is experimental and may change in the future.
:::

The [OpenFeature cli](https://github.com/open-feature/cli) allow code generation for flag accessors for OpenFeature.  
This allows to have a type safe way to access your flags in your code based on a flag manifest you can provide to the OpenFeature cli.

To generate the OpenFeature flag manifest you can use the `generate manifest` command of the `go-feature-flag-cli`,
it will generate a flag manifest file compatible with the **`OpenFeature cli`** that you can use.

## Add information to your flags
To generate the flag manifest you need to provide information about your flags in your configuration file.  
To do this we leverage the `metadata` field of your flags in your configuration file, to provide information about the flag.

You need to provide 2 fields in the `metadata` to allow the flag to be generated in the flag manifest:
- `defaultValue`: **(mandatory)** The default value of the flag.
- `description`: A description of the flag.

:::info
If you don't provide a `defaultValue` field in your metadata the flag will be ignored in the flag manifest.
:::

```yaml title="flags.goff.yaml"
enableFeatureA:
  variations:
    var_a: false
    var_b: true
  defaultRule:
    variation: var_a
  metadata:
    # highlight-next-line
    description: Controls whether Feature A is enabled.
    # highlight-next-line
    defaultValue: false

usernameMaxLength:
  variations:
    var_a: 20
    var_b: 30
  defaultRule:
    variation: var_a
  metadata:
    # highlight-next-line
    description: Maximum allowed length for usernames.
    # highlight-next-line
    defaultValue: 50

# ...
```




## Install the Command Line
Check the [installation guide](./cli) to install the `go-feature-flag-cli`.

## Generate the flag manifest

### Option 1: Generate locally with go-feature-flag-cli

```shell
go-feature-flag-cli generate manifest \
  --config="<location_of_your_flag_configuration_file>" \
  --flag_manifest_destination="<destination_of_your_flag_manifest>"
```

| param                         | description                                                                                                                                                            |
|-------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--flag_manifest_destination` | (mandatory) The destination where your manifest will be stored.                                                                                                        |
| `--config`                    | The location of your configuration file. _(if not provided we will search a file named `flags.goff.yaml`_ in one of this directories `./`, `/goff/`, `/etc/opt/goff/`. |
| `--format`                    | The format of your configuration flag _(acceptable values:`yaml`, `json`, `toml`)_.<br/>Default: **`yaml`**                                                            |

### Example of flag manifest
```json title="flag_manifest.json"
{
  "flags": {
    "enableFeatureA": {
      "flagType": "boolean",
      "defaultValue": false,
      "description": "Controls whether Feature A is enabled."
    },
    "greetingMessage": {
      "flagType": "string",
      "defaultValue": "Hello there!",
      "description": "The message to use for greeting users."
    }
  }
}
```

### Option 2: Use the relay proxy endpoint

If you are running the GO Feature Flag relay proxy, you can fetch the flag manifest directly from it
without needing to generate a local file. The relay proxy exposes a dedicated endpoint:

```
GET /openfeature/v0/manifest
```

The response is fully compatible with the OpenFeature CLI manifest format.

:::info
Flags without a `defaultValue` in their metadata are excluded from the response, same as when generating locally.
:::


You can pull the file directly with the openfeature cli:
```shell
openfeature pull --provider-url http://localhost:1031
```
If authentication is enabled on your relay proxy, pass your API key via the `Authorization: Bearer <key>` header.
If authentication is not enabled, no header is needed. The cli can also pass the token:

```shell
openfeature pull --provider-url http://localhost:1031 --auth-token secret-token
```

## Use the OpenFeature cli to generate the flag accessors

To install the OpenFeature cli you can follow the [installation guide](https://github.com/open-feature/cli) from the repository.

Once you have the OpenFeature cli installed you can use the `generate` command to generate the flag accessors.
You can provide the manifest either as a local file or directly as a URL pointing to the relay proxy endpoint.

```shell
openfeature generate go
```

:::info 
Here the example is using code generation in go, but react, typescript, C# and other languages are supported.

Refer to the [OpenFeature cli documentation](https://github.com/open-feature/cli) for more information.
:::


As the result of this command, you will have a file with the flag accessors generated based on the flag manifest you provided.
You will be able to use a type-safe way to access your flags in your code.

```go
// Code generated by OpenFeature CLI. DO NOT EDIT.
// CLI version: 0.4.1

// Package openfeature contains generated code produced by the OpenFeature CLI.
package openfeature

import (
        "context"
        "fmt"

        "github.com/open-feature/go-sdk/openfeature"
)

// stringer transforms a string to a Stringer
type stringer string

// String implements the fmt.Stringer interface
func (s stringer) String() string {
        return string(s)
}

type (
        evaluationValue[T any]   func(context.Context, openfeature.EvaluationContext) T
        evaluationDetails[T any] func(context.Context, openfeature.EvaluationContext) (openfeature.GenericEvaluationDetails[T], error)
)

var client = openfeature.NewDefaultClient()

// TestFlag returns the value of the "test-flag" feature flag.
// this is a simple feature flag
//
// The flag is a type of boolean and defaults to false.
var TestFlag = struct {
        fmt.Stringer
        // Value returns the value of the [TestFlag] flag.
        Value evaluationValue[bool]

        // ValueWithDetails returns the evaluation details of the [TestFlag] flag
        // and the evaluation error, if any.
        ValueWithDetails evaluationDetails[bool]
}{
        Stringer: stringer("test-flag"),
        Value: func(ctx context.Context, evalCtx openfeature.EvaluationContext) bool {
                return client.Boolean(ctx, "test-flag", false, evalCtx)
        },
        ValueWithDetails: func(ctx context.Context, evalCtx openfeature.EvaluationContext) (openfeature.GenericEvaluationDetails[bool], error) {
                return client.BooleanValueDetails(ctx, "test-flag", false, evalCtx)
        },
}

// TestFlag2 returns the value of the "test-flag2" feature flag.
//
// The flag is a type of boolean and defaults to false.
var TestFlag2 = struct {
        fmt.Stringer
        // Value returns the value of the [TestFlag2] flag.
        Value evaluationValue[bool]

        // ValueWithDetails returns the evaluation details of the [TestFlag2] flag
        // and the evaluation error, if any.
        ValueWithDetails evaluationDetails[bool]
}{
        Stringer: stringer("test-flag2"),
        Value: func(ctx context.Context, evalCtx openfeature.EvaluationContext) bool {
                return client.Boolean(ctx, "test-flag2", false, evalCtx)
        },
        ValueWithDetails: func(ctx context.Context, evalCtx openfeature.EvaluationContext) (openfeature.GenericEvaluationDetails[bool], error) {
                return client.BooleanValueDetails(ctx, "test-flag2", false, evalCtx)
        },
}
```
