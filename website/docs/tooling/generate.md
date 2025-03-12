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
    "usernameMaxLength": {
      "flagType": "integer",
      "defaultValue": 50,
      "description": "Maximum allowed length for usernames."
    },
    "greetingMessage": {
      "flagType": "string",
      "defaultValue": "Hello there!",
      "description": "The message to use for greeting users."
    }
  }
}
```

## Use the OpenFeature cli to generate the flag accessors

To install the OpenFeature cli you can follow the [installation guide](https://github.com/open-feature/cli) from the repository.

Once you have the OpenFeature cli installed you can use the `generate` command to generate the flag accessors.

```shell
openfeature-cli generate go \
  --flag_manifest_path="<destination_of_your_flag_manifest>" \
  --package_name="<name of your GO package>" \
  --output_path="<destination_go_file>"
```

As the result of this command, you will have a file with the flag accessors generated based on the flag manifest you provided.
You will be able to use a type-safe way to access your flags in your code.

:::info 
Here the example is using code generation in GO, but react and other languages are supported.

Refer to the [OpenFeature cli documentation](https://github.com/open-feature/cli) for more information.
:::
